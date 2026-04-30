# Story 1.8: Frontend Scaffolding with Vite + React Flow

Status: ready-for-dev

<!-- Validation: Run validate-create-story for quality check before dev-story. -->

## Story

As a developer,
I want the React + Vite + React Flow frontend scaffolded and served via `embed.FS`,
so that the UI is ready for canvas development.

## Acceptance Criteria

**Given** the `frontend/` directory is set up
**When** the developer runs `npx degit xyflow/vite-react-flow-template frontend` and `npm install`
**Then** the React Flow starter template is cloned with TypeScript support and Vite build system

**And** the built output (`frontend/dist/`) is embeddable via Go's `embed.FS` in `main.go`

**And** the frontend serves a basic React Flow canvas at `http://localhost:8080` when `nforge serve` is running

## Tasks / Subtasks

- [ ] Task 1: Scaffold React frontend with Vite + React Flow (AC: 1)
  - [ ] Subtask 1.1: Run `npx degit xyflow/vite-react-flow-template frontend` (official starter, Vite + TypeScript)
  - [ ] Subtask 1.2: Run `cd frontend && npm install` to install dependencies
  - [ ] Subtask 1.3: Verify TypeScript configuration (`tsconfig.json`) and Vite config (`vite.config.ts`)
  - [ ] Subtask 1.4: Verify React Flow components are available (`@xyflow/react` in package.json)

- [ ] Task 2: Configure frontend build for embed.FS compatibility (AC: 2)
  - [ ] Subtask 2.1: Update `vite.config.ts` to set base: `./` and build.outDir: `dist`
  - [ ] Subtask 2.2: Configure TypeScript for production build (strict mode, no ununsed vars)
  - [ ] Subtask 2.3: Run `npm run build` and verify `frontend/dist/` output exists with index.html
  - [ ] Subtask 2.4: Verify build output is self-contained (no external CDN dependencies)

- [ ] Task 3: Integrate with Go embed.FS and Gin server (AC: 2,3)
  - [ ] Subtask 3.1: Update `main.go` to import `embed` package and define `//go:embed frontend/dist/*`
  - [ ] Subtask 3.2: Serve embedded frontend via Gin static file server or `embed.FS` handler
  - [ ] Subtask 3.3: Verify `nforge serve` starts Gin server and serves React Flow canvas at `http://localhost:8080`
  - [ ] Subtask 3.4: Test that page loads with basic React Flow canvas (empty canvas with no nodes)

- [ ] Task 4: Set up frontend project structure for future development (AC: 1,3)
  - [ ] Subtask 4.1: Create directory structure: `frontend/src/components/{canvas,panels,ui}/`
  - [ ] Subtask 4.2: Create `frontend/src/workers/` directory for Web Worker offloading
  - [ ] Subtask 4.3: Create `frontend/src/types/` directory for TypeScript type definitions
  - [ ] Subtask 4.4: Create `frontend/src/hooks/` directory for custom React hooks

- [ ] Task 5: Verify end-to-end frontend serving (AC: 3)
  - [ ] Subtask 5.1: Run `cd frontend && npm run dev` — verify Vite dev server starts on port 5173
  - [ ] Subtask 5.2: Run `go build -o nforge main.go` — verify binary compiles with embedded frontend
  - [ ] Subtask 5.3: Run `./nforge serve` — verify canvas loads at `http://localhost:8080`
  - [ ] Subtask 5.4: Test that both Vite dev mode and embedded production mode work

## Dev Notes

### Architecture Patterns and Constraints

**Frontend Stack (Non-Negotiable):**
- **Vite + React + @xyflow/react** (TypeScript) — official `vite-react-flow-template` — [Source: architecture.md#Frontend-Architecture]
- **Tailwind CSS + Radix UI Primitives** — accessibility-first, WCAG 2.1 AA compliance — [Source: ux-design-specification.md#Design-System-Choice]
- **React Flow** as base, custom NodeTypes/EdgeTypes for n8n/TouchDesigner/DaVinci hybrid visuals — [Source: ux-design-specification.md#Component-Strategy]

**Build & Deployment:**
- **Go embed.FS** for serving React build from Go binary — `frontend/dist/` embedded in `main.go` — [Source: architecture.md#Starter-Template-Evaluation]
- **Vite Build System** — fast HMR (Hot Module Replacement) in dev, optimized production build — [Source: prd.md#MVP]
- **Multi-stage Docker**: `golang:1.24` builder → `gcr.io/distroless/static-debian12` runtime — [Source: architecture.md#Infrastructure-Deployment]

**Frontend Structure (from Architecture):**
```
frontend/
├── src/
│   ├── components/
│   │   ├── canvas/       # React Flow custom nodes/edges
│   │   │   ├── NodeTypes.tsx
│   │   │   ├── EdgeTypes.tsx
│   │   │   └── CanvasControls.tsx
│   │   ├── panels/        # Side panels
│   │   │   ├── MonologuePanel.tsx
│   │   │   ├── NodeConfig.tsx
│   │   │   └── SessionExplorer.tsx
│   │   └── ui/            # Shared UI components
│   │       ├── themes/
│   │       └── i18n/
│   ├── hooks/             # Custom React hooks
│   ├── workers/           # Web Worker offloading
│   ├── types/             # TypeScript definitions
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── tsconfig.json
├── vite.config.ts
└── index.html
```
— [Source: architecture.md#Complete-Project-Directory-Structure]

### Source Tree Components to Touch

**New Files (CREATE):**
- `frontend/` — Complete React scaffold from `vite-react-flow-template`
- `frontend/src/components/canvas/` — Canvas components (created in later stories)
- `frontend/src/components/panels/` — Panel components (created in later stories)
- `frontend/src/components/ui/` — Shared UI components
- `frontend/src/hooks/` — Custom React hooks
- `frontend/src/workers/` — Web Worker files
- `frontend/src/types/` — TypeScript type definitions
- `main.go` — Update to add `embed.FS` for serving frontend (if not done in Story 1.3)

**Updated Files (UPDATE):**
- `main.go` — Add `//go:embed frontend/dist/*` and embed.FS handler
- `frontend/vite.config.ts` — Configure base path and build output
- `frontend/package.json` — Verify dependencies (@xyflow/react, tailwind, radix-ui)

**Naming Conventions (CRITICAL — Must Follow):**
- TypeScript files: `kebab-case.tsx` — `monologue-panel.tsx`, `node-types.tsx`
- TypeScript components: `PascalCase` — `MonologuePanel`, `NodeTypes`
- TypeScript variables: `camelCase` — `graphData`, `executeNode()`
- API endpoints: `snake_case` plural — `/api/v1/sessions`
- JSON fields: `camelCase` — `{"sessionId": "...", "graphJson": {...}}`
— [Source: architecture.md#Naming-Patterns]

### Testing Standards

- **TypeScript**: Vitest (Vite-native) + React Testing Library — co-located `*.test.tsx` files
- **Build Verification**: `npm run build` succeeds, `frontend/dist/` contains valid index.html
- **Integration Test**: `go build` with embed.FS succeeds, `./nforge serve` serves frontend
— [Source: architecture.md#Starter-Template-Evaluation]

### Previous Story Learnings

**Story 1.3 (Gin Server with nforge serve):**
- Gin server is set up and running on port 8080
- REST API (`/api/v1/*`) and WebSocket hub (`/ws`) are implemented
- Health check available at `/healthz` and metrics at `/metrics`
- Frontend serving via embed.FS should integrate with existing Gin setup
— [Source: 1-3-gin-server-with-nforge-serve.md]

**Story 1.1 (Project Scaffolding):**
- Project structure established: `frontend/` directory at root level
- Go module initialized with Go 1.24+
- Dependencies installed: Gin, Cobra, protobuf
— [Source: 1-1-project-scaffolding-and-module-init.md]

## Project Structure Notes

### Alignment with Unified Project Structure

The frontend directory must exactly match the architecture specification:

```
frontend/
├── src/
│   ├── components/
│   │   ├── canvas/       # NodeTypes.tsx, EdgeTypes.tsx, CanvasControls.tsx
│   │   ├── panels/       # MonologuePanel.tsx, NodeConfig.tsx, SessionExplorer.tsx
│   │   └── ui/           # themes/, i18n/
│   ├── hooks/             # useWebSocket.ts, useGraphState.ts, useSession.ts
│   ├── workers/           # layout.worker.ts
│   ├── types/             # nodes.ts, edges.ts
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── tsconfig.json
├── vite.config.ts
└── index.html
```

— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**None identified** — This story establishes the frontend scaffold that subsequent stories (3.x) will build upon.

**Critical Reminder:** The `frontend/dist/` output MUST be embeddable via Go's `embed.FS`. Configure Vite to output relative paths (base: `./`) so the embedded files work correctly when served from the Go binary.

## References

- [Story 1.8 Definition: epics.md#Story-1.8](_bmad-output/planning-artifacts/epics.md#Story-1.8)
- [Architecture Decisions: architecture.md#Frontend-Architecture](_bmad-output/planning-artifacts/architecture.md#Frontend-Architecture)
- [Technical Requirements: prd.md#Technical-Requirements](_bmad-output/planning-artifacts/prd.md#Technical-Requirements)
- [Design System: ux-design-specification.md#Design-System-Choice](_bmad-output/planning-artifacts/ux-design-specification.md#Design-System-Choice)
- [vite-react-flow-template](https://github.com/xyflow/vite-react-flow-template) — Official React Flow + Vite + TypeScript starter
- [Story 1.3: Gin Server](_bmad-output/implementation-artifacts/1-3-gin-server-with-nforge-serve.md)
- [Story 1.1: Project Scaffolding](_bmad-output/implementation-artifacts/1-1-project-scaffolding-and-module-init.md)
- [Vite Documentation](https://vitejs.dev/config/) — Configuration reference
- [React Flow Documentation](https://reactflow.dev/) — NodeTypes, EdgeTypes, Canvas controls

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List
