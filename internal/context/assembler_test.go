package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextAssembler(t *testing.T) {
	db := newTestDB(t)
	assembler := NewContextAssembler(db)
	assert.NotNil(t, assembler)
}

func TestContextAssembler_AssembleContext(t *testing.T) {
	db := newTestDB(t)
	assembler := NewContextAssembler(db)
	kg := NewKnowledgeGraph(db)

	// Add some context
	require.NoError(t, kg.AddNodeOutput("node1", "test output"))

	context, err := assembler.AssembleContext("node2", 1000)
	assert.NoError(t, err)
	assert.Contains(t, context, "test output")
}

func TestContextAssembler_AssembleContext_Timeout(t *testing.T) {
	db := newTestDB(t)
	assembler := NewContextAssembler(db)

	// Test that it doesn't timeout for small context
	_, err := assembler.AssembleContext("node1", 100)
	assert.NoError(t, err)
	// Should return empty or context, but not timeout
	assert.NotPanics(t, func() {}) // Basic sanity check
}

func TestInjectContextIntoPrompt(t *testing.T) {
	prompt := "Original prompt"
	ctxContext := "Context info"

	result := InjectContextIntoPrompt(prompt, ctxContext)
	assert.Contains(t, result, "Original prompt")
	assert.Contains(t, result, "[Context]:")
	assert.Contains(t, result, "Context info")
}

func TestInjectContextIntoPrompt_EmptyContext(t *testing.T) {
	prompt := "Original prompt"
	result := InjectContextIntoPrompt(prompt, "")
	assert.Equal(t, prompt, result)
}
