package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type openAIProvider struct {
	cfg *Config
}

func (p *openAIProvider) Name() string {
	return "openai"
}

func (p *openAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"model":    p.cfg.Model,
		"messages": req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
		"stream":    req.Stream,
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}

func (p *openAIProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan string, <-chan error) {
	ch := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		req.Stream = true
		payload, err := json.Marshal(map[string]interface{}{
			"model":       p.cfg.Model,
			"messages":    req.Messages,
			"temperature": req.Temperature,
			"max_tokens":  req.MaxTokens,
			"stream":      true,
		})
		if err != nil {
			errCh <- err
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.getBaseURL()+"/chat/completions", strings.NewReader(string(payload)))
		if err != nil {
			errCh <- err
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if p.cfg.APIKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
		}

		client := &http.Client{Timeout: p.cfg.Timeout}
		resp, err := client.Do(httpReq)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()

		// Simplified streaming: read full response (full streaming implementation would parse SSE)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			errCh <- err
			return
		}

		ch <- string(body)
	}()

	return ch, errCh
}

func (p *openAIProvider) getBaseURL() string {
	if p.cfg.BaseURL != "" {
		return p.cfg.BaseURL
	}
	return "https://api.openai.com/v1"
}
