package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type openAIProvider struct {
	cfg *ProviderConfig
}

func (p *openAIProvider) Name() string { return "openai" }

func (p *openAIProvider) Complete(ctx context.Context, prompt string) (<-chan string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *openAIProvider) Chat(ctx context.Context, messages []Message) (<-chan string, error) {
	ch := make(chan string, 10)

	go func() {
		defer close(ch)

		payload, err := json.Marshal(map[string]interface{}{
			"model":       p.cfg.Model,
			"messages":    messages,
			"stream":      true,
			"temperature":  0.7,
			"max_tokens":  4096,
		})
		if err != nil {
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.getBaseURL()+"/chat/completions", strings.NewReader(string(payload)))
		if err != nil {
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if p.cfg.APIKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
		}

		client := &http.Client{Timeout: p.cfg.Timeout}
		resp, err := client.Do(httpReq)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// Close body on context cancellation to unblock the read loop
		go func() {
			<-ctx.Done()
			resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			return
		}

		// Parse SSE stream
		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
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

func (p *openAIProvider) getBaseURL() string {
	if p.cfg.BaseURL != "" {
		return p.cfg.BaseURL
	}
	return "https://api.openai.com/v1"
}
