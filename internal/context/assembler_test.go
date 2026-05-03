package context

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewAssembler(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	assembler := NewAssembler(store)
	assert.NotNil(t, assembler)
}

func TestAssembleContext_NoStore(t *testing.T) {
	assembler := NewAssembler(nil)
	ctx := context.Background()

	result, err := assembler.AssembleContext(ctx, ContextQuery{
		NodeType: "Goal",
		Prompt:   "test prompt",
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAssembleContext_Timing(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	assembler := NewAssembler(store)
	ctx := context.Background()

	start := time.Now()
	_, err := assembler.AssembleContext(ctx, ContextQuery{
		NodeType: "Goal",
		Prompt:   "test prompt",
	})
	dur := time.Since(start)

	assert.NoError(t, err)
	// NFR-04: context assembly <100ms
	assert.Less(t, dur, 100*time.Millisecond, "context assembly should complete in <100ms")
}

func TestAssembleContext_TimingBudget(t *testing.T) {
	// Verify that context assembly doesn't break budget pre-flight <10ms
	store, cleanup := newTestStore(t)
	defer cleanup()

	assembler := NewAssembler(store)
	ctx := context.Background()

	// Budget pre-flight check includes context assembly
	// Simulate token estimation + context assembly
	start := time.Now()

	// 1. Estimate tokens
	tokenCount := (len("test prompt") + 3) / 4

	// 2. Assemble context
	result, err := assembler.AssembleContext(ctx, ContextQuery{
		NodeType:  "Goal",
		Prompt:    "test prompt",
		MaxTokens: 1000,
	})
	assert.NoError(t, err)
	if result != nil {
		tokenCount += result.TokenCount
	}

	dur := time.Since(start)
	// Budget pre-flight must be <10ms (NFR-05)
	assert.Less(t, dur, 10*time.Millisecond, "budget pre-flight + context assembly should be <10ms")
	assert.Greater(t, tokenCount, 0)
}

// newTestStore creates a test BadgerDB store
func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open BadgerDB: %v", err)
	}
	store := &Store{db: db}
	return store, func() { store.Close() }
}
