package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"

	contextpkg "github.com/nnlgsakib/nodeforge/internal/context"
)

// Feedback represents execution feedback for a prompt
type Feedback struct {
	Prompt      string  `json:"prompt"`
	NodeType    string  `json:"nodeType"`
	Success     bool    `json:"success"`
	TokenUsage  int     `json:"tokenUsage"`
	QualityScore float64 `json:"qualityScore"`
	Timestamp   int64   `json:"timestamp"`
}

// PromptOptimizer optimizes prompts based on historical feedback
type PromptOptimizer struct {
	db         *badger.DB
	assembler  *contextpkg.ContextAssembler
}

// NewPromptOptimizer creates a new PromptOptimizer with BadgerDB storage and optional assembler
func NewPromptOptimizer(db *badger.DB, assembler *contextpkg.ContextAssembler) *PromptOptimizer {
	return &PromptOptimizer{db: db, assembler: assembler}
}

// SaveFeedback stores execution feedback in BadgerDB
func (po *PromptOptimizer) SaveFeedback(ctx context.Context, feedback Feedback) error {
	if po.db == nil {
		return fmt.Errorf("cannot save feedback: database is nil")
	}
	return po.db.Update(func(txn *badger.Txn) error {
		feedback.Timestamp = time.Now().UnixNano()
		data, err := json.Marshal(feedback)
		if err != nil {
			return fmt.Errorf("failed to marshal feedback: %w", err)
		}
		key := []byte(fmt.Sprintf("feedback:%s:%d", feedback.NodeType, feedback.Timestamp))
		return txn.Set(key, data)
	})
}

// GetFeedback retrieves historical feedback for a node type
func (po *PromptOptimizer) GetFeedback(ctx context.Context, nodeType string, limit int) ([]Feedback, error) {
	if po.db == nil {
		return nil, fmt.Errorf("cannot get feedback: database is nil")
	}
	if nodeType == "" {
		return nil, fmt.Errorf("nodeType must not be empty")
	}

	var feedbacks []Feedback

	err := po.db.View(func(txn *badger.Txn) error {
		// Prefix scan for feedback of this node type
		prefix := []byte(fmt.Sprintf("feedback:%s:", nodeType))
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var fb Feedback
				if err := json.Unmarshal(val, &fb); err != nil {
					return err
				}
				feedbacks = append(feedbacks, fb)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(feedbacks, func(i, j int) bool {
		return feedbacks[i].Timestamp > feedbacks[j].Timestamp
	})

	// Apply limit
	if limit > 0 && len(feedbacks) > limit {
		feedbacks = feedbacks[:limit]
	}

	return feedbacks, nil
}

// OptimizePrompt enhances a prompt based on historical feedback for the node type
// Falls back to original prompt if optimization fails (never blocks execution)
func (po *PromptOptimizer) OptimizePrompt(ctx context.Context, originalPrompt string, nodeType string) (result string) {
	result = originalPrompt // default to original

	defer func() {
		// Recover from panics to never block execution
		if r := recover(); r != nil {
			// Return original prompt on panic
			result = originalPrompt
		}
	}()

	// If no DB available, return original prompt with template
	if po.db == nil && po.assembler == nil {
		return applyTemplate(originalPrompt, nodeType)
	}

	// Assemble context from knowledge graph (if assembler available)
	contextStr := ""
	if po.assembler != nil {
		ctxContext, err := po.assembler.AssembleContext(nodeType, 1000)
		if err == nil {
			contextStr = ctxContext
		}
	}

	// Get historical feedback for this node type
	feedbacks, err := po.GetFeedback(ctx, nodeType, 10)
	if err != nil {
		// Fall back to original prompt on error (not templated)
		return originalPrompt
	}

	// Build optimized prompt with context
	optimizedPrompt := originalPrompt
	if contextStr != "" {
		optimizedPrompt = fmt.Sprintf("Context:\n%s\n\nPrompt:\n%s", contextStr, originalPrompt)
	}

	// Apply template if no feedback yet
	if len(feedbacks) ==0 {
		return applyTemplate(optimizedPrompt, nodeType)
	}

	// Build optimization hints from successful feedback
	hints := buildOptimizationHints(feedbacks)

	// Apply hints to optimized prompt
	finalPrompt := applyOptimizationHints(optimizedPrompt, hints, nodeType)
	return finalPrompt
}

// applyTemplate wraps the original prompt with node-type-specific template
func applyTemplate(prompt string, nodeType string) string {
	template := getPromptTemplate(nodeType)
	if template == "" {
		return prompt
	}
	return strings.ReplaceAll(template, "{{prompt}}", prompt)
}

// buildOptimizationHints extracts optimization hints from historical feedback
func buildOptimizationHints(feedbacks []Feedback) map[string]string {
	hints := make(map[string]string)

	// Find most successful feedback patterns
	successCount := 0
	var avgTokenUsage float64
	for _, fb := range feedbacks {
		if fb.Success {
			successCount++
		}
		avgTokenUsage += float64(fb.TokenUsage)
	}

	if len(feedbacks) >0 {
		avgTokenUsage /= float64(len(feedbacks))
	}

	if successCount >0 {
		hints["hasSuccessfulHistory"] = "true"
	}
	if avgTokenUsage >0 {
		hints["avgTokenUsage"] = fmt.Sprintf("%.0f", avgTokenUsage)
	}

	return hints
}

// applyOptimizationHints applies hints to the original prompt
func applyOptimizationHints(prompt string, hints map[string]string, nodeType string) string {
	// For now, return template-applied prompt
	// In the future, this could use more sophisticated optimization
	return applyTemplate(prompt, nodeType)
}

// getPromptTemplate returns the template for a node type
func getPromptTemplate(nodeType string) string {
	templates := map[string]string{
		"Goal": `You are a goal analysis expert. Analyze the following goal and prepare it for execution:

{{prompt}}

Provide a clear, structured analysis.`,
		"Spec": `You are a specification writer. Create a detailed specification based on:

{{prompt}}

Include all requirements, constraints, and acceptance criteria.`,
		"Plan": `You are a technical planner. Create an execution plan for:

{{prompt}}

Break down into clear, actionable steps.`,
		"Implement": `You are a software developer. Implement the following:

{{prompt}}

Write clean, tested, production-ready code.`,
		"Test": `You are a QA engineer. Create tests for:

{{prompt}}

Cover all edge cases and success paths.`,
		"Review": `You are a code reviewer. Review the following:

{{prompt}}

Provide constructive feedback and suggestions for improvement.`,
	}
	return templates[nodeType]
}
