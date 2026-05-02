# Story 2.2: LLM Provider Abstraction & Race Mode

Status: ready-for-dev

## Story

As a user,
I want to configure multiple LLM providers with race mode and auto-fallback,
so that I get the fastest/cheapest response with reliable connectivity.

## Acceptance Criteria

1. **Given** the LLM provider interface is defined (`type LLMProvider interface`)
   **When** the system initializes the LLM subsystem
   **Then** the `LLMProvider` interface is implemented with `Complete(ctx, prompt) (<-chan string, error)` and `Chat(ctx, messages) (<-chan string, error)` methods
   **And** the interface supports streaming responses via Go channels

2. **Given** the user configures providers via `nforge config set llm.openai-key <key>` (FR10)
   **When** the system reads configuration from `~/.nforge/config.yaml`
   **Then** supported keys include: `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `llm.default-model`
   **And** `cmd/nforge/config.go` already supports these keys in `supportedKeys` map

3. **Given** providers are configured
   **When** the system requests LLM completion
   **Then** the system supports: OpenAI (github.com/openai/openai-go), Anthropic SDK, DeepSeek, OpenRouter, Ollama local (via Ollama Go client)
   **And** each provider implements the `LLMProvider` interface (FR16)

4. **Given** race mode is enabled (default behavior)
   **When** a completion request is made
   **Then** multiple providers run simultaneously via goroutines
   **And** the fastest first token wins — slower providers are cancelled via context cancellation (FR11, NFR-03: sub-200ms wins)
   **And** the winning response is collected and returned; losing providers are cleanly cancelled

5. **Given** a provider returns rate limit or error
   **When** the primary provider fails
   **Then** auto-fallback triggers through the chain: Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter (FR12, NFR-19: 99.9% uptime)
   **And** fallback uses semantic matching: rate limit → cheaper/similar model
   **And** fallback decisions are logged for debugging

6. **Given** a session is starting
   **When** the WebSocket connection initializes or `nforge run` starts
   **Then** provider connectivity is pre-fetched asynchronously for zero-wait checks (FR56)
   **And** provider status (online/offline, latency) is reported to the UI via WebSocket message `{"type": "provider_status", "providers": {...}}`

## Tasks / Subtasks

- [ ] Task 1: Define LLMProvider Interface (AC: #1)
  - [ ] Create `internal/llm/provider.go` with `LLMProvider` interface
  - [ ] Define `Message` struct for chat conversations
  - [ ] Define `ProviderConfig` struct with timeout, model, baseURL fields

- [ ] Task 2: Implement Provider Clients (AC: #3)
  - [ ] Create `internal/llm/openai.go` — OpenAI client using `github.com/openai/openai-go`
  - [ ] Create `internal/llm/anthropic.go` — Anthropic client (when SDK available)
  - [ ] Create `internal/llm/deepseek.go` — DeepSeek client (OpenAI-compatible API)
  - [ ] Create `internal/llm/openrouter.go` — OpenRouter client (OpenAI-compatible API)
  - [ ] Create `internal/llm/ollama.go` — Ollama local client (Ollama Go client)

- [ ] Task 3: Implement Race Mode (AC: #4, NFR-03)
  - [ ] Create `internal/llm/race.go` with `RaceMode` struct
  - [ ] Implement `Complete()` method: launch goroutines per provider, collect fastest token
  - [ ] Implement context cancellation for losing providers
  - [ ] Add metrics: track which provider wins, latency per provider

- [ ] Task 4: Implement Auto-Fallback Chain (AC: #5, NFR-19)
  - [ ] Create `internal/llm/fallback.go` with fallback chain logic
  - [ ] Implement semantic matching: rate limit → cheaper model
  - [ ] Add fallback logging for debugging

- [ ] Task 5: Provider Status Pre-fetch (AC: #6, FR56)
  - [ ] Implement connectivity check in session init
  - [ ] Send `provider_status` WebSocket message to UI
  - [ ] Cache status results to avoid repeated checks

- [ ] Task 6: Integration & Config Integration (AC: #2)
  - [ ] Wire providers into `internal/llm/` package initialization
  - [ ] Read provider config from `llm.*` keys in `~/.nforge/config.yaml`
  - [ ] Add `llm.deepseek-key` and `llm.openrouter-key` to `cmd/nforge/config.go` supportedKeys

- [ ] Task 7: Unit Tests
  - [ ] Test LLMProvider interface compliance for each provider
  - [ ] Test race mode: verify fastest wins, slowest cancelled
  - [ ] Test fallback chain: verify order and semantic matching
  - [ ] Test provider status pre-fetch with mocked providers

## Dev Notes

### Architecture Patterns and Constraints

**Provider Interface (from `architecture.md` lines 216-258):**
```go
type LLMProvider interface {
    Complete(ctx context.Context, prompt string) (<-chan string, error)
    Chat(ctx context.Context, messages []Message) (<-chan string, error)
}
```

**Race Mode Pattern (from `architecture.md` lines 232-237):**
```go
type RaceMode struct {
    providers []LLMProvider
    timeout   time.Duration
}
// Fastest token wins, cancel losers
func (r *RaceMode) Complete(ctx context.Context, prompt string) (string, error)
```

**Key Constraints:**
- Use Go channels (`<-chan string`) for streaming tokens — NOT HTTP response bodies
- Race mode cancellation via `context.WithCancel()` — losing providers must stop cleanly
- Fallback chain: Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter
- BadgerDB (`internal/context/`) is NOT used here — that's for Smart Context Engine (Story 2.5)
- Gin WebSocket hub (`internal/canvas/` or `main.go`) sends `provider_status` messages

### Source Tree Components to Touch

| File | Action | Purpose |
|------|--------|---------|
| `internal/llm/provider.go` | NEW | LLMProvider interface definition |
| `internal/llm/race.go` | NEW | Race mode implementation |
| `internal/llm/openai.go` | NEW | OpenAI client |
| `internal/llm/anthropic.go` | NEW | Anthropic client |
| `internal/llm/deepseek.go` | NEW | DeepSeek client |
| `internal/llm/openrouter.go` | NEW | OpenRouter client |
| `internal/llm/ollama.go` | NEW | Ollama local client |
| `internal/llm/fallback.go` | NEW | Auto-fallback chain |
| `internal/llm/budget.go` | NEW (deferred to 2.4) | Token budget enforcer |
| `cmd/nforge/config.go` | UPDATE | Add `llm.deepseek-key`, `llm.openrouter-key` to supportedKeys |
| `internal/llm/llm_test.go` | NEW | Unit tests |

### Testing Standards Summary

**Framework:** Ginkgo + Testify (from `architecture.md` line 136)

**Test Patterns (from `architecture.md` lines 586-596):**
- Sentinel errors: `ErrLLMRateLimit`, `ErrLLMTimeout`
- Structured error responses via Gin handlers (for API endpoints)
- Co-locate tests: `llm_test.go` alongside source files

**Required Test Cases:**
1. Each provider correctly implements `LLMProvider` interface
2. Race mode: fastest provider wins, slowest receives cancellation
3. Fallback: rate limit on OpenAI → falls back to Anthropic
4. Provider status pre-fetch: returns correct online/offline status
5. Config integration: reads keys from `~/.nforge/config.yaml`

### Naming Conventions (from `architecture.md` lines 429-446)

**Go Code (`internal/llm/`):**
- **Packages**: `llm` (single word, snake_case not needed)
- **Functions**: `camelCase` — `raceProviders(prompt)`, `completeWithFallback(ctx, prompt)`
- **Structs**: `PascalCase` — `RaceMode`, `OpenAIClient`, `FallbackChain`
- **Variables**: `camelCase` — `providers`, `winningResp`
- **Constants**: `UPPER_SNAKE` — `DefaultTimeout`, `MaxRetries`

**JSON Fields (API/WebSocket messages):**
- `camelCase` — `{"providerStatus": {...}, "firstTokenLatency": 150}`

## Project Structure Notes

### Alignment with Unified Project Structure

**Expected structure (from `architecture.md` lines 656-673):**
```
internal/
├── llm/
│   ├── provider.go     # LLMProvider interface
│   ├── race.go         # Race mode (goroutines, fastest wins)
│   ├── openai.go       # OpenAI client
│   ├── anthropic.go   # Anthropic client
│   ├── deepseek.go     # DeepSeek client
│   ├── openrouter.go   # OpenRouter client
│   ├── ollama.go       # Ollama local client
│   ├── fallback.go     # Semantic fallback chain
│   ├── budget.go       # Token budget enforcer (story 2.4)
│   └── llm_test.go
```

### Detected Conflicts or Variances

- **None detected** — this is the first story in Epic 2, and `internal/llm/` directory does not yet exist
- Config keys in `cmd/nforge/config.go` already support `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `llm.default-model` — need to add `llm.deepseek-key` and `llm.openrouter-key`

## References

- [Epic 2 Context: architecture.md#LLM Integration Architecture](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/architecture.md#216-258)
- [Provider Interface Definition: architecture.md#LLM Integration Architecture](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/architecture.md#226-238)
- [Race Mode Pattern: architecture.md#LLM Integration Architecture](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/architecture.md#246-249)
- [Smart Context Engine (separate, story 2.5): architecture.md#Smart Context Engine](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/architecture.md#240-245)
- [Naming Conventions: architecture.md#Naming Patterns](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/architecture.md#429-446)
- [Project Structure: architecture.md#Complete Project Directory Structure](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/architecture.md#639-771)
- [FR10: epics.md#Story 2.2](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#423-436)
- [FR11: epics.md#Story 2.2](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#423-436)
- [FR12: epics.md#Story 2.2](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#423-436)
- [FR16: epics.md#Story 2.2](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#423-436)
- [FR56: epics.md#Story 2.2](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#423-436)
- [NFR-03: epics.md#NonFunctional Requirements](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#89)
- [NFR-19: epics.md#NonFunctional Requirements](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/epics.md#105)
- [Existing Config: cmd/nforge/config.go](D:/projects/go/research26/nfv2/cmd/nforge/config.go)
- [PRD LLM Section: prd.md#LLM Integration Capabilities](D:/projects/go/research26/nfv2/_bmad-output/planning-artifacts/prd.md#222-231)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
