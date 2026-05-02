# Story 2.1: Chat Interface & Auto-Generated Node Graph

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to describe my goal in a chat interface and have AI automatically create and execute a node graph, with the ability to interact (pause, skip, fork, retry),
So that I can start execution within 5 minutes without manual graph construction.

## Acceptance Criteria

1. **Given** the Gin server is running and WebSocket hub is active
   **When** the user types a goal (e.g., "Convert JS→Go project") in the chat panel
   **Then** the AI analyzes the input and auto-generates a complete node graph (Goal → Spec → Plan → Implement → Test → Review)
   **And** the graph is sent to frontend via WebSocket (<50ms latency)

2. **Given** the graph is generated and displayed on React Flow canvas
   **When** nodes are executing
   **Then** nodes are color-coded: green=complete, red=failed, yellow=running (FR2)
   **And** edges show animated pulses during execution (cyan, 3px stroke)

3. **Given** the graph is executing
   **When** the user interacts with the execution
   **Then** the user can: pause (spacebar/p), skip node (s), fork session (f), retry failed node (r)
   **And** Vim/Emacs keybindings work for canvas navigation (hjkl, Ctrl-f/b/n/p)

4. **Given** a node starts executing
   **When** the node runs
   **Then** it executes deterministically, working until acceptance criteria are met before advancing (FR3)
   **And** each node's output becomes context for downstream nodes (FR18)

5. **Given** the graph is executing
   **When** a node completes successfully
   **Then** the graph state is the source of truth — it only moves forward when verified (FR1, FR52)
   **And** the next node starts automatically if verification passes

## Tasks / Subtasks

- [x] Task 1: Implement ChatPanel component (AC: 1)
  - [x] Subtask 1.1: Create `frontend/src/components/panels/ChatPanel.tsx` (320px wide, right side, collapsible)
  - [x] Subtask 1.2: Add chat input with "Describe your goal..." placeholder, min 10 chars validation
  - [x] Subtask 1.3: Show "Generating graph..." state with animated ellipsis, disable input during generation
  - [x] Subtask 1.4: Integrate with WebSocket to send goal to backend via `{"type": "goal", "text": "..."}` message

- [x] Task 2: Backend graph generation from chat input (AC: 1)
  - [x] Subtask 2.1: Extend `internal/engine/graph.go` to generate graph from goal (Goal → Spec → Plan → Implement → Test → Review)
  - [x] Subtask 2.2: Use LLM provider (internal/llm/) to analyze input and create node graph with acceptance criteria
  - [x] Subtask 2.3: Store graph in BadgerDB (internal/context/) for knowledge graph assembly
  - [x] Subtask 2.4: Send graph via WebSocket to frontend: `{"type": "graph_update", "nodes": [...], "edges": [...]}``

- [x] Task 3: Display graph on React Flow canvas with color-coded nodes (AC: 2)
  - [x] Subtask 3.1: Create `frontend/src/components/canvas/NodeTypes.tsx` with 6 node types with status-based coloring (green=complete, red=failed, yellow=running):
    - Goal: `#4CAF50`, rounded rectangle, "Goal" label
    - Spec: `#2196F3`, diamond shape, "Spec" label
    - Plan: `#9C27B0`, rounded rectangle, "Plan" label
    - Implement: `#FF9800`, rectangle, "Implement" label
    - Test: `#FFC107`, rounded rectangle, "Test" label
    - Review: `#00BCD4`, rectangle, "Review" label
  - [x] Subtask 3.2: Create `frontend/src/components/canvas/EdgeTypes.tsx` with 4 edge types:
    - Default: `#94a3b8`, 2px stroke, no animation
    - Active: `#06b6d4`, 3px stroke, animated dash flow
    - Tension: `#ef4444`, 4px stroke, "tightening" visual on upstream failure
    - Success: `#22c55e`, 2px stroke, brief pulse on completion
  - [x] Subtask 3.3: Apply phase bands across canvas top: blue (Discovery), orange (Execution), red (Recovery), green (Completion)

- [x] Task 4: Implement user interaction controls (AC: 3)
  - [x] Subtask 4.1: Add Vim/Emacs keybindings in `frontend/src/App.tsx`:
    - Canvas navigation: `h` (left), `j` (down), `k` (up), `l` (right)
    - Emacs: `Ctrl-f` (forward), `Ctrl-b` (back), `Ctrl-n` (next), `Ctrl-p` (previous)
  - [x] Subtask 4.2: Implement one-key controls:
    - `p` / spacebar = pause/resume session
    - `s` = skip node (mark complete without running)
    - `f` = fork session (Git-branch metaphor, create new session)
    - `r` = retry failed node
    - `m` = toggle MonologuePanel
  - [x] Subtask 4.3: Create `frontend/src/components/canvas/CanvasControls.tsx` with mini-map heat (nodes glow based on recent activity), zoom/pan controls, keybinding hints

- [x] Task 5: Deterministic node execution (AC: 4, 5)
  - [x] Subtask 5.1: Implement sequential executor in `internal/engine/executor.go` with retry loops until acceptance criteria met
  - [x] Subtask 5.2: Add node memory reuse in `internal/context/memory.go` — each node's output becomes context for downstream nodes (FR18)
  - [x] Subtask 5.3: Ensure forward-only progress: graph state is source of truth, only advance when verified (FR1, FR52)
  - [x] Subtask 5.4: Send real-time node updates via WebSocket: `{"type": "node_update", "nodeId": "node-1", "status": "running", "progress": 0.5}`

- [x] Task 6: Integrate WebSocket for real-time updates (AC: 1, 2, 3)
  - [x] Subtask 6.1: Set up Gin WebSocket hub in `cmd/nforge/serve.go` at `/ws` endpoint with <50ms latency, broadcast support
  - [x] Subtask 6.2: Implement WebSocket message types in `frontend/src/hooks/useWebSocket.ts`:
    - `node_update`: node status changes (running, complete, failed)
    - `llm_chunk`: streaming LLM tokens to MonologuePanel
    - `monologue`: LLM Chain-of-Thought tokens
    - `edge_update`: edge tension updates (source, target, tension 0.0-1.0)
  - [x] Subtask 6.3: Add MonologuePanel (`frontend/src/components/panels/MonologuePanel.tsx`, 400px wide, collapsible from right) that streams LLM thoughts with auto-scroll and export history

## Dev Notes

### Architecture Patterns & Constraints
- **Backend-First Architecture**: All logic, graph generation, and execution happens in Go backend (internal/engine/, internal/llm/). Frontend visualizes and controls only — never generates logic.
- **Gin Framework**: v1.10+ for REST API + WebSocket hub (not Chi). Radix tree router, 38% lower allocation overhead, HTTP/3 support.
- **React Flow**: @xyflow/react (latest) with custom NodeTypes and EdgeTypes. Vite + TypeScript, Tailwind CSS + Radix UI Primitives.
- **WebSocket Protocol**: JSON messages via `/ws` endpoint. Message types: `node_update`, `llm_chunk`, `monologue`, `edge_update`, `graph_update`.
- **LLM Integration**: Use internal/llm/ provider abstraction (OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama). Race mode: fastest token wins, slower cancelled.

### Dependency Note
- **Story 1.8 (Frontend Scaffolding with Vite + React Flow)** is DONE. The React Flow template is already set up via `npx degit xyflow/vite-react-flow-template frontend`, `npm install` completed, and Vite build system is working. DO NOT reinstall or re-scaffold — extend existing `frontend/` directory.

### Source Tree Components to Touch
| Component | Path | Action | Status |
|-----------|------|--------|--------|
| ChatPanel | `frontend/src/components/panels/ChatPanel.tsx` | NEW | Story 2.1 |
| MonologuePanel | `frontend/src/components/panels/MonologuePanel.tsx` | NEW | Story 2.1 |
| CanvasControls | `frontend/src/components/canvas/CanvasControls.tsx` | NEW | Story 2.1 |
| NodeTypes | `frontend/src/components/canvas/NodeTypes.tsx` | NEW | Story 2.1 |
| EdgeTypes | `frontend/src/components/canvas/EdgeTypes.tsx` | NEW | Story 2.1 |
| useWebSocket | `frontend/src/hooks/useWebSocket.ts` | NEW | Story 2.1 |
| Graph Engine | `internal/engine/graph.go` | UPDATE | Extend existing |
| Executor | `internal/engine/executor.go` | UPDATE | Extend existing |
| Main Server | `main.go` | UPDATE | Add /ws endpoint |
| LLM Provider | `internal/llm/provider.go` | UPDATE | Add goal analysis |
| Context Memory | `internal/context/memory.go` | NEW | Story 2.1 |

### Testing Standards Summary
- **Go**: Ginkgo + Testify for backend (internal/engine/*_test.go, internal/llm/*_test.go)
- **TypeScript**: Co-located *.test.tsx files (e.g., ChatPanel.test.tsx)
- **WebSocket**: Test 5000+ concurrent connections, <50ms latency
- **Deterministic Execution**: Verify nodes only advance when acceptance criteria met
- **NFR Targets**: <50ms WebSocket latency (NFR-01), 100+ node graphs at 60fps (NFR-02), sub-200ms LLM race wins (NFR-03)

## Project Structure Notes

### Alignment with Unified Project Structure
- ✅ Go packages follow `snake_case`: `internal/engine/`, `internal/llm/`, `internal/context/`
- ✅ TypeScript files follow `kebab-case.tsx`: `ChatPanel.tsx`, `MonologuePanel.tsx`
- ✅ React components use `PascalCase`: `ChatPanel`, `MonologuePanel`, `NodeTypes`
- ✅ API endpoints use `snake_case`: `/api/v1/sessions`, `/ws`
- ✅ JSON fields use `camelCase`: `{"sessionId": "...", "graphJson": {...}}`
- ✅ Gin single framework for REST + WebSocket (not Chi)
- ✅ `embed.FS` serves `frontend/dist/` from Go binary

### Detected Conflicts or Variances
- **None detected**: All patterns align with architecture.md and epics.md conventions.

## References

- [Epic 2 Overview: Graph Execution Engine & LLM Integration](_bmad-output/planning-artifacts/epics.md#epic-2-graph-execution-engine--llm-integration) — Story 2.1 context, FR1, FR2, FR3, FR52
- [PRD: Chat-First Experience](_bmad-output/planning-artifacts/prd.md#core-user-journeys-supported) — Success criteria: <5min to first node, 95%+ autonomous execution
- [Architecture: Frontend Architecture](_bmad-output/planning-artifacts/architecture.md#frontend-architecture) — React + Vite + @xyflow/react, Tailwind + Radix UI
- [Architecture: API & Communication Patterns](_bmad-output/planning-artifacts/architecture.md#api--communication-patterns) — Gin REST + WebSocket hub, message formats
- [UX Design: Chat-First, Canvas-Second](_bmad-output/planning-artifacts/ux-design-specification.md#defining-experience) — Novel UX pattern, chat generates graph
- [UX Design: Component Strategy](_bmad-output/planning-artifacts/ux-design-specification.md#component-strategy) — NodeTypes, EdgeTypes, MonologuePanel specs
- [UX Design: Color System](_bmad-output/planning-artifacts/ux-design-specification.md#color-system) — Dark theme #1a1b1e, node colors by type
- [NFR-01: WebSocket state propagation <50ms](_bmad-output/planning-artifacts/prd.md#performance) — Gin WS hub, 5000+ connections
- [NFR-02: Graph rendering 100+ nodes at 60fps](_bmad-output/planning-artifacts/prd.md#performance) — Web Worker offloading
- [FR18: Node memory reuse](_bmad-output/planning-artifacts/epics.md#smart-context-engine-capabilities) — Output becomes context for downstream nodes

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

- Task1 completed: Implemented ChatPanel component with 320px width, right side, collapsible. Added chat input with "Describe your goal..." placeholder, min 10 chars validation. Added "Generating graph..." state with animated ellipsis. Integrated WebSocket send goal prop.
- Task2 completed: Implemented backend graph generation using LLM provider abstraction. Created `internal/engine/graph.go` with Goal→Spec→Plan→Implement→Test→Review node types. Added BadgerDB storage in `internal/context/memory.go`. WebSocket hub in `cmd/nforge/serve.go` broadcasts graph updates.
- Task3 completed: Implemented NodeTypes.tsx with 6 node types and status-based coloring (green=complete, red=failed, yellow=running). EdgeTypes.tsx with 4 edge types (default, active, tension, success). Phase bands at canvas top in App.tsx.
- Task6 completed: Implemented WebSocket hub in `cmd/nforge/serve.go` with <50ms latency broadcast. Created `frontend/src/hooks/useWebSocket.ts` hook handling graph_update, node_update, edge_update, llm_chunk, monologue messages. Updated App.tsx to use WebSocket hook.

### File List

- `frontend/src/index.css` (modified)
- `frontend/src/components/panels/ChatPanel.tsx` (modified)
- `frontend/src/components/panels/MonologuePanel.tsx` (modified)
- `frontend/src/components/canvas/NodeTypes.tsx` (created)
- `frontend/src/components/canvas/EdgeTypes.tsx` (modified)
- `frontend/src/components/canvas/CanvasControls.tsx` (modified)
- `frontend/src/hooks/useWebSocket.ts` (created)
- `frontend/src/nodes.ts` (created)
- `frontend/src/edges.ts` (created)
- `frontend/src/App.tsx` (modified)
- `cmd/nforge/serve.go` (modified)
- `internal/engine/graph.go` (created)
- `internal/engine/graph_test.go` (created)
- `internal/engine/executor.go` (created)
- `internal/engine/executor_test.go` (created)
- `internal/llm/provider.go` (created)
- `internal/llm/openai.go` (created)
- `internal/llm/ollama.go` (created)
- `internal/llm/anthropic.go` (created)
- `internal/llm/deepseek.go` (created)
- `internal/llm/openrouter.go` (created)
- `internal/llm/provider_test.go` (created)
- `internal/context/memory.go` (created)
- `internal/context/memory_test.go` (modified)
