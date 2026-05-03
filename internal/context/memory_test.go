package context

import (
	"context"
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

func TestSaveAndGetMonologueHistory(t *testing.T) {
	store, err := NewStore(t.TempDir())
	assert.NoError(t, err)
	defer store.Close()

	sessionID := "session-123"

	// Test retrieving when no history exists (should return empty slice)
	msgs, err := store.GetMonologueHistory(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, msgs)
	assert.Equal(t, 0, len(msgs))

	// Test saving monologue history
	history := []MonologueMessage{
		{ID: "1", Text: "Thinking...", Timestamp: 1714500000000},
		{ID: "2", Text: "Analyzing...", Timestamp: 1714500001000},
	}
	err = store.SaveMonologueHistory(context.Background(), sessionID, history)
	assert.NoError(t, err)

	// Test retrieving saved history
	msgs, err = store.GetMonologueHistory(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, "Thinking...", msgs[0].Text)
	assert.Equal(t, "Analyzing...", msgs[1].Text)
	assert.Equal(t, "1", msgs[0].ID)
	assert.Equal(t, int64(1714500000000), msgs[0].Timestamp)
}

func TestSaveMonologueHistoryEmpty(t *testing.T) {
	store, err := NewStore(t.TempDir())
	assert.NoError(t, err)
	defer store.Close()

	err = store.SaveMonologueHistory(context.Background(), "session-456", []MonologueMessage{})
	assert.NoError(t, err)

	msgs, err := store.GetMonologueHistory(context.Background(), "session-456")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(msgs))
}
