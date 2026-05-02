package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExecutor(t *testing.T) {
	graph := &Graph{
		ID:   "test-graph-1",
		Goal: "Test goal",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusPending},
		},
	}
	exec := NewExecutor(graph, nil, nil)
	assert.NotNil(t, exec)
	assert.Equal(t, graph, exec.graph)
}

func TestNodeStatusUpdate(t *testing.T) {
	graph := &Graph{
		ID:   "test-graph-2",
		Goal: "Test goal",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusPending},
		},
	}
	exec := NewExecutor(graph, nil, nil)

	// Update status
	exec.updateNodeStatus("node-1", NodeStatusRunning, 0.5)

	// Check if node was updated
	assert.Equal(t, NodeStatusRunning, graph.Nodes[0].Status)
	assert.Equal(t, 0.5, graph.Nodes[0].Progress)
}

func TestCheckAcceptanceCriteria(t *testing.T) {
	exec := &Executor{}
	node := &Node{
		AcceptanceCriteria: []string{"Output must be non-empty"},
	}

	// Empty output should fail
	assert.False(t, exec.checkAcceptanceCriteria(node, ""))

	// Short output should fail
	assert.False(t, exec.checkAcceptanceCriteria(node, "short"))

	// Output containing the criterion should pass
	assert.True(t, exec.checkAcceptanceCriteria(node, "Output must be non-empty: This is a long enough output for the test"))
}

func TestCheckAcceptanceCriteriaNoCriteria(t *testing.T) {
	exec := &Executor{}
	node := &Node{
		AcceptanceCriteria: []string{},
	}

	// No criteria - any reasonable output should pass
	assert.True(t, exec.checkAcceptanceCriteria(node, "This is a long enough output"))
}

func TestBuildContext(t *testing.T) {
	graph := &Graph{
		ID: "test-graph-3",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusComplete},
		},
	}
	exec := NewExecutor(graph, nil, nil)

	// Build context for node at index 1 (should be empty since store is nil)
	ctx := exec.buildContext(context.Background(), 1)
	// Context will be empty because store is nil
	assert.Equal(t, "", ctx)
}
