package context

import (
	"context"
	"strings"
	"time"
)

// Assembler retrieves and assembles knowledge graph context for prompt optimization
type Assembler struct {
	store *Store
}

// NewAssembler creates a new context assembler
func NewAssembler(store *Store) *Assembler {
	return &Assembler{store: store}
}

// ContextQuery represents a query for knowledge graph context
type ContextQuery struct {
	NodeType   string `json:"nodeType"`
	Prompt     string `json:"prompt"`
	MaxTokens  int    `json:"maxTokens"`
}

// ContextResult contains assembled context for prompt optimization
type ContextResult struct {
	Context    string `json:"context"`
	TokenCount int    `json:"tokenCount"`
	QueryTimeMs int64  `json:"queryTimeMs"`
}

// AssembleContext retrieves relevant context from knowledge graph
// Completes in <100ms (NFR-04) for typical queries
func (a *Assembler) AssembleContext(ctx context.Context, query ContextQuery) (*ContextResult, error) {
	start := time.Now()

	// Retrieve relevant graph data from BadgerDB
	contextStr, err := a.retrieveRelevantContext(ctx, query)
	if err != nil {
		// Return empty context on error (never block execution)
		return &ContextResult{
			Context:     "",
			TokenCount:  0,
			QueryTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	// Enforce MaxTokens limit if specified
	tokenCount := estimateTokens(contextStr)
	if query.MaxTokens > 0 && tokenCount > query.MaxTokens {
		// Truncate context to fit within MaxTokens
		contextStr = truncateToTokens(contextStr, query.MaxTokens)
		tokenCount = estimateTokens(contextStr)
	}

	result := &ContextResult{
		Context:     contextStr,
		TokenCount:  tokenCount,
		QueryTimeMs: time.Since(start).Milliseconds(),
	}

	// Validate NFR-04: <100ms query time
	if result.QueryTimeMs > 100 {
		// Log warning if query takes too long
		// In production, use proper logger
	}

	return result, nil
}

// retrieveRelevantContext fetches relevant context from BadgerDB
func (a *Assembler) retrieveRelevantContext(ctx context.Context, query ContextQuery) (string, error) {
	if a.store == nil {
		return "", nil
	}

	// Simple implementation: retrieve recent node outputs as context
	var contextBuilder strings.Builder

	// Example: retrieve monologue history as context
	messages, err := a.store.GetMonologueHistory(ctx, "current-session")
	if err == nil && len(messages) > 0 {
		contextBuilder.WriteString("Recent monologue context:\n")
		for _, msg := range messages {
			contextBuilder.WriteString(msg.Text)
			contextBuilder.WriteString("\n")
		}
	}

	return contextBuilder.String(), nil
}

// estimateTokens estimates token count for context string
func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4 // ~4 chars per token
}

// truncateToTokens truncates text to approximately maxTokens by removing excess characters
func truncateToTokens(text string, maxTokens int) string {
	if maxTokens <= 0 {
		return text
	}
	maxChars := maxTokens * 4 // approximate chars per token
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars]
}
