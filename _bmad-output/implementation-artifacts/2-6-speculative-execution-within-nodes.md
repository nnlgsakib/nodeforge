# Story 2.6: Speculative Execution within Nodes

Status: ready-for-dev

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

- [ ] Task 1: Implement speculative execution in node executor (AC: 1)
  - [ ] Subtask 1.1: Create `internal/engine/executor.go` with support for parallel execution paths
  - [ ] Subtask 1.2: Add goroutines for multiple parallel attempts within a single node
  - [ ] Subtask 1.3: Implement channel-based result collection from parallel attempts
  - [ ] Subtask 1.4: Add best result selection logic based on acceptance criteria verification

- [ ] Task 2: Implement AI Swarm per Node (AC: 1)
  - [ ] Subtask 2.1: Create `internal/llm/swarm.go` for multiple LLM agents negotiation within a node
  - [ ] Subtask 2.2: Integrate with LLM provider interface (`internal/llm/provider.go`) to spawn parallel LLM calls
  - [ ] Subtask 2.3: Add result comparison logic to select the best output based on acceptance criteria
  - [ ] Subtask 2.4: Log failed attempts for debugging without blocking progress (save to session logs)

- [ ] Task 3: Integrate speculative execution with existing systems (AC: 1)
  - [ ] Subtask 3.1: Update node execution flow to check if speculative execution is enabled per node type
  - [ ] Subtask 3.2: Add configuration option for speculative execution (enable/disable, max parallel attempts)
  - [ ] Subtask 3.3: Ensure acceptance criteria verification works correctly with speculative execution results
  - [ ] Subtask 3.4: Send speculative execution progress via WebSocket: `{"type": "node_update", "nodeId": "node-1", "status": "running", "speculative": true, "attempts": 3}`

- [ ] Task 4: Create/Update LLM provider interface for swarm support (AC: 1)
  - [ ] Subtask 4.1: Define `LLMProvider` interface in `internal/llm/provider.go` with `Complete` and `Chat` methods returning channels
  - [ ] Subtask 4.2: Implement provider clients (OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama) with channel-based streaming
  - [ ] Subtask 4.3: Add race mode logic in `internal/llm/race.go` (fastest token wins, cancel slower)
  - [ ] Subtask 4.4: Integrate token budget enforcer (`internal/llm/budget.go`) with speculative execution

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

### File List
