package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InitProjectWorkspace initializes the .nforge/ directory structure for a project
func InitProjectWorkspace(projectDir, projectName string) error {
	nforgeDir := filepath.Join(projectDir, ".nforge")

	// Check if .nforge already exists to avoid silent overwrite
	if info, err := os.Stat(nforgeDir); err == nil && info.IsDir() {
		return fmt.Errorf(".nforge/ directory already exists in %s", projectDir)
	}

	if err := os.MkdirAll(nforgeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .nforge directory: %w", err)
	}

	// Create config.yaml with actual timestamp
	config := fmt.Sprintf("project_name: %s\ncreated_at: \"%s\"\nversion: 0.1.0\n",
		projectName, time.Now().UTC().Format(time.RFC3339))
	configPath := filepath.Join(nforgeDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	// Create README.md
	readme := fmt.Sprintf("# %s\n\nNodeForge project workspace\n", projectName)
	readmePath := filepath.Join(nforgeDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Create .gitignore
	gitignore := ".nforge/workspace/data/\n.nforge/workspace/tmp/\n"
	gitignorePath := filepath.Join(nforgeDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}
