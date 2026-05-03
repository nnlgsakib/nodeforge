package context

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDB creates an in-memory BadgerDB for testing
func newTestDB(t *testing.T) *badger.DB {
	t.Helper()
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestKnowledgeGraph_AddNodeOutput(t *testing.T) {
	db := newTestDB(t)
	kg := NewKnowledgeGraph(db)

	// Test adding first node
	err := kg.AddNodeOutput("node1", "output1")
	assert.NoError(t, err)

	// Test adding second node (should create edge from node1 to node2)
	err = kg.AddNodeOutput("node2", "output2")
	assert.NoError(t, err)

	// Verify nodes were stored
	var nodes []ContextNode
	err = db.View(func(txn *badger.Txn) error {
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
				return err
			}
			nodes = append(nodes, node)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Len(t, nodes, 2)
}

func TestKnowledgeGraph_BuildContext(t *testing.T) {
	db := newTestDB(t)
	kg := NewKnowledgeGraph(db)

	// Add some nodes
	require.NoError(t, kg.AddNodeOutput("node1", "output1"))
	require.NoError(t, kg.AddNodeOutput("node2", "output2"))
	time.Sleep(10 * time.Millisecond) // Ensure timestamps differ

	// Build context for node3 with max 100 tokens
	context, err := kg.BuildContext(context.Background(),"node3", 100)
	assert.NoError(t, err)
	assert.Contains(t, context, "output1")
	assert.Contains(t, context, "output2")
}

func TestKnowledgeGraph_BuildContext_TokenBudget(t *testing.T) {
	db := newTestDB(t)
	kg := NewKnowledgeGraph(db)

	// Add large output
	largeOutput := string(make([]byte, 400)) // ~100 tokens
	require.NoError(t, kg.AddNodeOutput("node1", largeOutput))
	require.NoError(t, kg.AddNodeOutput("node2", "small output"))

	// Build context with 50 token limit (should only include part of large output)
	context, err := kg.BuildContext(context.Background(),"node3", 50)
	assert.NoError(t, err)
	// Should not include the large output (exceeds budget)
	assert.NotContains(t, context, largeOutput)
}

func TestKnowledgeGraph_GetDownstreamContext(t *testing.T) {
	db := newTestDB(t)
	kg := NewKnowledgeGraph(db)

	// Add nodes
	require.NoError(t, kg.AddNodeOutput("node1", "output1"))
	require.NoError(t, kg.AddNodeOutput("node2", "output2"))

	// Get downstream context for node2 (should include node1 as upstream)
	contextNodes := kg.GetDownstreamContext("node2")
	assert.Len(t, contextNodes, 1)
	assert.Equal(t, "node1", contextNodes[0].NodeID)
}

func TestKnowledgeGraph_BuildContext_Performance(t *testing.T) {
	db := newTestDB(t)
	kg := NewKnowledgeGraph(db)

	// Add 100 nodes
	for i := 0; i < 100; i++ {
		require.NoError(t, kg.AddNodeOutput(fmt.Sprintf("node%d", i), fmt.Sprintf("output%d", i)))
	}

	// Build context should complete in <100ms
	start := time.Now()
	_, err := kg.BuildContext(context.Background(),"node100", 10000)
	dur := time.Since(start)
	assert.NoError(t, err)
	assert.Less(t, dur, 100*time.Millisecond, "BuildContext took %v, exceeds 100ms NFR-04", dur)
}

func TestNodeMemory_StoreMemory(t *testing.T) {
	db := newTestDB(t)
	nm := NewNodeMemory(db)

	err := nm.StoreMemory("node1", "output", "test output")
	assert.NoError(t, err)

	// Verify it was stored
	val, ok := nm.GetMemory("node1", "output")
	assert.True(t, ok)
	assert.Equal(t, "test output", val)
}

func TestNodeMemory_GetMemory_NotFound(t *testing.T) {
	db := newTestDB(t)
	nm := NewNodeMemory(db)

	_, ok := nm.GetMemory("nonexistent", "key")
	assert.False(t, ok)
}

func TestNodeMemory_InjectMemoryIntoPrompt(t *testing.T) {
	db := newTestDB(t)
	nm := NewNodeMemory(db)

	// Store some memories
	require.NoError(t, nm.StoreMemory("node1", "mem1", "memory 1"))
	require.NoError(t, nm.StoreMemory("node1", "mem2", "memory 2"))

	prompt := "Original prompt"
	result := nm.InjectMemoryIntoPrompt(prompt, "node1")
	assert.Contains(t, result, "Original prompt")
	assert.Contains(t, result, "Memory Context:")
	assert.Contains(t, result, "memory 1")
	assert.Contains(t, result, "memory 2")
}

func TestNodeMemory_InjectMemoryIntoPrompt_NoMemories(t *testing.T) {
	db := newTestDB(t)
	nm := NewNodeMemory(db)

	prompt := "Original prompt"
	result := nm.InjectMemoryIntoPrompt(prompt, "node1")
	assert.Equal(t, prompt, result)
}

func TestSpecGenerator_GenerateSpec(t *testing.T) {
	db := newTestDB(t)
	sg := NewSpecGenerator(db)

	spec, err := sg.GenerateSpec("node1", "test output")
	assert.NoError(t, err)
	assert.Contains(t, spec, "node1")
	assert.Contains(t, spec, "test output")
}

func TestSpecGenerator_AddSystemReferences(t *testing.T) {
	db := newTestDB(t)
	sg := NewSpecGenerator(db)

	// Add initial references
	err := sg.AddSystemReferences("node1", []string{"ref1", "ref2"})
	assert.NoError(t, err)

	// Add more references
	err = sg.AddSystemReferences("node1", []string{"ref3"})
	assert.NoError(t, err)

	// Verify references were stored (by checking DB directly)
	var refs []string
	_ = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("sysref:node1"))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &refs)
		})
	})

	assert.Len(t, refs, 3)
	assert.Contains(t, refs, "ref1")
	assert.Contains(t, refs, "ref3")
}

func TestSpecGenerator_SpecToGraph(t *testing.T) {
	sg := NewSpecGenerator(nil) // DB not needed for SpecToGraph

	spec := `Node:goal1
Node:spec1
Node:plan1`

	nodes, edges, err := sg.SpecToGraph(spec)
	assert.NoError(t, err)
	assert.Len(t, nodes, 3)
	assert.Len(t, edges, 2) // Two edges connecting the three nodes
	assert.Equal(t, "goal1", nodes[0].ID)
}

func TestGraphSplitter_SplitGraphIfNeeded_NoSplit(t *testing.T) {
	db := newTestDB(t)
	gs := NewGraphSplitter(db)
	kg := NewKnowledgeGraph(db)

	// Add small nodes
	require.NoError(t, kg.AddNodeOutput("node1", "small output"))
	require.NoError(t, kg.AddNodeOutput("node2", "small output 2"))

	// Should not split (total tokens < 100)
	subGraphs, err := gs.SplitGraphIfNeeded("node3", 10000)
	assert.NoError(t, err)
	assert.Len(t, subGraphs, 1)
}

func TestGraphSplitter_SplitGraphIfNeeded_Split(t *testing.T) {
	db := newTestDB(t)
	gs := NewGraphSplitter(db)
	kg := NewKnowledgeGraph(db)

	// Add large output (400 chars = ~100 tokens)
	largeOutput := string(make([]byte, 400))
	require.NoError(t, kg.AddNodeOutput("node1", largeOutput))
	require.NoError(t, kg.AddNodeOutput("node2", largeOutput))
	require.NoError(t, kg.AddNodeOutput("node3", "small"))

	// Split with 150 token limit (should split into 2 sub-graphs)
	subGraphs, err := gs.SplitGraphIfNeeded("node4", 150)
	assert.NoError(t, err)
	assert.Greater(t, len(subGraphs), 1)
}

func TestEstimateSubGraphTokens(t *testing.T) {
	sg := SubGraph{
		Nodes: []ContextNode{
			{Output: "12345678"}, // 8 chars = 2 tokens
			{Output: "1234"},      // 4 chars = 1 token
		},
	}

	tokens := EstimateSubGraphTokens(sg)
	assert.Equal(t, 3, tokens) // 12 chars /4 = 3 tokens
}
