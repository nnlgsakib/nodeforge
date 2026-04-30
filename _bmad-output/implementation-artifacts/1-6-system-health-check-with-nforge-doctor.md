# Story 1.6: System Health Check with `nforge doctor`

Status: ready-for-dev

<!-- Validation: Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to check system health with `nforge doctor`,
so that I can verify all dependencies and connectivity are working.

## Acceptance Criteria

**Given** the CLI includes the `doctor` subcommand (story 1.2 completed, stub registered)
**When** the user runs `nforge doctor`
**Then** the system checks: Go version (1.24+), Gin framework availability, frontend build status, SQLite/BadgerDB connectivity, LLM provider connectivity (Ollama, OpenAI, Anthropic)
**And** results are displayed with clear pass/fail indicators for each check
**And** exit code is 0 if all checks pass, non-zero if any fail

## Tasks / Subtasks

- [ ] Task 1: Create doctor.go with full doctor subcommand (AC: 1)
  - [ ] Subtask 1.1: Create `cmd/nforge/doctor.go` with Cobra command definition
    ```go
    var doctorCmd = &cobra.Command{
        Use:   "doctor",
        Short: "Check system health and connectivity",
        RunE:  func(cmd *cobra.Command, args []string) error {
            return runHealthChecks()
        },
    }
    func init() { rootCmd.AddCommand(doctorCmd) }
    ```
  - [ ] Subtask 1.2: Implement `runHealthChecks()` function returning error if any check fails
  - [ ] Subtask 1.3: Use `cmd/nforge/root.go` persistent flags (`--verbose`, `--config-path`) — inherited automatically

- [ ] Task 2: Implement Go version check (AC: 1)
  - [ ] Subtask 2.1: Parse `runtime.Version()` to extract Go version
  - [ ] Subtask 2.2: Verify version >= Go 1.24 (compare major.minor: `go1.24` or higher)
  - [ ] Subtask 2.3: Display ✅ "Go 1.24+" or ❌ "Go X.XX detected, require 1.24+" with exit code 1 if fail

- [ ] Task 3: Implement Gin framework availability check (AC: 1)
  - [ ] Subtask 3.1: Verify `github.com/gin-gonic/gin` is in `go.mod` dependencies
  - [ ] Subtask 3.2: Attempt `gin.Default()` to confirm import works (or check module cache)
  - [ ] Subtask 3.3: Display ✅ "Gin v1.11.0 available" or ❌ "Gin framework not found" with exit code 1 if fail

- [ ] Task 4: Implement frontend build status check (AC: 1)
  - [ ] Subtask 4.1: Check if `frontend/dist/` directory exists (embedded build output)
  - [ ] Subtask 4.2: Verify `frontend/dist/index.html` exists (valid build)
  - [ ] Subtask 4.3: Display ✅ "Frontend build found (frontend/dist/)" or ❌ "Frontend not built. Run: cd frontend && npm run build" with exit code 1 if fail

- [ ] Task 5: Implement database connectivity checks (AC: 1)
  - [ ] Subtask 5.1: Check SQLite connectivity — `mattn/go-sqlite3` import, attempt `sql.Open("sqlite3", ":memory:")` and ping
  - [ ] Subtask 5.2: Check BadgerDB connectivity — `dgraph-io/badger/v4` import, attempt `badger.Open(badger.DefaultOptions(""))` with temp directory
  - [ ] Subtask 5.3: Display ✅ "SQLite: OK" / "BadgerDB: OK" or ❌ with specific error, exit code 1 if any fail

- [ ] Task 6: Implement LLM provider connectivity checks (AC: 1)
  - [ ] Subtask 6.1: Load config via Viper (`--config-path` flag or default `~/.nforge/config.yaml`)
  - [ ] Subtask 6.2: Check Ollama local — HTTP GET `http://localhost:11434/api/tags` (configurable via `llm.ollama-url`), expect 200 or 400 (if Ollama running)
  - [ ] Subtask 6.3: Check OpenAI — if `llm.openai-key` configured, HTTP GET `https://api.openai.com/v1/models` with `Authorization: Bearer <key>`, expect 200
  - [ ] Subtask 6.4: Check Anthropic — if `llm.anthropic-key` configured, HTTP GET `https://api.anthropic.com/v1/messages` with auth header, expect 200 or 401 (key validation)
  - [ ] Subtask 6.5: Display per-provider: ✅ "Ollama: Connected (http://localhost:11434)" or ⚠️ "Ollama: Not running (start with: ollama serve)" or ❌ "OpenAI: Invalid API key"
  - [ ] Subtask 6.6: LLM checks are WARNING only (non-fatal) — do NOT set exit code 1 for LLM failures (user may not have configured all providers)

- [ ] Task 7: Replace stub doctor command in root.go (AC: 1)
  - [ ] Subtask 7.1: Update `cmd/nforge/root.go` to replace stub doctor command with full implementation from `doctor.go`
  - [ ] Subtask 7.2: Verify `nforge doctor --help` shows full description (not stub message)

- [ ] Task 8: Verify end-to-end (AC: 1, 2, 3)
  - [ ] Subtask 8.1: Run `nforge doctor` — all checks pass → exit code 0
  - [ ] Subtask 8.2: Run `nforge doctor` with Gin not in go.mod → exit code 1, shows ❌ "Gin framework not found"
  - [ ] Subtask 8.3: Run `nforge doctor` with frontend/dist/ missing → exit code 1, shows ❌ "Frontend not built"
  - [ ] Subtask 8.4: Run `nforge doctor --verbose` → shows detailed check output (persistent flag works)
  - [ ] Subtask 8.5: Run `nforge doctor` with Ollama not running → exit code 0 (LLM checks are warnings only), shows ⚠️ "Ollama: Not running"

## Dev Notes

### Architecture Patterns and Constraints

**Go Version:** Go 1.24+ (62% GC pause reduction, improved reflect.Blueprint support) — [Source: architecture.md#Project-Context-Analysis]

**Framework Choices (Non-Negotiable):**
- **Cobra v1.10.2** for CLI — `cmd/nforge/doctor.go` subcommand — [Source: epics.md#Epic-1]
- **Gin 1.11.0** for REST API + WebSocket hub — doctor checks Gin availability — [Source: architecture.md#API-Communication-Patterns]
- **Viper** for config loading (`--config-path` flag from story 1.2, config keys from story 1.5) — [Source: 1-5-configuration-management-with-nforge-config.md]

**Database Dependencies:**
- **SQLite** (`mattn/go-sqlite3`) — check connectivity via `sql.Open("sqlite3", ":memory:")` — [Source: architecture.md#Data-Architecture]
- **BadgerDB** (`dgraph-io/badger/v4`) — check connectivity via `badger.Open()` with temp dir — [Source: architecture.md#Data-Architecture]

**LLM Provider Endpoints for Connectivity Checks:**
- **Ollama**: `http://localhost:11434/api/tags` (default, configurable via `llm.ollama-url` in config) — [Source: prd.md#FR10, FR16]
- **OpenAI**: `https://api.openai.com/v1/models` — requires `llm.openai-key` in config — [Source: prd.md#FR10]
- **Anthropic**: `https://api.anthropic.com/v1/messages` — requires `llm.anthropic-key` in config — [Source: prd.md#FR10]

### Source Tree Components to Touch

**New Files (CREATE):**
- `cmd/nforge/doctor.go` — Full doctor subcommand with all health checks
- `cmd/nforge/doctor_test.go` — Tests for health check logic

**UPDATE Files (MODIFY):**
- `cmd/nforge/root.go` — Replace stub doctor command with full implementation from doctor.go
  - Current state: root.go has stub doctor command registered (from story 1.2): `RunE: func() error { fmt.Println("doctor: not yet implemented"); return nil }`
  - What this story changes: Stub replaced with full doctor command from doctor.go
  - What must be preserved: All other subcommand registrations (serve, run, new, config, skill, session, graph) must remain unchanged

**Files That Must NOT Be Modified:**
- `main.go` — already wired to Cobra root command (from story 1.2)
- `go.mod` — all dependencies (Gin, Cobra, Viper, SQLite, Badger) already present
- `internal/` — no changes needed yet (database checks use standard library interfaces)
- `frontend/` — only checking if `frontend/dist/` exists, no modifications

### Testing Standards

- **Go**: Ginkgo + Testify (from cli-go-project-template pattern) — co-located `*_test.go` files
- **Test for this story:**
  - `cmd/nforge/doctor_test.go` — Test each health check independently
  - Test Go version check: mock `runtime.Version()` → verify pass/fail logic
  - Test Gin availability: verify module check logic
  - Test frontend build check: mock filesystem → verify pass/fail
  - Test SQLite connectivity: `sql.Open("sqlite3", ":memory:")` → expect success
  - Test BadgerDB connectivity: `badger.Open()` with temp dir → expect success
  - Test LLM checks: mock HTTP server → verify Ollama/OpenAI/Anthropic check logic
  - Test exit code: all pass → 0, any critical fail → 1, LLM fail → 0 (warning only)
— [Source: architecture.md#Starter-Template-Evaluation]

### Previous Story Intelligence (from Story 1.5, 1.4, 1.3, 1.2, 1.1)

**What Story 1.5 Established:**
- Viper configured for config persistence with config file path (default: `~/.nforge/config.yaml`)
- Supported config keys: `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `server.port`, `llm.default-model`
- `cmd/nforge/config.go` has full set/get implementation
- Persistent `--config-path` flag (from story 1.2) overrides default config location

**What Story 1.3 Established:**
- Gin 1.11.0 server with REST API (`/api/v1/*`) and WebSocket hub (`/ws`)
- `embed.FS` serves `frontend/dist/` at root path
- Health endpoint at `/healthz`, metrics at `/metrics`

**What Story 1.2 Established:**
- `cmd/nforge/root.go` has Cobra root command with persistent flags: `--verbose` (bool), `--config-path` (string)
- All 8 subcommands registered as stubs: serve, run, new, config, skill, session, doctor, graph
- `main.go` calls `cmd/nforge/root.go` Execute()

**What Story 1.1 Established:**
- Go module: `github.com/nlg/nfv2` with Go 1.24+
- Dependencies in go.mod: `gin-gonic/gin v1.11.0`, `spf13/cobra v1.10.2`, `google.golang.org/protobuf`
- Directory structure: `cmd/nforge/`, `internal/` (engine, llm, context, session, skills, canvas, security, devops), `frontend/`, `proto/`

**Key Learnings Across Stories:**
- Cobra v1.10.2 uses `go.yaml.in/yaml/v3` (not `gopkg.in/yaml.v3`) — Viper v1.19.0+ is compatible
- Naming conventions: `camelCase` functions, `PascalCase` structs, `snake_case` packages, `kebab-case` CLI flags
- All subcommands must be registered in root.go first as stubs (story 1.2), then fully implemented in their own files
- Doctor command was registered as **stub** in story 1.2 — this story delivers the full implementation

**Integration Points:**
- Story 1.6 reads config via Viper (established in story 1.5) to check LLM provider keys/URLs
- Story 1.6 checks frontend build status (`frontend/dist/`) which is created by `npm run build` (story 1.1)
- Story 1.6 checks Gin availability which was installed in story 1.1 and used in story 1.3
- Doctor is a CLI-only tool (no UI equivalent needed — FR28 is CLI-specific)

## Project Structure Notes

### Alignment with Unified Project Structure

The `cmd/nforge/` directory must contain:
- `root.go` — UPDATE: replace stub doctor with full command
- `doctor.go` — NEW: full health check implementation
- `serve.go`, `run.go`, `new.go`, `config.go`, `skill.go`, `session.go`, `graph.go` — other subcommands (unchanged)

```
nfv2/
├── cmd/nforge/          # CLI entrypoint (Cobra commands)
│   ├── root.go           # ← UPDATE: replace stub doctor with full command
│   ├── doctor.go         # ← NEW: nforge doctor (THIS STORY)
│   ├── serve.go         # nforge serve (story 1.3)
│   ├── run.go           # nforge run <spec-file> (story 1.8)
│   ├── new.go           # nforge new <project-name> (story 1.4)
│   ├── config.go        # nforge config set/get (story 1.5)
│   ├── skill.go         # nforge skill list/install (story 5.1)
│   ├── session.go       # nforge session resume/export (story 4.5)
│   └── graph.go         # nforge graph viz (story 1.7)
│
├── frontend/
│   └── dist/              # ← CHECKED: must exist (npm run build)
│
├── go.mod              # ← CHECKED: Gin, Cobra, SQLite, Badger dependencies
├── main.go              # ← UNCHANGED: calls root.go Execute()
...
```
— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**None** — this story delivers the full doctor implementation that was stubbed in story 1.2.

**Important Notes:**
- LLM connectivity checks are **warnings only** (exit code 0 if only LLM fails) — user may not have all providers configured
- Critical checks (Go version, Gin, frontend, databases) set exit code 1 if failed
- Use `--config-path` persistent flag (from root.go) to load config for LLM provider checks
- Ollama check uses `llm.ollama-url` from config (default: `http://localhost:11434`)
- OpenAI/Anthropic checks only run if respective API keys are configured in config

## References

- [Story 1.6 Definition: epics.md#Story-1.6](_bmad-output/planning-artifacts/epics.md#Story-1.6)
- [Story 1.5 (Previous): 1-5-configuration-management-with-nforge-config.md](_bmad-output/implementation-artifacts/1-5-configuration-management-with-nforge-config.md)
- [Story 1.2 (Stub): 1-2-cli-root-command-with-cobra-framework.md](_bmad-output/implementation-artifacts/1-2-cli-root-command-with-cobra-framework.md)
- [Story 1.1 (Foundation): 1-1-project-scaffolding-and-module-init.md](_bmad-output/implementation-artifacts/1-1-project-scaffolding-and-module-init.md)
- [Architecture - CLI Structure: architecture.md#Starter-Template-Evaluation](_bmad-output/planning-artifacts/architecture.md#Starter-Template-Evaluation)
- [Architecture - Data Architecture: architecture.md#Data-Architecture](_bmad-output/planning-artifacts/architecture.md#Data-Architecture)
- [Architecture - Naming Patterns: architecture.md#Naming-Patterns](_bmad-output/planning-artifacts/architecture.md#Naming-Patterns)
- [PRD - CLI Capabilities: prd.md#CLI-Capabilities](_bmad-output/planning-artifacts/prd.md#CLI-Capabilities)
- [PRD - FR28: prd.md#FR28](_bmad-output/planning-artifacts/prd.md#FR28)
- [Cobra v1.10.2 Documentation](https://github.com/spf13/cobra/releases/tag/v1.10.2)
- [Viper v1.19.0 Documentation](https://github.com/spf13/viper/releases/tag/v1.19.0)
- [Ollama API Docs](https://github.com/ollama/ollama/blob/main/docs/api.md)

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List

- `cmd/nforge/doctor.go` (NEW — full doctor implementation)
- `cmd/nforge/doctor_test.go` (NEW — tests for health checks)
- `cmd/nforge/root.go` (UPDATE — replace stub doctor command with full implementation)
