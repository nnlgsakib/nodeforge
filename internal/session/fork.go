package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ForkSession creates a new session by forking an existing one.
// It creates a new session entry with a new ID, copies the workspace state,
// and creates a Git branch from the current commit if the workspace is a git repo.
// The forked session inherits the goal and name (with "-fork" suffix) from the parent.
func (m *Manager) ForkSession(ctx context.Context, parentID string) (*Session, error) {
	if err := validateSessionID(parentID); err != nil {
		return nil, fmt.Errorf("fork: %w", err)
	}

	// Use background context if nil provided
	if ctx == nil {
		ctx = context.Background()
	}

	// Load parent session (brief lock)
	m.mu.Lock()
	parent, err := m.GetSession(ctx, parentID)
	m.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("fork: %w", err)
	}

	parentWorkspace := parent.Workspace

	// Generate fork session ID
	forkID := generateSessionID()
	// Strip existing -fork suffixes to prevent parent-fork-fork-fork... cascade
	baseName := strings.TrimSuffix(parent.Name, "-fork")
	forkName := baseName + "-fork"

	// Create fork directory structure
	forkProjectDir := filepath.Join(m.workspaceRoot, ".nforge", "sessions", forkID)
	forkWorkspaceDir := filepath.Join(forkProjectDir, "workspace")

	// Copy workspace directory (files only, not .git — we'll branch)
	if err := copyDir(parentWorkspace, forkWorkspaceDir); err != nil {
		return nil, fmt.Errorf("fork: failed to copy workspace: %w", err)
	}

	// Initialize .nforge workspace metadata for fork
	if err := InitProjectWorkspace(forkProjectDir, forkName); err != nil {
		if !isAlreadyExistsErr(err) {
			return nil, fmt.Errorf("fork: failed to init workspace: %w", err)
		}
	}

	// If parent workspace is a git repo, create a branch for the fork
	if isGitRepo(parentWorkspace) {
		// Copy .git directory to fork workspace
		parentGitDir := filepath.Join(parentWorkspace, ".git")
		forkGitDir := filepath.Join(forkWorkspaceDir, ".git")
		if err := copyDir(parentGitDir, forkGitDir); err != nil {
			return nil, fmt.Errorf("fork: failed to copy git directory: %w", err)
		}

		// Create a new git branch for this fork
		branchName := "fork-" + forkID
		branchCmd := exec.Command("git", "checkout", "-b", branchName)
		branchCmd.Dir = forkWorkspaceDir
		if _, err := branchCmd.CombinedOutput(); err != nil {
			// Branch might already exist — try to checkout instead
			checkoutCmd := exec.Command("git", "checkout", branchName)
			checkoutCmd.Dir = forkWorkspaceDir
			if out2, err2 := checkoutCmd.CombinedOutput(); err2 != nil {
				return nil, fmt.Errorf("fork: git branch/checkout failed: %s: %w", string(out2), err2)
			}
		}
	}

	now := time.Now().UTC()

	forkSess := &Session{
		ID:           forkID,
		Name:         forkName,
		Status:       StatusRunning,
		Goal:         parent.Goal,
		Workspace:    forkWorkspaceDir,
		GraphJSON:    parent.GraphJSON,
		ChatLog:      parent.ChatLog,
		CreatedAt:    now,
		LastActiveAt: now,
	}

	// Persist fork session to SQLite (brief lock during save)
	m.mu.Lock()
	err = m.saveSession(forkSess)
	m.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("fork: failed to save session: %w", err)
	}

	return forkSess, nil
}

// copyDir recursively copies src directory to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Skip symlinks to prevent infinite recursion and unbounded data copies
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				// On Windows, some .git files may be read-only or locked.
				// Skip individual file copy errors during fork to avoid
				// failing the entire operation for a single inaccessible object.
				continue
			}
		}
	}
	return nil
}

// copyFile copies a single file from src to dst, preserving original permissions
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	return os.WriteFile(dst, data, info.Mode())
}

// isAlreadyExistsErr checks if the error indicates a directory already exists
func isAlreadyExistsErr(err error) bool {
	return errors.Is(err, ErrWorkspaceExists)
}
