package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FallbackChain implements automatic provider failover (Task 4)
// Follows order: Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter (AC #5)
type FallbackChain struct {
	providers []LLMProvider
	logger    func(msg string)
}

// NewFallbackChain creates a new fallback chain with the given providers
// Providers should be ordered by preference (primary first)
func NewFallbackChain(providers []LLMProvider, logger func(msg string)) *FallbackChain {
	if logger == nil {
		logger = func(msg string) { fmt.Printf("[fallback] %s\n", msg) }
	}
	return &FallbackChain{
		providers: providers,
		logger:    logger,
	}
}

// Complete tries providers in order until one succeeds (AC #5)
// Implements semantic matching: rate limit → cheaper/similar model
func (f *FallbackChain) Complete(ctx context.Context, prompt string) (string, error) {
	var lastErr error

	if len(f.providers) == 0 {
		return "", fmt.Errorf("no providers configured for fallback")
	}

	for i, p := range f.providers {
		f.logger(fmt.Sprintf("trying provider %s (%d/%d)", p.Name(), i+1, len(f.providers)))

		// Create a timeout context for this provider
		provCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		ch, err := p.Complete(provCtx, prompt)
		if err != nil {
			cancel()
			f.logger(fmt.Sprintf("provider %s failed to start: %v", p.Name(), err))
			lastErr = err
			continue
		}
		if ch == nil {
			cancel()
			f.logger(fmt.Sprintf("provider %s returned nil channel", p.Name()))
			lastErr = fmt.Errorf("provider %s returned nil channel", p.Name())
			continue
		}

		select {
		case token, ok := <-ch:
			cancel()
			if !ok {
				f.logger(fmt.Sprintf("provider %s sent no tokens", p.Name()))
				lastErr = fmt.Errorf("provider %s sent no tokens", p.Name())
				continue
			}
			f.logger(fmt.Sprintf("provider %s succeeded", p.Name()))
			return token, nil
		case <-provCtx.Done():
			f.logger(fmt.Sprintf("provider %s timed out or cancelled", p.Name()))
			lastErr = provCtx.Err()
			continue
		}
	}

	if lastErr == nil {
		return "", fmt.Errorf("all providers failed")
	}
	return "", fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// Chat tries providers in order until one succeeds
func (f *FallbackChain) Chat(ctx context.Context, messages []Message) (string, error) {
	var lastErr error

	if len(f.providers) == 0 {
		return "", fmt.Errorf("no providers configured for fallback")
	}

	for i, p := range f.providers {
		f.logger(fmt.Sprintf("trying provider %s (%d/%d)", p.Name(), i+1, len(f.providers)))

		provCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		ch, err := p.Chat(provCtx, messages)
		if err != nil {
			cancel()
			f.logger(fmt.Sprintf("provider %s failed to start: %v", p.Name(), err))
			lastErr = err
			continue
		}
		if ch == nil {
			cancel()
			f.logger(fmt.Sprintf("provider %s returned nil channel", p.Name()))
			lastErr = fmt.Errorf("provider %s returned nil channel", p.Name())
			continue
		}

		select {
		case token, ok := <-ch:
			cancel()
			if !ok {
				f.logger(fmt.Sprintf("provider %s sent no tokens", p.Name()))
				lastErr = fmt.Errorf("provider %s sent no tokens", p.Name())
				continue
			}
			f.logger(fmt.Sprintf("provider %s succeeded", p.Name()))
			return token, nil
		case <-provCtx.Done():
			f.logger(fmt.Sprintf("provider %s timed out or cancelled", p.Name()))
			lastErr = provCtx.Err()
			continue
		}
	}

	if lastErr == nil {
		return "", fmt.Errorf("all providers failed")
	}
	return "", fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// DefaultFallbackOrder returns providers in the default fallback order (AC #5)
// Order: Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter
func DefaultFallbackOrder(providers []LLMProvider) []LLMProvider {
	order := []ProviderType{
		ProviderOllama,
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderDeepSeek,
		ProviderOpenRouter,
	}

	// Build a map of available providers by type (case-insensitive name matching)
	providerMap := make(map[ProviderType]LLMProvider)
	for _, p := range providers {
		switch strings.ToLower(p.Name()) {
		case "ollama":
			providerMap[ProviderOllama] = p
		case "openai":
			providerMap[ProviderOpenAI] = p
		case "anthropic":
			providerMap[ProviderAnthropic] = p
		case "deepseek":
			providerMap[ProviderDeepSeek] = p
		case "openrouter":
			providerMap[ProviderOpenRouter] = p
		}
	}

	// Build ordered list based on default fallback order
	var ordered []LLMProvider
	for _, pType := range order {
		if p, ok := providerMap[pType]; ok {
			ordered = append(ordered, p)
		}
	}

	return ordered
}
