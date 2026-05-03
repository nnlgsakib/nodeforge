package context

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Store manages graph storage in BadgerDB
type Store struct {
	db *badger.DB
}

// NewStore creates a new context store
func NewStore(dbPath string) (*Store, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // disable default logger

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the BadgerDB connection
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveGraph stores a graph in BadgerDB (accepts any type, marshals to JSON)
func (s *Store) SaveGraph(ctx context.Context, graphID string, graphData interface{}) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("graph:" + graphID)
		data, err := json.Marshal(graphData)
		if err != nil {
			return fmt.Errorf("failed to marshal graph: %w", err)
		}
		return txn.Set(key, data)
	})
}

// GetGraph retrieves a graph by ID as raw JSON
func (s *Store) GetGraph(ctx context.Context, graphID string) (json.RawMessage, error) {
	var data json.RawMessage
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("graph:" + graphID))
		if err != nil {
			return fmt.Errorf("graph not found: %w", err)
		}

		return item.Value(func(val []byte) error {
			if len(val) == 0 {
				return fmt.Errorf("graph %s has empty value", graphID)
			}
			// Copy the data since val is only valid during this callback
			data = make(json.RawMessage, len(val))
			copy(data, val)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return data, nil
}

// SaveNodeOutput stores node output for downstream context reuse (FR18)
func (s *Store) SaveNodeOutput(ctx context.Context, graphID, nodeID, output string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("output:%s:%s", graphID, nodeID))
		return txn.Set(key, []byte(output))
	})
}

// GetNodeOutput retrieves node output for context reuse
func (s *Store) GetNodeOutput(ctx context.Context, graphID, nodeID string) (string, error) {
	var output []byte
	err := s.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("output:%s:%s", graphID, nodeID))
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			if len(val) == 0 {
				return fmt.Errorf("node output %s:%s is empty", graphID, nodeID)
			}
			// Copy the data since val is only valid during this callback
			output = make([]byte, len(val))
			copy(output, val)
			return nil
		})
	})

	if err != nil {
		return "", err
	}
	return string(output), nil
}

// MonologueMessage represents a single LLM inner monologue entry
type MonologueMessage struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}

// SaveMonologueHistory stores monologue messages for a session
func (s *Store) SaveMonologueHistory(ctx context.Context, sessionID string, messages []MonologueMessage) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID cannot be empty")
	}
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte("monologue:" + sessionID)
		data, err := json.Marshal(messages)
		if err != nil {
			return fmt.Errorf("failed to marshal monologue history: %w", err)
		}
		return txn.Set(key, data)
	})
}

// GetMonologueHistory retrieves monologue messages for a session
func (s *Store) GetMonologueHistory(ctx context.Context, sessionID string) ([]MonologueMessage, error) {
	var messages []MonologueMessage
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("monologue:" + sessionID))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				// Return empty slice if not found
				messages = []MonologueMessage{}
				return nil
			}
			return err
		}

		return item.Value(func(val []byte) error {
			if len(val) == 0 {
				messages = []MonologueMessage{}
				return nil
			}
			return json.Unmarshal(val, &messages)
		})
	})

	if err != nil {
		return nil, err
	}
	return messages, nil
}

// NodeMemory manages node memory reuse (FR18)
type NodeMemory struct {
	db *badger.DB
}

// NewNodeMemory creates a new NodeMemory instance
func NewNodeMemory(db *badger.DB) *NodeMemory {
	return &NodeMemory{db: db}
}

// StoreMemory stores node output for downstream use (Subtask 2.2)
func (nm *NodeMemory) StoreMemory(nodeID, key, value string) error {
	return nm.db.Update(func(txn *badger.Txn) error {
		memKey := []byte(fmt.Sprintf("mem:%s:%s", nodeID, key))
		return txn.Set(memKey, []byte(value))
	})
}

// GetMemory retrieves upstream memory for current node (Subtask 2.3)
func (nm *NodeMemory) GetMemory(nodeID, key string) (string, bool) {
	var value []byte
	err := nm.db.View(func(txn *badger.Txn) error {
		memKey := []byte(fmt.Sprintf("mem:%s:%s", nodeID, key))
		item, err := txn.Get(memKey)
		if err != nil {
			return fmt.Errorf("failed to get memory: %w", err)
		}
		return item.Value(func(val []byte) error {
			value = make([]byte, len(val))
			copy(value, val)
			return nil
		})
	})

	if err != nil {
		return "", false
	}
	return string(value), true
}

// InjectMemoryIntoPrompt adds memory context to LLM prompts (Subtask 2.4)
func (nm *NodeMemory) InjectMemoryIntoPrompt(prompt string, nodeID string) string {
	// Get all memory entries for this node
	var memories []string
	_ = nm.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("mem:%s:", nodeID))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				memories = append(memories, string(val))
				return nil
			})
			if err != nil {
				continue
			}
		}
		return nil
	})

	if len(memories) == 0 {
		return prompt
	}

	memoryContext := "Memory Context:\n" + strings.Join(memories, "\n")
	return prompt + "\n\n" + memoryContext
}

// DefaultStorePath returns the default BadgerDB path
func DefaultStorePath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".nforge", "context.db")
}
