package llm

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		hasError bool
	}{
		{"OpenAI", &Config{Type: ProviderOpenAI, APIKey: "key"}, false},
		{"Anthropic", &Config{Type: ProviderAnthropic, APIKey: "key"}, false},
		{"Ollama", &Config{Type: ProviderOllama}, false},
		{"Invalid", &Config{Type: "invalid"}, true},
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
	openai := &openAIProvider{cfg: &Config{Type: ProviderOpenAI}}
	assert.Equal(t, "openai", openai.Name())

	ollama := &ollamaProvider{cfg: &Config{Type: ProviderOllama}}
	assert.Equal(t, "ollama", ollama.Name())
}
