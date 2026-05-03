package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTokenBudget(t *testing.T) {
	budget := DefaultTokenBudget()
	assert.Equal(t, 100000, budget.TotalBudgetPerSession)
	assert.Equal(t, 4096, budget.MaxTokensPerRequest)
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty string", "", 0},
		{"4 chars (1 token)", "test", 1}, // (4+3)/4 = 1
		{"5 chars (2 tokens)", "hello", 2}, // (5+3)/4 = 2
		{"8 chars (2 tokens)", "hellooo", 2}, // (8+3)/4 = 2
		{"typical prompt", "Write a function that adds two numbers", 10}, // (38+3)/4 = 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateTokens(tt.text)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateTokens10k(t *testing.T) {
	// Separate test for 10k chars to avoid creating large string in table
	longText := make([]byte, 10000)
	for i := range longText {
		longText[i] = 'a'
	}
	expected := (10000 + 3) / 4 // 2500
	result := EstimateTokens(string(longText))
	assert.Equal(t, expected, result)
}

func TestEstimateTokensPerformance(t *testing.T) {
	// Test that estimation completes in <10ms for 10k char prompt
	longText := make([]byte, 10000)
	for i := range longText {
		longText[i] = 'a'
	}

	start := time.Now()
	_ = EstimateTokens(string(longText))
	dur := time.Since(start)

	assert.Less(t, dur, 10*time.Millisecond, "EstimateTokens should complete in <10ms for 10k chars")
}

func TestNewBudgetEnforcer(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 50000,
		MaxTokensPerRequest:   2048,
	}

	enforcer := NewBudgetEnforcer(budget)
	assert.NotNil(t, enforcer)

	used, remaining, total := enforcer.BudgetStatus()
	assert.Equal(t, 0, used)
	assert.Equal(t, 50000, remaining)
	assert.Equal(t, 50000, total)
}

func TestNewBudgetEnforcerNilBudget(t *testing.T) {
	enforcer := NewBudgetEnforcer(nil)
	assert.NotNil(t, enforcer)

	_, remaining, total := enforcer.BudgetStatus()
	assert.Equal(t, 100000, total)
	assert.Equal(t, 100000, remaining)
}

func TestCheckBudget_Success(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 10000,
		MaxTokensPerRequest:   1000,
	}
	enforcer := NewBudgetEnforcer(budget)
	ctx := context.Background()

	err := enforcer.CheckBudget(ctx, 500)
	assert.NoError(t, err)
}

func TestCheckBudget_ExceedsPerRequest(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 10000,
		MaxTokensPerRequest:   1000,
	}
	enforcer := NewBudgetEnforcer(budget)
	ctx := context.Background()

	err := enforcer.CheckBudget(ctx, 1001)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenBudgetExceeded)
}

func TestCheckBudget_ExceedsSessionBudget(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 1000,
		MaxTokensPerRequest:   5000,
	}
	enforcer := NewBudgetEnforcer(budget)
	ctx := context.Background()

	// First request should succeed
	err := enforcer.CheckBudget(ctx, 600)
	assert.NoError(t, err)

	// Track the usage
	enforcer.TrackUsage(600)

	// Next request should fail (remaining is 400, request is 500)
	err = enforcer.CheckBudget(ctx, 500)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenBudgetExceeded)
}

func TestTrackUsage(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 10000,
		MaxTokensPerRequest:   5000,
	}
	enforcer := NewBudgetEnforcer(budget)

	enforcer.TrackUsage(1500)
	used, remaining, total := enforcer.BudgetStatus()
	assert.Equal(t, 1500, used)
	assert.Equal(t, 8500, remaining)
	assert.Equal(t, 10000, total)
}

func TestTrackUsage_Negative(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 10000,
		MaxTokensPerRequest:   5000,
	}
	enforcer := NewBudgetEnforcer(budget)

	// Negative values should be ignored
	enforcer.TrackUsage(-500)
	used, remaining, _ := enforcer.BudgetStatus()
	assert.Equal(t, 0, used)
	assert.Equal(t, 10000, remaining)
}

func TestBudgetStatus(t *testing.T) {
	enforcer := NewBudgetEnforcer(DefaultTokenBudget())

	enforcer.TrackUsage(2500)
	used, remaining, total := enforcer.BudgetStatus()

	assert.Equal(t, 2500, used)
	assert.Equal(t, 97500, remaining)
	assert.Equal(t, 100000, total)
}

func TestEstimateAndCheck(t *testing.T) {
	budget := &TokenBudget{
		TotalBudgetPerSession: 10000,
		MaxTokensPerRequest:   1000,
	}
	enforcer := NewBudgetEnforcer(budget)
	ctx := context.Background()

	// Test with short prompt
	tokens, err := enforcer.EstimateAndCheck(ctx, "Hello world")
	assert.NoError(t, err)
	assert.Greater(t, tokens, 0)

	// Test with prompt that exceeds per-request budget
	longPrompt := make([]byte, 5000)
	for i := range longPrompt {
		longPrompt[i] = 'a'
	}
	tokens, err = enforcer.EstimateAndCheck(ctx, string(longPrompt))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenBudgetExceeded)
	assert.Greater(t, tokens, 1000)
}
