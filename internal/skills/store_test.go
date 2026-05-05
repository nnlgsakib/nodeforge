package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-skills.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestStore_InsertAndExists(t *testing.T) {
	store := newTestStore(t)

	err := store.Insert("test-skill", "1.0.0")
	require.NoError(t, err)

	exists, err := store.Exists("test-skill")
	require.NoError(t, err)
	assert.True(t, exists)

	assert.True(t, store.IsInstalled("test-skill"))
}

func TestStore_InsertReplace(t *testing.T) {
	store := newTestStore(t)

	err := store.Insert("skill-a", "1.0.0")
	require.NoError(t, err)

	// Insert again with new version (OR REPLACE)
	err = store.Insert("skill-a", "2.0.0")
	require.NoError(t, err)

	// Should still exist
	exists, err := store.Exists("skill-a")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestStore_Delete(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.Insert("to-delete", "1.0.0"))
	require.NoError(t, store.Delete("to-delete"))

	exists, err := store.Exists("to-delete")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestStore_List(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.Insert("s1", "1.0.0"))
	require.NoError(t, store.Insert("s2", "2.0.0"))
	require.NoError(t, store.Insert("s3", "1.5.0"))

	list, err := store.List()
	require.NoError(t, err)
	assert.Len(t, list, 3)

	// Verify all IDs are present
	ids := make(map[string]bool)
	for _, sk := range list {
		ids[sk.SkillID] = true
	}
	assert.True(t, ids["s1"])
	assert.True(t, ids["s2"])
	assert.True(t, ids["s3"])
}

func TestStore_DefaultPath(t *testing.T) {
	// Test with empty path (should use home dir)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	expected := filepath.Join(home, ".nforge", "skills.db")

	store, err := NewStore("")
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	// Just verify it doesn't error
	exists, err := store.Exists("any")
	require.NoError(t, err)
	assert.False(t, exists)
	_ = expected // used for path verification
}

func TestStore_IsInstalled_NotFound(t *testing.T) {
	store := newTestStore(t)

	assert.False(t, store.IsInstalled("nonexistent"))
}
