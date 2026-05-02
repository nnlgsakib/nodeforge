package llm

import "context"

type openRouterProvider struct {
	cfg *Config
}

func (p *openRouterProvider) Name() string { return "openrouter" }

func (p *openRouterProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// TODO: implement OpenRouter API
	return nil, nil
}

func (p *openRouterProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan string, <-chan error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	close(ch)
	close(errCh)
	return ch, errCh
}
