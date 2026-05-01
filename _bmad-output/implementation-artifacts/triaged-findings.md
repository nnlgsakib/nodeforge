# Triaged Findings — Story 1.4 Code Review

## Summary
- **Total findings before triage:** 29
- **Dismissed (false positives / handled):** 2
- **Deferred (pre-existing):** 1
- **Decision Needed:** 5
- **Patch (fixable):** 21
- **Failed layers:** none (all three layers executed successfully)

---

## DECISION NEEDED (requires human input)

### D1. [decision] WebSocket CheckOrigin rejects empty Origin
- **Sources:** blind
- **Detail:** `serve.go:websocketUpgrader.CheckOrigin` rejects requests with empty `Origin` header (non-browser clients: curl, mobile apps, some JS frameworks). Current policy only allows `localhost:5173` and `localhost:<servePort>`.
- **Location:** `cmd/nforge/serve.go:websocketUpgrader`
- **Question:** Should non-browser clients be allowed to connect via WebSocket? If yes, consider allowing empty Origin or making it configurable.

### D2. [decision] Frontend hardcoded workspaceDir — CLI `--workspace-dir` has no effect on UI
- **Sources:** edge + auditor
- **Detail:** `frontend/src/App.tsx:createProject()` always sends `workspaceDir: '.'`. The API parses `req.WorkspaceDir` but never uses it — `CreateSessionWithName` uses the manager's `workspaceRoot`. The CLI `--workspace-dir` flag works for CLI but UI always creates in `.`.
- **Location:** `frontend/src/App.tsx:createProject()`, `internal/canvas/api.go:createSession()`
- **Question:** Should the UI accept a workspaceDir parameter? Should `CreateSessionWithName` accept a workspace dir override? Is the hardcoded `'.'` intentional for Story 1.4?

### D3. [decision] `session.NewManager(".")` uses relative path — configurable?
- **Sources:** edge
- **Detail:** `serve.go` calls `session.NewManager(".")` creating sessions relative to cwd. If server started from unexpected directory, sessions created in wrong location.
- **Location:** `cmd/nforge/serve.go:session.NewManager(".")`
- **Question:** Should workspace root be a configurable flag (e.g., `--workspace-root`) or left to Story 1.5 (Configuration Management)?

### D4. [decision] `--template` flag ignored — implement or defer?
- **Sources:** auditor
- **Detail:** `cmd/nforge/new.go` defines `--template` flag and passes `template` to `runNewProject`, but the parameter is ignored. No template logic is implemented.
- **Location:** `cmd/nforge/new.go:runNewProject()`
- **Question:** Was template support intended for Story 1.4 or deferred? If 1.4, need to implement; if not, remove the flag to avoid confusion.

### D5. [decision] Frontend filenames not kebab-case per spec
- **Sources:** auditor
- **Detail:** Spec ("Naming Conventions") requires `kebab-case.tsx` files (e.g., `chat-panel.tsx`) with `PascalCase` components. Current files are `ChatPanel.tsx` and `SessionExplorer.tsx` (PascalCase filenames).
- **Location:** `frontend/src/components/panels/ChatPanel.tsx`, `SessionExplorer.tsx`
- **Question:** Rename files to `chat-panel.tsx` / `session-explorer.tsx`, or update spec to allow PascalCase filenames?

---

## PATCH (fixable without human input)

### P1. [patch] Predictable session ID via PID
- **Sources:** blind + edge + auditor
- **Detail:** `generateSessionID()` returns `sess-<pid>` which is predictable, reused across restarts, and not unique. Should use UUID or timestamp+random.
- **Location:** `internal/session/manager.go:generateSessionID()`

### P2. [patch] Port validation missing range check (1-65535)
- **Sources:** blind
- **Detail:** `validatePort()` only checks numeric characters, accepts "0" or "99999" which are invalid. Should validate 1 ≤ port ≤ 65535.
- **Location:** `cmd/nforge/serve.go:validatePort()`

### P3. [patch] SetDistFS logs to stdout instead of returning error
- **Sources:** blind
- **Detail:** On `fs.Sub` failure, error is printed via `fmt.Printf` and `frontendFS` set to nil. Caller cannot detect failure. Should return error.
- **Location:** `cmd/nforge/serve.go:SetDistFS()`

### P4. [patch] Data race on servePort in websocketUpgrader closure
- **Sources:** blind
- **Detail:** `servePort` captured by closure in `websocketUpgrader` init. Although practically safe (set before server starts), technically a data race. Move upgrader initialization into `runServer()` after flag parsing.
- **Location:** `cmd/nforge/serve.go:websocketUpgrader`

### P5. [patch] Path traversal check incomplete (URL-encoded `..`)
- **Sources:** blind + edge
- **Detail:** NoRoute handler only checks literal `..` via `strings.Contains`. URL-encoded variants like `..%2F` or `%2E%2E%2F` bypass the check.
- **Location:** `cmd/nforge/serve.go:NoRoute handler`

### P6. [patch] Case-sensitive Content-Type detection
- **Sources:** blind
- **Detail:** Extension checks use lowercase `strings.HasSuffix` only. Files like `.JS`, `.CSS`, `.PNG` get `application/octet-stream`.
- **Location:** `cmd/nforge/serve.go:NoRoute handler`

### P7. [patch] WebSocket close frame not echoed
- **Sources:** blind
- **Detail:** Read loop breaks on close frame without sending a close response frame. Per WebSocket spec, should echo the close frame.
- **Location:** `cmd/nforge/serve.go:WebSocket handler`

### P8. [patch] Empty `created_at` in config.yaml
- **Sources:** blind + auditor
- **Detail:** `InitProjectWorkspace` writes `created_at: ""` (empty string). Should populate with actual RFC3339 timestamp.
- **Location:** `internal/session/workspace.go:InitProjectWorkspace()`

### P9. [patch] Readiness endpoint unaware of shutdown
- **Sources:** blind
- **Detail:** `/readyz` returns 200 even during graceful shutdown. Should track shutdown state and return 503 when shutting down.
- **Location:** `cmd/nforge/serve.go:/readyz`

### P10. [patch] No WebSocket connection tracking for graceful shutdown
- **Sources:** blind
- **Detail:** WebSocket connections not tracked in a map/slice. Graceful shutdown cannot close them, leading to hung connections during shutdown.
- **Location:** `cmd/nforge/serve.go:WebSocket handler`

### P11. [patch] CORS middleware hardcoded to localhost:5173
- **Sources:** blind
- **Detail:** CORS middleware only allows `http://localhost:5173`. Not configurable for production or alternative dev setups. Should read from config or flag.
- **Location:** `cmd/nforge/serve.go:CORS middleware`

### P12. [patch] Frontend uses `alert()` for user feedback
- **Sources:** blind
- **Detail:** `handleCreateProject` uses `alert()` which blocks JS thread and provides poor UX. Should use a toast notification or inline message.
- **Location:** `frontend/src/App.tsx:handleCreateProject()`

### P13. [patch] Project name not sanitized — path traversal via project name
- **Sources:** edge + blind
- **Detail:** `projectName` is not validated/sanitized. Names like `../../etc` could traverse out of workspace root via `filepath.Join`. Should reject names containing path separators or special chars.
- **Location:** `internal/session/manager.go:CreateSessionWithName()`, `cmd/nforge/new.go`

### P14. [patch] Concurrent session creation race
- **Sources:** edge
- **Detail:** `CreateSessionWithName` has no locking. Two simultaneous requests with same project name could both pass directory-exists checks and create duplicate sessions.
- **Location:** `internal/session/manager.go:CreateSessionWithName()`

### P15. [patch] Frontend rapid double-click creates duplicate projects
- **Sources:** edge
- **Detail:** No debounce or loading state on "New Project" button. Rapid clicks send multiple POST requests.
- **Location:** `frontend/src/components/panels/SessionExplorer.tsx`

### P16. [patch] WebSocket 60s idle timeout kills active connections
- **Sources:** edge
- **Detail:** `conn.SetReadDeadline(60s)` kills connections with no messages for 60s, even if client is active but not sending. Should implement ping/pong properly.
- **Location:** `cmd/nforge/serve.go:WebSocket handler`

### P17. [patch] InitProjectWorkspace overwrites existing .nforge/ without warning
- **Sources:** edge
- **Detail:** `InitProjectWorkspace` uses `os.WriteFile` which overwrites existing config.yaml, README.md, .gitignore without checking if `.nforge/` already exists. Should return error or warn.
- **Location:** `internal/session/workspace.go:InitProjectWorkspace()`

### P18. [patch] Frontend error handling assumes JSON response
- **Sources:** edge
- **Detail:** `createProject()` calls `response.json()` which throws `SyntaxError` on non-JSON error responses (e.g., 500 HTML). Should check content-type or wrap in try/catch with fallback.
- **Location:** `frontend/src/App.tsx:createProject()`

### P19. [patch] ChatPanel "new" command accepts invalid project names
- **Sources:** blind
- **Detail:** Regex `/^new\s+(\S+)$/i` captures any non-whitespace after "new", including special chars, paths, etc. Should validate project name format.
- **Location:** `frontend/src/components/panels/ChatPanel.tsx:handleSubmit()`

### P20. [patch] SessionRequest.WorkspaceDir required but unused (dead code)
- **Sources:** auditor
- **Detail:** `req.WorkspaceDir` is parsed with `binding:"required"` but never used in `createSession`. `CreateSessionWithName` uses manager's `workspaceRoot`. Either use the field or remove it.
- **Location:** `internal/canvas/api.go:createSession()`

### P21. [patch] No frontend tests (Vitest + RTL)
- **Sources:** auditor
- **Detail:** Spec requires "Vitest + React Testing Library (co-located `*.test.tsx` files)" but no test files exist for `ChatPanel.tsx` or `SessionExplorer.tsx`.
- **Location:** `frontend/src/components/panels/`

---

## DEFERRED (pre-existing, not caused by this change)

### D1. [defer] Missing //go:embed directive for distFS
- **Sources:** blind
- **Detail:** `var distFS embed.FS` without `//go:embed` directive (typically in `main.go`). If missing, `distFS` is always empty and `frontendFS` is nil. This is a pre-existing issue — not introduced by Story 1.4 changes.
- **Location:** `main.go` (not in diff) or `cmd/nforge/serve.go`

---

## DISMISSED (false positives / handled) — Count: 2

1. **Whitespace-only project name** (edge #6) — `SessionExplorer.tsx` checks `projectName.trim()` before calling `onCreateProject`. `ChatPanel.tsx` regex `(\S+)` requires at least one non-whitespace char. Both paths are safe.

2. **NoRoute directory falls through to index.html** (edge #11) — For SPA routing, falling through to `index.html` on directory requests is correct behavior, not a bug.

---

## Failed Layers
None — all three layers (Blind Hunter, Edge Case Hunter, Acceptance Auditor) completed successfully.
