package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// hashNode computes a SHA-256 hash for a single node based on its type, label, and acceptance criteria.
// The hash is deterministic: identical nodes produce identical hashes.
// Output is excluded so that execution results don't trigger false change detection.
func hashNode(node *Node) string {
	if node == nil {
		return ""
	}
	// Build a canonical string representation of the node
	var parts []string
	parts = append(parts, string(node.Type))
	parts = append(parts, node.Label)

	// Sort acceptance criteria for deterministic ordering
	sortedAC := make([]string, len(node.AcceptanceCriteria))
	copy(sortedAC, node.AcceptanceCriteria)
	sort.Strings(sortedAC)
	parts = append(parts, sortedAC...)

	// Include metadata keys and values (sorted for determinism)
	if node.Metadata != nil {
		keys := make([]string, 0, len(node.Metadata))
		for k := range node.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%v", k, node.Metadata[k]))
		}
	}

	joined := strings.Join(parts, "|")
	h := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(h[:])
}

// hashEdge computes a SHA-256 hash for an edge.
func hashEdge(edge *Edge) string {
	joined := fmt.Sprintf("%s|%s|%s", edge.Source, edge.Target, edge.Type)
	h := sha256.Sum256([]byte(joined))
	return hex.EncodeToString(h[:])
}

// computeGraphHash computes the Merkle root hash for the entire graph.
// It concatenates all per-node hashes and edge hashes, then SHA-256 hashes the result.
func computeGraphHash(nodes []*Node, edges []*Edge) string {
	var nodeHashes []string
	for _, node := range nodes {
		nodeHashes = append(nodeHashes, hashNode(node))
	}

	var edgeHashes []string
	for _, edge := range edges {
		edgeHashes = append(edgeHashes, hashEdge(edge))
	}

	// Sort for deterministic ordering
	sort.Strings(nodeHashes)
	sort.Strings(edgeHashes)

	combined := strings.Join(nodeHashes, "") + "|" + strings.Join(edgeHashes, "")
	h := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(h[:])
}

// detectChangedNodes compares the old Merkle root hash against the current graph state
// and returns the list of node IDs that have changed. If oldHash is empty, all nodes
// are considered changed.
func detectChangedNodes(oldHash string, oldNodeHashes map[string]string, nodes []*Node, edges []*Edge) ([]string, error) {
	if oldHash == "" {
		// No previous hash — all nodes are new
		changed := make([]string, 0, len(nodes))
		for _, node := range nodes {
			changed = append(changed, node.ID)
		}
		return changed, nil
	}

	// Compute new per-node hashes
	newNodeHashes := make(map[string]string, len(nodes))
	for _, node := range nodes {
		newNodeHashes[node.ID] = hashNode(node)
	}

	// Compare per-node hashes to find changed nodes
	var changed []string
	for _, node := range nodes {
		if oldNodeHashes == nil || oldNodeHashes[node.ID] != newNodeHashes[node.ID] {
			changed = append(changed, node.ID)
		}
	}

	// Also detect structural changes (edges changed)
	newGraphHash := computeGraphHash(nodes, edges)
	if newGraphHash != oldHash {
		if len(changed) == 0 {
			// Edge-only change: mark all nodes as needing re-verification
			for _, node := range nodes {
				changed = append(changed, node.ID)
			}
		}
	}

	return changed, nil
}
