package llm

import "context"

type deepSeekProvider struct {
	cfg *Config
}

func (p *deepSeekProvider) Name() string { return "deepseek" }

func (p *deepSeekProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// TODO: implement DeepSeek API
	return nil, nil
}

func (p *deepSeekProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan string, <-chan error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	close(ch)
	close(errCh)
	return ch, errCh
}
