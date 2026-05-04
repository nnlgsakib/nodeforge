package skills

import (
	"math/rand"
	"sync"
	"time"
)

// ABTestVariant represents a single variant in an A/B test.
type ABTestVariant struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Weight float64 `json:"weight"` // Traffic allocation (0-1), must sum to 1.0 across variants
}

// ABTestConfig holds the configuration for an A/B test.
type ABTestConfig struct {
	SkillID  string           `json:"skillId"`
	Variants []ABTestVariant  `json:"variants"`
}

// ABTestMetrics collects metrics for a single variant.
type ABTestMetrics struct {
	Executions    int     `json:"executions"`
	Successes     int     `json:"successes"`
	TotalTimeMs   float64 `json:"totalTimeMs"`
	TokenUsage    int     `json:"tokenUsage"`
}

// ABTestRunner manages A/B test routing and metrics collection.
type ABTestRunner struct {
	mu      sync.RWMutex
	tests   map[string]*ABTestConfig
	metrics map[string]map[string]*ABTestMetrics // testID -> variantID -> metrics
	rng     *rand.Rand
}

// NewABTestRunner creates a new A/B test runner.
func NewABTestRunner() *ABTestRunner {
	return &ABTestRunner{
		tests:   make(map[string]*ABTestConfig),
		metrics: make(map[string]map[string]*ABTestMetrics),
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetAllTests returns all registered test skill IDs.
func (r *ABTestRunner) GetAllTests() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]bool)
	for id := range r.tests {
		result[id] = true
	}
	return result
}

// RegisterTest adds an A/B test configuration.
func (r *ABTestRunner) RegisterTest(config *ABTestConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate weights sum to ~1.0
	total := 0.0
	for _, v := range config.Variants {
		total += v.Weight
	}
	if total < 0.99 || total > 1.01 {
		// Normalize weights
		if total > 0 {
			for i := range config.Variants {
				config.Variants[i].Weight /= total
			}
		}
	}

	r.tests[config.SkillID] = config
	r.metrics[config.SkillID] = make(map[string]*ABTestMetrics)
	for _, v := range config.Variants {
		r.metrics[config.SkillID][v.ID] = &ABTestMetrics{}
	}
}

// SelectVariant returns the variant ID to use for a given skill ID based on weighted random selection.
func (r *ABTestRunner) SelectVariant(skillID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.tests[skillID]
	if !ok || len(config.Variants) == 0 {
		return "" // No A/B test registered, use default
	}

	// Weighted random selection
	roll := r.rng.Float64()
	cumulative := 0.0
	for _, v := range config.Variants {
		cumulative += v.Weight
		if roll <= cumulative {
			return v.ID
		}
	}
	// Fallback to last variant
	return config.Variants[len(config.Variants)-1].ID
}

// RecordMetrics records execution metrics for a variant.
func (r *ABTestRunner) RecordMetrics(skillID, variantID string, success bool, durationMs float64, tokens int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.metrics[skillID]
	if !ok {
		return
	}
	vm, ok := m[variantID]
	if !ok {
		return
	}

	vm.Executions++
	if success {
		vm.Successes++
	}
	vm.TotalTimeMs += durationMs
	vm.TokenUsage += tokens
}

// GetMetrics returns a deep copy of all metrics for a given skill ID.
func (r *ABTestRunner) GetMetrics(skillID string) map[string]*ABTestMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ABTestMetrics)
	for k, v := range r.metrics[skillID] {
		// Deep copy to prevent external mutation
		cp := *v
		result[k] = &cp
	}
	return result
}
