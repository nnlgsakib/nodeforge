package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TokenBudget holds token budget configuration
type TokenBudget struct {
	TotalBudgetPerSession    int `json:"totalBudgetPerSession" yaml:"totalBudgetPerSession"`
	MaxTokensPerRequest      int `json:"maxTokensPerRequest" yaml:"maxTokensPerRequest"`
}

// DefaultTokenBudget returns the default token budget configuration
func DefaultTokenBudget() *TokenBudget {
	return &TokenBudget{
		TotalBudgetPerSession: 100000,
		MaxTokensPerRequest:   4096,
	}
}

// ErrTokenBudgetExceeded is returned when a request would exceed the token budget
var ErrTokenBudgetExceeded = fmt.Errorf("token budget exceeded")

// BudgetEnforcer enforces token budgets and tracks usage
type BudgetEnforcer struct {
	mu                    sync.RWMutex
	budget                *TokenBudget
	usedSessionTokens     int
	sessionBudgetRemaining int
}

// NewBudgetEnforcer creates a new BudgetEnforcer with the given budget
func NewBudgetEnforcer(budget *TokenBudget) *BudgetEnforcer {
	if budget == nil {
		budget = DefaultTokenBudget()
	}
	return &BudgetEnforcer{
		budget:                budget,
		sessionBudgetRemaining: budget.TotalBudgetPerSession,
		usedSessionTokens:     0,
	}
}

// EstimateTokens estimates the number of tokens in a text string using a fast heuristic
// (~4 characters per token for English text). Designed to complete in <10ms for typical prompts.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	// Fast heuristic: ~4 characters per token for English text
	return (len(text) + 3) / 4
}

// CheckBudget checks if a request with the given estimated tokens can proceed
// Returns ErrTokenBudgetExceeded if the request would exceed the budget
// Completes in <10ms as it only checks in-memory state
func (be *BudgetEnforcer) CheckBudget(ctx context.Context, estimatedTokens int) error {
	if be.budget == nil {
		return fmt.Errorf("budget enforcer not properly initialized")
	}
	if estimatedTokens < 0 {
		return fmt.Errorf("estimated tokens cannot be negative: %d", estimatedTokens)
	}

	be.mu.RLock()
	defer be.mu.RUnlock()

	if estimatedTokens > be.budget.MaxTokensPerRequest {
		return fmt.Errorf("%w: request tokens %d exceed max per request %d",
			ErrTokenBudgetExceeded, estimatedTokens, be.budget.MaxTokensPerRequest)
	}

	if estimatedTokens > be.sessionBudgetRemaining {
		return fmt.Errorf("%w: request tokens %d exceed remaining session budget %d",
			ErrTokenBudgetExceeded, estimatedTokens, be.sessionBudgetRemaining)
	}

	return nil
}

// CheckAndTrack atomically checks the budget and tracks usage under a single write lock.
// This prevents TOCTOU race conditions between check and track.
func (be *BudgetEnforcer) CheckAndTrack(ctx context.Context, estimatedTokens int) error {
	if be.budget == nil {
		return fmt.Errorf("budget enforcer not properly initialized")
	}
	if estimatedTokens < 0 {
		return fmt.Errorf("estimated tokens cannot be negative: %d", estimatedTokens)
	}

	be.mu.Lock()
	defer be.mu.Unlock()

	if estimatedTokens > be.budget.MaxTokensPerRequest {
		return fmt.Errorf("%w: request tokens %d exceed max per request %d",
			ErrTokenBudgetExceeded, estimatedTokens, be.budget.MaxTokensPerRequest)
	}

	if estimatedTokens > be.sessionBudgetRemaining {
		return fmt.Errorf("%w: request tokens %d exceed remaining session budget %d",
			ErrTokenBudgetExceeded, estimatedTokens, be.sessionBudgetRemaining)
	}

	// Track usage atomically with the check
	be.usedSessionTokens += estimatedTokens
	be.sessionBudgetRemaining -= estimatedTokens
	if be.sessionBudgetRemaining < 0 {
		be.sessionBudgetRemaining = 0
	}

	return nil
}

// TrackUsage records actual token consumption after an LLM call
func (be *BudgetEnforcer) TrackUsage(actualTokens int) {
	if actualTokens < 0 {
		return // ignore negative values
	}

	be.mu.Lock()
	defer be.mu.Unlock()

	be.usedSessionTokens += actualTokens
	be.sessionBudgetRemaining -= actualTokens
	if be.sessionBudgetRemaining < 0 {
		be.sessionBudgetRemaining = 0
	}
}

// BudgetStatus returns the current budget status
func (be *BudgetEnforcer) BudgetStatus() (used int, remaining int, total int) {
	be.mu.RLock()
	defer be.mu.RUnlock()

	return be.usedSessionTokens, be.sessionBudgetRemaining, be.budget.TotalBudgetPerSession
}

// EstimateAndCheck combines token estimation and budget check in one call
// Returns the estimated token count and any budget error
func (be *BudgetEnforcer) EstimateAndCheck(ctx context.Context, text string) (int, error) {
	start := time.Now()
	estimated := EstimateTokens(text)
	dur := time.Since(start)
	if dur > 10*time.Millisecond {
		// Estimation took longer than expected - log warning
		// In production, use proper logger: log.Printf("WARN: token estimation took %v (>10ms)", dur)
	}
	return estimated, be.CheckBudget(ctx, estimated)
}
