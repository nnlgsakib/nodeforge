# Story 2.8: Headless CLI Execution

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to run headless execution with `nforge run <spec-file>`,
So that I can execute graphs identically in terminal without a browser.

## Acceptance Criteria

1. **Given** the CLI includes the `run` subcommand
   **When** the user runs `nforge run <spec-file>` in terminal
   **Then** the same graph execution logic runs as the browser UI — identical nodes, identical results (FR22, FR30)
   **And** the spec file is parsed to either extract the goal (auto-generating the graph) or load a pre-defined graph structure
   **And** execution uses `internal/engine/executor.go` (same as browser UI, no WebSocket needed for headless)

2. **Given** the `graph` subcommand includes `viz` subcommand
   **When** the user runs `nforge graph viz [spec-file]`
   **Then** an ASCII art graph is displayed in terminal (FR27)
   **And** nodes are shown as boxes with status colors (green=complete, red=failed, yellow=running)
   **And** edges are shown as ASCII arrows (→)

3. **Given** headless mode is executing a graph
   **When** all nodes complete successfully (all green)
   **Then** `nforge run` exits with code 0 (CI/CD integration)
   **And** if any node fails (red), exits with code 1
   **And** exit code 2 for usage errors (missing spec file, invalid format)

4. **Given** the CLI and UI are compared
   **When** the same spec file is run via `nforge run` and via browser UI
   **Then** identical nodes execute with identical results (FR30 CLI/UI parity)
   **And** LLM provider calls, token usage, and output artifacts are identical

## Tasks / Subtasks

- [ ] Task 1: Implement `nforge run` subcommand (AC: 1, 3)
  - [ ] Subtask 1.1: Update `cmd/nforge/run.go` to parse spec file argument (Cobra `Args: cobra.ExactArgs(1)`)
  - [ ] Subtask 1.2: Define spec file format (YAML) with two modes:
        - Mode 1: Goal-only (`goal: "Convert JS→Go"`) → auto-generate graph via `internal/engine/graph.go` (same as Story 2.1 chat interface)
        - Mode 2: Pre-defined graph (`nodes: [...]`, `edges: [...]`) → load graph directly
  - [ ] Subtask 1.3: Implement spec file parser in `internal/engine/spec.go` (NEW) — parse YAML with `goccy/go-yaml`, validate structure
  - [ ] Subtask 1.4: Wire `nforge run` to use `internal/engine/executor.go` for headless execution (same executor as browser UI, no WS)
  - [ ] Subtask 1.5: Add exit code logic: 0=all green, 1=red node, 2=usage error (os.Exit codes)
  - [ ] Subtask 1.6: Add `--ascii` flag to `nforge run` to display ASCII graph during execution (reuse Task 2 renderer)

- [ ] Task 2: Implement `nforge graph viz` subcommand (AC: 2)
  - [ ] Subtask 2.1: Update `cmd/nforge/graph.go` to add `viz` subcommand with optional spec file argument
  - [ ] Subtask 2.2: Implement ASCII graph renderer in `internal/engine/ascii.go` (NEW)
  - [ ] Subtask 2.3: Render nodes as `[Type: Label]` with status suffix: (G)=green, (Y)=yellow, (R)=red
  - [ ] Subtask 2.4: Render edges as ` → ` arrows, layout left-to-right (topological order)

- [ ] Task 3: Ensure CLI/UI parity (AC: 4)
  - [ ] Subtask 3.1: Verify `internal/engine/executor.go` is called by both `nforge run` and browser UI WebSocket handler
  - [ ] Subtask 3.2: Ensure LLM provider calls (`internal/llm/`) are identical regardless of execution mode (same `LLMProvider` interface)
  - [ ] Subtask 3.3: Add CI/CD example to docs: `nforge run spec.yaml && echo "SUCCESS" || echo "FAILED"`

- [ ] Task 4: Add testing for headless execution (AC: 1, 3)
  - [ ] Subtask 4.1: Test `nforge run` with goal-only spec file (auto-generates graph) using `testify`
  - [ ] Subtask 4.2: Test `nforge run` with pre-defined graph spec file
  - [ ] Subtask 4.3: Test exit codes: 0 (success), 1 (failure), 2 (usage error) via `exec.Command("nforge", "run", ...).CombinedOutput()`
  - [ ] Subtask 4.4: Test ASCII graph output matches expected format in `internal/engine/ascii_test.go`

## Dev Notes

### Architecture Patterns & Constraints

- **Backend-First Architecture**: All execution logic in Go backend (`internal/engine/`, `internal/llm/`). Headless mode uses same executor as browser UI — zero code duplication.
- **Cobra CLI**: v1.10.2 (go.mod). `run` subcommand takes 1 argument (spec file path). `graph viz` takes optional spec file argument.
- **Spec File Format**: YAML (matches config.yaml format, uses `goccy/go-yaml` already in go.mod as indirect dependency; add as direct dependency). Two modes:
  - **Goal Mode**: Minimal, contains `goal: <string>`. Triggers auto-graph generation via `internal/engine/graph.go` (same as Story 2.1 chat interface).
  - **Graph Mode**: Contains `nodes: []` and `edges: []` defining complete graph structure. Bypasses auto-generation.
  - Example goal-mode spec (`spec.yaml`):
    ```yaml
    goal: "Convert JS→Go project"
    ```
  - Example graph-mode spec:
    ```yaml
    nodes:
      - id: goal-1
        type: Goal
        label: "Convert JS→Go"
      - id: spec-1
        type: Spec
        label: "Generate Spec"
    edges:
      - source: goal-1
        target: spec-1
    ```
- **Execution Engine**: `internal/engine/executor.go` — sequential/parallel execution with retry loops until acceptance criteria met. Headless mode calls executor directly (no WebSocket, no frontend). Executor must not require `context.Context` with WS; support both modes.
- **Exit Codes**: 0=success (all nodes green), 1=failure (any node red), 2=usage error (missing file, invalid YAML, unsupported spec mode). Use `os.Exit(code)` in `run.go`.
- **ASCII Graph**: Rendered left-to-right in topological order. Nodes: `[Type: Label] (G|Y|R)`. Edges: ` → `. Example:
  ```
  [Goal: Convert JS→Go] (G) → [Spec: Generate Spec] (Y) → [Plan: Create Plan] (G)
  ```
- **Dependencies**: `goccy/go-yaml` (YAML parsing) — add to go.mod direct dependencies. Cobra v1.10.2 already present.

### Source Tree Components to Touch

| Component | Path | Action | Status |
|-----------|------|--------|--------|
| Run Cmd | `cmd/nforge/run.go` | UPDATE | Placeholder exists ("story 1.8" comment is wrong — should be 2.8) |
| Graph Cmd | `cmd/nforge/graph.go` | UPDATE | Placeholder exists, add `viz` subcommand |
| Spec Parser | `internal/engine/spec.go` | NEW | Parse YAML spec files, validate structure |
| ASCII Renderer | `internal/engine/ascii.go` | NEW | Render graph as ASCII art |
| Executor | `internal/engine/executor.go` | UPDATE | Ensure headless-compatible (no WS dependencies) |
| Graph Engine | `internal/engine/graph.go` | UPDATE | Add goal-to-graph generation for headless mode |
| LLM Provider | `internal/llm/provider.go` | REFERENCE | Same provider interface for headless/UI |
| Session Mgr | `internal/session/manager.go` | REFERENCE | Headless uses same session management |

### Testing Standards Summary

- **Go**: Testify + Ginkgo for backend (`internal/engine/*_test.go`, `cmd/nforge/*_test.go`)
- **Exit Code Tests**: Use `exec.Command` to run `nforge run` binary, check `wait().ExitCode()` for 0, 1, 2
- **Spec Parser Tests**: Valid/invalid YAML, goal-mode vs graph-mode, missing fields
- **ASCII Renderer Tests**: Compare output string against expected ASCII format in `ascii_test.go`
- **Parity Tests**: Run same spec via headless executor with mock LLM, verify node execution order and results

### Project Structure Notes

- **Alignment with Unified Project Structure**:
  - ✅ Go packages follow `snake_case`: `internal/engine/`, `internal/llm/`
  - ✅ CLI commands in `cmd/nforge/` with Cobra (root.go, run.go, graph.go)
  - ✅ Spec files are YAML (matches config.yaml format, uses same YAML library)
  - ✅ Exit codes follow CLI conventions (0=success, 1=error, 2=usage)
- **Detected Conflicts or Variances**:
  - `run.go` placeholder comment says "story 1.8" — incorrect, should reference story 2.8
  - `goccy/go-yaml` is indirect dependency — need to add as direct dependency in go.mod

### References

- [Epic 2 Overview: Graph Execution Engine & LLM Integration](_bmad-output/planning-artifacts/epics.md#epic-2-graph-execution-engine--llm-integration) — Story 2.8 context, FR22, FR30
- [PRD: Headless CLI Execution](_bmad-output/planning-artifacts/prd.md#fr22) — FR22: `nforge run <spec-file>`
- [PRD: ASCII Art Graph](_bmad-output/planning-artifacts/prd.md#fr27) — FR27: `nforge graph viz`
- [PRD: CLI/UI Parity](_bmad-output/planning-artifacts/prd.md#fr30) — FR30: same graphs execute identically
- [Architecture: CLI Framework](_bmad-output/planning-artifacts/architecture.md#api--communication-patterns) — Cobra subcommands, `run` is defined
- [Architecture: Graph Engine](_bmad-output/planning-artifacts/architecture.md#core-graph) — `internal/engine/` executor, spec generation
- [Architecture: LLM Integration](_bmad-output/planning-artifacts/architecture.md#llm-integration-architecture) — Provider abstraction, same for headless/UI
- [UX Design: Headless Platform](_bmad-output/planning-artifacts/ux-design-specification.md#platform-strategy) — Secondary platform: CLI headless, `nforge run <spec>`
- [UX Design: User Journey 3](_bmad-output/planning-artifacts/ux-design-specification.md#journey-3-jordan--cicd-integration) — Jordan's CI/CD headless flow
- [Story 2.1: Chat Interface](_bmad-output/implementation-artifacts/2-1-chat-interface-and-auto-generated-node-graph.md) — Auto-graph generation from goal (reused in spec goal-mode)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
