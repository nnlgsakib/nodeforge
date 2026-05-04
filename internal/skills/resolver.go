package skills

import (
	"errors"
	"fmt"
)

// ErrSkillNotFound is returned when a skill is not found in the registry.
var ErrSkillNotFound = errors.New("skills: skill not found")

// ResolveDependencies returns the full ordered list of skills to install,
// including all transitive dependencies, using depth-first resolution.
// The registry lookup function returns a manifest for a given skill ID.
func ResolveDependencies(skillID string, registry func(id string) (*SkillManifest, error)) ([]string, error) {
	var resolved []string
	visited := make(map[string]bool)

	if err := resolveDFS(skillID, registry, visited, &resolved); err != nil {
		return nil, err
	}

	return resolved, nil
}

func resolveDFS(id string, registry func(id string) (*SkillManifest, error), visited map[string]bool, resolved *[]string) error {
	if visited[id] {
		return nil
	}
	visited[id] = true

	m, err := registry(id)
	if err != nil {
		return fmt.Errorf("skills: resolve dependency %q: %w", id, err)
	}

	for _, dep := range m.Dependencies {
		if err := resolveDFS(dep, registry, visited, resolved); err != nil {
			return err
		}
	}

	*resolved = append(*resolved, id)
	return nil
}
