# Story 2.4: Prompt Optimization & Token Budget

Status: ready-for-dev

## Story

As a system,
I want to auto-optimize prompts based on execution feedback and enforce token budgets,
so that costs are controlled and prompts improve over time.

## Acceptance Criteria

1. **[AC1]** Given the prompt optimization system is implemented, when a node prepares to call an LLM provider, then the system optimizes prompts automatically based on past execution feedback (FR14)

2. **[AC2]** Given the token budget enforcer is implemented, when a node prepares to call an LLM provider, then token budget pre-flight estimation completes in <10ms before the LLM call is dispatched (FR15, NFR-05)

3. **[AC3]** Given a token budget is configured, when an LLM request would exceed the budget, then the request is rejected before any LLM call is made

4. **[AC4]** Given budget tracking is implemented, when LLM calls complete, then budget usage is tracked and reported back to the user via the API and WebSocket

## Tasks / Subtasks

- [ ] Task 1: Implement Token Budget Enforcer (AC: 2, 3, 4)
  - [ ] Subtask 1.1: Create `internal/llm/budget.go` with `BudgetEnforcer` struct and `TokenBudget` config struct
  - [ ] Subtask 1.2: Implement `EstimateTokens(text string) int` function with <10ms performance for typical prompts (<10k chars)
  - [ ] Subtask 1.3: Implement `CheckBudget(ctx, estimatedTokens) error` that rejects requests exceeding budget
  - [ ] Subtask 1.4: Implement `TrackUsage(actualTokens int)` to record token consumption
  - [ ] Subtask 1.5: Add budget configuration keys to `cmd/nforge/config.go`: `llm.token-budget`, `llm.token-budget-per-request`
  - [ ] Subtask 1.6: Write unit tests for budget enforcer (estimation accuracy, budget enforcement, performance <10ms)

- [ ] Task 2: Implement Prompt Optimization System (AC: 1)
  - [ ] Subtask 2.1: Create `internal/llm/optimizer.go` with `PromptOptimizer` struct
  - [ ] Subtask 2.2: Implement feedback storage using BadgerDB to persist prompt execution results (success/failure, token usage, quality scores)
  - [ ] Subtask 2.3: Implement `OptimizePrompt(ctx, originalPrompt string, nodeType string) string` that enhances prompts based on historical feedback
  - [ ] Subtask 2.4: Add prompt templates for each node type (Goal, Spec, Plan, Implement, Test, Review) that improve over time
  - [ ] Subtask 2.5: Write unit tests for prompt optimizer (optimization produces valid prompts, feedback storage/retrieval)

- [ ] Task 3: Integrate with LLM Provider Interface (AC: 1, 2, 3, 4)
  - [ ] Subtask 3.1: Modify `internal/llm/provider.go` to call budget enforcer before LLM invocation
  - [ ] Subtask 3.2: Add prompt optimization hook in provider call path (optimize prompt before sending to LLM)
  - [ ] Subtask 3.3: Add budget tracking after LLM response (track actual tokens used)
  - [ ] Subtask 3.4: Emit WebSocket events for budget status updates (`budget_update` message type)
  - [ ] Subtask 3.5: Write integration tests for the full flow (optimize → budget check → LLM call → track usage)

- [ ] Task 4: Smart Context Integration (AC: 1, 2)
  - [ ] Subtask 4.1: Integrate with `internal/context/assembler.go` to include knowledge graph context in optimized prompts
  - [ ] Subtask 4.2: Ensure context assembly completes in <100ms (NFR-04) when building optimized prompts
  - [ ] Subtask 4.3: Write tests verifying context integration doesn't break budget timing (<10ms pre-flight)

## Dev Notes

### Architecture Compliance

**Files to CREATE (new for this story):**
- `internal/llm/budget.go` — Token budget enforcer with <10ms estimation
- `internal/llm/optimizer.go` — Prompt optimization using execution feedback
- `internal/llm/budget_test.go` — Unit tests for budget enforcer
- `internal/llm/optimizer_test.go` — Unit tests for prompt optimizer

**Files to MODIFY (existing, must preserve current behavior):**
- `internal/llm/provider.go` — Add budget check and prompt optimization hooks (file doesn't exist yet; will be created in Epic 2, Story 2.2 — this story should define the interface)
- `cmd/nforge/config.go` — Add budget config keys to `supportedKeys` map (lines 13-19)

**Key Architecture Patterns (from architecture.md):**
- LLM Provider interface: `type LLMProvider interface { Complete(ctx, prompt) (<-chan string, error) }`
- Budget enforcer must be fast: <10ms pre-flight estimation (NFR-05)
- Token estimation: use simple heuristic (~4 chars per token) for speed; avoid API calls during estimation
- BadgerDB for feedback storage (knowledge graph in `internal/context/`)
- Integration with Smart Context Engine for prompt enhancement

### Technical Stack & Versions

| Component | Technology | Version |
|-----------|-------------|---------|
| LLM Provider Abstraction | `internal/llm/` package | Go 1.24+ |
| Token Budget Storage | In-memory + config persistence | N/A |
| Prompt Feedback Store | BadgerDB (dgraph-io/badger/v4) | v4 latest |
| Token Estimation | Heuristic (~4 chars/token) | N/A |
| Config Management | Viper (via Cobra) | latest |

### Code Organization (from architecture.md)

```
internal/llm/
├── provider.go     # LLMProvider interface (created in story 2.2, stub here)
├── race.go         # Race mode (created in story 2.2)
├── openai.go       # OpenAI client (created in story 2.2)
├── anthropic.go   # Anthropic client (created in story 2.2)
├── budget.go       # Token budget enforcer ← NEW (this story)
├── optimizer.go    # Prompt optimization ← NEW (this story)
├── budget_test.go  # ← NEW
└── optimizer_test.go # ← NEW
```

### Performance Requirements

- **NFR-05**: Token budget pre-flight estimation <10ms (measure with `time.Now()` benchmarking)
- **NFR-04**: Smart context assembly <100ms when building optimized prompts
- **Token estimation heuristic**: ~4 characters per token for English text (fast approximation)

### Testing Standards

- **Framework**: Go standard `testing` package + `testify` for assertions
- **Coverage**: All public functions in `budget.go` and `optimizer.go`
- **Performance tests**: Benchmark tests for <10ms budget estimation
- **Integration tests**: Full flow from prompt optimization → budget check → usage tracking
- **File location**: Co-located `*_test.go` files in `internal/llm/`

### API & WebSocket Integration

**WebSocket message for budget updates (add to architecture.md patterns):**
```json
{"type": "budget_update", "budgetRemaining": 8500, "budgetTotal": 10000, "lastRequestTokens": 500}
```

**Config keys to add in `cmd/nforge/config.go`:**
- `llm.token-budget` — Total token budget per session (default: 100000)
- `llm.token-budget-per-request` — Max tokens per individual request (default: 4096)

### Security & Error Handling

- Budget enforcement MUST happen before any LLM API call (no exceptions)
- Failed budget checks return sentinel error: `ErrTokenBudgetExceeded`
- Prompt optimization failures MUST fall back to original prompt (never block execution)
- All errors logged with context (node ID, estimated tokens, budget remaining)

### Cross-Story Dependencies

- **Story 2.2 (LLM Provider Abstraction)**: Creates `internal/llm/provider.go` with `LLMProvider` interface. This story defines the budget/optimizer integration points.
- **Story 2.5 (Smart Context Engine)**: Creates `internal/context/` with BadgerDB knowledge graph. This story integrates context assembly into prompt optimization.
- **Story 1.5 (Configuration)**: Established `cmd/nforge/config.go` pattern for adding config keys.

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

## References

- [Source: epics.md#Story2.4] Story definition and acceptance criteria
- [Source: architecture.md#LLM Integration Architecture] Provider interface, budget enforcer, Smart Context integration
- [Source: architecture.md#API & Communication Patterns] WebSocket message formats
- [Source: architecture.md#Naming Patterns] Go naming conventions (snake_case packages, camelCase functions)
- [Source: PRD.md#FR14-FR15] Functional requirements for prompt optimization and token budgets
- [Source: PRD.md#NFR-05] Non-functional requirement for <10ms pre-flight estimation
- [Source: cmd/nforge/config.go] Config key registration pattern (supportedKeys map)
