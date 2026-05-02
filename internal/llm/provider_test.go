package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProviderConfig
		hasError bool
	}{
		{"OpenAI", &ProviderConfig{Type: ProviderOpenAI, APIKey: "key"}, false},
		{"Anthropic", &ProviderConfig{Type: ProviderAnthropic, APIKey: "key"}, false},
		{"Ollama", &ProviderConfig{Type: ProviderOllama}, false},
		{"Invalid", &ProviderConfig{Type: "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov, err := NewProvider(tt.cfg)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, prov)
			}
		})
	}
}

func TestProviderNames(t *testing.T) {
	openai := &openAIProvider{cfg: &ProviderConfig{Type: ProviderOpenAI}}
	assert.Equal(t, "openai", openai.Name())

	ollama := &ollamaProvider{cfg: &ProviderConfig{Type: ProviderOllama}}
	assert.Equal(t, "ollama", ollama.Name())
}

func TestLLMProviderInterface(t *testing.T) {
	// Verify all providers implement LLMProvider interface
	providers := []LLMProvider{
		&openAIProvider{cfg: &ProviderConfig{Type: ProviderOpenAI, APIKey: "test"}},
		&anthropicProvider{cfg: &ProviderConfig{Type: ProviderAnthropic, APIKey: "test"}},
		&deepSeekProvider{cfg: &ProviderConfig{Type: ProviderDeepSeek, APIKey: "test"}},
		&openRouterProvider{cfg: &ProviderConfig{Type: ProviderOpenRouter, APIKey: "test"}},
		&ollamaProvider{cfg: &ProviderConfig{Type: ProviderOllama, BaseURL: "http://localhost:11434"}},
	}

	for _, p := range providers {
		assert.NotNil(t, p)
		assert.NotEmpty(t, p.Name())
	}
}

func TestRaceMode(t *testing.T) {
	// Test that RaceMode can be created
	cfg := &ProviderConfig{
		Type:    ProviderOllama,
		BaseURL: "http://localhost:11434",
		Model:   "llama3",
		Timeout: 30 * time.Second,
	}
	prov, err := NewProvider(cfg)
	assert.NoError(t, err)

	rm := NewRaceMode([]LLMProvider{prov}, 30*time.Second)
	assert.NotNil(t, rm)
}

func TestFallbackChain(t *testing.T) {
	// Test that FallbackChain can be created
	cfg := &ProviderConfig{
		Type:    ProviderOllama,
		BaseURL: "http://localhost:11434",
		Model:   "llama3",
		Timeout: 30 * time.Second,
	}
	prov, err := NewProvider(cfg)
	assert.NoError(t, err)

	fc := NewFallbackChain([]LLMProvider{prov}, nil)
	assert.NotNil(t, fc)
}

func TestDefaultFallbackOrder(t *testing.T) {
	// Create providers
	ollamaCfg := &ProviderConfig{Type: ProviderOllama, BaseURL: "http://localhost:11434"}
	openaiCfg := &ProviderConfig{Type: ProviderOpenAI, APIKey: "key"}

	ollamaProv, _ := NewProvider(ollamaCfg)
	openaiProv, _ := NewProvider(openaiCfg)

	providers := []LLMProvider{openaiProv, ollamaProv} // Intentionally wrong order
	ordered := DefaultFallbackOrder(providers)

	// Should be ordered: Ollama, OpenAI
	assert.Len(t, ordered, 2)
	assert.Equal(t, "ollama", ordered[0].Name())
	assert.Equal(t, "openai", ordered[1].Name())
}

// TestStatusChecker tests the provider status checker
func TestStatusChecker(t *testing.T) {
	cfg := &ProviderConfig{
		Type:    ProviderOllama,
		BaseURL: "http://localhost:11434",
		Model:   "llama3",
		Timeout: 5 * time.Second,
	}
	prov, err := NewProvider(cfg)
	assert.NoError(t, err)

	sc := NewStatusChecker([]LLMProvider{prov})
	assert.NotNil(t, sc)

	// Test GetStatuses
	statuses := sc.GetStatuses(context.Background())
	assert.NotNil(t, statuses)
	assert.Equal(t, "provider_status", statuses["type"])
}

// TestRaceModeFastestWins tests that race mode returns the fastest token
// Note: This is a basic test since we can't easily mock streaming responses
func TestRaceModeCreation(t *testing.T) {
	cfg := &ProviderConfig{
		Type:    ProviderOllama,
		BaseURL: "http://localhost:11434",
		Model:   "llama3",
		Timeout: 30 * time.Second,
	}
	prov, err := NewProvider(cfg)
	assert.NoError(t, err)

	rm := NewRaceMode([]LLMProvider{prov}, 30*time.Second)
	assert.NotNil(t, rm)
}

func TestNewProviderAllTypes(t *testing.T) {
	tests := []struct {
		name string
		cfg  *ProviderConfig
	}{
		{"OpenAI", &ProviderConfig{Type: ProviderOpenAI, APIKey: "test-key"}},
		{"Anthropic", &ProviderConfig{Type: ProviderAnthropic, APIKey: "test-key"}},
		{"DeepSeek", &ProviderConfig{Type: ProviderDeepSeek, APIKey: "test-key"}},
		{"OpenRouter", &ProviderConfig{Type: ProviderOpenRouter, APIKey: "test-key"}},
		{"Ollama", &ProviderConfig{Type: ProviderOllama, BaseURL: "http://localhost:11434"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov, err := NewProvider(tt.cfg)
			assert.NoError(t, err)
			assert.NotNil(t, prov)
			assert.Equal(t, string(tt.cfg.Type), prov.Name())
		})
	}
}
