# Story 3.9: Interactive Wires & Novel UX Patterns

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want Web Worker offloading for 60fps rendering, interactive wires as health indicators, LLM "flight recorder" pattern, and forward-only progress with color-coded phase bands,
so that the canvas is fast, informative, and revolutionary.

## Acceptance Criteria

1. **Given** Web Worker and interactive wire components are implemented
   **When** 100+ node graphs render on canvas
   **Then** layout is offloaded to Web Worker (layout.worker.ts) for smooth 60fps rendering with zero main-thread blocking (UX-DR26, FR55, NFR-02, NFR-16)

2. **Given** interactive wire components are implemented
   **When** the user interacts with edges between nodes
   **Then** interactive wires show:
   - TouchDesigner-style pluck for metadata bubble (long-press edge → floating bubble with latency, data flow, status) (UX-DR27, FR4)
   - Edge tension visualization (stroke-width based on upstream health, red/thickening when upstream fails) (FR49)
   - Heartbeat monitor metaphor for edge activity (UX-DR27)

3. **Given** LLM Inner Monologue panel is implemented
   **When** a node executes with LLM processing
   **Then** LLM Inner Monologue acts as "flight recorder":
   - Side panel streams thinking in real-time (existing: `monologue-panel.tsx`)
   - Saves history for debugging (export functionality exists, verify persistence)
   - Auto-scroll with user toggle (UX-DR28)

4. **Given** chat panel and canvas are implemented
   **When** the user interacts with the application
   **Then** "Chat-First, Canvas-Second" novel UX pattern is properly implemented:
   - Chat generates graph, canvas becomes monitor/controller (not builder) (UX-DR29)
   - Familiar metaphor: "Chat is the spec, Canvas is the execution dashboard"
   - ChatPanel (320px narrow) right-side, Canvas full-screen, MonologuePanel (400px) right-side (verify in App.tsx)

5. **Given** phase bands are implemented in App.tsx
   **When** nodes execute on the canvas
   **Then** forward-only progress shows color-coded phase bands:
   - Nodes only advance when verified (UX-DR30, FR3)
   - Blue (Discovery), Orange (Execution), Red (Recovery), Green (Completion) bands across canvas top
   - Git-branch mental model: graph advances like a successful pipeline (FR48)
   - Verify implementation in App.tsx lines 48-53, 269-302

## Tasks / Subtasks

- [ ] Task 1: Optimize Web Worker for 60fps rendering (AC: 1)
  - [ ] Subtask 1.1: Profile layout.worker.ts with 100+ nodes, ensure layout calculation <16ms (60fps threshold)
  - [ ] Subtask 1.2: Implement performance metrics in worker: report layout time to main thread via `LayoutResponse.metrics`
  - [ ] Subtask 1.3: Optimize dagre layout for large graphs: limit iterations, simplify if nodes >100
  - [ ] Subtask 1.4: Ensure zero main-thread blocking: all layout computation in worker, use `requestAnimationFrame` for DOM updates in App.tsx

- [ ] Task 2: Implement interactive wires with TouchDesigner-style pluck (AC: 2)
  - [ ] Subtask 2.1: Add long-press interaction to EdgeTypes.tsx: detect press >500ms on edge path → show metadata bubble
  - [ ] Subtask 2.2: Create metadata bubble component: floating tooltip showing edge ID, latency, data flow rate, source/target status (TouchDesigner-style)
  - [ ] Subtask 2.3: Implement edge tension visualization: stroke-width = f(tension) where tension ∈ [0,1] from `edge_update` messages
  - [ ] Subtask 2.4: Add heartbeat monitor metaphor: animated pulse frequency increases with edge activity (higher tension = faster pulse)

- [ ] Task 3: Enhance LLM "flight recorder" pattern (AC: 3)
  - [ ] Subtask 3.1: Verify monologue-panel.tsx streams tokens in real-time via WebSocket `monologue`/`llm_chunk` messages
  - [ ] Subtask 3.2: Ensure history persistence: messages array maintained in useWebSocket.ts, sliced to last 500 entries
  - [ ] Subtask 3.3: Verify export functionality: "Export" button calls `exportMonologueAsMarkdown()` with all messages (not just displayMessages)
  - [ ] Subtask 3.4: Add "flight recorder" visual indicator: recording light (red dot, pulse animation) in panel header during streaming

- [ ] Task 4: Verify "Chat-First, Canvas-Second" pattern (AC: 4)
  - [ ] Subtask 4.1: Verify App.tsx layout: ChatPanel (right, 320px) + Canvas (full-screen) + MonologuePanel (right, 400px)
  - [ ] Subtask 4.2: Confirm chat generates graph: `handleSendGoal()` sends WebSocket message type `goal`, canvas updates via `graphUpdateQueue`
  - [ ] Subtask 4.3: Confirm canvas is monitor/controller: nodes not manually created in UI, all graph generation via backend
  - [ ] Subtask 4.4: Verify keyboard shortcuts: 'm' toggles MonologuePanel, 'p'/space toggles pause, 'r' retries, 'f' forks, 's' skips

- [ ] Task 5: Verify forward-only progress with phase bands (AC: 5)
  - [ ] Subtask 5.1: Confirm phase bands render at canvas top: Discovery (blue #3B82F6), Execution (orange #F97316), Recovery (red #EF4444), Completion (green #22C55E)
  - [ ] Subtask 5.2: Verify nodes advance only when verified: check App.tsx `handleNodeUpdate()` only updates status from backend `node_update` messages
  - [ ] Subtask 5.3: Test forward-only progress: node status never reverts from green→yellow/red via UI interaction
  - [ ] Subtask 5.4: Confirm phase band widths match story 3.3 spec: Discovery=25%, Execution=25%, Recovery=25%, Completion=25%

- [ ] Task 6: Write tests for all new/modified components (AC: 1-5)
  - [ ] Subtask 6.1: Unit tests for layout.worker.ts: verify layout completes in <16ms for 100 nodes
  - [ ] Subtask 6.2: Unit tests for EdgeTypes.tsx: verify tension edge renders with correct stroke-width, metadata bubble appears on long-press
  - [ ] Subtask 6.3: Unit tests for monologue-panel.tsx: verify streaming tokens, auto-scroll, export functionality
  - [ ] Subtask 6.4: Integration test: simulate graph generation from chat, verify canvas updates, phase bands render correctly

## Dev Notes

- Relevant architecture patterns and constraints:
  - Frontend: React + Vite + @xyflow/react (TypeScript) with Web Worker offloading (Architecture.md Lines 260-307)
  - Web Worker pattern: `layout.worker.ts` uses dagre for graph layout, communicates via `postMessage` (Architecture.md Lines 290-293)
  - Edge types: React Flow custom edges with TouchDesigner-inspired interactive wires (UX Design Spec Lines 934-967)
  - LLM streaming: WebSocket messages `llm_chunk` and `monologue` feed into useWebSocket hook (useWebSocket.ts Lines 74-88)
  - Design tokens: Tailwind config with dark theme (#1a1b1e background), edge colors (default: #94a3b8, active: #06b6d4, tension: #ef4444, success: #22c55e) (UX Design Spec Lines 304-316)
  - Performance: 100+ node graphs at 60fps, zero main-thread blocking (NFR-02, NFR-16)

- Source tree components to touch:
  - **UPDATE**: `frontend/src/workers/layout.worker.ts` — Optimize for 60fps (AC1), add performance metrics
  - **UPDATE**: `frontend/src/components/canvas/EdgeTypes.tsx` — Add TouchDesigner-style pluck interaction, metadata bubble, tension visualization (AC2)
  - **UPDATE**: `frontend/src/components/panels/monologue-panel.tsx` — Enhance "flight recorder" pattern, verify history persistence (AC3)
  - **UPDATE**: `frontend/src/App.tsx` — Verify "Chat-First, Canvas-Second" pattern, verify phase bands (AC4, AC5)
  - **UPDATE**: `frontend/src/hooks/useWebSocket.ts` — Verify message handling for `monologue` and `edge_update` (AC2, AC3)
  - **CREATE**: `frontend/src/components/canvas/EdgeMetadataBubble.tsx` — New component for TouchDesigner-style metadata bubble (AC2)
  - **UPDATE**: `frontend/src/components/canvas/CanvasControls.tsx` — May need refinements for phase band integration (AC5)

- Testing standards summary:
  - Framework: Jest + React Testing Library (Architecture.md Line 131)
  - Co-located test files: `*.test.tsx` next to component files (Architecture.md Line 485)
  - Test patterns: Render component, simulate user interactions, assert edge tension/metadata bubble behavior
  - Performance tests: Measure layout time in worker, assert <16ms for 100 nodes (60fps threshold)

### Project Structure Notes

- Alignment with unified project structure (paths, modules, naming):
  - Frontend files follow `kebab-case.tsx` naming (TypeScript convention from Architecture.md Lines 441-446)
  - Components use `PascalCase` naming (EdgeTypes, MonologuePanel, CanvasControls)
  - Variables/functions use `camelCase` (TypeScript convention)
  - Matches architecture: `frontend/src/components/canvas/` for EdgeTypes, `frontend/src/components/panels/` for MonologuePanel
  - Web Worker in `frontend/src/workers/` (Architecture.md Line 291)

- Detected conflicts or variances (with rationale):
  - No conflicts detected — follows established patterns from existing components (EdgeTypes.tsx, MonologuePanel.tsx, CanvasControls.tsx)
  - Note: Story 3.3 (CanvasControls.tsx) and 3.2 (MonologuePanel.tsx) are in `ready-for-dev` status — this story builds upon those implementations
  - layout.worker.ts already exists from Story 2.7 (Incremental Execution & Web Worker Offloading) — this story optimizes it for 60fps

### References

- Epic 3 overview: [Source: epics.md#Epic3](_bmad-output/planning-artifacts/epics.md#Epic3) (Lines 528-665)
- Story 3.9 details: [Source: epics.md#Story3.9](_bmad-output/planning-artifacts/epics.md#Story3.9) (Lines 650-664)
- UX Design Spec (Interactive Wires): [Source: ux-design-specification.md#InteractiveWires](_bmad-output/planning-artifacts/ux-design-specification.md#InteractiveWires) (Lines 407-431)
- UX Design Spec (Flight Recorder): [Source: ux-design-specification.md#FlightRecorder](_bmad-output/planning-artifacts/ux-design-specification.md#FlightRecorder) (Lines 427-431)
- UX Design Spec (Chat-First Pattern): [Source: ux-design-specification.md#ChatFirst](_bmad-output/planning-artifacts/ux-design-specification.md#ChatFirst) (Lines 409-413)
- UX Design Spec (Phase Bands): [Source: ux-design-specification.md#PhaseBands](_bmad-output/planning-artifacts/ux-design-specification.md#PhaseBands) (Lines 415-419)
- UX Design Spec (Edge Types): [Source: ux-design-specification.md#EdgeTypes](_bmad-output/planning-artifacts/ux-design-specification.md#EdgeTypes) (Lines 934-967)
- Architecture (Frontend): [Source: architecture.md#FrontendArchitecture](_bmad-output/planning-artifacts/architecture.md#FrontendArchitecture) (Lines 260-307)
- Architecture (Naming Conventions): [Source: architecture.md#NamingPatterns](_bmad-output/planning-artifacts/architecture.md#NamingPatterns) (Lines 429-446)
- Architecture (Performance): [Source: architecture.md#Performance](_bmad-output/planning-artifacts/architecture.md#Performance) (Lines 898-914)
- UX-DR26: Web Worker offloading (epics.md Line 660)
- UX-DR27: Interactive wires as health indicators (epics.md Line 661)
- UX-DR28: LLM "flight recorder" pattern (epics.md Line 662)
- UX-DR29: "Chat-First, Canvas-Second" pattern (epics.md Line 663)
- UX-DR30: Forward-only progress with phase bands (epics.md Line 664)
- FR55: Web Worker offloading for 60fps (epics.md Line 55)
- FR4: Interactive node connections n8n-style (epics.md Line 176)
- FR49: Reactive edge tension (epics.md Line 221)
- FR3: Deterministic node execution (epics.md Line 175)
- FR48: Color-coded node bands by lifecycle phase (epics.md Line 220)
- NFR-02: 100+ node graphs at 60fps (epics.md Line 88)
- NFR-16: Smooth interaction with 100+ node graphs (epics.md Line 102)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
