package context

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// SpecGenerator handles auto-spec generation from node outputs (FR19)
type SpecGenerator struct {
	db *badger.DB
}

// NewSpecGenerator creates a new SpecGenerator
func NewSpecGenerator(db *badger.DB) *SpecGenerator {
	return &SpecGenerator{db: db}
}

// GenerateSpec auto-generates a spec from node output (Subtask 3.2)
func (sg *SpecGenerator) GenerateSpec(nodeID string, output string) (string, error) {
	spec := fmt.Sprintf("Spec for node %s:\n%s\nGenerated at: %d",
		nodeID, output, time.Now().Unix())

	// Save spec to BadgerDB
	err := sg.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("spec:%s:%d", nodeID, time.Now().UnixNano()))
		return txn.Set(key, []byte(spec))
	})

	if err != nil {
		return "", fmt.Errorf("failed to save spec: %w", err)
	}

	return spec, nil
}

// AddSystemReferences adds system references to the graph (Subtask 3.3, FR19)
func (sg *SpecGenerator) AddSystemReferences(nodeID string, refs []string) error {
	return sg.db.Update(func(txn *badger.Txn) error {
		// Get existing references
		key := []byte(fmt.Sprintf("sysref:%s", nodeID))
		var existingRefs []string

		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return fmt.Errorf("failed to get system refs: %w", err)
		}

		if err == nil {
			err = item.Value(func(val []byte) error {
				return json.Unmarshal(val, &existingRefs)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal refs: %w", err)
			}
		}

		// Add new references
		existingRefs = append(existingRefs, refs...)

		// Save updated references
		data, err := json.Marshal(existingRefs)
		if err != nil {
			return fmt.Errorf("failed to marshal refs: %w", err)
		}

		return txn.Set(key, data)
	})
}

// SpecToGraph converts generated specs into executable graph nodes (Subtask 3.4)
func (sg *SpecGenerator) SpecToGraph(spec string) ([]Node, []Edge, error) {
	// Simple parser: each line starting with "Node:" becomes a node
	var nodes []Node
	var edges []Edge

	lines := strings.Split(spec, "\n")
	nodeID := ""
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Node:") {
			nodeID = strings.TrimPrefix(line, "Node:")
			nodes = append(nodes, Node{
				ID:   nodeID,
				Type: "spec",
				Data: map[string]interface{}{"content": line},
			})
			// Add edge from previous node if exists
			if i > 0 && len(nodes) > 1 {
				edges = append(edges, Edge{
					From: nodes[len(nodes)-2].ID,
					To:   nodeID,
				})
			}
		}
	}

	return nodes, edges, nil
}

// Node represents a graph node for SpecToGraph
type Node struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// Edge represents a graph edge for SpecToGraph
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}
