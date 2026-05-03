package context

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// ContextAssembler assembles context for LLM calls (Subtask 5.1)
type ContextAssembler struct {
	kg      *KnowledgeGraph // Uses KnowledgeGraph for context assembly
	splitter *GraphSplitter // Handles context overflow (AC4)
}

// NewContextAssembler creates a new ContextAssembler
func NewContextAssembler(db *badger.DB) *ContextAssembler {
	return &ContextAssembler{
		kg:       NewKnowledgeGraph(db),
		splitter:  NewGraphSplitter(db),
	}
}

// AssembleContext assembles context for LLM call (called by LLM provider before invocation)
// Must complete in <100ms (NFR-04), fallback to naive prompt on timeout (Subtask 5.4)
func (ca *ContextAssembler) AssembleContext(nodeID string, maxTokens int) (string, error) {
	start := time.Now()
	defer func() {
		if dur := time.Since(start); dur > 100*time.Millisecond {
			fmt.Printf("WARN: AssembleContext took %v (>100ms threshold)\n", dur)
		}
	}()

	// Check if graph needs splitting (AC4 - FR20)
	subGraphs, err := ca.splitter.SplitGraphIfNeeded(nodeID, maxTokens)
	if err != nil {
		fmt.Printf("WARN: SplitGraphIfNeeded failed: %v\n", err)
	}
	if len(subGraphs) > 1 {
		fmt.Printf("INFO: Graph split into %d sub-graphs for node %s\n", len(subGraphs), nodeID)
	}

	// Create context with 100ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	contextStr, err := ca.kg.BuildContext(ctx, nodeID, maxTokens)
	if err != nil {
		// Check if it was a timeout
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("WARN: Context assembly timed out, falling back to naive prompt")
			return "", nil
		}
		return "", fmt.Errorf("failed to assemble context: %w", err)
	}

	return contextStr, nil
}

// InjectContextIntoPrompt injects context into LLM prompt (Subtask 5.3)
// Used by story 2.4's PromptOptimizer
func InjectContextIntoPrompt(prompt, ctxContext string) string {
	if ctxContext == "" {
		return prompt
	}
	return fmt.Sprintf("%s\n\n[Context]:\n%s", prompt, ctxContext)
}

// Documented integration interface (Subtask 5.2):
// To integrate with LLM provider (story 2.2):
// 1. In internal/llm/provider.go, before calling LLM API:
//    contextAssembler := context.NewContextAssembler(db)
//    ctxContext, err := contextAssembler.AssembleContext(nodeID, maxTokens)
//    if err == nil {
//        prompt = context.InjectContextIntoPrompt(originalPrompt, ctxContext)
//    }
// 2. Pass the enhanced prompt to LLM provider's Chat() or ChatStream() method
//
// Integration with Prompt Optimizer (story 2.4):
// - PromptOptimizer.OptimizePrompt() can use the context from AssembleContext()
// - Both have separate timing requirements: <10ms for budget, <100ms for context
