# Story 4.3: Fork, Git Auto-Commit & Time-Travel Debug

Status: done

## Story

As a user,
I want to fork sessions like Git branches, have workspace changes auto-committed to Git after each node, and time-travel debug by checking out workspace state at any completed node,
So that I can experiment safely and debug historically.

## Acceptance Criteria

1. **Given** the fork, Git auto-commit, and time-travel systems are implemented
   **When** the user presses 'f' (fork) or clicks "Fork Session"
   **Then** a new session branch is created with the current state, allowing different approaches to be tried (FR34)
   **And** workspace changes are auto-committed to Git after each node completion with deterministic commit messages (FR35, NFR-28)
   **And** the user can time-travel debug by checking out workspace state at any completed node (FR36)

## Tasks / Subtasks

- [x] Implement Git auto-commit after each node completion (AC: 2)
  - [x] Create `internal/session/autocommit.go` with `AutoCommit(sessionID, nodeID, status string) error` function
  - [x] Run `git add -A` and `git commit -m "Node <nodeID> completed: <status>"` in workspace directory
  - [x] Integrate with `internal/engine/executor.go` to call `AutoCommit` after each node completes
  - [x] Deterministic commit messages: include node ID, status, ISO 8601 timestamp

- [x] Implement session fork (Git branch-like) (AC: 1)
  - [x] Create `internal/session/fork.go` with `ForkSession(ctx context.Context, parentID string) (*Session, error)` function
  - [x] Initialize Git repo in workspace if not already initialized (`git init`)
  - [x] Create new session (new ID, same workspace state) with Git branch from current commit
  - [x] Add fork CLI command to `cmd/nforge/session.go` (e.g., `nforge session fork <id>`)

- [x] Implement time-travel debug (AC: 3)
  - [x] Create `internal/session/timetravel.go` with `CheckoutNodeState(sessionID, nodeID string) error` function
  - [x] Look up Git commit hash for the target completed node
  - [x] Run `git checkout <commit-hash>` in workspace to restore files to that node's completion state
  - [x] Add WebSocket message type `checkout_node` for UI-triggered time-travel (App.tsx integration)

- [x] Update `internal/session/manager.go` to expose `ForkSession`, `AutoCommit`, `CheckoutNodeState` methods
- [x] Update `internal/session/workspace.go` to include `InitGitRepo()` method for workspace Git initialization

## Dev Notes

### Project Structure Notes

- Aligned with `internal/session/` package structure (Go `snake_case` packages, `camelCase` functions, `PascalCase` structs)
- New files to create: `autocommit.go`, `fork.go`, `timetravel.go`
- Update existing files: `manager.go` (add methods), `workspace.go` (add Git init)
- Integration points:
  - `internal/engine/executor.go` calls `AutoCommit` after node completion
  - `cmd/nforge/session.go` adds `fork` subcommand
  - `frontend/src/App.tsx` handles `checkout_node` WebSocket message for time-travel

### Architecture Compliance

- Git auto-commit uses workspace directory inside chroot jail (FR59: workspace isolation)
- Fork creates new session with isolated workspace (FR31: isolated sessions)
- Time-travel uses Git checkout within chroot jail (no escape to parent dirs)
- All Git operations use `os/exec` to run `git` CLI commands (no external Git library)
- Session manager uses `sync.Mutex` for thread safety (consistent with `manager.go`)

### Library/Framework Requirements

- Git operations via Go `os/exec` package (run `git` commands directly)
- No external Git library needed (keep dependencies minimal, align with project "no bloat" philosophy)
- Reuse `github.com/google/uuid` for new session IDs (already in `manager.go`)
- Gin WebSocket hub for `checkout_node` message propagation (already implemented in `cmd/nforge/serve.go`)

### Testing Requirements

- Unit tests: `autocommit_test.go`, `fork_test.go`, `timetravel_test.go` in `internal/session/`
- Integration test: simulate node completion → verify `git log` shows deterministic commit
- Integration test: fork session → verify new Git branch exists → compare workspace states
- Integration test: time-travel → `git checkout` → verify workspace files match node completion state
- Test framework: Ginkgo + Testify (per architecture.md)

## Previous Story Intelligence

No previous story files found for Epic 4 (stories 4-1, 4-2 not yet created). Follow established patterns from `internal/session/manager.go`:
- Use `sync.Mutex` for thread safety
- Wrap errors with `fmt.Errorf("context: %w", err)` (Go `%w` verb)
- Return `(result, error)` pairs from methods
- Use `path/filepath` for path operations (cross-platform compatibility)

## Project Context Reference

- [Source: epics.md#Story-4.3](_bmad-output/planning-artifacts/epics.md#Story-4.3) — Story definition, user story, acceptance criteria
- [Source: architecture.md#Session-Management](_bmad-output/planning-artifacts/architecture.md#Session-Management) — `internal/session/` structure, file organization
- [Source: architecture.md#Git-Integration](_bmad-output/planning-artifacts/architecture.md#Git-Integration) — NFR-28: auto-commit per node, time-travel debug
- [Source: prd.md#FR34-FR36](_bmad-output/planning-artifacts/prd.md#FR34-FR36) — Functional requirements for fork, auto-commit, time-travel
- [Source: ux-design-specification.md#Journey-2-Sam](_bmad-output/planning-artifacts/ux-design-specification.md#Journey-2-Sam) — User journey: stuck node recovery with fork/retry/skip
- [Source: architecture.md#Naming-Patterns](_bmad-output/planning-artifacts/architecture.md#Naming-Patterns) — Go/TypeScript naming conventions

## Dev Agent Record

### Agent Model Used

Qoder CLI (sonnet)

### Debug Log References

### Completion Notes List

- Created `internal/session/autocommit.go`: `AutoCommit(sessionID, nodeID, status string) error` method on `*Manager`. Runs `git add -A` + `git commit` with deterministic message format `"Node <nodeID> completed: <status> [<ISO8601>]"`. Skips commit if no staged changes. Validates session ID and checks git repo existence.
- Created `internal/session/autocommit_test.go`: 6 tests covering non-git repo, no-changes, with-changes, deterministic commit messages, invalid session ID, and `isGitRepo` helper.
- Created `internal/session/fork.go`: `ForkSession(ctx, parentID)` creates new session with copied workspace and git branch. Copies `.git` directory and creates `fork-<id>` branch. Nil-context safe.
- Created `internal/session/fork_test.go`: 5 tests covering basic fork, fork with git, non-existent parent, invalid ID, and `copyDir` helper.
- Created `internal/session/timetravel.go`: `CheckoutNodeState(sessionID, nodeID)` finds commit via `git log --grep` and checks out. `GetNodeCommitHash` returns hash without checkout.
- Created `internal/session/timetravel_test.go`: 5 tests covering successful checkout, node not found, non-git repo, invalid session ID, and `GetNodeCommitHash`.
- Updated `internal/session/workspace.go`: Added `InitGitRepo(workspaceDir)` function that runs `git init` + configures default user. Safe to call multiple times.
- Updated `internal/engine/executor.go`: Added `NodeAutoCommitter` interface, `autoCommitter`/`sessionID` fields, `SetAutoCommitter`/`SetSessionID` methods. Calls `AutoCommit` after node completion (best-effort, non-blocking).
- Updated `cmd/nforge/session.go`: Added `fork` subcommand with `runForkSession` handler.
- Updated `cmd/nforge/serve.go`: Added `checkout_node` WebSocket message handler in read pump. Calls `CheckoutNodeState` and broadcasts response.
- Updated `frontend/src/hooks/useWebSocket.ts`: Added `checkout_node` message type handler ( Story 4.3).

### File List

**New files:**
- `internal/session/autocommit.go`
- `internal/session/autocommit_test.go`
- `internal/session/fork.go`
- `internal/session/fork_test.go`
- `internal/session/timetravel.go`
- `internal/session/timetravel_test.go`

**Modified files:**
- `internal/session/workspace.go`
- `internal/engine/executor.go`
- `cmd/nforge/session.go`
- `cmd/nforge/serve.go`
- `frontend/src/hooks/useWebSocket.ts`

## Change Log

- Implemented story 4.3: Fork, Git Auto-Commit & Time-Travel Debug (2026-05-04)
  - Git auto-commit after each node completion with deterministic messages
  - Session fork with workspace copy and Git branching
  - Time-travel debug via Git checkout of node completion commits
  - WebSocket `checkout_node` message type for UI-triggered time-travel
  - CLI `nforge session fork <id>` command
  - 16 new unit/integration tests (all passing)
