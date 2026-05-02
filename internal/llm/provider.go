package llm

import (
	"context"
	"fmt"
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

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Messages    []Message     `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Provider defines the interface for LLM providers
type Provider interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan string, <-chan error)
	Name() string
}

// Config holds provider configuration
type Config struct {
	Type     ProviderType
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  time.Duration
}

// NewProvider creates a new LLM provider based on config
func NewProvider(cfg *Config) (Provider, error) {
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

// RaceResult holds the result of a race between providers
type RaceResult struct {
	ProviderName string
	Response     *ChatResponse
	Duration     time.Duration
}

// Race executes multiple providers and returns the fastest response
func Race(ctx context.Context, providers []Provider, req *ChatRequest) (*RaceResult, error) {
	type result struct {
		resp     *ChatResponse
		duration time.Duration
		err      error
		name     string
	}

	results := make(chan result, len(providers))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, p := range providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					if r != nil {
						results <- result{err: fmt.Errorf("provider %s panicked: %v", prov.Name(), r)}
					}
				}
			}()
			start := time.Now()
			resp, err := prov.Chat(ctx, req)
			dur := time.Since(start)
			if ctx.Err() != nil {
				return
			}
			results <- result{resp: resp, duration: dur, err: err, name: prov.Name()}
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
					Response:     r.resp,
					Duration:     r.duration,
				}, nil
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("race cancelled: %w", ctx.Err())
		}
	}
}
