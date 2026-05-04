package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SkillManifest represents a skill's metadata and configuration.
type SkillManifest struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Version     string   `json:"version" yaml:"version"`
	Description string   `json:"description" yaml:"description"`
	Author      string   `json:"author" yaml:"author"`
	Category    string   `json:"category" yaml:"category"`
	Rating      float64  `json:"rating" yaml:"rating"`
	RatingCount int      `json:"ratingCount" yaml:"rating_count"`
	Downloads   int      `json:"downloads" yaml:"downloads"`
	Icon        string   `json:"icon" yaml:"icon"`
	Tags        []string `json:"tags" yaml:"tags"`
	// Dependencies is a list of skill IDs that must be installed first.
	Dependencies []string `json:"dependencies" yaml:"dependencies"`
	// MainFile is the entrypoint for the skill (relative to skill directory).
	MainFile string `json:"mainFile" yaml:"main_file"`
}

// LoadManifest reads and parses a skill.json manifest file from the given directory.
func LoadManifest(dir string) (*SkillManifest, error) {
	path := filepath.Join(dir, "skill.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("skills: read manifest %q: %w", path, err)
	}

	var m SkillManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("skills: parse manifest %q: %w", path, err)
	}

	if m.ID == "" {
		return nil, fmt.Errorf("skills: manifest %q missing required field: id", path)
	}
	if m.Name == "" {
		return nil, fmt.Errorf("skills: manifest %q missing required field: name", path)
	}

	return &m, nil
}
