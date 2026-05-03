package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nnlgsakib/nodeforge/internal/llm"
	nfContext "github.com/nnlgsakib/nodeforge/internal/context"
)

// NodeType represents the type of a graph node
type NodeType string

const (
	NodeTypeGoal   NodeType = "goal"
	NodeTypeSpec   NodeType = "spec"
	NodeTypePlan   NodeType = "plan"
	NodeTypeImpl   NodeType = "implement"
	NodeTypeTest   NodeType = "test"
	NodeTypeReview NodeType = "review"
)

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusComplete  NodeStatus = "complete"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusSkipped   NodeStatus = "skipped"
)

// Node represents a graph node
type Node struct {
	ID                string                 `json:"id"`
	Type              NodeType              `json:"type"`
	Label             string                `json:"label"`
	Status            NodeStatus            `json:"status"`
	Progress          float64               `json:"progress"`
	AcceptanceCriteria []string              `json:"acceptanceCriteria"`
	Output            string                `json:"output,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// Edge represents a graph edge
type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type,omitempty"`
	Tension float64 `json:"tension,omitempty"`
}

// GraphMetadata stores metadata about a graph execution (Story 2.7)
type GraphMetadata struct {
	MerkleRoot string            `json:"merkleRoot"`
	NodeHashes map[string]string `json:"nodeHashes"` // Per-node hashes for change detection
}

// Graph represents a node graph
type Graph struct {
	ID        string         `json:"id"`
	Goal      string         `json:"goal"`
	Nodes     []*Node        `json:"nodes"`
	Edges     []*Edge        `json:"edges"`
	Status    string         `json:"status"`
	CreatedAt string         `json:"createdAt"`
	Metadata  *GraphMetadata `json:"metadata,omitempty"` // Story 2.7: Merkle tree metadata
}

// Generator creates graphs from goals
type Generator struct {
	llmProvider llm.LLMProvider
	store        *nfContext.Store
}

// NewGenerator creates a new graph generator
func NewGenerator(provider llm.LLMProvider, store *nfContext.Store) *Generator {
	return &Generator{llmProvider: provider, store: store}
}

// Generate creates a node graph from a user goal
func (g *Generator) Generate(ctx context.Context, goal string) (*Graph, error) {
	graph := &Graph{
		ID:        generateID(),
		Goal:      goal,
		Status:    "pending",
		CreatedAt: currentTime(),
	}

	// Generate graph structure using LLM
	prompt := fmt.Sprintf(`Analyze the following goal and generate a node graph with acceptance criteria for each node.

Goal: %s

Generate a graph with nodes: Goal → Spec → Plan → Implement → Test → Review.
For each node, provide acceptance criteria.

Output JSON format:
{
  "nodes": [
    {"type": "goal", "label": "Goal", "acceptanceCriteria": ["criteria1"]},
    {"type": "spec", "label": "Spec", "acceptanceCriteria": ["criteria1"]},
    {"type": "plan", "label": "Plan", "acceptanceCriteria": ["criteria1"]},
    {"type": "implement", "label": "Implement", "acceptanceCriteria": ["criteria1"]},
    {"type": "test", "label": "Test", "acceptanceCriteria": ["criteria1"]},
    {"type": "review", "label": "Review", "acceptanceCriteria": ["criteria1"]}
  ]
}`, goal)

	if g.llmProvider == nil {
		// Fallback: generate default graph without LLM
		log.Println("[WARN] LLM provider is nil, falling back to default graph")
		graph := g.generateDefaultGraph(goal)
		g.saveGraph(ctx, graph)
		return graph, nil
	}

	messages := []llm.Message{
		{Role: "system", Content: "You are a graph generation assistant. Output only valid JSON."},
		{Role: "user", Content: prompt},
	}

	ch, err := g.llmProvider.Chat(ctx, messages)
	if err != nil {
		// Fallback to default graph on LLM failure
		log.Printf("[WARN] LLM chat failed: %v, falling back to default graph", err)
		graph := g.generateDefaultGraph(goal)
		g.saveGraph(ctx, graph)
		return graph, nil
	}

	// Collect streamed response
	var respBuilder strings.Builder
	for token := range ch {
		respBuilder.WriteString(token)
	}

	resp := respBuilder.String()
	if resp == "" {
		log.Println("[WARN] LLM returned empty response, falling back to default graph")
		graph := g.generateDefaultGraph(goal)
		g.saveGraph(ctx, graph)
		return graph, nil
	}

	// Parse LLM response
	var llmOutput struct {
		Nodes []struct {
			Type              string   `json:"type"`
			Label             string   `json:"label"`
			AcceptanceCriteria []string `json:"acceptanceCriteria"`
		} `json:"nodes"`
	}

	if err := json.Unmarshal([]byte(resp), &llmOutput); err != nil {
		log.Printf("[WARN] Failed to parse LLM response: %v, falling back to default graph", err)
		graph := g.generateDefaultGraph(goal)
		g.saveGraph(ctx, graph)
		return graph, nil
	}

	// Validate that we got valid nodes
	if len(llmOutput.Nodes) == 0 {
		log.Printf("[WARN] LLM returned empty nodes, falling back to default graph")
		graph := g.generateDefaultGraph(goal)
		g.saveGraph(ctx, graph)
		return graph, nil
	}

	// Build nodes
	for i, n := range llmOutput.Nodes {
		nodeType := NodeType(n.Type)
		graph.Nodes = append(graph.Nodes, &Node{
			ID:                fmt.Sprintf("node-%d", i+1),
			Type:              nodeType,
			Label:             n.Label,
			Status:            NodeStatusPending,
			AcceptanceCriteria: n.AcceptanceCriteria,
		})
	}

	// Build edges
	for i := 0; i < len(graph.Nodes)-1; i++ {
		graph.Edges = append(graph.Edges, &Edge{
			ID:     fmt.Sprintf("edge-%d", i+1),
			Source: graph.Nodes[i].ID,
			Target: graph.Nodes[i+1].ID,
			Type:   "default",
		})
	}

	// Persist graph to BadgerDB
	g.saveGraph(ctx, graph)

	return graph, nil
}

// saveGraph persists a graph to BadgerDB if store is available
func (g *Generator) saveGraph(ctx context.Context, graph *Graph) {
	if g.store == nil {
		return
	}
	if graph.ID == "" {
		log.Printf("[WARN] Skipping save: graph ID is empty")
		return
	}
	if err := g.store.SaveGraph(ctx, graph.ID, graph); err != nil {
		log.Printf("[WARN] Failed to save graph %s to BadgerDB: %v", graph.ID, err)
	}
}

// generateDefaultGraph creates a default graph without LLM
func (g *Generator) generateDefaultGraph(goal string) *Graph {
	graph := &Graph{
		ID:        generateID(),
		Goal:      goal,
		Status:    "pending",
		CreatedAt: currentTime(),
	}

	types := []NodeType{NodeTypeGoal, NodeTypeSpec, NodeTypePlan, NodeTypeImpl, NodeTypeTest, NodeTypeReview}
	labels := []string{"Goal", "Spec", "Plan", "Implement", "Test", "Review"}

	for i, t := range types {
		graph.Nodes = append(graph.Nodes, &Node{
			ID:     fmt.Sprintf("node-%d", i+1),
			Type:   t,
			Label:  labels[i],
			Status: NodeStatusPending,
			AcceptanceCriteria: []string{
				fmt.Sprintf("%s node completed successfully", labels[i]),
			},
		})
	}

	for i := 0; i < len(graph.Nodes)-1; i++ {
		graph.Edges = append(graph.Edges, &Edge{
			ID:     fmt.Sprintf("edge-%d", i+1),
			Source: graph.Nodes[i].ID,
			Target: graph.Nodes[i+1].ID,
			Type:   "default",
		})
	}

	return graph
}

func generateID() string {
	return fmt.Sprintf("graph-%d-%d", currentTimeMillis(), fastRand())
}

// fastRand returns a pseudo-random number for ID generation
func fastRand() int {
	return int(time.Now().UnixNano() % 100000)
}

func currentTime() string {
	return currentTimeMillisStr()
}

func currentTimeMillis() int64 {
	return time.Now().UnixMilli()
}

func currentTimeMillisStr() string {
	return time.Now().UTC().Format(time.RFC3339)
}
