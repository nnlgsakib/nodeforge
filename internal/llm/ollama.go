package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ollamaProvider struct {
	cfg *ProviderConfig
}

func (p *ollamaProvider) Name() string { return "ollama" }

func (p *ollamaProvider) Complete(ctx context.Context, prompt string) (<-chan string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *ollamaProvider) Chat(ctx context.Context, messages []Message) (<-chan string, error) {
	ch := make(chan string, 10)

	payload, err := json.Marshal(map[string]interface{}{
		"model":    p.cfg.Model,
		"messages": messages,
		"stream":   true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.getBaseURL()+"/api/chat", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}

			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			if chunk.Message.Content != "" {
				select {
				case ch <- chunk.Message.Content:
				case <-ctx.Done():
					return
				}
			}

			if chunk.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			// Stream read error - channel will be closed without further tokens
			// Error is silently absorbed as the channel is already being closed
		}
	}()

	return ch, nil
}

func (p *ollamaProvider) getBaseURL() string {
	if p.cfg.BaseURL != "" {
		return p.cfg.BaseURL
	}
	return "http://localhost:11434"
}
