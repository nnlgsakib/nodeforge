package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// validSessionID matches the expected session ID format: sess-<uuid>
var validSessionID = regexp.MustCompile(`^sess-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// validateSessionID returns an error if the session ID is malformed
func validateSessionID(id string) error {
	if !validSessionID.MatchString(id) {
		return fmt.Errorf("invalid session ID format: %q", id)
	}
	return nil
}

// InitProjectWorkspace initializes the .nforge/ directory structure for a project
func InitProjectWorkspace(projectDir, projectName string) error {
	nforgeDir := filepath.Join(projectDir, ".nforge")

	// Check if .nforge already exists to avoid silent overwrite
	if info, err := os.Stat(nforgeDir); err == nil && info.IsDir() {
		return fmt.Errorf(".nforge/ directory already exists in %s: %w", projectDir, ErrWorkspaceExists)
	}

	if err := os.MkdirAll(nforgeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .nforge directory: %w", err)
	}

	// Sanitize project name for config/README (remove control characters)
	safeName := strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return -1
		}
		return r
	}, projectName)

	// Create config.yaml with actual timestamp
	config := fmt.Sprintf("project_name: %s\ncreated_at: \"%s\"\nversion: 0.1.0\n",
		safeName, time.Now().UTC().Format(time.RFC3339))
	configPath := filepath.Join(nforgeDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	// Create README.md
	readme := fmt.Sprintf("# %s\n\nNodeForge project workspace\n", safeName)
	readmePath := filepath.Join(nforgeDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Create .gitignore
	gitignore := "data/\ntmp/\n"
	gitignorePath := filepath.Join(nforgeDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

// WorkspacePath returns the workspace directory path for a session.
// Validates sessionID to prevent path traversal.
func (m *Manager) WorkspacePath(sessionID string) (string, error) {
	if err := validateSessionID(sessionID); err != nil {
		return "", err
	}
	return filepath.Join(m.workspaceRoot, ".nforge", "sessions", sessionID, "workspace"), nil
}

// EnsureWorkspaceDir creates the workspace directory for a session if it doesn't exist
func (m *Manager) EnsureWorkspaceDir(sessionID string) error {
	workspaceDir, err := m.WorkspacePath(sessionID)
	if err != nil {
		return err
	}
	return os.MkdirAll(workspaceDir, 0755)
}

// WriteWorkspaceFile writes a file to the session's workspace with directory traversal protection
func (m *Manager) WriteWorkspaceFile(sessionID, relativePath string, content []byte) error {
	// Check workspace size quota before writing (NFR-17)
	// Lock during check to narrow TOCTOU window
	m.mu.Lock()
	err := m.checkWorkspaceSize(sessionID)
	m.mu.Unlock()
	if err != nil {
		return err
	}

	workspaceDir, err := m.WorkspacePath(sessionID)
	if err != nil {
		return err
	}

	// Validate relativePath to prevent directory traversal
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: directory traversal not allowed")
	}
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("invalid file path: absolute paths not allowed")
	}

	resolvedPath := filepath.Join(workspaceDir, cleanPath)

	// Ensure workspace dir exists before containment check
	absWorkspace := workspaceDir
	if !filepath.IsAbs(absWorkspace) {
		absWorkspace, _ = filepath.Abs(absWorkspace)
	}

	resolvedDir := filepath.Dir(resolvedPath)
	if !filepath.IsAbs(resolvedDir) {
		resolvedDir, _ = filepath.Abs(resolvedDir)
	}

	// Use EvalSymlinks on existing directories; fall back to absolute path check
	if resolvedSymlink, err := filepath.EvalSymlinks(resolvedDir); err == nil {
		resolvedDir = resolvedSymlink
	}
	if workspaceSymlink, err := filepath.EvalSymlinks(absWorkspace); err == nil {
		absWorkspace = workspaceSymlink
	}

	if !strings.HasPrefix(resolvedDir, absWorkspace) {
		return fmt.Errorf("invalid file path: resolved path outside workspace")
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	return os.WriteFile(resolvedPath, content, 0644)
}

// ReadWorkspaceFile reads a file from the session's workspace with directory traversal protection
func (m *Manager) ReadWorkspaceFile(sessionID, relativePath string) ([]byte, error) {
	workspaceDir, err := m.WorkspacePath(sessionID)
	if err != nil {
		return nil, err
	}

	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: directory traversal not allowed")
	}
	if filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("invalid file path: absolute paths not allowed")
	}

	resolvedPath := filepath.Join(workspaceDir, cleanPath)

	// Ensure containment check uses resolved symlinks where possible
	absWorkspace := workspaceDir
	if !filepath.IsAbs(absWorkspace) {
		absWorkspace, _ = filepath.Abs(absWorkspace)
	}
	if resolved, err := filepath.EvalSymlinks(absWorkspace); err == nil {
		absWorkspace = resolved
	}

	absResolved, err := filepath.EvalSymlinks(resolvedPath)
	if err != nil {
		// File doesn't exist yet — check directory containment
		dirAbs, _ := filepath.EvalSymlinks(filepath.Dir(resolvedPath))
		if dirAbs == "" {
			dirAbs = filepath.Dir(resolvedPath)
		}
		if !filepath.IsAbs(dirAbs) {
			dirAbs, _ = filepath.Abs(dirAbs)
		}
		if !strings.HasPrefix(dirAbs, absWorkspace) {
			return nil, fmt.Errorf("invalid file path: resolved path outside workspace")
		}
		return nil, fmt.Errorf("failed to read workspace file: %w", err)
	}

	if !strings.HasPrefix(absResolved, absWorkspace) {
		return nil, fmt.Errorf("invalid file path: resolved path outside workspace")
	}

	return os.ReadFile(resolvedPath)
}

// InitGitRepo initializes a Git repository in the session's workspace directory.
// It runs `git init` and configures basic settings. Safe to call multiple times.
func InitGitRepo(workspaceDir string) error {
	if _, err := os.Stat(workspaceDir); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory does not exist: %s", workspaceDir)
	}

	// Skip if already a git repo
	if isGitRepo(workspaceDir) {
		return nil
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = workspaceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %s: %w", string(output), err)
	}

	// Configure default user for commits (can be overridden by user)
	cmd = exec.Command("git", "config", "user.email", "nforge@local")
	cmd.Dir = workspaceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config email failed: %s: %w", string(output), err)
	}
	cmd = exec.Command("git", "config", "user.name", "NodeForge")
	cmd.Dir = workspaceDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config name failed: %s: %w", string(output), err)
	}

	return nil
}
