# Story 4.5: Session Resume/Export CLI

Status: done

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

- [x] Implement session resume CLI command (AC: 1)
  - [x] Add `resume` subcommand to `cmd/nforge/session.go` with Cobra
  - [x] Implement `Manager.ResumeSession(ctx, id) error` in `internal/session/manager.go`
  - [x] Restore session state: graph JSON from SQLite, workspace files from disk
  - [x] Broadcast resume event via WebSocket hub for UI parity (FR30)

- [x] Implement session export CLI command (AC: 2)
  - [x] Add `export` subcommand to `cmd/nforge/session.go` with Cobra
  - [x] Create `internal/session/export.go` with tarball generation logic
  - [x] Include: graph JSON, workspace source files, README
  - [x] Exclude: API keys, `.env`, `config.yaml` with secrets, `.git` directory (NFR-10)
  - [x] Save tarball to current directory with naming: `session-<id>-<timestamp>.tar.gz`

- [x] Implement session list CLI command (AC: 3)
  - [x] Add `list` subcommand to `cmd/nforge/session.go` with Cobra
  - [x] Use existing `Manager.ListSessions(ctx)` from Story 4.1
  - [x] Calculate workspace size via `Manager.GetSessionStats(ctx, id)`
  - [x] Format output: table with columns ID, Name, Status, Created, Size

- [x] Add API endpoints for UI parity (AC: 1, 2, FR30)
  - [x] `POST /api/v1/sessions/:id/resume` — restore session, returns session data
  - [x] `POST /api/v1/sessions/:id/export` — download tarball, sets Content-Disposition header
  - [x] Verify `GET /api/v1/sessions` (list) exists from Story 4.1, update if needed

- [x] Testing (all AC)
  - [x] Unit tests for `ResumeSession`, `ExportSession`, `GetSessionStats` in `internal/session/session_test.go`
  - [x] CLI tests: `nforge session resume/export/list` with table-driven tests
  - [x] Integration test: full export-extract-verify cycle
  - [x] Verify no API keys in export tarball (security test)

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

1. **CLI Commands Already Existed**: `session list`, `session resume`, and `session export` subcommands were already implemented in `cmd/nforge/session.go` from previous stories. `Manager.ResumeSession` was already in `manager.go`. `ExportSession` and `ExportSessionToWriter` were already in `export.go`. API endpoints for resume and export were already registered in `serve.go`.

2. **New: GetSessionStats Method**: Added `Manager.GetSessionStats(ctx, id)` to `internal/session/manager.go` which combines session metadata with workspace size calculation. Returns `SessionStats` struct with ID, Name, Status, CreatedAt, LastActiveAt, and WorkspaceSize fields.

3. **Updated CLI List Output**: Modified `runListSessions` in `cmd/nforge/session.go` to display workspace size using `GetSessionStats`. Added `formatBytes` helper for human-readable size formatting (B, KB, MB, GB).

4. **Enhanced Export Security**: Added `.env.*` pattern to `excludedPatterns` in `export.go` to exclude `.env.local`, `.env.production`, etc.

5. **Comprehensive Tests Added**:
   - `TestResumeSession_NotFound`, `TestResumeSession_Success`, `TestResumeSession_PreservesGraphAndChat`
   - `TestResumeSession_UpdatesTimestamps`, `TestResumeSession_NonCompleteSession`
   - `TestGetSessionStats_NotFound`, `TestGetSessionStats_ReturnsWorkspaceSize`, `TestGetSessionStats_EmptyWorkspace`
   - `TestExportExtractVerifyCycle` — full integration test
   - `TestExportExcludesAllSensitiveKeys` — comprehensive graph sanitization
   - `TestExportExcludesAllSecretFiles` — file exclusion verification
   - `TestFormatBytes` — table-driven test for size formatting

6. **UI Parity Verified**: `SessionExplorer.tsx` already has resume, export, and fork buttons. `useSession.ts` hook has `resumeSession` and `exportSession` methods. No UI changes needed.

7. **All session package tests pass** (55+ tests): `go test ./internal/session/ -v`

### File List

| File | Action | Description |
|------|--------|-------------|
| `internal/session/manager.go` | UPDATE | Added `GetSessionStats` method and `SessionStats` struct |
| `internal/session/export.go` | UPDATE | Added `.env.*` to exclusion patterns |
| `internal/session/session_test.go` | UPDATE | Added tests for ResumeSession, GetSessionStats, export-extract cycle, security tests |
| `cmd/nforge/session.go` | UPDATE | Updated list output to show workspace size with `formatBytes` helper |

### Review Findings

#### Decision Needed

- [x] [Review][Decision] Tarball naming convention mismatch — Resolved: updated to `session-<id>-<timestamp>.tar.gz` format to match spec. [`internal/session/export.go`]

#### Patches

- [x] [Review][Patch] `.git` directory not excluded from export tarballs — Fixed: added `.git` to `excludedPatterns`. [`internal/session/export.go`]
- [x] [Review][Patch] `formatBytes` panics on values >= 1 exabyte — Fixed: added guard for `exp >= 6` returning EB unit. [`cmd/nforge/session.go:114`]
- [x] [Review][Patch] N+1 query pattern in `runListSessions` — Fixed: added `GetSessionStatsFromSession()` method that reuses already-fetched session data, eliminating redundant DB query. [`cmd/nforge/session.go:99`, `internal/session/manager.go:190`]
- [x] [Review][Patch] `sanitizeGraphJSON` silently returns `"{}"` on JSON array input — Fixed: now tries array fallback and sanitizes via `sanitizeSlice`. [`internal/session/export.go`]
- [x] [Review][Patch] `isExcluded` ignores `filepath.Match` errors — Fixed: logs warning on invalid glob patterns. [`internal/session/export.go:33,41`]
- [x] [Review][Patch] No CLI command tests — Fixed: added `cmd/nforge/session_test.go` with tests for `runListSessions`, `runResumeSession`, `runExportSession`, `formatBytes`, and exabyte edge case. [`cmd/nforge/session_test.go`]

#### Deferred

- [x] [Review][Defer] `formatBytes` produces nonsense for negative values — `formatBytes(-1)` returns `"-1 B"`. Workspace size from `GetWorkspaceSize` cannot be negative; defensive but not required for this change. — deferred, pre-existing
- [x] [Review][Defer] `sanitizeMap` misses double-encoded JSON strings — If a sensitive value is nested inside a JSON string (e.g., `{"output": "{\"api_key\": \"sk-123\"}"}`), it is not parsed or sanitized. Pre-existing sanitization limitation, not introduced by this change. — deferred, pre-existing
- [x] [Review][Defer] Symlink check in `addWorkspaceToTar` may not work on Windows — `filepath.Walk` symlink detection behavior differs on Windows. Pre-existing, not introduced by this change. — deferred, pre-existing
- [x] [Review][Defer] Export API uses POST instead of spec's GET — `serve.go` registers `POST /api/v1/sessions/:id/export`, spec task says GET. Dev intentionally changed to POST; pre-existing decision. — deferred, pre-existing
