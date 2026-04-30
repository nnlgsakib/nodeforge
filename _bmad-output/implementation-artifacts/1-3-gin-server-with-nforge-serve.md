# Story 1.3: Gin Server with `nforge serve`

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to start the web UI + API with `nforge serve`,
So that I can access the React Flow canvas in my browser.

## Acceptance Criteria

**Given** the Gin framework is installed and `main.go` is set up
**When** the user runs `nforge serve`
**Then** the Gin server starts on the configured port (default :8080) with REST API (`/api/v1/*`) and WebSocket hub (`/ws`)
**And** the React build (from `frontend/dist/`) is served via `embed.FS` at the root path
**And** health check is available at `/healthz` and metrics at `/metrics`

## Tasks / Subtasks

- [ ] Task 1: Create `nforge serve` Cobra subcommand (AC: 1)
  - [ ] Subtask 1.1: Create `cmd/nforge/serve.go` with `serve` command definition
  - [ ] Subtask 1.2: Add `--port` flag (default: 8080) and `--frontend-dir` flag (default: `frontend/dist`)
  - [ ] Subtask 1.3: Register `serve` command in `cmd/nforge/root.go` (depends on Story 1.2 being implemented)

- [ ] Task 2: Implement Gin server with REST API and WebSocket hub (AC: 1)
  - [ ] Subtask 2.1: Initialize Gin router with default middleware (Logger, Recovery)
  - [ ] Subtask 2.2: Define REST API route group `/api/v1/` with placeholder endpoints for sessions, skills, config
  - [ ] Subtask 2.3: Implement WebSocket hub at `/ws` endpoint with Gorilla WebSocket upgrade
  - [ ] Subtask 2.4: Add CORS middleware for frontend dev server compatibility (localhost:5173 during dev)

- [ ] Task 3: Implement `embed.FS` for frontend serving (AC: 2)
  - [ ] Subtask 3.1: Import `embed` package and declare `//go:embed frontend/dist/*` in `main.go`
  - [ ] Subtask 3.2: Serve embedded React build at root path (`/*`) — fallthrough to `index.html` for SPA routing
  - [ ] Subtask 3.3: Verify `frontend/dist/` exists; if missing, serve a placeholder "Build frontend first" message
  - [ ] Subtask 3.4: Test that `http://localhost:8080` serves the React Flow canvas after `npm run build`

- [ ] Task 4: Add health check and metrics endpoints (AC: 3)
  - [ ] Subtask 4.1: Implement `/healthz` endpoint returning JSON with status, timestamp, and version
  - [ ] Subtask 4.2: Implement `/metrics` endpoint (Prometheus format placeholder — full implementation in Story 6.5)
  - [ ] Subtask 4.3: Add readiness check (Gin server up, frontend embedded) and liveness check (basic uptime)

- [ ] Task 5: Integration and startup verification (AC: 1, 2, 3)
  - [ ] Subtask 5.1: `nforge serve` starts without errors on default port 8080
  - [ ] Subtask 5.2: `curl http://localhost:8080/api/v1/` returns 200 OK (REST API alive)
  - [ ] Subtask 5.3: WebSocket connection to `ws://localhost:8080/ws` succeeds (upgrade to WS protocol)
  - [ ] Subtask 5.4: `curl http://localhost:8080/` serves React index.html from embedded dist
  - [ ] Subtask 5.5: `curl http://localhost:8080/healthz` returns JSON with status "ok"

## Dev Notes

### Architecture Patterns and Constraints

**Gin Version:** Gin 1.11.0 (radix tree router, 38% lower allocation overhead, HTTP/3 support) — NOT Chi. Single framework for REST API + WebSocket hub. — [Source: architecture.md#API-Communication-Patterns]

**WebSocket Hub Requirements:**
- Gorilla WebSocket for `github.com/gorilla/websocket` (Gin doesn't have built-in WS; use Gorilla for upgrade)
- Support 5000+ concurrent connections (NFR-01: <50ms state propagation)
- Message format: `{"type": "node_update", "nodeId": "...", "status": "..."}` — [Source: architecture.md#API-Communication-Patterns]

**embed.FS Pattern (CRITICAL):**
- Declare `//go:embed frontend/dist/*` at package level in `main.go`
- Serve SPA: any non-API, non-WS route falls through to `index.html` (React Router / React Flow)
- During development: frontend runs on `localhost:5173` (Vite HMR); Gin proxies or CORS allows cross-origin
- In production: only the embedded build is served; no filesystem access needed

**API Design:**
- REST: `/api/v1/sessions`, `/api/v1/skills`, `/api/v1/config` — [Source: architecture.md#API-Boundaries]
- WebSocket: `/ws` — real-time graph state, LLM streaming, monologue
- Health: `/healthz` — session stats, LLM connectivity (NFR-28)
- Metrics: `/metrics` — Prometheus (NFR-30, full impl in Epic 6)

**Cobra Integration (Prerequisite: Story 1.2):**
- `cmd/nforge/serve.go` defines the `serve` subcommand
- Persistent flags from root command (`--verbose`, `--config-path`) apply
- `--port` flag overrides config file setting
- Server runs until interrupted (Ctrl+C) with graceful shutdown hook

### Source Tree Components to Touch

**New Files (CREATE):**
- `cmd/nforge/serve.go` — Cobra subcommand for `nforge serve`
- Health/metrics endpoint handlers (in `internal/devops/` or inline in serve.go for now)

**Files to Modify (UPDATE):**
- `cmd/nforge/root.go` — Register `serve` command (after Story 1.2 creates this file)
- `main.go` — Add `//go:embed frontend/dist/*` directive, serve embedded content, start Gin router

**Current State of Files Being Modified:**
- `main.go` (from Story 1.1): Currently a scaffold with basic Gin setup and embed.FS foundation. Needs server startup logic added.
- `cmd/nforge/root.go` (from Story 1.2 — NOT YET CREATED): Will define root command with persistent flags. Story 1.3 depends on 1.2 being done first.

**IMPLEMENTATION ORDER WARNING:**
Story 1.2 (CLI Root Command) MUST be implemented before 1.3 can be completed. The `serve.go` subcommand must be registered in `root.go`. If 1.2 is not done, create the serve command structure but note it cannot be registered until root.go exists.

### Testing Standards

**Go Testing (Ginkgo + Testify):**
- Co-located `*_test.go` files
- Test that `nforge serve --port 9090` starts server on port 9090
- Test that `/healthz` returns 200 with valid JSON
- Test WebSocket upgrade at `/ws` endpoint
- Test that embedded frontend is served at `/` (requires `frontend/dist/` to exist)

**Test Pattern:**
```go
// cmd/nforge/serve_test.go
func TestServeCommand(t *testing.T) {
    // Test that serve command registers correctly
    // Test that --port flag is respected
    // Test that invalid port returns error
}
```

## Project Structure Notes

### Alignment with Unified Project Structure

The server implementation must align with the architecture specification:

```
nfv2/
├── cmd/nforge/
│   ├── root.go           # Cobra root command (Story 1.2)
│   └── serve.go         # nforge serve (THIS STORY)
├── internal/
│   ├── devops/
│   │   ├── health.go    # /healthz endpoint (created in this story)
│   │   └── metrics.go   # /metrics Prometheus (placeholder in this story)
│   └── ...
├── frontend/
│   └── dist/            # React build output (embedded by main.go)
└── main.go              # Gin server + embed.FS (modified in this story)
```

— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**Dependency on Story 1.2:** Story 1.2 (CLI Root Command with Cobra Framework) is currently in "backlog" status. This story (1.3) cannot be fully completed until 1.2 is done. The developer should:
1. Note the dependency
2. Create `serve.go` with the command definition
3. Add a TODO: "Register in root.go after Story 1.2 is implemented"
4. Or implement Story 1.2 first, then complete 1.3

**Gin Version Consistency:** Architecture specifies Gin 1.10+; Story 1.1 installed Gin v1.11.0 (compatible with Go 1.24). Do NOT upgrade to v1.12.0+ (requires Go 1.25+).

## References

- [Story 1.3 Definition: epics.md#Story-1.3](_bmad-output/planning-artifacts/epics.md#Story-1.3)
- [Story 1.1 (Prerequisite): implementation-artifacts/1-1-project-scaffolding-and-module-init.md](_bmad-output/implementation-artifacts/1-1-project-scaffolding-and-module-init.md)
- [Story 1.2 (Prerequisite - backlog): epics.md#Story-1.2](_bmad-output/planning-artifacts/epics.md#Story-1.2)
- [Gin Router: architecture.md#API-Communication-Patterns](_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns)
- [WebSocket Hub: architecture.md#API-Communication-Patterns](_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns)
- [embed.FS Pattern: architecture.md#Starter-Template-Evaluation](_bmad-output/planning-artifacts/architecture.md#Starter-Template-Evaluation)
- [Health & Metrics: prd.md#DevOps-Capabilities](_bmad-output/planning-artifacts/prd.md#DevOps-Capabilities)
- [NFR-01: architecture.md#Performance](_bmad-output/planning-artifacts/architecture.md#Performance)
- [Gorilla WebSocket: https://github.com/gorilla/websocket](https://github.com/gorilla/websocket) — Required for WS support in Gin
- [Gin v1.11.0: https://github.com/gin-gonic/gin/releases/tag/v1.11.0](https://github.com/gin-gonic/gin/releases/tag/v1.11.0)

## Dev Agent Record

### Agent Model Used

tencent/hy3-review:free

### Debug Log References

### Completion Notes List

### File List

- `cmd/nforge/serve.go` (NEW)
- `cmd/nforge/root.go` (MODIFY - after Story 1.2)
- `main.go` (MODIFY - add embed directive, server startup)
- `internal/devops/health.go` (NEW - or inline in serve.go)
- `internal/devops/metrics.go` (NEW - placeholder)
