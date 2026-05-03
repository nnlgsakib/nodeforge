package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSpec_GoalMode(t *testing.T) {
	data := []byte(`goal: "Convert JS->Go project"`)

	spec, err := ParseSpec(data)
	require.NoError(t, err)
	assert.Equal(t, SpecModeGoal, spec.Mode)
	assert.Equal(t, "Convert JS->Go project", spec.Goal)
	assert.Nil(t, spec.Nodes)
}

func TestParseSpec_GraphMode(t *testing.T) {
	data := []byte(`
nodes:
  - id: goal-1
    type: Goal
    label: "Convert JS->Go"
  - id: spec-1
    type: Spec
    label: "Generate Spec"
edges:
  - source: goal-1
    target: spec-1
`)

	spec, err := ParseSpec(data)
	require.NoError(t, err)
	assert.Equal(t, SpecModeGraph, spec.Mode)
	assert.Len(t, spec.Nodes, 2)
	assert.Len(t, spec.Edges, 1)

	assert.Equal(t, "goal-1", spec.Nodes[0].ID)
	assert.Equal(t, "Goal", spec.Nodes[0].Type)
	assert.Equal(t, "Convert JS->Go", spec.Nodes[0].Label)

	assert.Equal(t, "goal-1", spec.Edges[0].Source)
	assert.Equal(t, "spec-1", spec.Edges[0].Target)
}

func TestParseSpec_MissingNodeID(t *testing.T) {
	data := []byte(`
nodes:
  - type: Goal
    label: "Test"
`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field 'id'")
}

func TestParseSpec_MissingNodeType(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    label: "Test"
`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field 'type'")
}

func TestParseSpec_MissingNodeLabel(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    type: Goal
`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field 'label'")
}

func TestParseSpec_MissingEdgeSource(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges:
  - target: n2
`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field 'source'")
}

func TestParseSpec_MissingEdgeTarget(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges:
  - source: n1
`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field 'target'")
}

func TestParseSpec_InvalidYAML(t *testing.T) {
	data := []byte(`: invalid: yaml: [`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestParseSpec_EmptySpec(t *testing.T) {
	data := []byte(`{}`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must contain either a 'goal' field")
}

func TestParseSpec_GoalWithNodes(t *testing.T) {
	// When both goal and nodes are present, treat as graph mode
	data := []byte(`
goal: "Test"
nodes:
  - id: n1
    type: Goal
    label: "Test"
`)

	spec, err := ParseSpec(data)
	require.NoError(t, err)
	assert.Equal(t, SpecModeGraph, spec.Mode)
}

func TestParseSpec_DuplicateNodeID(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
  - id: n1
    type: Spec
    label: "Duplicate"
`)

	_, err := ParseSpec(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate node ID")
}

func TestParseSpec_EdgeReferencesNonExistentNode(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges:
  - source: n1
    target: nonexistent
`)

	spec, err := ParseSpec(data)
	require.NoError(t, err)

	_, err = spec.ToGraph()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "references non-existent node")
}

func TestParseSpec_EdgeReferencesNonExistentSource(t *testing.T) {
	data := []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges:
  - source: missing
    target: n1
`)

	spec, err := ParseSpec(data)
	require.NoError(t, err)

	_, err = spec.ToGraph()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "references non-existent node")
}

func TestSpecFile_ToGraph_GraphMode(t *testing.T) {
	data := []byte(`
nodes:
  - id: goal-1
    type: Goal
    label: "Convert"
  - id: spec-1
    type: Spec
    label: "Spec"
edges:
  - source: goal-1
    target: spec-1
`)

	spec, err := ParseSpec(data)
	require.NoError(t, err)

	graph, err := spec.ToGraph()
	require.NoError(t, err)

	assert.Len(t, graph.Nodes, 2)
	assert.Equal(t, "goal-1", graph.Nodes[0].ID)
	assert.Equal(t, NodeType("Goal"), graph.Nodes[0].Type)
	assert.Equal(t, NodeStatusPending, graph.Nodes[0].Status)

	assert.Len(t, graph.Edges, 1)
	assert.Equal(t, "goal-1", graph.Edges[0].Source)
	assert.Equal(t, "spec-1", graph.Edges[0].Target)
}

func TestSpecFile_ToGraph_GoalMode_Error(t *testing.T) {
	spec := &SpecFile{Mode: SpecModeGoal, Goal: "test"}
	_, err := spec.ToGraph()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "goal-mode spec cannot be converted directly")
}

func TestParseSpecFile_FileNotFound(t *testing.T) {
	_, err := ParseSpecFile("/nonexistent/path/spec.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read spec file")
}

func TestParseSpecFile_ValidFile(t *testing.T) {
	// Create temp spec file
	dir := t.TempDir()
	specPath := filepath.Join(dir, "test-spec.yaml")
	err := os.WriteFile(specPath, []byte(`goal: "Test goal"`), 0644)
	require.NoError(t, err)

	spec, err := ParseSpecFile(specPath)
	require.NoError(t, err)
	assert.Equal(t, SpecModeGoal, spec.Mode)
	assert.Equal(t, "Test goal", spec.Goal)
}
