# Story 4.3: Fork, Git Auto-Commit & Time-Travel Debug

Status: ready-for-dev

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

- [ ] Implement Git auto-commit after each node completion (AC: 2)
  - [ ] Create `internal/session/autocommit.go` with `AutoCommit(sessionID, nodeID, status string) error` function
  - [ ] Run `git add -A` and `git commit -m "Node <nodeID> completed: <status>"` in workspace directory
  - [ ] Integrate with `internal/engine/executor.go` to call `AutoCommit` after each node completes
  - [ ] Deterministic commit messages: include node ID, status, ISO 8601 timestamp

- [ ] Implement session fork (Git branch-like) (AC: 1)
  - [ ] Create `internal/session/fork.go` with `ForkSession(ctx context.Context, parentID string) (*Session, error)` function
  - [ ] Initialize Git repo in workspace if not already initialized (`git init`)
  - [ ] Create new session (new ID, same workspace state) with Git branch from current commit
  - [ ] Add fork CLI command to `cmd/nforge/session.go` (e.g., `nforge session fork <id>`)

- [ ] Implement time-travel debug (AC: 3)
  - [ ] Create `internal/session/timetravel.go` with `CheckoutNodeState(sessionID, nodeID string) error` function
  - [ ] Look up Git commit hash for the target completed node
  - [ ] Run `git checkout <commit-hash>` in workspace to restore files to that node's completion state
  - [ ] Add WebSocket message type `checkout_node` for UI-triggered time-travel (App.tsx integration)

- [ ] Update `internal/session/manager.go` to expose `ForkSession`, `AutoCommit`, `CheckoutNodeState` methods
- [ ] Update `internal/session/workspace.go` to include `InitGitRepo()` method for workspace Git initialization

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

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
