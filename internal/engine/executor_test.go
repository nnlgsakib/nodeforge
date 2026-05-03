package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nnlgsakib/nodeforge/internal/llm"
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
	exec := NewExecutor(graph, nil, nil, nil, nil)
	assert.NotNil(t, exec)
	assert.Equal(t, graph, exec.graph)
	assert.NotNil(t, exec.swarmConfig)
	assert.False(t, exec.swarmConfig.Enabled)
}

func TestExecutorSetSwarm(t *testing.T) {
	graph := &Graph{
		ID:   "test-graph-swarm",
		Goal: "Test goal",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusPending},
		},
	}
	exec := NewExecutor(graph, nil, nil, nil, nil)

	swarm := llm.NewSwarm(nil, nil, nil)
	config := &llm.SwarmConfig{Enabled: true, MaxAttempts: 5}
	exec.SetSwarm(swarm, config)

	assert.Equal(t, swarm, exec.swarm)
	assert.True(t, exec.swarmConfig.Enabled)
	assert.Equal(t, 5, exec.swarmConfig.MaxAttempts)
}

func TestExecutorSetSwarmNilConfig(t *testing.T) {
	graph := &Graph{ID: "test", Goal: "test", Nodes: []*Node{{ID: "n1", Type: NodeTypeGoal}}}
	exec := NewExecutor(graph, nil, nil, nil, nil)

	swarm := llm.NewSwarm(nil, nil, nil)
	exec.SetSwarm(swarm, nil)

	assert.Equal(t, swarm, exec.swarm)
	assert.NotNil(t, exec.swarmConfig) // should retain default
}

// mockBroadcaster captures broadcast messages for testing
type mockBroadcaster struct {
	mu            sync.Mutex
	rawMessages   [][]byte
	nodeUpdates   []nodeUpdate
	edgeUpdates   []edgeUpdate
}

type nodeUpdate struct {
	nodeID   string
	status   string
	progress float64
}

type edgeUpdate struct {
	source  string
	target  string
	tension float64
}

func (m *mockBroadcaster) BroadcastNodeUpdate(nodeID, status string, progress float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodeUpdates = append(m.nodeUpdates, nodeUpdate{nodeID, status, progress})
}

func (m *mockBroadcaster) BroadcastEdgeUpdate(source, target string, tension float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.edgeUpdates = append(m.edgeUpdates, edgeUpdate{source, target, tension})
}

func (m *mockBroadcaster) BroadcastRaw(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rawMessages = append(m.rawMessages, data)
}

func TestBroadcastSpeculativeStart(t *testing.T) {
	graph := &Graph{
		ID:   "test-graph-spec",
		Goal: "Test goal",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusPending},
		},
	}
	exec := NewExecutor(graph, nil, nil, nil, nil)

	hub := &mockBroadcaster{}
	exec.SetHub(hub)

	config := &llm.SwarmConfig{Enabled: true, MaxAttempts: 4}
	exec.swarmConfig = config

	exec.broadcastSpeculativeStart("node-1")

	assert.Len(t, hub.rawMessages, 1)
	// Message should contain speculative: true and attempts: 4
	msg := string(hub.rawMessages[0])
	assert.Contains(t, msg, `"speculative":true`)
	assert.Contains(t, msg, `"attempts":4`)
	assert.Contains(t, msg, `"nodeId":"node-1"`)
	assert.Contains(t, msg, `"status":"running"`)
}

func TestBroadcastSpeculativeStartNoHub(t *testing.T) {
	exec := &Executor{
		swarmConfig: &llm.SwarmConfig{Enabled: true, MaxAttempts: 3},
	}
	// Should not panic when hub is nil
	exec.broadcastSpeculativeStart("node-1")
}

func TestExecutorRunWithSimulatedSpeculative(t *testing.T) {
	graph := &Graph{
		ID:   "test-graph-exec",
		Goal: "Test goal",
		Nodes: []*Node{
			{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Status: NodeStatusPending,
				AcceptanceCriteria: []string{"Simulated output"}},
		},
	}
	exec := NewExecutor(graph, nil, nil, nil, nil)

	hub := &mockBroadcaster{}
	exec.SetHub(hub)

	// With nil llmProv, executor simulates output
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := exec.Run(ctx)
	assert.NoError(t, err)

	// Node should be complete
	assert.Equal(t, NodeStatusComplete, graph.Nodes[0].Status)
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
	exec := NewExecutor(graph, nil, nil, nil, nil)

	// Build context for node at index 1 (should be empty since store is nil)
	ctx := exec.buildContext(context.Background(), 1)
	// Context will be empty because store is nil
	assert.Equal(t, "", ctx)
}
