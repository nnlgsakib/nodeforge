# Story 2.6: Speculative Execution within Nodes

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a system,
I want to run multiple attempts in parallel within a node and select the best result,
so that execution quality improves without user intervention.

## Acceptance Criteria

1. **Given** the node executor supports parallel execution paths
   **When** a node starts executing with speculative execution enabled (FR53)
   **Then** multiple LLM agents negotiate within the node (AI Swarm per Node)
   **And** parallel attempts run via goroutines with results collected via channels
   **And** the best result (based on acceptance criteria verification) wins and proceeds
   **And** failed attempts are logged for debugging but don't block progress

## Tasks / Subtasks

- [x] Task 1: Implement speculative execution in node executor (AC: 1)
  - [x] Subtask 1.1: Create `internal/engine/executor.go` with support for parallel execution paths
  - [x] Subtask 1.2: Add goroutines for multiple parallel attempts within a single node
  - [x] Subtask 1.3: Implement channel-based result collection from parallel attempts
  - [x] Subtask 1.4: Add best result selection logic based on acceptance criteria verification

- [x] Task 2: Implement AI Swarm per Node (AC: 1)
  - [x] Subtask 2.1: Create `internal/llm/swarm.go` for multiple LLM agents negotiation within a node
  - [x] Subtask 2.2: Integrate with LLM provider interface (`internal/llm/provider.go`) to spawn parallel LLM calls
  - [x] Subtask 2.3: Add result comparison logic to select the best output based on acceptance criteria
  - [x] Subtask 2.4: Log failed attempts for debugging without blocking progress (save to session logs)

- [x] Task 3: Integrate speculative execution with existing systems (AC: 1)
  - [x] Subtask 3.1: Update node execution flow to check if speculative execution is enabled per node type
  - [x] Subtask 3.2: Add configuration option for speculative execution (enable/disable, max parallel attempts)
  - [x] Subtask 3.3: Ensure acceptance criteria verification works correctly with speculative execution results
  - [x] Subtask 3.4: Send speculative execution progress via WebSocket: `{"type": "node_update", "nodeId": "node-1", "status": "running", "speculative": true, "attempts": 3}`

- [x] Task 4: Create/Update LLM provider interface for swarm support (AC: 1)
  - [x] Subtask 4.1: Define `LLMProvider` interface in `internal/llm/provider.go` with `Complete` and `Chat` methods returning channels
  - [x] Subtask 4.2: Implement provider clients (OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama) with channel-based streaming
  - [x] Subtask 4.3: Add race mode logic in `internal/llm/race.go` (fastest token wins, cancel slower)
  - [x] Subtask 4.4: Integrate token budget enforcer (`internal/llm/budget.go`) with speculative execution

## Dev Notes

### Architecture Patterns & Constraints

- **AI Swarm per Node (FR53)**: Multiple LLM agents negotiate within single node; speculative execution where best result wins (architecture.md lines 39, 247-249)
- **Goroutines + Channels**: Parallel attempts via Go goroutines, results collected via channels (architecture.md line 249)
- **LLM Provider Abstraction**: Use `internal/llm/provider.go` interface; providers return `<-chan string` for streaming tokens
- **Race Mode**: Fastest token wins, slower providers cancelled immediately (NFR-03: sub-200ms race wins)
- **Token Budget**: Pre-flight estimation completes in <10ms (NFR-05); enforce budget per speculative attempt
- **Acceptance Criteria Verification**: Each speculative result must pass node's acceptance criteria; only verified results are considered "best"

### Source Tree Components to Touch

| Component | Path | Action | Status |
|-----------|------|--------|--------|
| Executor | `internal/engine/executor.go` | NEW | Story 2.6 |
| Swarm | `internal/llm/swarm.go` | NEW | Story 2.6 |
| LLM Provider | `internal/llm/provider.go` | NEW | Story 2.6 |
| Race Mode | `internal/llm/race.go` | NEW | Story 2.6 |
| Token Budget | `internal/llm/budget.go` | NEW | Story 2.6 |
| Context Graph | `internal/context/graph.go` | NEW | Story 2.5 (prerequisite) |
| WebSocket Hub | `main.go` (Gin `/ws`) | UPDATE | Add speculative execution status messages |

### Testing Standards Summary

- **Go**: Ginkgo + Testify for backend (`internal/engine/executor_test.go`, `internal/llm/swarm_test.go`)
- **Test Parallel Execution**: Verify multiple goroutines spawn, results collected via channel, no race conditions
- **Test Best Result Selection**: Mock LLM responses, verify acceptance criteria filters results correctly
- **Test Failed Attempts**: Ensure failed attempts are logged but don't block progress
- **NFR Targets**:
  - Sub-200ms provider race wins (NFR-03)
  - <10ms token budget pre-flight estimation (NFR-05)
  - 5000+ WebSocket connections with <50ms latency (NFR-01)

### Concurrency Guidelines

- Use `sync.WaitGroup` for goroutine synchronization
- Use buffered channels to prevent goroutine leaks
- Cancel all parallel attempts when best result is selected (context cancellation)
- Protect shared state with mutexes or channels (no shared mutable state)
- Run `go test -race` to detect race conditions

## Project Structure Notes

### Alignment with Unified Project Structure

- ✅ Go packages follow `snake_case`: `internal/engine/`, `internal/llm/`, `internal/context/`
- ✅ Go functions use `camelCase`: `executeNode(ctx)`, `runSpeculative(node)`
- ✅ Go structs use `PascalCase`: `type Executor struct`, `type Swarm struct`
- ✅ JSON fields use `camelCase`: `{"nodeId": "...", "speculative": true}`
- ✅ API endpoints use `snake_case`: `/api/v1/sessions`, `/api/v1/config`
- ✅ Gin single framework for REST + WebSocket (not Chi)
- ✅ `embed.FS` serves `frontend/dist/` from Go binary

### Detected Conflicts or Variances

- **None detected**: All patterns align with architecture.md and epics.md conventions.
- **Note**: Stories 2.1-2.4 are in `ready-for-dev` status; Story 2.5 (Smart Context Engine) is a prerequisite for speculative execution (BadgerDB knowledge graph needed for context assembly).

## References

- [Epic 2 Story 2.6: Speculative Execution within Nodes](_bmad-output/planning-artifacts/epics.md#story26-speculative-execution-within-nodes) — Story context, FR53 coverage
- [PRD: FR53 Speculative Execution](_bmad-output/planning-artifacts/prd.md#execution--performance-capabilities) — FR53 requirement details
- [Architecture: AI Swarm per Node](_bmad-output/planning-artifacts/architecture.md#llm-integration-architecture) — Design for multiple LLM agents per node
- [Architecture: Goroutines + Channels](_bmad-output/planning-artifacts/architecture.md#llm-integration-architecture) — Parallel execution implementation pattern
- [Architecture: LLM Provider Interface](_bmad-output/planning-artifacts/architecture.md#llm-integration-architecture) — `LLMProvider` interface definition
- [NFR-03: LLM Race Mode <200ms](_bmad-output/planning-artifacts/prd.md#performance) — Race mode performance target
- [NFR-05: Token Budget <10ms](_bmad-output/planning-artifacts/prd.md#performance) — Budget estimation performance target
- [UX Design: AI Swarm per Node](_bmad-output/planning-artifacts/ux-design-specification.md#novel-ux-patterns) — "AI Swarm per Node" as nice-to-have capability

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

- **Swarm Implementation** (`internal/llm/swarm.go`): Created `Swarm` struct that orchestrates multiple parallel LLM attempts within a single node. Uses goroutines with `sync.WaitGroup` and buffered channels for result collection. Implements `SwarmResult` scoring based on acceptance criteria verification. Failed attempts are logged via `log.Printf` but don't block progress. `SwarmConfig` provides enable/disable and max parallel attempts configuration.
- **Executor Integration** (`internal/engine/executor.go`): Added `swarm` and `swarmConfig` fields to `Executor`. Added `SetSwarm()` method for configuration. Updated `executeNode()` to check if speculative execution is enabled and route to swarm or single execution. Added `broadcastSpeculativeStart()` for WebSocket progress messages matching the specified format.
- **LLM Provider Interface**: Already existed from previous stories with `Complete` and `Chat` methods returning `<-chan string`. Race mode (`internal/llm/race.go`) and budget enforcement (`internal/llm/budget.go`) already implemented. Swarm integrates with `BudgetEnforcer` for token tracking per attempt.
- **Tests**: Added `internal/llm/swarm_test.go` with 14 tests covering single/speculative execution, best result selection, failures, context cancellation, budget tracking, and scoring. Updated `internal/engine/executor_test.go` with 6 new tests for SetSwarm, speculative broadcast, and simulated execution. All tests pass. `go build ./...` clean.

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/llm/swarm.go` | NEW | AI Swarm implementation with speculative execution, scoring, and best result selection |
| `internal/llm/swarm_test.go` | NEW | Unit tests for Swarm (14 tests) |
| `internal/engine/executor.go` | UPDATE | Added speculative execution integration, SetSwarm, broadcastSpeculativeStart |
| `internal/engine/executor_test.go` | UPDATE | Added 6 new tests for speculative execution features |

### Review Findings

**Decision-Needed (require user input):**

- [x] [Review][Decision] Single LLM provider instead of multiple agents — RESOLVED: Keep single provider, multiple attempts. "Multiple LLM agents" means parallel goroutines (speculative attempts), not necessarily different provider instances. Swarm's purpose is quality comparison via acceptance criteria, not provider racing.
- [x] [Review][Decision] No first-token race mode implementation — RESOLVED: Keep current design. NFR-03 Race Mode is a separate feature handled by `llm.Race()`. Swarm is about quality (acceptance criteria scoring), not speed. Racing on first token is incompatible with needing full output to score against AC.

**Patch (fixable without human input):**

- [x] [Review][Patch] No cancellation of parallel attempts when best result is selected [internal/llm/swarm.go:Execute()] — FIXED: added early cancellation via `context.WithCancel` and `sync.Once` — when a passing result is found, remaining goroutines are cancelled.
- [x] [Review][Patch] Non-verified results can be selected as best [internal/llm/swarm.go:selectBestResult()] — FIXED: `selectBestResult()` now sorts by `PassedAC` first (true > false), then by score descending.
- [x] [Review][Patch] Missing pre-flight token budget estimation [internal/llm/swarm.go:Execute()] — FIXED: added pre-flight check using `budget.CheckBudget()` before launching goroutines.
- [x] [Review][Patch] Budget enforcement is optional [internal/llm/swarm.go:runAttempt(), NewSwarm()] — FIXED: added warning log when budget is nil; `NewSwarm` allows nil but warns at execution time.

**Deferred (pre-existing or out of scope):**

- [x] [Review][Defer] (none)
