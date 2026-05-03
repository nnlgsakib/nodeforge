package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashNodeConsistency(t *testing.T) {
	node := &Node{
		ID:               "node-1",
		Type:             NodeTypeGoal,
		Label:            "Test Goal",
		Output:           "Test output",
		AcceptanceCriteria: []string{"criteria1", "criteria2"},
	}

	hash1 := hashNode(node)
	hash2 := hashNode(node)

	assert.Equal(t, hash1, hash2, "hashNode should be deterministic")
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 64) // SHA-256 hex length
}

func TestHashNodeDifferentNodes(t *testing.T) {
	node1 := &Node{
		ID:     "node-1",
		Type:   NodeTypeGoal,
		Label:  "Goal A",
		Output: "Output A",
	}
	node2 := &Node{
		ID:     "node-2",
		Type:   NodeTypeGoal,
		Label:  "Goal B",
		Output: "Output B",
	}

	hash1 := hashNode(node1)
	hash2 := hashNode(node2)

	assert.NotEqual(t, hash1, hash2, "different nodes should have different hashes")
}

func TestHashNodeACOrderIndependence(t *testing.T) {
	node1 := &Node{
		ID:     "node-1",
		Type:   NodeTypeGoal,
		Label:  "Goal",
		Output: "Output",
		AcceptanceCriteria: []string{"criteria1", "criteria2"},
	}
	node2 := &Node{
		ID:     "node-1",
		Type:   NodeTypeGoal,
		Label:  "Goal",
		Output: "Output",
		AcceptanceCriteria: []string{"criteria2", "criteria1"},
	}

	assert.Equal(t, hashNode(node1), hashNode(node2),
		"hash should be independent of AC slice order")
}

func TestHashEdge(t *testing.T) {
	edge := &Edge{
		ID:     "edge-1",
		Source: "node-1",
		Target: "node-2",
		Type:   "default",
	}

	hash := hashEdge(edge)
	assert.Len(t, hash, 64)
	assert.NotEmpty(t, hash)
}

func TestComputeGraphHash(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal"},
		{ID: "node-2", Type: NodeTypeSpec, Label: "Spec"},
	}
	edges := []*Edge{
		{ID: "edge-1", Source: "node-1", Target: "node-2", Type: "default"},
	}

	hash1 := computeGraphHash(nodes, edges)
	hash2 := computeGraphHash(nodes, edges)

	assert.Equal(t, hash1, hash2, "computeGraphHash should be deterministic")
	assert.Len(t, hash1, 64)
}

func TestComputeGraphHashEmpty(t *testing.T) {
	hash := computeGraphHash(nil, nil)
	assert.Len(t, hash, 64)
}

func TestDetectChangedNodesNoOldHash(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal"},
		{ID: "node-2", Type: NodeTypeSpec, Label: "Spec"},
	}
	edges := []*Edge{}

	changed, err := detectChangedNodes("", nil, nodes, edges)
	require.NoError(t, err)
	assert.Len(t, changed, 2)
	assert.Contains(t, changed, "node-1")
	assert.Contains(t, changed, "node-2")
}

func TestDetectChangedNodesAllUnchanged(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Output: "output1"},
		{ID: "node-2", Type: NodeTypeSpec, Label: "Spec", Output: "output2"},
	}
	edges := []*Edge{
		{Source: "node-1", Target: "node-2", Type: "default"},
	}

	// Compute initial hashes
	rootHash := computeGraphHash(nodes, edges)
	nodeHashes := make(map[string]string)
	for _, node := range nodes {
		nodeHashes[node.ID] = hashNode(node)
	}

	changed, err := detectChangedNodes(rootHash, nodeHashes, nodes, edges)
	require.NoError(t, err)
	assert.Empty(t, changed, "no nodes should be changed when nothing changed")
}

func TestDetectChangedNodesOneChanged(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Output: "output1"},
		{ID: "node-2", Type: NodeTypeSpec, Label: "Spec", Output: "output2"},
	}
	edges := []*Edge{}

	rootHash := computeGraphHash(nodes, edges)
	nodeHashes := make(map[string]string)
	for _, node := range nodes {
		nodeHashes[node.ID] = hashNode(node)
	}

	// Change only node-2's label (part of hash)
	nodes[1].Label = "New Spec"

	changed, err := detectChangedNodes(rootHash, nodeHashes, nodes, edges)
	require.NoError(t, err)
	assert.Len(t, changed, 1)
	assert.Contains(t, changed, "node-2")
}

func TestDetectChangedNodesEdgeChange(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Output: "output1"},
		{ID: "node-2", Type: NodeTypeSpec, Label: "Spec", Output: "output2"},
	}
	edges := []*Edge{
		{Source: "node-1", Target: "node-2", Type: "default"},
	}

	rootHash := computeGraphHash(nodes, edges)
	nodeHashes := make(map[string]string)
	for _, node := range nodes {
		nodeHashes[node.ID] = hashNode(node)
	}

	// Change edge
	edges[0].Type = "tension"

	changed, err := detectChangedNodes(rootHash, nodeHashes, nodes, edges)
	require.NoError(t, err)
	// Edge change should mark all nodes as needing re-verification
	assert.NotEmpty(t, changed)
}

func TestDetectChangedNodesSingleNodeGraph(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal", Output: "output1"},
	}
	edges := []*Edge{}

	rootHash := computeGraphHash(nodes, edges)
	nodeHashes := map[string]string{"node-1": hashNode(nodes[0])}

	// No change
	changed, err := detectChangedNodes(rootHash, nodeHashes, nodes, edges)
	require.NoError(t, err)
	assert.Empty(t, changed)

	// Change
	nodes[0].Label = "New Goal"
	changed, err = detectChangedNodes(rootHash, nodeHashes, nodes, edges)
	require.NoError(t, err)
	assert.Len(t, changed, 1)
	assert.Contains(t, changed, "node-1")
}

func TestDetectChangedNodesNilOldNodeHashes(t *testing.T) {
	nodes := []*Node{
		{ID: "node-1", Type: NodeTypeGoal, Label: "Goal"},
	}
	edges := []*Edge{}

	// Old hash provided but no old node hashes
	changed, err := detectChangedNodes("some-hash", nil, nodes, edges)
	require.NoError(t, err)
	// Should detect all nodes as changed
	assert.Len(t, changed, 1)
}

// Benchmark: 100-node graph with 95% unchanged nodes re-executes in <2s (NFR-06)
func BenchmarkMerkleReexecution100Nodes(b *testing.B) {
	// Build 100-node graph
	nodes := make([]*Node, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = &Node{
			ID:               "node-" + string(rune('0'+i%10)),
			Type:             NodeTypeGoal,
			Label:            "Goal",
			Output:           "Stable output " + string(rune(i)),
			AcceptanceCriteria: []string{"criteria"},
		}
	}
	edges := make([]*Edge, 99)
	for i := 0; i < 99; i++ {
		edges[i] = &Edge{Source: nodes[i].ID, Target: nodes[i+1].ID, Type: "default"}
	}

	// Compute initial hash
	rootHash := computeGraphHash(nodes, edges)
	nodeHashes := make(map[string]string)
	for _, node := range nodes {
		nodeHashes[node.ID] = hashNode(node)
	}

	// Change 5% of nodes (5 out of 100)
	for i := 0; i < 5; i++ {
		nodes[i*20].Output = "Changed output"
	}

	b.ResetTimer()
	b.ReportAllocs()

	start := time.Now()
	for i := 0; i < b.N; i++ {
		_, err := detectChangedNodes(rootHash, nodeHashes, nodes, edges)
		if err != nil {
			b.Fatal(err)
		}
	}

	elapsed := time.Since(start)
	b.Logf("100-node Merkle re-execution: %v for %d iterations", elapsed, b.N)

	// Verify single iteration is well under 2s
	if b.N == 1 && elapsed > 2*time.Second {
		b.Fatalf("Merkle re-execution took %v, expected <2s", elapsed)
	}
}
