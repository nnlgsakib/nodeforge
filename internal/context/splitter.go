package context

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

// SubGraph represents a sub-graph after splitting (Subtask 4.3)
type SubGraph struct {
	Nodes          []ContextNode `json:"nodes"`
	Edges          []string      `json:"edges"` // Edge IDs as strings
	EstimatedTokens int          `json:"estimatedTokens"`
}

// GraphSplitter handles context overflow by splitting graphs (FR20)
type GraphSplitter struct {
	db *badger.DB
}

// NewGraphSplitter creates a new GraphSplitter
func NewGraphSplitter(db *badger.DB) *GraphSplitter {
	return &GraphSplitter{db: db}
}

// SplitGraphIfNeeded auto-splits graphs on overflow (Subtask 4.2, FR20)
func (gs *GraphSplitter) SplitGraphIfNeeded(nodeID string, maxTokens int) ([]SubGraph, error) {
	var contextNodes []ContextNode
	err := gs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("ctx:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
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
		return nil, fmt.Errorf("failed to query context nodes: %w", err)
	}

	// Calculate total tokens using shared estimation
	totalTokens := EstimateSubGraphTokens(SubGraph{Nodes: contextNodes})

	// If under budget, return single sub-graph
	if totalTokens <= maxTokens {
		return []SubGraph{buildSubGraph(contextNodes, 4)}, nil
	}

	// Split into sub-graphs
	var subGraphs []SubGraph
	var currentNodes []ContextNode
	currentTokens := 0
	charPerToken := 4 // ~4 chars/token heuristic (story 2.4)

	for _, node := range contextNodes {
		nodeTokens := len(node.Output) / charPerToken
		if currentTokens+nodeTokens > maxTokens && len(currentNodes) > 0 {
			// Save current sub-graph and start new one
			subGraphs = append(subGraphs, buildSubGraph(currentNodes, charPerToken))
			currentNodes = []ContextNode{}
			currentTokens = 0
		}
		currentNodes = append(currentNodes, node)
		currentTokens += nodeTokens
	}

	// Add remaining nodes
	if len(currentNodes) > 0 {
		subGraphs = append(subGraphs, buildSubGraph(currentNodes, charPerToken))
	}

	return subGraphs, nil
}

// EstimateSubGraphTokens estimates tokens for a sub-graph (Subtask 4.4, <10ms estimation)
func EstimateSubGraphTokens(sg SubGraph) int {
	totalChars := 0
	for _, node := range sg.Nodes {
		totalChars += len(node.Output)
	}
	return totalChars / 4 // ~4 chars/token heuristic
}

// buildSubGraph creates a SubGraph from nodes, populating Edges from node relationships
func buildSubGraph(nodes []ContextNode, charPerToken int) SubGraph {
	edges := []string{}
	tokens := 0
	for _, node := range nodes {
		tokens += len(node.Output) / charPerToken
		edges = append(edges, node.EdgesFrom...)
		edges = append(edges, node.EdgesTo...)
	}
	// Deduplicate edges
	seen := make(map[string]bool)
	uniqueEdges := []string{}
	for _, e := range edges {
		if !seen[e] {
			seen[e] = true
			uniqueEdges = append(uniqueEdges, e)
		}
	}
	return SubGraph{
		Nodes:          nodes,
		Edges:          uniqueEdges,
		EstimatedTokens: tokens,
	}
}
