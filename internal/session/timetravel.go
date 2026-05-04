package session

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// validNodeID matches typical node ID format (alphanumeric, hyphens, underscores)
var validNodeID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// CheckoutNodeState restores the workspace to the state it was in after a specific node completed.
// It finds the git commit corresponding to the node completion and checks out that commit.
// The workspace is put in detached HEAD state.
func (m *Manager) CheckoutNodeState(sessionID, nodeID string) error {
	if err := validateSessionID(sessionID); err != nil {
		return fmt.Errorf("timetravel: %w", err)
	}
	if nodeID == "" {
		return fmt.Errorf("timetravel: nodeID cannot be empty")
	}

	workspaceDir, err := m.WorkspacePath(sessionID)
	if err != nil {
		return fmt.Errorf("timetravel: %w", err)
	}

	if !isGitRepo(workspaceDir) {
		return fmt.Errorf("timetravel: workspace %s is not a git repository", workspaceDir)
	}

	// Find the commit hash for the node
	commitHash, err := findNodeCommit(workspaceDir, nodeID)
	if err != nil {
		return fmt.Errorf("timetravel: %w", err)
	}

	// git checkout -f <commit-hash> (force to handle dirty working tree)
	checkoutCmd := exec.Command("git", "checkout", "-f", commitHash)
	checkoutCmd.Dir = workspaceDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("timetravel: git checkout failed: %s: %w", string(output), err)
	}

	return nil
}

// findNodeCommit searches the git log for the commit corresponding to a node completion
func findNodeCommit(workspaceDir, nodeID string) (string, error) {
	// Search git log for commits matching the node completion pattern.
	// Commit messages are: "Node <nodeID> completed: <status> [<timestamp>]"
	// Use regex with escaped nodeID to prevent false positives (node-1 matching node-10)
	escapedID := regexp.QuoteMeta(nodeID)
	pattern := "Node " + escapedID + " completed:"
	logCmd := exec.Command("git", "log", "--oneline", "-E", "--grep="+pattern)
	logCmd.Dir = workspaceDir
	output, err := logCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("find commit: git log failed: %s", string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", fmt.Errorf("no commit found for node %s", nodeID)
	}

	// Take the most recent matching commit (first line)
	parts := strings.SplitN(lines[0], " ", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", fmt.Errorf("could not parse commit hash from: %s", lines[0])
	}

	return parts[0], nil
}

// GetNodeCommitHash returns the git commit hash for a specific node completion without checking out.
// Returns the most recent commit matching the node ID.
func (m *Manager) GetNodeCommitHash(sessionID, nodeID string) (string, error) {
	if err := validateSessionID(sessionID); err != nil {
		return "", fmt.Errorf("timetravel: %w", err)
	}
	if nodeID == "" {
		return "", fmt.Errorf("timetravel: nodeID cannot be empty")
	}

	workspaceDir, err := m.WorkspacePath(sessionID)
	if err != nil {
		return "", fmt.Errorf("timetravel: %w", err)
	}

	if !isGitRepo(workspaceDir) {
		return "", fmt.Errorf("timetravel: workspace is not a git repository")
	}

	return findNodeCommit(workspaceDir, nodeID)
}
