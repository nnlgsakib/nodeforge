---
project_name: 'nfv2'
user_name: 'NLG'
date: '2026-05-02'
sections_completed: ['technology_stack', 'language_specific_rules', 'framework_specific_rules', 'testing_rules', 'code_quality_rules', 'development_workflow_rules', 'critical_dont_miss_rules']
status: 'complete'
rule_count: 65
optimized_for_llm: true
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

## Quick Reference

| Area | Key Rule |
|------|----------|
| Go naming | `snake_case` packages, `camelCase` functions, `PascalCase` structs |
| TS naming | `kebab-case.tsx` files, `PascalCase` components, `camelCase` variables |
| JSON | `camelCase` fields in Go struct tags: `` `json:"sessionId"` `` |
| API | `snake_case` endpoints: `/api/v1/sessions`, WebSocket on `/ws` |
| Data | SQLite for sessions/skills, BadgerDB for knowledge graph |
| Serving | `embed.FS` for `frontend/dist/`, Vite `base: './'` required |
| Security | chroot + eBPF + Argon2, plugins via gRPC Unix socket |
| Testing Go | `go test ./...` + testify, CGO for SQLite |
| Testing TS | Vitest (not Jest!), @testing-library/react |
| Errors | Go: `%w` verb for wrapping, never return directly from Gin handlers |

## Technology Stack & Versions

**Backend (Go 1.26.2):**
- github.com/gin-gonic/gin v1.11.0 — REST API + WebSocket hub
- github.com/spf13/cobra v1.10.2 — CLI framework
- github.com/spf13/viper v1.19.0 — Configuration
- github.com/dgraph-io/badger/v4 v4.9.1 — Knowledge graph (BadgerDB)
- github.com/mattn/go-sqlite3 v1.14.44 — SQLite (requires CGO_ENABLED=1)
- github.com/gorilla/websocket v1.5.3 — WebSocket
- github.com/openai/openai-go v1.12.0 — OpenAI provider
- github.com/stretchr/testify v1.11.1 — Testing

**Frontend (TypeScript 5.3.3 + React 18.2.0):**
- @xyflow/react ^12.10.0 — React Flow canvas
- vite ^5.0.12, vitest ^3.1.1 — Build + testing
- eslint ^8.56.0, @testing-library/react ^15.0.7

**Key Constraints:**
- Go 1.26.2, TypeScript `"strict": true`
- Vite `base: './'` required for Go `embed.FS` serving
- go-sqlite3 requires `CGO_ENABLED=1` on Linux

---

## Critical Implementation Rules

### Language-Specific Rules

**Go (backend/internal/):**
- Packages: `snake_case` (`internal/engine/`), functions: `camelCase`, structs: `PascalCase`
- Errors: Always use `%w` verb — `fmt.Errorf("op: %w", err)`
- SQLite: `snake_case` tables/columns; JSON: `camelCase` field tags `` `json:"sessionId"` ``
- CGO required for go-sqlite3 — `CGO_ENABLED=1`

**TypeScript (frontend/src/):**
- Files: `kebab-case.tsx`, components: `PascalCase`, variables: `camelCase`
- Strict mode: No implicit `any`, no unused locals/params
- JSON fields: `camelCase`; interfaces: `PascalCase`

### Framework-Specific Rules

**Gin (Go Backend):**
- REST: `router.GET("/api/v1/sessions", handler)` — `snake_case` endpoints
- WebSocket: `gorilla/websocket.Upgrader` on `/ws` (not Gin-native)
- Handlers: Use `c.JSON()`, `c.AbortWithStatusJSON()` — never return errors directly
- `embed.FS` serves `frontend/dist/` — requires Vite `base: './'`

**React + React Flow:**
- Node/edge data: `any[]` (cast from `unknown`)
- WebSocket: `useWebSocket` hook with queue-based state
- Message types: `graph_update`, `node_update`, `edge_update`, `llm_chunk`, `monologue`

**Cobra CLI:**
- Commands: `cmd/nforge/{command}.go`, subcommands via `rootCmd.AddCommand()`
- Config: `viper.GetString("key")`, `viper.Set("key", value)`

**Vite:**
- `base: './'` + `outDir: 'dist'` — consumed by Go `embed.FS`

### Testing Rules

**Go:**
- `go test ./...` + testify assertions, `*_test.go` co-located
- Table-driven tests preferred, `go.uber.org/mock` for mocks
- CGO required for SQLite tests

**TypeScript:**
- Vitest (not Jest!) + @testing-library/react, `*.test.tsx` co-located
- `npx vitest run` (CI) or `npx vitest` (watch)
- TypeScript checks: `npx tsc --noEmit`

**Boundaries:**
- Unit: Individual functions/components; Integration (Go): Real BadgerDB/SQLite
- WebSocket: Mock `gorilla/websocket.Conn` via interfaces

### Code Quality & Style Rules

**ESLint (Frontend):**
- `eslint . --ext ts,tsx --max-warnings 0` (zero warnings policy)
- Plugins: `react-hooks`, `react-refresh`, `@typescript-eslint`

**Go:**
- `go fmt ./...` before commits, package-level comments only
- Sentinel errors: `var ErrNotFound = errors.New("not found")`

**Organization:**
- Go: `types.go`, `manager.go`, `handler.go`, `*_test.go` per package
- TS: `*.tsx` (components), `*.ts` (utilities), `workers/` (Web Workers), `hooks/` (hooks)

### Development Workflow Rules

**Git:**
- Stories: `{epic#}-{story#}-{slug}`, commits: imperative mood
- Status: `backlog` → `ready-for-dev` → `in-progress` → `review` → `done`

**Build:**
- Go: `go build -o nforge ./cmd/nforge/` or `make build`
- Frontend: `cd frontend && npm run build` → `frontend/dist/`
- Windows: `GOOS=windows GOARCH=amd64 go build -o nforge.exe`

**Dev Servers:**
- Frontend: `cd frontend && npm run dev` (Vite HMR, :5173)
- Backend: `go run ./cmd/nforge/ serve` (`NFORGE_VERBOSE=true` for debug)

**Deploy:**
- Single binary: `./nforge serve` (API + WS + frontend via embed.FS)
- Docker: `docker build -t nforge:latest .` (multi-arch)

### Critical Don't-Miss Rules

**Anti-Patterns [AVOID]:**
- [AVOID] Go: `func Create_Node()` or `node_type := "goal"` (snake_case in Go)
- [AVOID] TypeScript: `export const goal_node: NodeType` (snake_case in TS)
- [AVOID] Production: Serve frontend via filesystem — always use `embed.FS`
- [AVOID] Plugins: Direct function calls — use gRPC Unix socket + chroot
- [AVOID] Security: Skip eBPF syscall filter — block exec, mount, reboot
- [AVOID] JSON tags: `` `json:"session_id"` `` — use `` `json:"sessionId"` ``
- [AVOID] Testing: Jest for frontend — use Vitest
- [AVOID] Data: Swap SQLite/BadgerDB roles — SQLite for sessions, Badger for graph
- [AVOID] Gin: Return errors directly — use `c.JSON()`, `c.AbortWithStatusJSON()`
- [AVOID] Vite: `base: '/'` — must be `base: './'` for embed.FS`

**Edge Cases:**
- WebSocket reconnection (`useWebSocket` queue-based state)
- LLM fallback chain: Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter
- React Flow 100+ nodes: Web Worker layout offload
- BadgerDB: Close transactions properly, handle conflicts
- SQLite: WAL mode for concurrent writes
- Chroot jail: Validate paths to prevent escape

**Security (MUST):**
- Chroot jail per session, eBPF syscall filter, Argon2 + AES-256-GCM
- Rate limiting (token bucket), Ed25519 graph signing
- Never include API keys in session exports
- Plugins: gRPC + chroot + resource limits

**Performance:**
- WebSocket: Queue-based state (prevent overwrite), <50ms latency, 5000+ connections
- React Flow: Web Worker for 100+ nodes (60fps)
- BadgerDB: 30%+ token reduction via graph traversal
- LLM Race mode: Fastest token wins, cancel slower providers

---

## Usage Guidelines

**For AI Agents:**
- Read this file before implementing any code
- Follow ALL rules exactly as documented
- When in doubt, prefer the more restrictive option
- Update this file if new patterns emerge

**For Humans:**
- Keep this file lean and focused on agent needs
- Update when technology stack changes
- Review periodically for outdated rules
- Remove rules that become obvious over time

Last Updated: 2026-05-02
