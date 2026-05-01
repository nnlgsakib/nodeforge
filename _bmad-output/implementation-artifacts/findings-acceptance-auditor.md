## Acceptance Auditor Findings

### Acceptance Criteria Check

**AC1: `nforge new <project-name>` CLI command creates workspace with `.nforge/` structure**
- ✅ `cmd/nforge/new.go` — Cobra command with `Use: "new <project-name>"` present
- ✅ `internal/session/manager.go` — `CreateSessionWithName` present, creates project dir and calls `InitProjectWorkspace`
- ✅ `internal/session/workspace.go` — `InitProjectWorkspace` creates `.nforge/` with config.yaml, README.md, .gitignore
- **Finding:** `--workspace-dir` flag documented in spec (Subtask 1.1) but `runNewProject` passes `newWorkspaceDir` to function, yet `CreateSessionWithName` doesn't accept workspace dir — manager's `workspaceRoot` is set at creation time (`NewManager(workspaceDir)`)

**AC2: UI "New Project" button and chat panel create project workspace**
- ✅ `frontend/src/components/panels/SessionExplorer.tsx` — "New Project" button present, calls `onCreateProject`
- ✅ `frontend/src/components/panels/ChatPanel.tsx` — Parses "new <project-name>" from chat input
- ✅ `frontend/src/App.tsx` — `createProject()` POSTs to `/api/v1/sessions`
- ✅ `internal/canvas/api.go` — `POST /api/v1/sessions` endpoint registered and functional
- **Finding:** Frontend always sends `workspaceDir: '.'` (hardcoded in `createProject`), ignoring any CLI `--workspace-dir` equivalent for UI

**AC3: FR30 CLI/UI parity — both interfaces create identical project structures**
- ✅ Both paths use `mgr.CreateSessionWithName()` → `InitProjectWorkspace()` → identical `.nforge/` structure
- **PASS** — parity satisfied

### Spec Constraint Violations

- **[Workspace dir not passed through API]** — Violates: Spec Task 2.3 "Call REST API POST /api/v1/sessions to create session with specified name" — `SessionRequest.WorkspaceDir` is parsed (`binding:"required"`) but **never used** in `createSession` handler. `CreateSessionWithName` doesn't accept workspace dir. The `workspaceDir` field in the request is dead data.

- **[generateSessionID uses PID]** — Violates: Spec "Naming Conventions: JSON fields: camelCase" is followed, but session ID generation (`sess-<pid>`) is not unique-safe. Spec references session management patterns that imply durable session identifiers.

- **[created_at empty in config.yaml]** — Violates: Spec Task 1.3 "Initialize `.nforge/` directory structure" — `config.yaml` has `created_at: ""` (empty). Should be populated with actual timestamp per FR23.

- **[new_test.go only tests happy path]** — Violates: Spec "Testing Standards: Go: Ginkgo + Testify" — Only one test exists, only happy path, no tests for `--workspace-dir`, `--template` flags, no error cases.

- **[Frontend components not in kebab-case filenames]** — Violates: Spec "Naming Conventions: TypeScript: kebab-case.tsx files (chat-panel.tsx), PascalCase components (ChatPanel)" — Files are `ChatPanel.tsx` and `SessionExplorer.tsx` (PascalCase filenames), not kebab-case as specified.

- **[API route registration via canvas.RegisterAPIRoutes]** — Spec says `internal/canvas/api.go` should have POST /api/v1/sessions. Current code: `serve.go` calls `canvas.RegisterAPIRoutes(r, sessionMgr)` which registers routes. This is **correct** — routes defined in `api.go`, registered from `serve.go`. (Previous review finding about "wrong file" was stale.)

### Missing from Diff (Spec-Required)

- **No `getEnvBool` or environment variable support** — Spec references `root.go` with env var patterns, but no env var handling in new code (acceptable if deferred to Story 1.5)
- **No `--template` flag implementation logic** — `newTemplate` variable exists and `--template` flag is defined in `init()`, but `runNewProject` receives `template` and **ignores it** — always uses "default" template path but never actually applies template

### Test Coverage Gaps

- `manager_test.go` — Tests `CreateSessionWithName` happy path only
- `workspace_test.go` — Tests `InitProjectWorkspace` happy path only
- `new_test.go` — Tests `runNewProject` happy path only
- **No tests for:** invalid project names, permission errors, existing `.nforge/` directory, concurrent creation, `--workspace-dir` flag, `--template` flag
- **No frontend tests** — Spec requires "Vitest + React Testing Library (co-located `*.test.tsx` files)" — no test files for `ChatPanel.tsx` or `SessionExplorer.tsx`
