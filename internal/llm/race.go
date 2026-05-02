package llm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RaceMode runs multiple providers simultaneously and returns the fastest response (Task 3)
// Implements AC #4: race mode with context cancellation for losing providers
type RaceMode struct {
	providers []LLMProvider
	timeout   time.Duration
}

// NewRaceMode creates a new RaceMode with the given providers
func NewRaceMode(providers []LLMProvider, timeout time.Duration) *RaceMode {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &RaceMode{
		providers: providers,
		timeout:   timeout,
	}
}

// Complete launches goroutines per provider and returns the fastest first token (AC #4)
// Losing providers are cancelled via context cancellation
func (r *RaceMode) Complete(ctx context.Context, prompt string) (string, error) {
	type result struct {
		token    string
		duration time.Duration
		err      error
		name     string
	}

	results := make(chan result, len(r.providers))
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range r.providers {
		wg.Add(1)
		go func(prov LLMProvider) {
			defer wg.Done()
			defer func() {
				if rec := recover(); rec != nil {
					results <- result{err: fmt.Errorf("provider %s panicked: %v", prov.Name(), rec)}
				}
			}()

			start := time.Now()
			ch, err := prov.Complete(ctx, prompt)
			if err != nil {
				results <- result{err: fmt.Errorf("provider %s failed to start: %w", prov.Name(), err)}
				return
			}
			if ch == nil {
				results <- result{err: fmt.Errorf("provider %s returned nil channel", prov.Name())}
				return
			}

			select {
			case token, ok := <-ch:
				if !ok {
					results <- result{err: fmt.Errorf("provider %s sent no tokens", prov.Name())}
					return
				}
				dur := time.Since(start)
				results <- result{token: token, duration: dur, name: prov.Name()}
			case <-ctx.Done():
				return
			}
		}(p)
	}

	if len(r.providers) == 0 {
		return "", fmt.Errorf("no providers available for race")
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for {
		select {
		case r, ok := <-results:
			if !ok {
				return "", fmt.Errorf("all providers failed")
			}
			if r.err == nil {
				cancel()
				return r.token, nil
			}
		case <-ctx.Done():
			return "", fmt.Errorf("race timed out: %w", ctx.Err())
		}
	}
}

// Chat launches goroutines per provider and returns the fastest chat response
func (r *RaceMode) Chat(ctx context.Context, messages []Message) (string, error) {
	type result struct {
		token    string
		duration time.Duration
		err      error
		name     string
	}

	results := make(chan result, len(r.providers))
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range r.providers {
		wg.Add(1)
		go func(prov LLMProvider) {
			defer wg.Done()
			defer func() {
				if rec := recover(); rec != nil {
					results <- result{err: fmt.Errorf("provider %s panicked: %v", prov.Name(), rec)}
				}
			}()

			start := time.Now()
			ch, err := prov.Chat(ctx, messages)
			if err != nil {
				results <- result{err: fmt.Errorf("provider %s failed to start: %w", prov.Name(), err)}
				return
			}
			if ch == nil {
				results <- result{err: fmt.Errorf("provider %s returned nil channel", prov.Name())}
				return
			}

			select {
			case token, ok := <-ch:
				if !ok {
					results <- result{err: fmt.Errorf("provider %s sent no tokens", prov.Name())}
					return
				}
				dur := time.Since(start)
				results <- result{token: token, duration: dur, name: prov.Name()}
			case <-ctx.Done():
				return
			}
		}(p)
	}

	if len(r.providers) == 0 {
		return "", fmt.Errorf("no providers available for race")
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for {
		select {
		case r, ok := <-results:
			if !ok {
				return "", fmt.Errorf("all providers failed")
			}
			if r.err == nil {
				cancel()
				return r.token, nil
			}
		case <-ctx.Done():
			return "", fmt.Errorf("race timed out: %w", ctx.Err())
		}
	}
}
