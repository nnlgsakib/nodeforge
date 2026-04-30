# Story 1.4: Project Creation via CLI & UI

Status: ready-for-dev

<!-- Validation: Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to create new projects with `nforge new <project-name>` via CLI and from the UI,
so that I can quickly scaffold a workspace for my goals from either interface.

## Acceptance Criteria

**Given** the CLI and UI are functional
**When** the user runs `nforge new <project-name>` via CLI
**Then** a new session workspace is created with the specified project name, initialized with a `.nforge/` directory structure

**And** from the UI, the user can click "New Project" button (or type a project name in the chat panel) to create a new project workspace

**And** FR30 (CLI/UI parity) is satisfied: both interfaces create identical project structures

## Tasks / Subtasks

- [ ] Task 1: Implement `nforge new <project-name>` CLI command (AC: 1,2,3,5)
  - [ ] Subtask 1.1: Create `cmd/nforge/new.go` with Cobra command (flags: --workspace-dir, --template)
  - [ ] Subtask 1.2: Implement session creation via `internal/session/manager.go` (method: CreateSessionWithName)
  - [ ] Subtask 1.3: Initialize `.nforge/` directory structure with workspace files (config.yaml, README.md, .gitignore)
  - [ ] Subtask 1.4: Verify CLI creates identical project structure as UI

- [ ] Task 2: Implement UI "New Project" button and chat panel integration (AC: 1,4,5)
  - [ ] Subtask 2.1: Add "New Project" button to ChatPanel header or SessionExplorer panel
  - [ ] Subtask 2.2: Implement chat panel input handling for project creation (parse "new <project-name>" or button click)
  - [ ] Subtask 2.3: Call REST API `POST /api/v1/sessions` to create session with specified name
  - [ ] Subtask 2.4: Verify UI creates identical project structure as CLI

- [ ] Task 3: Ensure CLI/UI parity (FR30) (AC:5)
  - [ ] Subtask 3.1: Verify both interfaces create same `.nforge/` directory structure (config.yaml, workspace files)
  - [ ] Subtask 3.2: Test that both create identical session workspace (session ID format, metadata)
  - [ ] Subtask 3.3: Document parity verification in Dev Notes

## Dev Notes

### Architecture Patterns and Constraints

**Cobra CLI Framework:**
- `cmd/nforge/new.go` — Cobra subcommand for `nforge new <project-name>` with persistent flags from root command — [Source: architecture.md#API-Communication-Patterns]

**Session Management:**
- `internal/session/manager.go` — Session creation with custom name, workspace initialization — [Source: architecture.md#Project-Structure]
- `internal/session/workspace.go` — Chroot jail setup, `.nforge/` directory structure — [Source: prd.md#FR31]

**Frontend Integration:**
- `frontend/src/components/panels/ChatPanel.tsx` — Handle project creation input, call API — [Source: ux-design-specification.md#Journey-1-Alex]
- `frontend/src/components/panels/SessionExplorer.tsx` — "New Project" button — [Source: architecture.md#Frontend-Architecture]

**API Endpoint:**
- `POST /api/v1/sessions` — Create session with name parameter — [Source: architecture.md#API-Communication-Patterns]

**CLI/UI Parity (FR30):**
- Both interfaces MUST create identical project structures (same `.nforge/` directory, same config.yaml format) — [Source: prd.md#FR30]

### Source Tree Components to Touch

**New Files (CREATE):**
- `cmd/nforge/new.go` — Cobra command for `nforge new <project-name>` with flags
- `frontend/src/components/panels/SessionExplorer.tsx` — Add "New Project" button (if not exists from Story 1.1)
- `frontend/src/components/panels/ChatPanel.tsx` — Handle project creation input

**Updated Files (UPDATE):**
- `internal/session/manager.go` — Add `CreateSessionWithName(ctx, name string) (*Session, error)` method
- `internal/session/workspace.go` — Add `InitProjectWorkspace(name string) error` to initialize `.nforge/` structure
- `internal/canvas/api.go` — Add `POST /api/v1/sessions` handler (if not exists)

**Naming Conventions (CRITICAL — Must Follow):**
- Go: `snake_case` packages (`internal/session/`), `camelCase` functions (`createSessionWithName`), `PascalCase` structs (`type Session struct`)
- TypeScript: `kebab-case.tsx` files (`chat-panel.tsx`), `PascalCase` components (`ChatPanel`)
- API endpoints: `snake_case` plural (`/api/v1/sessions`)
- JSON fields: `camelCase` (`{"sessionId": "...", "projectName": "..."}`)
— [Source: architecture.md#Naming-Patterns]

### Testing Standards

- **Go**: Ginkgo + Testify (co-located `*_test.go` files) — test `CreateSessionWithName`, `InitProjectWorkspace`
- **TypeScript**: Vitest + React Testing Library (co-located `*.test.tsx` files) — test "New Project" button, chat input handling
— [Source: architecture.md#Starter-Template-Evaluation]

## Project Structure Notes

### Alignment with Unified Project Structure

The implementation must follow the architecture specification:

```
cmd/nforge/
├── root.go           # Cobra root command (persistent flags)
├── new.go           # nforge new <project-name> (THIS STORY)
├── serve.go         # nforge serve (Story 1.3)
└── ...

internal/session/
├── manager.go       # Session CRUD (add CreateSessionWithName)
├── workspace.go     # Workspace init (add InitProjectWorkspace)
└── ...

frontend/src/components/
├── panels/
│   ├── ChatPanel.tsx      # Handle project creation input
│   └── SessionExplorer.tsx # "New Project" button
└── ...
```

— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**None identified** — This story builds on Story 1.1 (Project Scaffolding) which established the directory structure. Session management is defined in `internal/session/` and frontend components in `frontend/src/components/`.

**Critical Reminder:** Both CLI and UI must create identical project structures to satisfy FR30 (CLI/UI parity). The `.nforge/` directory must be identical regardless of interface used.

## References

- [Story 1.4 Definition: epics.md#Story-1.4](_bmad-output/planning-artifacts/epics.md#Story-1.4)
- [Architecture Decisions: architecture.md#Project-Structure](_bmad-output/planning-artifacts/architecture.md#Project-Structure)
- [API Patterns: architecture.md#API-Communication-Patterns](_bmad-output/planning-artifacts/architecture.md#API-Communication-Patterns)
- [Frontend Architecture: architecture.md#Frontend-Architecture](_bmad-output/planning-artifacts/architecture.md#Frontend-Architecture)
- [CLI/UI Parity: prd.md#FR30](_bmad-output/planning-artifacts/prd.md#FR30)
- [Project Creation: prd.md#FR23](_bmad-output/planning-artifacts/prd.md#FR23)
- [Chat-First Experience: ux-design-specification.md#Journey-1-Alex](_bmad-output/planning-artifacts/ux-design-specification.md#Journey-1-Alex)
- [Session Management: ux-design-specification.md#Journey-2-Sam](_bmad-output/planning-artifacts/ux-design-specification.md#Journey-2-Sam)
- [Previous Story 1.1: 1-1-project-scaffolding-and-module-init.md](_bmad-output/implementation-artifacts/1-1-project-scaffolding-and-module-init.md)

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List
