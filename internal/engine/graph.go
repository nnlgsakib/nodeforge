package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nnlgsakib/nodeforge/internal/llm"
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

// Graph represents a node graph
type Graph struct {
	ID        string  `json:"id"`
	Goal      string  `json:"goal"`
	Nodes     []*Node `json:"nodes"`
	Edges     []*Edge `json:"edges"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

// Generator creates graphs from goals
type Generator struct {
	llmProvider llm.Provider
}

// NewGenerator creates a new graph generator
func NewGenerator(provider llm.Provider) *Generator {
	return &Generator{llmProvider: provider}
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

	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a graph generation assistant. Output only valid JSON."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   2000,
	}

	if g.llmProvider == nil {
		// Fallback: generate default graph without LLM
		return g.generateDefaultGraph(goal), nil
	}

	resp, err := g.llmProvider.Chat(ctx, req)
	if err != nil {
		// Fallback to default graph on LLM failure
		return g.generateDefaultGraph(goal), nil
	}

	if len(resp.Choices) == 0 {
		return g.generateDefaultGraph(goal), nil
	}

	// Parse LLM response
	var llmOutput struct {
		Nodes []struct {
			Type              string   `json:"type"`
			Label             string   `json:"label"`
			AcceptanceCriteria []string `json:"acceptanceCriteria"`
		} `json:"nodes"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &llmOutput); err != nil {
		return g.generateDefaultGraph(goal), nil
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

	return graph, nil
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
	// Simple ID generation - in production use UUID
	return fmt.Sprintf("graph-%d", currentTimeMillis())
}

func currentTime() string {
	return currentTimeMillisStr()
}

func currentTimeMillis() int64 {
	// Placeholder - use actual time in production
	return 1714656000000 // static for now
}

func currentTimeMillisStr() string {
	return "2026-05-02T12:00:00Z"
}
