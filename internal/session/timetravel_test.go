package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutNodeState_Success(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-timetravel")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	// Create file and commit it with node completion message
	testFile := filepath.Join(workspaceDir, "v1.txt")
	err = os.WriteFile(testFile, []byte("version 1"), 0644)
	require.NoError(t, err)

	runGit(t, workspaceDir, "add", ".")
	runGit(t, workspaceDir, "commit", "-m", "Node node-1 completed: complete [2026-05-04T12:00:00Z]")

	// Modify file
	err = os.WriteFile(testFile, []byte("version 2"), 0644)
	require.NoError(t, err)
	runGit(t, workspaceDir, "add", ".")
	runGit(t, workspaceDir, "commit", "-m", "Node node-2 completed: complete [2026-05-04T13:00:00Z]")

	// Time-travel to node-1
	err = mgr.CheckoutNodeState(sess.ID, "node-1")
	require.NoError(t, err)

	// Verify file content is restored
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "version 1", string(content))
}

func TestCheckoutNodeState_NodeNotFound(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-timetravel-notfound")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	err = mgr.CheckoutNodeState(sess.ID, "nonexistent-node")
	assert.Error(t, err)
}

func TestCheckoutNodeState_NonGitRepo(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-timetravel-no-git")
	require.NoError(t, err)

	// Don't init git
	err = mgr.CheckoutNodeState(sess.ID, "node-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestCheckoutNodeState_InvalidSessionID(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)
	_ = mgr

	err := mgr.CheckoutNodeState("invalid-id", "node-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session ID")
}

func TestGetNodeCommitHash(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-get-commit")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	testFile := filepath.Join(workspaceDir, "data.txt")
	err = os.WriteFile(testFile, []byte("data"), 0644)
	require.NoError(t, err)

	runGit(t, workspaceDir, "add", ".")
	runGit(t, workspaceDir, "commit", "-m", "Node node-42 completed: complete [2026-05-04T12:00:00Z]")

	hash, err := mgr.GetNodeCommitHash(sess.ID, "node-42")
	require.NoError(t, err)
	assert.Len(t, hash, 7) // Short hash is 7 chars
}

func TestCheckoutNodeState_SimilarNodeIDs(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-timetravel-similar")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	// Create and commit node-1
	testFile := filepath.Join(workspaceDir, "data.txt")
	err = os.WriteFile(testFile, []byte("node-1 content"), 0644)
	require.NoError(t, err)
	runGit(t, workspaceDir, "add", ".")
	runGit(t, workspaceDir, "commit", "-m", "Node node-1 completed: complete [2026-05-04T12:00:00Z]")

	// Create and commit node-10 (should NOT match when looking for node-1)
	err = os.WriteFile(testFile, []byte("node-10 content"), 0644)
	require.NoError(t, err)
	runGit(t, workspaceDir, "add", ".")
	runGit(t, workspaceDir, "commit", "-m", "Node node-10 completed: complete [2026-05-04T13:00:00Z]")

	// Time-travel to node-1 — should get node-1 content, not node-10
	err = mgr.CheckoutNodeState(sess.ID, "node-1")
	require.NoError(t, err)

	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "node-1 content", string(content), "should restore node-1 content, not node-10")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(out))
}
