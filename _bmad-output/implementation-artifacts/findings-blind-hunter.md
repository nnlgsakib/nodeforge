## Blind Hunter Findings

- **Predictable session ID via PID** — `internal/session/manager.go:generateSessionID()` — Uses `os.Getpid()` which is predictable, reused across restarts, and not unique enough for session identifiers
- **Port validation missing range check** — `cmd/nforge/serve.go:validatePort()` — Only checks numeric characters, accepts "0" or "99999" which are invalid port numbers (valid range: 1-65535)
- **SetDistFS swallows error, logs to stdout** — `cmd/nforge/serve.go:SetDistFS()` — Error printed via `fmt.Printf` to stdout, `frontendFS` set to nil silently, caller cannot detect failure
- **Potential data race on servePort** — `cmd/nforge/serve.go:websocketUpgrader` — `servePort` captured by closure in `CheckOrigin`, read without synchronization while potentially written by flag parsing (technically a data race per Go race detector)
- **Path traversal check incomplete** — `cmd/nforge/serve.go:NoRoute handler` — Only checks literal `..` in path, does not catch URL-encoded variants like `..%2F` or `%2E%2E%2F`
- **Case-sensitive Content-Type detection** — `cmd/nforge/serve.go:NoRoute handler` — Extension checks use lowercase `HasSuffix` only; `.JS`, `.CSS`, `.PNG` files get `application/octet-stream`
- **WebSocket close frame not echoed** — `cmd/nforge/serve.go:WebSocket handler` — Read loop breaks on close frame without sending a close response frame back to client
- **Empty created_at in config.yaml** — `internal/session/workspace.go:InitProjectWorkspace()` — Writes `created_at: ""` (empty string) instead of actual timestamp
- **Readiness endpoint unaware of shutdown** — `cmd/nforge/serve.go:/readyz` — Returns 200 even during graceful shutdown, no shutdown state tracking
- **No WebSocket connection tracking** — `cmd/nforge/serve.go:WebSocket handler` — Connections not tracked in a map; graceful shutdown cannot close them, leading to hung connections
- **CORS middleware hardcoded to localhost:5173** — `cmd/nforge/serve.go:CORS middleware` — Only allows `http://localhost:5173` origin, not configurable for production or alternative dev setups
- **WebSocket CheckOrigin rejects empty Origin** — `cmd/nforge/serve.go:websocketUpgrader.CheckOrigin` — Non-browser clients (curl, mobile apps, non-browser JS) typically send empty Origin and get rejected
- **Missing //go:embed directive for distFS** — `cmd/nforge/serve.go:var distFS embed.FS` — Without `//go:embed` directive (typically in main.go), `distFS` is always empty and `frontendFS` is always nil
- **Frontend alert() for user feedback** — `frontend/src/App.tsx:handleCreateProject()` — Uses `alert()` which blocks the JS thread and provides poor UX
- **ChatPanel new command accepts any non-whitespace** — `frontend/src/components/panels/ChatPanel.tsx:handleSubmit()` — Regex `/^new\s+(\S+)$/i` captures anything after "new " including special chars, paths, etc.
