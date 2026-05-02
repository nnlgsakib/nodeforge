package engine

import (
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

	// Long enough output should pass
	assert.True(t, exec.checkAcceptanceCriteria(node, "This is a long enough output for the test"))
}

func TestBuildContext(t *testing.T) {
	graph := &Graph{
		ID: "test-graph-3",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusComplete},
		},
	}
	exec := NewExecutor(graph, nil, nil)

	// Build context for node at index 1 (should include node-1 output if available)
	ctx := exec.buildContext(1)
	assert.NotEmpty(t, ctx)
}
