package nforge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunNewProject(t *testing.T) {
	tmpDir := t.TempDir()
	projectName := "test-project"
	workspaceDir := tmpDir

	err := runNewProject(projectName, workspaceDir)
	if err != nil {
		t.Fatalf("runNewProject failed: %v", err)
	}

	// Check that project directory was created
	projectDir := filepath.Join(workspaceDir, projectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatalf("project directory not created at %s", projectDir)
	}

	// Check that .nforge directory was created
	nforgeDir := filepath.Join(projectDir, ".nforge")
	if _, err := os.Stat(nforgeDir); os.IsNotExist(err) {
		t.Fatalf(".nforge directory not created at %s", nforgeDir)
	}

	// Check for config.yaml
	configFile := filepath.Join(nforgeDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatalf("config.yaml not created at %s", configFile)
	}

	// Check for README.md
	readmeFile := filepath.Join(nforgeDir, "README.md")
	if _, err := os.Stat(readmeFile); os.IsNotExist(err) {
		t.Fatalf("README.md not created at %s", readmeFile)
	}

	// Check for .gitignore
	gitignoreFile := filepath.Join(nforgeDir, ".gitignore")
	if _, err := os.Stat(gitignoreFile); os.IsNotExist(err) {
		t.Fatalf(".gitignore not created at %s", gitignoreFile)
	}
}
