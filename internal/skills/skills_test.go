package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadManifest(t *testing.T) {
	tests := []struct {
		name        string
		jsonContent string
		expectErr   bool
		expectID    string
		expectName  string
	}{
		{
			name: "valid manifest",
			jsonContent: `{
				"id": "test-skill",
				"name": "Test Skill",
				"version": "1.0.0",
				"description": "A test skill",
				"author": "Test Author",
				"category": "Testing",
				"rating": 4.5,
				"ratingCount": 100,
				"downloads": 500
			}`,
			expectErr:  false,
			expectID:   "test-skill",
			expectName: "Test Skill",
		},
		{
			name: "missing id",
			jsonContent: `{
				"name": "Test Skill",
				"version": "1.0.0"
			}`,
			expectErr: true,
		},
		{
			name: "missing name",
			jsonContent: `{
				"id": "test-skill",
				"version": "1.0.0"
			}`,
			expectErr: true,
		},
		{
			name:        "invalid json",
			jsonContent: `{not valid json`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory with skill.json
			dir := t.TempDir()
			err := os.WriteFile(filepath.Join(dir, "skill.json"), []byte(tt.jsonContent), 0644)
			require.NoError(t, err)

			m, err := LoadManifest(dir)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectID, m.ID)
			assert.Equal(t, tt.expectName, m.Name)
		})
	}
}

func TestResolveDependencies(t *testing.T) {
	// Simple registry for testing
	registry := map[string]*SkillManifest{
		"skill-a": {ID: "skill-a", Dependencies: []string{"skill-b", "skill-c"}},
		"skill-b": {ID: "skill-b", Dependencies: []string{"skill-d"}},
		"skill-c": {ID: "skill-c", Dependencies: nil},
		"skill-d": {ID: "skill-d", Dependencies: nil},
		"skill-e": {ID: "skill-e", Dependencies: []string{"skill-e"}}, // circular
	}

	regFn := func(id string) (*SkillManifest, error) {
		s, ok := registry[id]
		if !ok {
			return nil, ErrSkillNotFound
		}
		return s, nil
	}

	t.Run("resolves transitive dependencies", func(t *testing.T) {
		result, err := ResolveDependencies("skill-a", regFn)
		require.NoError(t, err)
		// Expected order: skill-d, skill-b, skill-c, skill-a (DFS)
		assert.Contains(t, result, "skill-a")
		assert.Contains(t, result, "skill-b")
		assert.Contains(t, result, "skill-c")
		assert.Contains(t, result, "skill-d")
		// Verify dependency order: dependencies come before dependents
		assert.Equal(t, []string{"skill-d", "skill-b", "skill-c", "skill-a"}, result)
	})

	t.Run("skill with no dependencies", func(t *testing.T) {
		result, err := ResolveDependencies("skill-c", regFn)
		require.NoError(t, err)
		assert.Equal(t, []string{"skill-c"}, result)
	})

	t.Run("circular dependency handled", func(t *testing.T) {
		result, err := ResolveDependencies("skill-e", regFn)
		require.NoError(t, err)
		assert.Equal(t, []string{"skill-e"}, result)
	})

	t.Run("skill not found", func(t *testing.T) {
		_, err := ResolveDependencies("nonexistent", regFn)
		assert.Error(t, err)
	})
}

func TestErrSkillNotFound(t *testing.T) {
	assert.EqualError(t, ErrSkillNotFound, "skills: skill not found")
}
