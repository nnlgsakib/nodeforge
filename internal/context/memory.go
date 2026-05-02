package context

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

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

// DefaultStorePath returns the default BadgerDB path
func DefaultStorePath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, ".nforge", "context.db")
}
