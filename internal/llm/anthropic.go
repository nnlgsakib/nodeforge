package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type anthropicProvider struct {
	cfg *ProviderConfig
}

func (p *anthropicProvider) Name() string { return "anthropic" }

func (p *anthropicProvider) Complete(ctx context.Context, prompt string) (<-chan string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *anthropicProvider) Chat(ctx context.Context, messages []Message) (<-chan string, error) {
	ch := make(chan string, 10)

	// Convert messages to Anthropic format
	var anthropicMessages []map[string]interface{}
	for _, msg := range messages {
		anthropicMessages = append(anthropicMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	payload, err := json.Marshal(map[string]interface{}{
		"model":     p.cfg.Model,
		"messages":  anthropicMessages,
		"stream":    true,
		"max_tokens": 4096,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.getBaseURL()+"/v1/messages", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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
				Type  string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.Type == "content_block_delta" && chunk.Delta.Text != "" {
				select {
				case ch <- chunk.Delta.Text:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (p *anthropicProvider) getBaseURL() string {
	if p.cfg.BaseURL != "" {
		return p.cfg.BaseURL
	}
	return "https://api.anthropic.com"
}
