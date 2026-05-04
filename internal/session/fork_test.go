package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForkSession_Basic(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	parent, err := mgr.CreateSessionWithName(nil, "test-fork-parent")
	require.NoError(t, err)

	// Write a file in parent workspace
	testFile := filepath.Join(parent.Workspace, "parent-file.txt")
	err = os.WriteFile(testFile, []byte("parent content"), 0644)
	require.NoError(t, err)

	fork, err := mgr.ForkSession(nil, parent.ID)
	require.NoError(t, err)

	assert.NotEqual(t, parent.ID, fork.ID)
	assert.Equal(t, parent.Name+"-fork", fork.Name)
	assert.Equal(t, parent.Goal, fork.Goal)
	assert.Equal(t, StatusRunning, fork.Status)

	// Verify forked workspace has the file
	forkedFile := filepath.Join(fork.Workspace, "parent-file.txt")
	content, err := os.ReadFile(forkedFile)
	require.NoError(t, err)
	assert.Equal(t, "parent content", string(content))
}

func TestForkSession_WithGit(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	parent, err := mgr.CreateSessionWithName(nil, "test-fork-git")
	require.NoError(t, err)

	workspaceDir := parent.Workspace
	setupGitWorkspace(t, workspaceDir)

	// Create and commit a file
	testFile := filepath.Join(workspaceDir, "committed.txt")
	err = os.WriteFile(testFile, []byte("committed content"), 0644)
	require.NoError(t, err)

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = workspaceDir
	_, _ = cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "initial commit")
	cmd.Dir = workspaceDir
	_, _ = cmd.CombinedOutput()

	fork, err := mgr.ForkSession(nil, parent.ID)
	require.NoError(t, err)

	// Verify fork workspace has .git directory
	gitDir := filepath.Join(fork.Workspace, ".git")
	_, err = os.Stat(gitDir)
	assert.NoError(t, err, "fork workspace should have .git directory")

	// Verify git log in fork
	cmd = exec.Command("git", "log", "--oneline")
	cmd.Dir = fork.Workspace
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "initial commit")
}

func TestForkSession_NonExistent(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)
	_ = mgr

	_, err := mgr.ForkSession(nil, "sess-nonexistent")
	assert.Error(t, err)
}

func TestForkSession_InvalidID(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)
	_ = mgr

	_, err := mgr.ForkSession(nil, "invalid-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session ID")
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir() + "/dst"

	// Create nested structure
	os.MkdirAll(filepath.Join(src, "a", "b"), 0755)
	os.WriteFile(filepath.Join(src, "root.txt"), []byte("root"), 0644)
	os.WriteFile(filepath.Join(src, "a", "nested.txt"), []byte("nested"), 0644)
	os.WriteFile(filepath.Join(src, "a", "b", "deep.txt"), []byte("deep"), 0644)

	err := copyDir(src, dst)
	require.NoError(t, err)

	// Verify
	content, err := os.ReadFile(filepath.Join(dst, "root.txt"))
	require.NoError(t, err)
	assert.Equal(t, "root", string(content))

	content, err = os.ReadFile(filepath.Join(dst, "a", "nested.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested", string(content))

	content, err = os.ReadFile(filepath.Join(dst, "a", "b", "deep.txt"))
	require.NoError(t, err)
	assert.Equal(t, "deep", string(content))
}
