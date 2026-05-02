package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ollamaProvider struct {
	cfg *Config
}

func (p *ollamaProvider) Name() string {
	return "ollama"
}

func (p *ollamaProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"model":  p.cfg.Model,
		"messages": req.Messages,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": req.Temperature,
			"num_predict": req.MaxTokens,
		},
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, string(body))
	}

	var ollamaResp struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ChatResponse{
		ID: "",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}

func (p *ollamaProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan string, <-chan error) {
	ch := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		payload, _ := json.Marshal(map[string]interface{}{
			"model":  p.cfg.Model,
			"messages": req.Messages,
			"stream": true,
		})

		httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.getBaseURL()+"/api/chat", strings.NewReader(string(payload)))
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: p.cfg.Timeout}
		resp, err := client.Do(httpReq)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		ch <- string(body)
	}()

	return ch, errCh
}

func (p *ollamaProvider) getBaseURL() string {
	if p.cfg.BaseURL != "" {
		return p.cfg.BaseURL
	}
	return "http://localhost:11434"
}
