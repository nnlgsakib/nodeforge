package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// deepSeekProvider implements LLMProvider using DeepSeek API (OpenAI-compatible)
type deepSeekProvider struct {
	cfg *ProviderConfig
}

func (p *deepSeekProvider) Name() string { return "deepseek" }

func (p *deepSeekProvider) Complete(ctx context.Context, prompt string) (<-chan string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *deepSeekProvider) Chat(ctx context.Context, messages []Message) (<-chan string, error) {
	ch := make(chan string, 10)

	payload, err := json.Marshal(map[string]interface{}{
		"model":       p.cfg.Model,
		"messages":    messages,
		"stream":      true,
		"temperature": 0.7,
		"max_tokens":  4096,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.getBaseURL()+"/chat/completions", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	}

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		// Unblock reads on context cancellation
		go func() {
			<-ctx.Done()
			resp.Body.Close()
		}()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				select {
				case ch <- chunk.Choices[0].Delta.Content:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (p *deepSeekProvider) getBaseURL() string {
	if p.cfg.BaseURL != "" {
		return p.cfg.BaseURL
	}
	return "https://api.deepseek.com"
}
