package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestManagerWithGit(t *testing.T) (*Manager, string) {
	t.Helper()
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = mgr.Close() })
	return mgr, tmpDir
}

func setupGitWorkspace(t *testing.T, workspaceDir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = workspaceDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git init failed: %s", string(out))

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = workspaceDir
	_, _ = cmd.CombinedOutput()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = workspaceDir
	_, _ = cmd.CombinedOutput()
}

func TestAutoCommit_NonGitRepo(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	// Create a session to get a valid session ID
	sess, err := mgr.CreateSessionWithName(nil, "test-autocommit")
	require.NoError(t, err)

	// Do NOT initialize git — should fail
	err = mgr.AutoCommit(sess.ID, "node-1", "complete")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestAutoCommit_NoChanges(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-autocommit-no-changes")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	// No file changes — should succeed with no commit
	err = mgr.AutoCommit(sess.ID, "node-1", "complete")
	assert.NoError(t, err)
}

func TestAutoCommit_WithChanges(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-autocommit-changes")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	// Create a file in workspace
	testFile := filepath.Join(workspaceDir, "hello.txt")
	err = os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	// AutoCommit should commit the new file
	err = mgr.AutoCommit(sess.ID, "node-1", "complete")
	assert.NoError(t, err)

	// Verify commit exists in git log
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Node node-1 completed: complete")
}

func TestAutoCommit_DeterministicCommitMessage(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)

	sess, err := mgr.CreateSessionWithName(nil, "test-deterministic-msg")
	require.NoError(t, err)

	workspaceDir := sess.Workspace
	setupGitWorkspace(t, workspaceDir)

	testFile := filepath.Join(workspaceDir, "data.txt")
	err = os.WriteFile(testFile, []byte("test data"), 0644)
	require.NoError(t, err)

	err = mgr.AutoCommit(sess.ID, "node-42", "failed")
	assert.NoError(t, err)

	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	msg := string(output)
	assert.Contains(t, msg, "Node node-42 completed: failed")
	// Should contain ISO 8601 timestamp (bracketed)
	assert.Contains(t, msg, "[20")
}

func TestAutoCommit_InvalidSessionID(t *testing.T) {
	mgr, _ := setupTestManagerWithGit(t)
	_ = mgr

	err := mgr.AutoCommit("invalid-id", "node-1", "complete")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session ID")
}

func TestIsGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Not a git repo
	assert.False(t, isGitRepo(tmpDir))

	// Initialize git
	setupGitWorkspace(t, tmpDir)
	assert.True(t, isGitRepo(tmpDir))
}
