package context

import (
	"path/filepath"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore(t.TempDir())
	assert.NoError(t, err)
	assert.NotNil(t, store)
	defer store.Close()
}

func TestSaveAndGetGraph(t *testing.T) {
	store, err := NewStore(t.TempDir())
	assert.NoError(t, err)
	defer store.Close()

	// Test saving and retrieving graph
	// Note: requires engine.Graph type, which is in another package
	// This is a placeholder test
	t.Skip("requires cross-package graph type")
}

func TestDefaultStorePath(t *testing.T) {
	path := DefaultStorePath("/workspace")
	assert.Equal(t, filepath.Join("/workspace", ".nforge", "context.db"), path)
}
