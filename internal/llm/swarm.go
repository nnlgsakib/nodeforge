package llm

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// SwarmResult holds the result of a single speculative attempt
type SwarmResult struct {
	AttemptID    string
	ProviderName string
	Output       string
	TokensUsed   int
	Duration     time.Duration
	Score        float64
	PassedAC     bool
	Err          error
}

// SwarmConfig configures speculative execution within a node
type SwarmConfig struct {
	// Enabled turns speculative execution on/off
	Enabled bool
	// MaxAttempts is the maximum number of parallel attempts (default: 3)
	MaxAttempts int
	// Timeout is the maximum time to wait for all attempts (default: 60s)
	Timeout time.Duration
	// MinScore is the minimum score for a result to be considered acceptable
	MinScore float64
}

// DefaultSwarmConfig returns sensible defaults
func DefaultSwarmConfig() *SwarmConfig {
	return &SwarmConfig{
		Enabled:     false,
		MaxAttempts: 3,
		Timeout:     60 * time.Second,
		MinScore:    0.7,
	}
}

// Swarm orchestrates multiple LLM agents negotiating within a single node.
// It runs multiple speculative attempts in parallel, scores each result
// against acceptance criteria, and selects the best one.
type Swarm struct {
	config   *SwarmConfig
	provider LLMProvider
	budget   *BudgetEnforcer
}

// NewSwarm creates a new Swarm for speculative execution
func NewSwarm(config *SwarmConfig, provider LLMProvider, budget *BudgetEnforcer) *Swarm {
	if config == nil {
		config = DefaultSwarmConfig()
	}
	return &Swarm{
		config:   config,
		provider: provider,
		budget:   budget,
	}
}

// Execute runs multiple parallel attempts and returns the best result
// that passes acceptance criteria verification.
// Failed attempts are logged but don't block progress.
func (s *Swarm) Execute(
	ctx context.Context,
	nodeID string,
	messages []Message,
	acceptanceCriteria []string,
) (*SwarmResult, error) {
	if !s.config.Enabled {
		// Fall back to single execution
		return s.executeSingle(ctx, messages, acceptanceCriteria)
	}

	maxAttempts := s.config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	timeout := s.config.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	// Separate cancel for early termination when best result is found
	ctx, cancelEarly := context.WithCancel(ctx)
	defer cancelEarly()

	type attemptResult struct {
		index int
		result *SwarmResult
	}

	// Pre-flight budget check: ensure we have budget before launching goroutines
	if s.budget != nil {
		estimatedTokens := maxAttempts * 500 // rough estimate: 500 tokens per attempt
		if err := s.budget.CheckBudget(ctx, estimatedTokens); err != nil {
			return nil, fmt.Errorf("pre-flight budget check failed: %w", err)
		}
	} else {
		log.Printf("[WARN] budget enforcer is nil — token usage will not be tracked or limited")
	}

	resultsCh := make(chan attemptResult, maxAttempts)
	var wg sync.WaitGroup
	var once sync.Once

	for i := 0; i < maxAttempts; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					resultsCh <- attemptResult{
						index: idx,
						result: &SwarmResult{
							AttemptID: fmt.Sprintf("attempt-%d", idx),
							Err:       fmt.Errorf("attempt %d panicked: %v", idx, r),
						},
					}
				}
			}()

			res := s.runAttempt(ctx, idx, messages, acceptanceCriteria)
			resultsCh <- attemptResult{index: idx, result: res}

			// If this result passed AC with a good score, cancel remaining attempts
			if res != nil && res.Err == nil && res.PassedAC && res.Score >= s.config.MinScore {
				once.Do(func() {
					log.Printf("[INFO] early cancellation: attempt %d passed AC with score %.2f", idx, res.Score)
					cancelEarly()
				})
			}
		}(i)
	}

	// Close channel when all attempts complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect all results
	var results []*SwarmResult
	for ar := range resultsCh {
		results = append(results, ar.result)
		if ar.result.Err != nil {
			log.Printf("[WARN] speculative attempt %d failed: %v", ar.index, ar.result.Err)
		}
	}

	// Select best result
	best := s.selectBestResult(results)
	if best == nil {
		return nil, fmt.Errorf("all %d speculative attempts failed", maxAttempts)
	}
	if !best.PassedAC {
		log.Printf("[WARN] no speculative attempt passed acceptance criteria, returning best available")
	}

	return best, nil
}

// executeSingle runs a single LLM attempt (non-speculative path)
func (s *Swarm) executeSingle(
	ctx context.Context,
	messages []Message,
	acceptanceCriteria []string,
) (*SwarmResult, error) {
	start := time.Now()

	ch, err := s.provider.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("single execution failed: %w", err)
	}

	var output strings.Builder
	var totalTokens int
	for token := range ch {
		output.WriteString(token)
		totalTokens += EstimateTokens(token)
	}

	outputStr := output.String()
	if outputStr == "" {
		return nil, fmt.Errorf("single execution returned empty output")
	}

	passedAC, score := s.scoreResult(outputStr, acceptanceCriteria)

	return &SwarmResult{
		AttemptID:    "single",
		ProviderName: s.provider.Name(),
		Output:       outputStr,
		TokensUsed:   totalTokens,
		Duration:     time.Since(start),
		Score:        score,
		PassedAC:     passedAC,
	}, nil
}

// runAttempt executes a single speculative attempt
func (s *Swarm) runAttempt(
	ctx context.Context,
	idx int,
	messages []Message,
	acceptanceCriteria []string,
) *SwarmResult {
	attemptID := fmt.Sprintf("attempt-%d", idx)
	start := time.Now()

	ch, err := s.provider.Chat(ctx, messages)
	if err != nil {
		return &SwarmResult{
			AttemptID: attemptID,
			Err:       fmt.Errorf("chat failed: %w", err),
		}
	}
	if ch == nil {
		return &SwarmResult{
			AttemptID: attemptID,
			Err:       fmt.Errorf("provider returned nil channel"),
		}
	}

	var output strings.Builder
	var totalTokens int
	for {
		select {
		case token, ok := <-ch:
			if !ok {
				goto collectDone
			}
			output.WriteString(token)
			totalTokens += EstimateTokens(token)
		case <-ctx.Done():
			return &SwarmResult{
				AttemptID: attemptID,
				Err:       fmt.Errorf("attempt cancelled: %w", ctx.Err()),
			}
		}
	}

collectDone:
	outputStr := output.String()
	if outputStr == "" {
		return &SwarmResult{
			AttemptID: attemptID,
			Err:       fmt.Errorf("attempt produced empty output"),
		}
	}

	// Track token usage against budget
	if s.budget != nil {
		s.budget.TrackUsage(totalTokens)
	}

	passedAC, score := s.scoreResult(outputStr, acceptanceCriteria)

	return &SwarmResult{
		AttemptID:    attemptID,
		ProviderName: s.provider.Name(),
		Output:       outputStr,
		TokensUsed:   totalTokens,
		Duration:     time.Since(start),
		Score:        score,
		PassedAC:     passedAC,
	}
}

// scoreResult evaluates an output against acceptance criteria
// Returns (passedAllCriteria, score) where score is 0.0-1.0
func (s *Swarm) scoreResult(output string, acceptanceCriteria []string) (bool, float64) {
	if len(acceptanceCriteria) == 0 {
		// No criteria - accept any reasonable output
		if len(output) < 10 {
			return false, 0.0
		}
		return true, 0.8
	}

	metCount := 0
	for _, criterion := range acceptanceCriteria {
		if criterion == "" {
			continue
		}
		if containsSubstring(output, criterion) {
			metCount++
		}
	}

	score := float64(metCount) / float64(len(acceptanceCriteria))
	passed := score >= s.config.MinScore

	return passed, score
}

// selectBestResult picks the highest-scoring result that passed AC.
// PassedAC results are strictly preferred over non-passed results.
// If no result passed AC, returns the highest-scoring available.
func (s *Swarm) selectBestResult(results []*SwarmResult) *SwarmResult {
	if len(results) == 0 {
		return nil
	}

	// Filter out errored results
	var valid []*SwarmResult
	for _, r := range results {
		if r.Err == nil && r.Output != "" {
			valid = append(valid, r)
		}
	}
	if len(valid) == 0 {
		return nil
	}

	// Sort by PassedAC first (true > false), then by score descending
	sort.Slice(valid, func(i, j int) bool {
		if valid[i].PassedAC != valid[j].PassedAC {
			return valid[i].PassedAC // true sorts before false
		}
		return valid[i].Score > valid[j].Score
	})

	return valid[0]
}

// containsSubstring checks if substr appears in s (case-insensitive)
func containsSubstring(s, substr string) bool {
	if substr == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
