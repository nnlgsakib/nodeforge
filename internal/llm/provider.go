package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProviderType represents supported LLM providers
type ProviderType string

const (
	ProviderOpenAI     ProviderType = "openai"
	ProviderAnthropic  ProviderType = "anthropic"
	ProviderDeepSeek   ProviderType = "deepseek"
	ProviderOpenRouter ProviderType = "openrouter"
	ProviderOllama     ProviderType = "ollama"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMProvider defines the interface for LLM providers (AC #1)
// Streaming responses via Go channels; error is returned for immediate failures
type LLMProvider interface {
	Complete(ctx context.Context, prompt string) (<-chan string, error)
	Chat(ctx context.Context, messages []Message) (<-chan string, error)
	Name() string
}

// ProviderConfig holds provider configuration (Task 1)
type ProviderConfig struct {
	Type    ProviderType
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

// NewProvider creates a new LLM provider based on config
func NewProvider(cfg *ProviderConfig) (LLMProvider, error) {
	switch cfg.Type {
	case ProviderOpenAI:
		return &openAIProvider{cfg: cfg}, nil
	case ProviderAnthropic:
		return &anthropicProvider{cfg: cfg}, nil
	case ProviderDeepSeek:
		return &deepSeekProvider{cfg: cfg}, nil
	case ProviderOpenRouter:
		return &openRouterProvider{cfg: cfg}, nil
	case ProviderOllama:
		return &ollamaProvider{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", cfg.Type)
	}
}

// RaceResult holds the result of a race between providers (Task 3)
type RaceResult struct {
	ProviderName string
	FirstToken   string
	Duration     time.Duration
}

// Race executes multiple providers and returns the fastest first token (Task 3)
// Losers are cancelled via context cancellation
func Race(ctx context.Context, providers []LLMProvider, prompt string) (*RaceResult, error) {
	type result struct {
		firstToken string
		duration   time.Duration
		err        error
		name       string
	}

	results := make(chan result, len(providers))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range providers {
		wg.Add(1)
		go func(prov LLMProvider) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results <- result{err: fmt.Errorf("provider %s panicked: %v", prov.Name(), r)}
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
				results <- result{firstToken: token, duration: dur, name: prov.Name()}
			case <-ctx.Done():
				return
			}
		}(p)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Wait for first successful result
	for {
		select {
		case r, ok := <-results:
			if !ok {
				return nil, fmt.Errorf("all providers failed")
			}
			if r.err == nil {
				cancel() // cancel remaining requests
				return &RaceResult{
					ProviderName: r.name,
					FirstToken:  r.firstToken,
					Duration:    r.duration,
				}, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("race cancelled: %w", ctx.Err())
		}
	}
}

// RaceSimple is a simpler race that returns the fastest full response (Task 3)
// Unlike Race, this collects the complete response via Chat streaming.
func RaceSimple(ctx context.Context, providers []LLMProvider, messages []Message) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	type result struct {
		output string
		err    error
		name   string
	}

	results := make(chan result, len(providers))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range providers {
		wg.Add(1)
		go func(prov LLMProvider) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results <- result{err: fmt.Errorf("provider %s panicked: %v", prov.Name(), r)}
				}
			}()

			ch, err := prov.Chat(ctx, messages)
			if err != nil {
				results <- result{err: fmt.Errorf("provider %s failed to start: %w", prov.Name(), err)}
				return
			}
			if ch == nil {
				results <- result{err: fmt.Errorf("provider %s returned nil channel", prov.Name())}
				return
			}

			var sb strings.Builder
			for token := range ch {
				sb.WriteString(token)
			}

			if sb.Len() == 0 {
				results <- result{err: fmt.Errorf("provider %s sent no tokens", prov.Name())}
				return
			}
			results <- result{output: sb.String(), name: prov.Name()}
		}(p)
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
				return r.output, nil
			}
		case <-ctx.Done():
			return "", fmt.Errorf("race cancelled: %w", ctx.Err())
		}
	}
}
