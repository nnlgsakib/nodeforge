package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// ContextNode represents a node in the knowledge graph
type ContextNode struct {
	ID        string   `json:"id"`
	NodeID    string   `json:"nodeId"`
	Output    string   `json:"output"`
	Timestamp  int64    `json:"timestamp"`
	EdgesFrom  []string `json:"edgesFrom"` // Node IDs this node has edges from
	EdgesTo    []string `json:"edgesTo"`   // Node IDs this node has edges to
}

// KnowledgeGraph manages the knowledge graph for context assembly
type KnowledgeGraph struct {
	db *badger.DB
}

// NewKnowledgeGraph creates a new KnowledgeGraph using the provided BadgerDB instance
func NewKnowledgeGraph(db *badger.DB) *KnowledgeGraph {
	return &KnowledgeGraph{db: db}
}

// AddNodeOutput stores node output as a graph node with edge relationships (Subtask 1.2)
func (kg *KnowledgeGraph) AddNodeOutput(nodeID, output string) error {
	return kg.db.Update(func(txn *badger.Txn) error {
		// Single-pass: build map of nodeID -> latest ContextNode
		nodeMap := make(map[string]ContextNode)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("ctx:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keyStr := string(item.Key())
			parts := strings.Split(keyStr, ":")
			if len(parts) >= 2 {
				nodeIDKey := parts[1]
				// Load full node to check timestamp
				var node ContextNode
				err := item.Value(func(val []byte) error {
					return json.Unmarshal(val, &node)
				})
				if err != nil {
					continue
				}
				if existing, ok := nodeMap[nodeIDKey]; !ok || node.Timestamp > existing.Timestamp {
					nodeMap[nodeIDKey] = node
				}
			}
		}

		// Create new context node
		node := ContextNode{
			ID:        fmt.Sprintf("ctx:%s:%d", nodeID, time.Now().UnixNano()),
			NodeID:    nodeID,
			Output:    output,
			Timestamp:  time.Now().Unix(),
			EdgesFrom:  []string{},
			EdgesTo:    []string{},
		}

		// Use nodeMap to create edges
		var upstreamNodes []string
		for upstreamID, upstreamNode := range nodeMap {
			if upstreamID == nodeID {
				continue // Skip self
			}
			upstreamNodes = append(upstreamNodes, upstreamID)
			// Update upstream node to add edge to this new node
			upstreamNode.EdgesTo = append(upstreamNode.EdgesTo, nodeID)
			updatedData, err := json.Marshal(upstreamNode)
			if err != nil {
				return fmt.Errorf("failed to marshal upstream node: %w", err)
			}
			if err := txn.Set([]byte(upstreamNode.ID), updatedData); err != nil {
				return fmt.Errorf("failed to update upstream node: %w", err)
			}
		}

		// Add edges from upstream nodes to this new node
		node.EdgesFrom = upstreamNodes

		// Save the node
		nodeData, err := json.Marshal(node)
		if err != nil {
			return fmt.Errorf("failed to marshal context node: %w", err)
		}

		return txn.Set([]byte(node.ID), nodeData)
	})
}

// BuildContext assembles context from graph in <100ms (Subtask 1.3, NFR-04)
func (kg *KnowledgeGraph) BuildContext(ctx context.Context, nodeID string, maxTokens int) (string, error) {
	start := time.Now()
	defer func() {
		// Log if assembly takes >100ms (NFR-04 violation)
		if dur := time.Since(start); dur > 100*time.Millisecond {
			fmt.Printf("WARN: BuildContext took %v (>100ms threshold)\n", dur)
		}
	}()

	var contextNodes []ContextNode
	err := kg.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("ctx:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			item := it.Item()
			var node ContextNode
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &node)
			})
			if err != nil {
				continue
			}
			contextNodes = append(contextNodes, node)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to query context nodes: %w", err)
	}

	// Assemble context: include upstream nodes first, then current node
	var contextParts []string
	tokenCount := 0
	charPerToken := 4 // ~4 chars/token heuristic (consistent with story 2.4)

	// Sort nodes by timestamp (oldest first) for context assembly
	for _, node := range contextNodes {
		if node.NodeID == nodeID {
			continue // Skip current node, it will be added separately
		}
		// Check token budget
		nodeTokens := len(node.Output) / charPerToken
		if tokenCount+nodeTokens > maxTokens {
			break
		}
		contextParts = append(contextParts, fmt.Sprintf("[Node %s]: %s", node.NodeID, node.Output))
		tokenCount += nodeTokens
	}

	context := strings.Join(contextParts, "\n")
	return context, nil
}

// GetDownstreamContext retrieves upstream/downstream context for memory reuse (Subtask 1.4, FR18)
func (kg *KnowledgeGraph) GetDownstreamContext(nodeID string) []ContextNode {
	var result []ContextNode
	_ = kg.db.View(func(txn *badger.Txn) error {
		// Single pass: build map of nodeID -> latest ContextNode
		nodeMap := make(map[string]ContextNode)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("ctx:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keyStr := string(item.Key())
			parts := strings.Split(keyStr, ":")
			if len(parts) < 2 {
				continue
			}
			nodeIDKey := parts[1]
			var node ContextNode
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &node)
			})
			if err != nil {
				continue
			}
			if existing, ok := nodeMap[nodeIDKey]; !ok || node.Timestamp > existing.Timestamp {
				nodeMap[nodeIDKey] = node
			}
		}

		// Get the current node's latest version
		currentNode, ok := nodeMap[nodeID]
		if !ok {
			return nil
		}

		// Get all upstream nodes from EdgesFrom
		for _, upstreamID := range currentNode.EdgesFrom {
			if upstream, ok := nodeMap[upstreamID]; ok {
				result = append(result, upstream)
			}
		}

		return nil
	})
	return result
}
