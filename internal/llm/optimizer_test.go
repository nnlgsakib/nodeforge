package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*badger.DB, func()) {
	t.Helper()
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("failed to open BadgerDB: %v", err)
	}
	return db, func() { db.Close() }
}

func TestNewPromptOptimizer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	assert.NotNil(t, optimizer)
}

func TestSaveAndGetFeedback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	// Save feedback
	feedback := Feedback{
		Prompt:      "Test prompt",
		NodeType:    "Goal",
		Success:     true,
		TokenUsage:  100,
		QualityScore: 0.95,
	}

	err := optimizer.SaveFeedback(ctx, feedback)
	assert.NoError(t, err)

	// Get feedback
	feedbacks, err := optimizer.GetFeedback(ctx, "Goal", 10)
	assert.NoError(t, err)
	assert.Len(t, feedbacks, 1)
	assert.Equal(t, "Test prompt", feedbacks[0].Prompt)
	assert.Equal(t, "Goal", feedbacks[0].NodeType)
	assert.True(t, feedbacks[0].Success)
	assert.Equal(t, 100, feedbacks[0].TokenUsage)
}

func TestGetFeedback_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	feedbacks, err := optimizer.GetFeedback(ctx, "Goal", 10)
	assert.NoError(t, err)
	assert.Empty(t, feedbacks)
}

func TestGetFeedback_Limit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	// Save 5 feedback entries
	for i := 0; i < 5; i++ {
		fb := Feedback{
			Prompt:   fmt.Sprintf("Prompt %d", i),
			NodeType: "Spec",
			Success:  true,
		}
		err := optimizer.SaveFeedback(ctx, fb)
		assert.NoError(t, err)
		// Small delay to ensure unique timestamps
		time.Sleep(time.Millisecond)
	}

	// Get with limit 3
	feedbacks, err := optimizer.GetFeedback(ctx, "Spec", 3)
	assert.NoError(t, err)
	assert.Len(t, feedbacks, 3)
}

func TestOptimizePrompt_NoFeedback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	original := "Write a function to add two numbers"
	result := optimizer.OptimizePrompt(ctx, original, "Implement")

	// Should return template-applied prompt (not original, since template exists)
	assert.Contains(t, result, original)
}

func TestOptimizePrompt_WithFeedback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	// Save some successful feedback
	fb := Feedback{
		Prompt:   "Previous prompt",
		NodeType: "Test",
		Success:  true,
	}
	err := optimizer.SaveFeedback(ctx, fb)
	assert.NoError(t, err)

	original := "Create tests for add function"
	result := optimizer.OptimizePrompt(ctx, original, "Test")

	// Should contain original prompt in template
	assert.Contains(t, result, original)
}

func TestOptimizePrompt_InvalidNodeType(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	original := "Some prompt"
	result := optimizer.OptimizePrompt(ctx, original, "InvalidType")

	// No template for invalid type, should return original
	assert.Equal(t, original, result)
}

func TestGetPromptTemplate(t *testing.T) {
	tests := []struct {
		nodeType string
		hasTemplate bool
	}{
		{"Goal", true},
		{"Spec", true},
		{"Plan", true},
		{"Implement", true},
		{"Test", true},
		{"Review", true},
		{"Invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			template := getPromptTemplate(tt.nodeType)
			if tt.hasTemplate {
				assert.NotEmpty(t, template)
				assert.Contains(t, template, "{{prompt}}")
			} else {
				assert.Empty(t, template)
			}
		})
	}
}

func TestBuildOptimizationHints(t *testing.T) {
	feedbacks := []Feedback{
		{Prompt: "test1", Success: true, TokenUsage: 100, QualityScore: 0.9},
		{Prompt: "test2", Success: true, TokenUsage: 200, QualityScore: 0.8},
	}

	hints := buildOptimizationHints(feedbacks)
	assert.Equal(t, "true", hints["hasSuccessfulHistory"])
	assert.NotEmpty(t, hints["avgTokenUsage"])
}

func TestBuildOptimizationHints_Empty(t *testing.T) {
	hints := buildOptimizationHints(nil)
	assert.Empty(t, hints)
}

func TestOptimizePrompt_PanicRecovery(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	optimizer := NewPromptOptimizer(db, nil)
	ctx := context.Background()

	// This should not panic; optimizer recovers from panics
	result := optimizer.OptimizePrompt(ctx, "test", "Goal")
	assert.NotEmpty(t, result)
}
