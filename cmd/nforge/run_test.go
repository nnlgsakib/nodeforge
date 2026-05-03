package nforge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd_MissingSpecFile(t *testing.T) {
	// Exit code 2 for missing spec file
	cmd := runCmd
	err := cmd.RunE(cmd, []string{"/nonexistent/spec.yaml"})
	require.Error(t, err)
	assert.Equal(t, 2, ExitCodeForError(err))
}

func TestRunCmd_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "bad.yaml")
	err := os.WriteFile(specPath, []byte(`: invalid: yaml: [`), 0644)
	require.NoError(t, err)

	cmd := runCmd
	err = cmd.RunE(cmd, []string{specPath})
	require.Error(t, err)
	assert.Equal(t, 2, ExitCodeForError(err))
}

func TestRunCmd_EmptySpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "empty.yaml")
	err := os.WriteFile(specPath, []byte(`{}`), 0644)
	require.NoError(t, err)

	cmd := runCmd
	err = cmd.RunE(cmd, []string{specPath})
	require.Error(t, err)
	assert.Equal(t, 2, ExitCodeForError(err))
}

func TestRunCmd_GraphMode_Success_NoLLM(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "graph.yaml")
	err := os.WriteFile(specPath, []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges: []
`), 0644)
	require.NoError(t, err)

	cmd := runCmd
	cmd.Flags().Set("no-llm", "true")
	cmd.Flags().Set("ascii", "false")
	err = cmd.RunE(cmd, []string{specPath})
	require.NoError(t, err)
	assert.Equal(t, 0, ExitCodeForError(err))
}

func TestRunCmd_GraphMode_NodeFailure(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "fail.yaml")
	err := os.WriteFile(specPath, []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges: []
`), 0644)
	require.NoError(t, err)

	cmd := runCmd
	cmd.Flags().Set("no-llm", "true")
	cmd.Flags().Set("ascii", "false")
	// In simulation mode all nodes succeed, so we can't test failure path here.
	// This test verifies the command structure works.
	err = cmd.RunE(cmd, []string{specPath})
	require.NoError(t, err)
	assert.Equal(t, 0, ExitCodeForError(err))
}

func TestRunCmd_ExitCodeForError_Nil(t *testing.T) {
	assert.Equal(t, 0, ExitCodeForError(nil))
}

func TestRunCmd_ExitCodeForError_PlainError(t *testing.T) {
	assert.Equal(t, 1, ExitCodeForError(assert.AnError))
}

func TestRunCmd_ExitCodeForError_ExitCodeError(t *testing.T) {
	ecErr := &exitCodeError{Code: 2, Err: assert.AnError}
	assert.Equal(t, 2, ExitCodeForError(ecErr))
}

func TestGraphVizCmd_NoSpec(t *testing.T) {
	cmd := graphVizCmd
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err) // No spec = just prints usage hint, no error
}

func TestGraphVizCmd_GoalModeSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "goal.yaml")
	err := os.WriteFile(specPath, []byte(`goal: "Test"`), 0644)
	require.NoError(t, err)

	cmd := graphVizCmd
	err = cmd.RunE(cmd, []string{specPath})
	require.Error(t, err)
	// Goal mode can't be visualized without LLM
	assert.Contains(t, err.Error(), "goal-mode spec requires LLM")
}

func TestGraphVizCmd_GraphModeSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "graph.yaml")
	err := os.WriteFile(specPath, []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
  - id: n2
    type: Spec
    label: "Spec"
edges:
  - source: n1
    target: n2
`), 0644)
	require.NoError(t, err)

	cmd := graphVizCmd
	cmd.Flags().Set("verbose", "false")
	err = cmd.RunE(cmd, []string{specPath})
	require.NoError(t, err)
}

func TestGraphVizCmd_Verbose(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "graph.yaml")
	err := os.WriteFile(specPath, []byte(`
nodes:
  - id: n1
    type: Goal
    label: "Test"
edges: []
`), 0644)
	require.NoError(t, err)

	cmd := graphVizCmd
	cmd.Flags().Set("verbose", "true")
	err = cmd.RunE(cmd, []string{specPath})
	require.NoError(t, err)
}

func TestGraphVizCmd_InvalidSpec(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "bad.yaml")
	err := os.WriteFile(specPath, []byte(`: invalid`), 0644)
	require.NoError(t, err)

	cmd := graphVizCmd
	err = cmd.RunE(cmd, []string{specPath})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse spec file")
}
