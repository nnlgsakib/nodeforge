package session

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

// AutoCommit stages all workspace changes and commits them with a deterministic message.
// It runs git add -A and git commit in the session's workspace directory.
// The commit message includes nodeID, status, and ISO 8601 timestamp for traceability.
func (m *Manager) AutoCommit(sessionID, nodeID, status string) error {
	if err := validateSessionID(sessionID); err != nil {
		return fmt.Errorf("autocommit: %w", err)
	}

	workspaceDir, err := m.WorkspacePath(sessionID)
	if err != nil {
		return fmt.Errorf("autocommit: %w", err)
	}

	// Check if git is initialized in workspace
	if !isGitRepo(workspaceDir) {
		return fmt.Errorf("autocommit: workspace %s is not a git repository", workspaceDir)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	commitMsg := fmt.Sprintf("Node %s completed: %s [%s]", nodeID, status, timestamp)

	// git add -A
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = workspaceDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("autocommit: git add failed: %s: %w", string(output), err)
	}

	// Check if there are changes to commit
	diffCmd := exec.Command("git", "diff", "--cached", "--quiet")
	diffCmd.Dir = workspaceDir
	if err := diffCmd.Run(); err == nil {
		// Exit code 0 means no differences — nothing to commit
		return nil
	}

	// git commit -m "<message>"
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = workspaceDir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("autocommit: git commit failed: %s: %w", string(output), err)
	}

	return nil
}

// isGitRepo checks whether the directory is a git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	// git rev-parse --git-dir returns .git (or path to .git) on success
	clean := string(output)
	return clean == ".git\n" || clean == ".git" || filepath.Clean(clean) == gitDir
}
