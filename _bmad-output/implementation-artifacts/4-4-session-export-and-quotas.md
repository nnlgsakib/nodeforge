# Story 4.4: Session Export & Quotas

Status: ready-for-dev

## Story

As a user,
I want to export sessions as self-contained tarballs and have session quotas enforced (max sessions, max workspace size),
so that I can share results and the system stays within resource limits.

## Acceptance Criteria

1. **Given** the export and quota enforcement systems are implemented
   **When** the user runs `nforge session export <id>` or clicks "Export" in UI
   **Then** a self-contained tarball is generated (graph JSON + source code + README) (FR37)

2. **Given** session quotas are configured
   **When** a new session is created or workspace files are written
   **Then** session quotas are enforced: max sessions limit (configurable), max workspace size (500MB per session) (FR38, NFR-17)

3. **Given** the export system is generating a tarball
   **When** the tarball is created
   **Then** only necessary files are included — API keys and secrets are excluded (NFR-10)

4. **Given** the CLI and UI interfaces
   **When** the user exports a session via either interface
   **Then** identical tarballs are produced (CLI/UI parity, FR30)

5. **Given** a session has completed all nodes (all green)
   **When** the user triggers export
   **Then** the tarball includes: `graph.json` (node graph), `workspace/` (source files), `README.md` (auto-generated session summary)

## Tasks / Subtasks

- [ ] Task 1: Implement session export command (AC: 1, 4, 5)
  - [ ] Subtask 1.1: Add `export` subcommand to `cmd/nforge/session.go` with CLI signature `nforge session export <id> [--output ./path.tar.gz]`
  - [ ] Subtask 1.2: Create `internal/session/export.go` with `ExportSession(ctx, sessionID, outputPath) error` function
  - [ ] Subtask 1.3: Package graph JSON from BadgerDB (`internal/context/`), workspace files from chroot jail, and generate README.md
  - [ ] Subtask 1.4: Add UI export button in `frontend/src/components/panels/SessionExplorer.tsx` (visible when all nodes green)
  - [ ] Subtask 1.5: Add backend API endpoint `POST /api/v1/sessions/:id/export` returning tarball as streaming response

- [ ] Task 2: Enforce session quotas (AC: 2)
  - [ ] Subtask 2.1: Implement `internal/session/quota.go` with `CheckQuota(sessionID) error` function
  - [ ] Subtask 2.2: Enforce max sessions limit (default 100, configurable via `nforge config set session.max_sessions`)
  - [ ] Subtask 2.3: Enforce max workspace size (500MB per session, NFR-17) — check before each workspace write
  - [ ] Subtask 2.4: Add quota config to `internal/session/manager.go` session creation flow

- [ ] Task 3: Exclude secrets from export (AC: 3)
  - [ ] Subtask 3.1: Add exclusion list in `internal/session/export.go`: exclude `.env`, `config.yaml` (contains API keys), `.nforge/context.db/`, and any file matching `*secret*`, `*key*`, `*credential*`
  - [ ] Subtask 3.2: Sanitize `graph.json` to remove any embedded API keys or tokens from node outputs

- [ ] Task 4: Add tests for export and quotas (AC: 1-5)
  - [ ] Subtask 4.1: Unit tests for `ExportSession` — verify tarball contents, exclusions, README generation
  - [ ] Subtask 4.2: Unit tests for quota enforcement — verify max sessions, workspace size limits
  - [ ] Subtask 4.3: Integration test: CLI `nforge session export <id>` produces valid tarball
  - [ ] Subtask 4.4: Verify UI export button triggers download in browser

## Dev Notes

### Architecture Patterns & Constraints

- **Backend-First**: All export logic in Go (`internal/session/export.go`), frontend only triggers via WebSocket/REST
- **SQLite for Metadata**: Session metadata (ID, creation date, workspace size) stored in `internal/session/` SQLite tables
- **BadgerDB for Graph**: Graph JSON retrieved from `internal/context/` (BadgerDB knowledge graph)
- **Chroot Jail Workspaces**: Workspace files accessed via `internal/session/workspace.go` (chroot-isolated paths)
- **CLI/UI Parity**: `nforge session export` and UI export button must produce identical tarballs (FR30)

### Source Tree Components to Touch

| Component | Path | Action |
|-----------|------|--------|
| Session CLI command | `cmd/nforge/session.go` | UPDATE: add `export` subcommand |
| Session manager | `internal/session/manager.go` | UPDATE: integrate quota checks |
| Session export | `internal/session/export.go` | NEW: export logic |
| Session quota | `internal/session/quota.go` | NEW: quota enforcement |
| Session workspace | `internal/session/workspace.go` | UPDATE: add size tracking |
| REST API routes | `cmd/nforge/serve.go` | UPDATE: add `POST /api/v1/sessions/:id/export` |
| Session Explorer UI | `frontend/src/components/panels/SessionExplorer.tsx` | UPDATE: add export button |
| WebSocket hub | `cmd/nforge/serve.go` (hub) | UPDATE: broadcast export progress |

### Testing Standards Summary

- **Framework**: Ginkgo + Testify (Go backend), Vitest (frontend)
- **Coverage Target**: 80%+ for new files (`export.go`, `quota.go`)
- **Test Types**: Unit tests (mock DB), integration tests (real SQLite/Badger, CLI execution)
- **Key Test Cases**:
  - Export produces valid tarball with correct files
  - Export excludes secrets (API keys, config files)
  - Quota rejects when max sessions reached
  - Quota rejects when workspace exceeds 500MB
  - CLI and UI export produce identical results

## Project Structure Notes

### Alignment with Unified Project Structure

- Go packages follow `snake_case`: `internal/session/`, `internal/context/`
- CLI commands use Cobra subcommands: `nforge session export <id>`
- API endpoints use `snake_case`: `POST /api/v1/sessions/:id/export`
- JSON responses use `camelCase`: `{"sessionId": "...", "exportUrl": "..."}`
- Frontend files use `kebab-case`: `session-explorer.tsx`
- React components use `PascalCase`: `SessionExplorer`

### Detected Conflicts or Variances

- None detected — this story follows established patterns from Epic 4 (session management)

## References

- [Epic 4 Requirements: Story 4.4](_bmad-output/planning-artifacts/epics.md#story-44-session-export--quotas) — Acceptance criteria, user story
- [Architecture: Session Management](_bmad-output/planning-artifacts/architecture.md#session-mgmt--sqlitesessionsskills--badgerdb-knowledge-graph) — `internal/session/` structure, SQLite + BadgerDB usage
- [Architecture: Project Structure](_bmad-output/planning-artifacts/architecture.md#complete-project-directory-structure) — `cmd/nforge/`, `internal/session/` paths
- [UX: Journey 1 Alex](_bmad-output/planning-artifacts/ux-design-specification.md#journey-1-alex--jsgogo-migration-first-time-user) — Export trigger on all nodes green
- [UX: Component Strategy](_bmad-output/planning-artifacts/ux-design-specification.md#component-strategy) — SessionExplorer panel export button
- [FR37: Export session as self-contained tarball](_bmad-output/planning-artifacts/epics.md#fr37-user-can-export-session-as-self-contained-tarball-graph--source--readme)
- [FR38: Enforce session quotas](_bmad-output/planning-artifacts/epics.md#fr38-system-enforces-session-quotas-max-sessions-max-workspace-size)
- [NFR-10: Session secrets exclusion](_bmad-output/planning-artifacts/epics.md#nfr-10-session-secrets-vault-integration-for-secret-management-api-keys-never-logged-never-in-session-export-tarballs)
- [NFR-17: Session density quotas](_bmad-output/planning-artifacts/epics.md#nfr-17-session-density-100-concurrent-sessions-per-instance-each-with-isolated-workspace-max-500mb-workspace-size-quota)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

