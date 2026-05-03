package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// swarmMockProvider is a test double that returns predefined output
type swarmMockProvider struct {
	name   string
	output string
	delay  time.Duration
	fail   bool
}

func (m *swarmMockProvider) Complete(ctx context.Context, prompt string) (<-chan string, error) {
	return m.streamOutput(ctx)
}

func (m *swarmMockProvider) Chat(ctx context.Context, messages []Message) (<-chan string, error) {
	return m.streamOutput(ctx)
}

func (m *swarmMockProvider) Name() string {
	return m.name
}

func (m *swarmMockProvider) streamOutput(ctx context.Context) (<-chan string, error) {
	if m.fail {
		return nil, fmt.Errorf("mock provider %s failed", m.name)
	}
	ch := make(chan string, 1)
	go func() {
		defer close(ch)
		if m.delay > 0 {
			select {
			case <-time.After(m.delay):
			case <-ctx.Done():
				return
			}
		}
		// Stream output in chunks
		for i := 0; i < len(m.output); i += 10 {
			end := i + 10
			if end > len(m.output) {
				end = len(m.output)
			}
			select {
			case ch <- m.output[i:end]:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func TestNewSwarm(t *testing.T) {
	provider := &swarmMockProvider{name: "test", output: "hello world"}
	swarm := NewSwarm(nil, provider, nil)

	assert.NotNil(t, swarm)
	assert.False(t, swarm.config.Enabled)
	assert.Equal(t, 3, swarm.config.MaxAttempts)
}

func TestSwarmExecuteSingle(t *testing.T) {
	provider := &swarmMockProvider{
		name:   "test-provider",
		output: "This is a valid output that meets the acceptance criteria",
	}
	config := &SwarmConfig{
		Enabled:     false,
		MaxAttempts: 3,
		Timeout:     5 * time.Second,
		MinScore:    0.7,
	}
	swarm := NewSwarm(config, provider, nil)

	result, err := swarm.Execute(
		context.Background(),
		"node-1",
		[]Message{{Role: "user", Content: "test"}},
		[]string{"valid output"},
	)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Output, "valid output")
	assert.True(t, result.PassedAC)
	assert.Equal(t, "test-provider", result.ProviderName)
}

func TestSwarmExecuteSpeculative(t *testing.T) {
	provider := &swarmMockProvider{
		name:   "test-provider",
		output: "Result: This output addresses all acceptance criteria properly",
	}
	config := &SwarmConfig{
		Enabled:     true,
		MaxAttempts: 3,
		Timeout:     5 * time.Second,
		MinScore:    0.5,
	}
	swarm := NewSwarm(config, provider, nil)

	result, err := swarm.Execute(
		context.Background(),
		"node-1",
		[]Message{{Role: "user", Content: "test"}},
		[]string{"acceptance criteria"},
	)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "attempt-0", result.AttemptID)
	assert.True(t, result.Score > 0)
}

func TestSwarmSelectsBestResult(t *testing.T) {
	// Use a provider that produces consistent output
	provider := &swarmMockProvider{
		name:   "provider",
		output: "This output fully meets all stated acceptance criteria",
	}
	config := &SwarmConfig{
		Enabled:     true,
		MaxAttempts: 3,
		Timeout:     5 * time.Second,
		MinScore:    0.5,
	}
	budget := NewBudgetEnforcer(&TokenBudget{
		TotalBudgetPerSession:  100000,
		MaxTokensPerRequest:    4096,
	})
	swarm := NewSwarm(config, provider, budget)

	result, err := swarm.Execute(
		context.Background(),
		"node-1",
		[]Message{{Role: "user", Content: "test"}},
		[]string{"acceptance criteria", "meets all"},
	)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.Score, 0.0)
	assert.True(t, result.PassedAC)
}

func TestSwarmAllAttemptsFail(t *testing.T) {
	provider := &swarmMockProvider{
		name:   "failing",
		output: "",
		fail:   true,
	}
	config := &SwarmConfig{
		Enabled:     true,
		MaxAttempts: 2,
		Timeout:     5 * time.Second,
		MinScore:    0.7,
	}
	swarm := NewSwarm(config, provider, nil)

	_, err := swarm.Execute(
		context.Background(),
		"node-1",
		[]Message{{Role: "user", Content: "test"}},
		[]string{"criteria"},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestSwarmContextCancellation(t *testing.T) {
	provider := &swarmMockProvider{
		name:   "slow",
		output: "slow output",
		delay:  2 * time.Second,
	}
	config := &SwarmConfig{
		Enabled:     true,
		MaxAttempts: 3,
		Timeout:     5 * time.Second,
		MinScore:    0.7,
	}
	swarm := NewSwarm(config, provider, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := swarm.Execute(
		ctx,
		"node-1",
		[]Message{{Role: "user", Content: "test"}},
		[]string{"criteria"},
	)

	assert.Error(t, err)
}

func TestSwarmScoreResult(t *testing.T) {
	swarm := &Swarm{config: DefaultSwarmConfig()}

	// All criteria met
	passed, score := swarm.scoreResult("This output meets criterion one and criterion two", []string{"criterion one", "criterion two"})
	assert.True(t, passed)
	assert.Equal(t, 1.0, score)

	// Partial criteria met
	passed, score = swarm.scoreResult("This only meets criterion one", []string{"criterion one", "criterion two"})
	assert.False(t, passed)
	assert.Equal(t, 0.5, score)

	// No criteria - any reasonable output passes
	passed, score = swarm.scoreResult("This is a long enough output for testing purposes", []string{})
	assert.True(t, passed)
	assert.Equal(t, 0.8, score)

	// Short output fails
	passed, score = swarm.scoreResult("short", []string{})
	assert.False(t, passed)
	assert.Equal(t, 0.0, score)
}

func TestSwarmSelectBestResult(t *testing.T) {
	swarm := &Swarm{config: DefaultSwarmConfig()}

	results := []*SwarmResult{
		{AttemptID: "a1", Output: "out1", Score: 0.5, Err: nil},
		{AttemptID: "a2", Output: "out2", Score: 0.9, Err: nil},
		{AttemptID: "a3", Output: "out3", Score: 0.7, Err: nil},
	}

	best := swarm.selectBestResult(results)
	assert.NotNil(t, best)
	assert.Equal(t, "a2", best.AttemptID)
	assert.Equal(t, 0.9, best.Score)
}

func TestSwarmSelectBestResultWithErrors(t *testing.T) {
	swarm := &Swarm{config: DefaultSwarmConfig()}

	results := []*SwarmResult{
		{AttemptID: "a1", Output: "", Err: fmt.Errorf("failed")},
		{AttemptID: "a2", Output: "good output", Score: 0.8, Err: nil},
		{AttemptID: "a3", Output: "", Err: fmt.Errorf("panic")},
	}

	best := swarm.selectBestResult(results)
	assert.NotNil(t, best)
	assert.Equal(t, "a2", best.AttemptID)
}

func TestSwarmSelectBestResultEmpty(t *testing.T) {
	swarm := &Swarm{config: DefaultSwarmConfig()}

	best := swarm.selectBestResult([]*SwarmResult{})
	assert.Nil(t, best)

	best = swarm.selectBestResult([]*SwarmResult{
		{AttemptID: "a1", Output: "", Err: fmt.Errorf("failed")},
	})
	assert.Nil(t, best)
}

func TestSwarmBudgetTracking(t *testing.T) {
	provider := &swarmMockProvider{
		name:   "budget-test",
		output: "This is output that uses some tokens for budget testing purposes",
	}
	budget := NewBudgetEnforcer(&TokenBudget{
		TotalBudgetPerSession:  100000,
		MaxTokensPerRequest:    4096,
	})
	config := &SwarmConfig{
		Enabled:     true,
		MaxAttempts: 2,
		Timeout:     5 * time.Second,
		MinScore:    0.5,
	}
	swarm := NewSwarm(config, provider, budget)

	_, err := swarm.Execute(
		context.Background(),
		"node-1",
		[]Message{{Role: "user", Content: "test"}},
		[]string{"budget"},
	)

	require.NoError(t, err)
	used, remaining, total := budget.BudgetStatus()
	assert.Greater(t, used, 0)
	assert.Equal(t, 100000, total)
	assert.Less(t, remaining, total)
}

func TestDefaultSwarmConfig(t *testing.T) {
	cfg := DefaultSwarmConfig()
	assert.False(t, cfg.Enabled)
	assert.Equal(t, 3, cfg.MaxAttempts)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
	assert.Equal(t, 0.7, cfg.MinScore)
}

func TestContainsSubstring(t *testing.T) {
	assert.True(t, containsSubstring("Hello World", "world"))
	assert.True(t, containsSubstring("Hello World", ""))
	assert.False(t, containsSubstring("Hello", "world"))
	assert.True(t, containsSubstring("ACCEPTANCE CRITERIA MET", "acceptance"))
}
