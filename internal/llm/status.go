package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ProviderStatus represents the connectivity status of a provider
type ProviderStatus struct {
	Name      string        `json:"name"`
	Online    bool          `json:"online"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
}

// StatusChecker checks and caches provider connectivity status (Task 5)
type StatusChecker struct {
	providers []LLMProvider
	cache     map[string]ProviderStatus
	mu        sync.RWMutex
	cacheTime time.Duration
}

// NewStatusChecker creates a new status checker
// Takes a copy of providers to avoid data races with the caller's slice
func NewStatusChecker(providers []LLMProvider) *StatusChecker {
	providersCopy := make([]LLMProvider, len(providers))
	copy(providersCopy, providers)
	return &StatusChecker{
		providers: providersCopy,
		cache:     make(map[string]ProviderStatus),
		cacheTime: 5 * time.Minute,
	}
}

// CheckAll checks connectivity for all providers (AC #6)
// Returns status for each provider and sends via the provided callback
func (s *StatusChecker) CheckAll(ctx context.Context) map[string]ProviderStatus {
	results := make(map[string]ProviderStatus)
	var wg sync.WaitGroup

	for _, p := range s.providers {
		wg.Add(1)
		go func(prov LLMProvider) {
			defer wg.Done()
			status := s.checkProvider(ctx, prov)
			s.mu.Lock()
			results[prov.Name()] = status
			s.mu.Unlock()
		}(p)
	}

	wg.Wait()
	return results
}

// checkProvider checks if a single provider is reachable
func (s *StatusChecker) checkProvider(ctx context.Context, provider LLMProvider) ProviderStatus {
	start := time.Now()
	status := ProviderStatus{
		Name: provider.Name(),
	}

	// Create a short timeout context for the check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to get a response from the provider
	ch, err := provider.Complete(checkCtx, "test")
	if err != nil {
		status.Online = false
		status.Error = err.Error()
		status.Latency = time.Since(start)
		return status
	}
	if ch == nil {
		status.Online = false
		status.Error = "provider returned nil channel"
		status.Latency = time.Since(start)
		return status
	}

	// Wait for first token or timeout
	select {
	case _, ok := <-ch:
		status.Latency = time.Since(start)
		if !ok {
			status.Online = false
			status.Error = "no tokens received"
		} else {
			status.Online = true
		}
	case <-checkCtx.Done():
		status.Latency = time.Since(start)
		status.Online = false
		status.Error = "timeout"
	}

	return status
}

// GetStatuses returns provider statuses in WebSocket message format (AC #6)
func (s *StatusChecker) GetStatuses(ctx context.Context) map[string]interface{} {
	statuses := s.CheckAll(ctx)

	result := map[string]interface{}{
		"type":       "provider_status",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"providers":  make(map[string]interface{}),
	}

	providersMap := result["providers"].(map[string]interface{})
	for name, status := range statuses {
		providersMap[name] = map[string]interface{}{
			"online": status.Online,
			"latency_ms": status.Latency.Milliseconds(),
			"error":     status.Error,
		}
	}

	return result
}

// FormatWebSocketMessage formats provider status as a JSON WebSocket message
func (s *StatusChecker) FormatWebSocketMessage(ctx context.Context) ([]byte, error) {
	statuses := s.GetStatuses(ctx)
	return json.Marshal(statuses)
}

// CheckAndBroadcast checks all providers and sends status via callback
// Uses the provided context for cancellation (e.g., server shutdown)
func (s *StatusChecker) CheckAndBroadcast(ctx context.Context, broadcastFn func(data []byte)) {
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		data, err := s.FormatWebSocketMessage(ctx)
		if err != nil {
			fmt.Printf("[status] Failed to format status message: %v\n", err)
			return
		}
		if broadcastFn != nil {
			broadcastFn(data)
		}
	}()
}
