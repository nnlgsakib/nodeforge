# Story 4.2: Session Resume & Graceful Shutdown

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to resume sessions after restart with graceful shutdown snapshots and auto-cleaned zombie sessions,
so that I never lose my work and stale sessions are removed.

## Acceptance Criteria

1. **Given** the session resume and heartbeat systems are implemented
   **When** the user runs `nforge session resume <id>` or clicks "Resume" in UI
   **Then** the session is restored with snapshot from graceful shutdown (FR33)
   **And** the WebSocket connection is re-established with the restored session state
   **And** the canvas reflects the restored graph state (nodes, edges, statuses)

2. **Given** the heartbeat system is operational
   **When** a session's heartbeat times out (no activity for configured threshold)
   **Then** the session is marked as zombie and auto-cleaned (FR39)
   **And** workspace resources are released (chroot jail removed, file handles closed)
   **And** the zombie session is removed from the session list in UI and CLI

3. **Given** the system health check is implemented
   **When** the user runs `nforge doctor`
   **Then** session health is verified (active sessions, heartbeat status, workspace integrity) (FR28)
   **And** LLM provider connectivity is checked (FR28, FR10-FR16)
   **And** results are displayed with clear pass/fail indicators for each check

4. **Given** the graceful shutdown handler is implemented in `main.go`
   **When** the server receives SIGINT or SIGTERM
   **Then** all active sessions are snapshotted (graph JSON, chat log, workspace files)
   **And** WebSocket connections are closed gracefully with notification to clients
   **And** the server exits after all sessions are saved (exit code 0)

5. **Given** the session manager implements snapshot/restore
   **When** a session is snapshotted during shutdown
   **Then** the snapshot includes: graph state (nodes, edges, statuses), chat history, workspace file manifest, Git commit hash (if auto-commit enabled)
   **And** the snapshot is saved to SQLite (`sessions` table) and workspace metadata to BadgerDB

## Tasks / Subtasks

- [ ] Task 1 (AC: #1, #4, #5) — Session Resume & Graceful Shutdown
  - [ ] Subtask 1.1: Implement `internal/session/manager.go` — `ResumeSession(id)` method that loads snapshot from SQLite + BadgerDB
  - [ ] Subtask 1.2: Implement `internal/devops/graceful.go` — Graceful shutdown handler (SIGINT/SIGTERM) that snapshots all active sessions
  - [ ] Subtask 1.3: Add `ResumeSession` REST endpoint in `cmd/nforge/serve.go` (GET `/api/v1/sessions/:id/resume`)
  - [ ] Subtask 1.4: Implement WebSocket reconnection logic in `frontend/src/hooks/useWebSocket.ts` to handle session restore

- [ ] Task 2 (AC: #2) — Zombie Session Auto-Cleanup
  - [ ] Subtask 2.1: Implement `internal/session/heartbeat.go` — Heartbeat goroutine per session with configurable timeout (default: 5min)
  - [ ] Subtask 2.2: Implement zombie detection loop in session manager that runs every 60s, finds timed-out sessions, and cleans them up
  - [ ] Subtask 2.3: Add zombie cleanup to `nforge doctor` health check output

- [ ] Task 3 (AC: #3) — Enhanced `nforge doctor` Health Check
  - [ ] Subtask 3.1: Extend `cmd/nforge/doctor.go` to verify session health (count active, list zombies, check workspace integrity)
  - [ ] Subtask 3.2: Add session-specific checks: SQLite connectivity, BadgerDB integrity, workspace file permissions
  - [ ] Subtask 3.3: Display session health section in `nforge doctor` output with pass/fail per check

## Dev Notes

- Relevant architecture patterns and constraints:
  - Session data stored in SQLite (`internal/session/`) with BadgerDB (`internal/context/`) for knowledge graph context
  - Graceful shutdown must use Go's `signal.Notify` + `sync.WaitGroup` to wait for all session snapshots
  - WebSocket hub (`cmd/nforge/serve.go`) must broadcast "session_resume" message to clients on restore
  - Heartbeat uses in-memory map (`map[string]time.Time`) keyed by session ID, updated on any session activity

- Source tree components to touch:
  - `internal/session/manager.go` — Add `ResumeSession()`, extend `Session` struct with `Snapshot` field
  - `internal/session/heartbeat.go` — New file for heartbeat goroutine + zombie cleanup
  - `internal/devops/graceful.go` — New file for SIGINT/SIGTERM handler
  - `cmd/nforge/serve.go` — Add resume endpoint, integrate graceful shutdown with `signal.Notify`
  - `cmd/nforge/session.go` — Add `resume` subcommand to `session` command
  - `cmd/nforge/doctor.go` — Extend with session health checks
  - `frontend/src/hooks/useWebSocket.ts` — Handle "session_resume" message type, restore canvas state
  - `frontend/src/components/panels/SessionExplorer.tsx` — Add "Resume" button, show zombie status

- Testing standards summary:
  - Go backend: Ginkgo + Testify, table-driven tests for `ResumeSession()`, heartbeat timeout, zombie cleanup
  - Frontend: Vitest + React Testing Library for WebSocket reconnection, SessionExplorer resume button
  - Integration: Test graceful shutdown by sending SIGTERM, verify snapshots created; test zombie cleanup with simulated timeout

### Project Structure Notes

- Alignment with unified project structure (paths, modules, naming):
  - All Go files in `internal/session/` use `snake_case` package naming, `camelCase` functions, `PascalCase` structs
  - Frontend files in `frontend/src/hooks/` use `camelCase` for functions, `PascalCase` for components/hooks
  - API endpoints: `snake_case` paths (`/api/v1/sessions/:id/resume`), `camelCase` JSON fields (`{"sessionId": "..."}`)
  - Database tables: `snake_case` plural (`sessions`, `session_snapshots`)

- Detected conflicts or variances (with rationale):
  - None — follows established patterns from Story 1.1-1.8 (project scaffolding, CLI, Gin server)

### References

- Epic 4 context: [Source: epics.md#Epic4] Session Management & Recovery — covers FR31-FR39
- Story 4.1: [Source: epics.md#Story4.1] Session Creation & Isolation — prerequisite for resume (session must exist first)
- Architecture: [Source: architecture.md#Session-Mgmt] `internal/session/` package — SQLite for sessions, workspace isolation
- Architecture: [Source: architecture.md#API-Communication] Gin REST + WebSocket hub — `/api/v1/sessions` endpoints, `/ws` for real-time updates
- Architecture: [Source: architecture.md#Infrastructure] Graceful shutdown with session snapshot (FR33)
- PRD: [Source: prd.md#FR33] Resume sessions after restart with graceful shutdown snapshot
- PRD: [Source: prd.md#FR39] Auto-clean zombie sessions (heartbeat timeout detection)
- PRD: [Source: prd.md#FR28] System health check with `nforge doctor`
- UX Design: [Source: ux-design-specification.md#Journey2] Sam — Stuck Node Recovery (power user resume/fork patterns)
- UX Design: [Source: ux-design-specification.md#MonologuePanel] Session state recovery via UI
- Go naming conventions: [Source: architecture.md#Naming-Patterns] `snake_case` packages, `camelCase` functions
- CLI framework: [Source: architecture.md#CLI-Framework] Cobra with `session` subcommand

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List
