package llm

import "context"

type anthropicProvider struct {
	cfg *Config
}

func (p *anthropicProvider) Name() string { return "anthropic" }

func (p *anthropicProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// TODO: implement Anthropic API
	return nil, nil
}

func (p *anthropicProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan string, <-chan error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	close(ch)
	close(errCh)
	return ch, errCh
}
