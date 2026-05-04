# Story 4.5: Session Resume/Export CLI

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to resume/export sessions with `nforge session resume/export <id>` via CLI,
so that I can manage sessions headlessly in terminal.

## Acceptance Criteria

1. **Given** the CLI includes the `session` subcommand
   **When** the user runs `nforge session resume <id>`
   **Then** the session is restored and the user can continue working via CLI or UI (FR26, FR30 CLI/UI parity)

2. **Given** the session export command is implemented
   **When** the user runs `nforge session export <id>`
   **Then** a self-contained tarball (graph + source + README) is generated in the current directory
   **And** API keys and secrets are excluded from the export (NFR-10)

3. **Given** sessions exist in the system
   **When** the user runs `nforge session list`
   **Then** all sessions are displayed with ID, status, creation date, and workspace size
   **And** output format is tabular with clear column headers

## Tasks / Subtasks

- [ ] Implement session resume CLI command (AC: 1)
  - [ ] Add `resume` subcommand to `cmd/nforge/session.go` with Cobra
  - [ ] Implement `Manager.ResumeSession(ctx, id) error` in `internal/session/manager.go`
  - [ ] Restore session state: graph JSON from SQLite, workspace files from disk
  - [ ] Broadcast resume event via WebSocket hub for UI parity (FR30)

- [ ] Implement session export CLI command (AC: 2)
  - [ ] Add `export` subcommand to `cmd/nforge/session.go` with Cobra
  - [ ] Create `internal/session/export.go` with tarball generation logic
  - [ ] Include: graph JSON, workspace source files, README
  - [ ] Exclude: API keys, `.env`, `config.yaml` with secrets, `.git` directory (NFR-10)
  - [ ] Save tarball to current directory with naming: `session-<id>-<timestamp>.tar.gz`

- [ ] Implement session list CLI command (AC: 3)
  - [ ] Add `list` subcommand to `cmd/nforge/session.go` with Cobra
  - [ ] Use existing `Manager.ListSessions(ctx)` from Story 4.1
  - [ ] Calculate workspace size via `Manager.GetSessionStats(ctx, id)`
  - [ ] Format output: table with columns ID, Name, Status, Created, Size

- [ ] Add API endpoints for UI parity (AC: 1, 2, FR30)
  - [ ] `POST /api/v1/sessions/:id/resume` — restore session, returns session data
  - [ ] `GET /api/v1/sessions/:id/export` — download tarball, sets Content-Disposition header
  - [ ] Verify `GET /api/v1/sessions` (list) exists from Story 4.1, update if needed

- [ ] Testing (all AC)
  - [ ] Unit tests for `ResumeSession`, `ExportSession`, `GetSessionStats` in `internal/session/session_test.go`
  - [ ] CLI tests: `nforge session resume/export/list` with table-driven tests
  - [ ] Integration test: full export-extract-verify cycle
  - [ ] Verify no API keys in export tarball (security test)

## Dev Notes

### Architecture Compliance
- **Package**: `internal/session/` (snake_case per Go convention, [Source: architecture.md#Project Structure & Boundaries])
- **Database**: SQLite via `mattn/go-sqlite3 v1.14.44` (CGO_ENABLED=1 required, [Source: project-context.md#Technology Stack & Versions])
- **API Endpoints**: `snake_case` REST paths (`/api/v1/sessions`), `camelCase` JSON fields (`sessionId`, `createdAt`), [Source: architecture.md#API & Communication Patterns]
- **CLI Framework**: Cobra subcommands with `session` parent, [Source: project-context.md#Framework-Specific Rules]
- **WebSocket**: Broadcast resume events for UI real-time updates, [Source: architecture.md#API & Communication Patterns]
- **Security**: Exclude API keys/secrets from exports (NFR-10, [Source: architecture.md#Security Architecture])

### Files to Create/Modify
| File | Action | Purpose |
|------|--------|---------|
| `cmd/nforge/session.go` | UPDATE | Add resume, export, list subcommands to Cobra |
| `internal/session/manager.go` | UPDATE | Add ResumeSession, ExportSession, GetSessionStats methods |
| `internal/session/export.go` | NEW | Tarball generation, exclude secrets |
| `internal/session/session_test.go` | UPDATE | Tests for resume/export/list |
| `cmd/nforge/serve.go` | UPDATE | Register resume/export API endpoints |
| `frontend/src/components/panels/SessionExplorer.tsx` | UPDATE | Ensure UI parity with CLI (FR30) |

### Testing Standards
- **Go**: `testify` assertions, table-driven tests, SQLite integration tests (CGO_ENABLED=1), [Source: project-context.md#Testing Rules]
- **CLI Testing**: Test Cobra commands via `Execute()` and output capture
- **Coverage**: Session resume (state restoration), export (tarball content + secret exclusion), list (formatting + stats)
- **Security Tests**: Verify API keys not in export, config.yaml secrets excluded

### Project Structure Alignment
- Follows `internal/session/` structure from [architecture.md#Complete Project Directory Structure]:
  ```
  internal/session/
    ├── manager.go      # Session CRUD (SQLite)
    ├── workspace.go    # Workspace file operations
    ├── export.go       # NEW: Tarball generation, exclude secrets
    ├── heartbeat.go    # Zombie cleanup (Story 4.2)
    ├── autocommit.go   # Git auto-commit (Story 4.3)
    ├── quota.go        # Session quotas (Story 4.4)
    └── session_test.go # Updated: resume/export/list tests
  ```

### Previous Story Intelligence
- Story 4.1 established `Manager.CreateSessionWithName()`, `Manager.ListSessions()`, SQLite schema for sessions
- Story 4.1 created `cmd/nforge/session.go` stub ("session: not yet implemented (story 4.5)")
- Reuse `Manager.ListSessions()` from 4.1 for `nforge session list`
- Follow existing Cobra pattern from `cmd/nforge/serve.go` and `cmd/nforge/new.go`

### References
- Epic 4 overview: [epics.md#Epic-4: Session Management & Recovery]
- Story 4.5 spec: [epics.md#Story 4.5: Session Resume/Export CLI]
- Session manager spec: [architecture.md#Session Mgmt (FR31-FR39)]
- CLI conventions: [project-context.md#Framework-Specific Rules (Cobra CLI)]
- API conventions: [project-context.md#API: snake_case endpoints, camelCase JSON fields]
- Security: Exclude API keys from exports [architecture.md#Security Architecture]
- Export requirements: [epics.md#Story 4.4: Session Export & Quotas] (tarball format)
- CLI/UI parity requirement: [project-context.md#API: CLI and UI have feature parity]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
