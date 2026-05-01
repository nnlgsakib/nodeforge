package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitProjectWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-workspace"

	err := InitProjectWorkspace(tmpDir, projectName)
	if err != nil {
		t.Fatalf("InitProjectWorkspace failed: %v", err)
	}

	// Check .nforge directory exists
	nforgeDir := filepath.Join(tmpDir, ".nforge")
	if _, err := os.Stat(nforgeDir); os.IsNotExist(err) {
		t.Fatalf(".nforge directory not created at %s", nforgeDir)
	}

	// Check config.yaml has valid created_at
	configPath := filepath.Join(nforgeDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config.yaml not created at %s", configPath)
	}

	// Check README.md
	readmePath := filepath.Join(nforgeDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Fatalf("README.md not created at %s", readmePath)
	}

	// Check .gitignore
	gitignorePath := filepath.Join(nforgeDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Fatalf(".gitignore not created at %s", gitignorePath)
	}
}

func TestInitProjectWorkspaceAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// First call should succeed
	err := InitProjectWorkspace(tmpDir, "first")
	if err != nil {
		t.Fatalf("first InitProjectWorkspace failed: %v", err)
	}

	// Second call should fail (directory already has .nforge/)
	err = InitProjectWorkspace(tmpDir, "second")
	if err == nil {
		t.Error("expected error for existing .nforge/ directory, got nil")
	}
}
