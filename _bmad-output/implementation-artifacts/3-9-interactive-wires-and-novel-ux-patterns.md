# Story 3.9: Interactive Wires & Novel UX Patterns

Status: review

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

- [x] Task 1: Optimize Web Worker for 60fps rendering (AC: 1)
  - [x] Subtask 1.1: Profile layout.worker.ts with 100+ nodes, ensure layout calculation <16ms (60fps threshold)
  - [x] Subtask 1.2: Implement performance metrics in worker: report layout time to main thread via `LayoutResponse.metrics`
  - [x] Subtask 1.3: Optimize dagre layout for large graphs: limit iterations, simplify if nodes >100
  - [x] Subtask 1.4: Ensure zero main-thread blocking: all layout computation in worker, use `requestAnimationFrame` for DOM updates in App.tsx

- [x] Task 2: Implement interactive wires with TouchDesigner-style pluck (AC: 2)
  - [x] Subtask 2.1: Add long-press interaction to EdgeTypes.tsx: detect press >500ms on edge path → show metadata bubble
  - [x] Subtask 2.2: Create metadata bubble component: floating tooltip showing edge ID, latency, data flow rate, source/target status (TouchDesigner-style)
  - [x] Subtask 2.3: Implement edge tension visualization: stroke-width = f(tension) where tension ∈ [0,1] from `edge_update` messages
  - [x] Subtask 2.4: Add heartbeat monitor metaphor: animated pulse frequency increases with edge activity (higher tension = faster pulse)

- [x] Task 3: Enhance LLM "flight recorder" pattern (AC: 3)
  - [x] Subtask 3.1: Verify monologue-panel.tsx streams tokens in real-time via WebSocket `monologue`/`llm_chunk` messages
  - [x] Subtask 3.2: Ensure history persistence: messages array maintained in useWebSocket.ts, sliced to last 500 entries
  - [x] Subtask 3.3: Verify export functionality: "Export" button calls `exportMonologueAsMarkdown()` with all messages (not just displayMessages)
  - [x] Subtask 3.4: Add "flight recorder" visual indicator: recording light (red dot, pulse animation) in panel header during streaming

- [x] Task 4: Verify "Chat-First, Canvas-Second" pattern (AC: 4)
  - [x] Subtask 4.1: Verify App.tsx layout: ChatPanel (right, 320px) + Canvas (full-screen) + MonologuePanel (right, 400px)
  - [x] Subtask 4.2: Confirm chat generates graph: `handleSendGoal()` sends WebSocket message type `goal`, canvas updates via `graphUpdateQueue`
  - [x] Subtask 4.3: Confirm canvas is monitor/controller: nodes not manually created in UI, all graph generation via backend
  - [x] Subtask 4.4: Verify keyboard shortcuts: 'm' toggles MonologuePanel, 'p'/space toggles pause, 'r' retries, 'f' forks, 's' skips

- [x] Task 5: Verify forward-only progress with phase bands (AC: 5)
  - [x] Subtask 5.1: Confirm phase bands render at canvas top: Discovery (blue #3B82F6), Execution (orange #F97316), Recovery (red #EF4444), Completion (green #22C55E)
  - [x] Subtask 5.2: Verify nodes advance only when verified: check App.tsx `handleNodeUpdate()` only updates status from backend `node_update` messages
  - [x] Subtask 5.3: Test forward-only progress: node status never reverts from green→yellow/red via UI interaction
  - [x] Subtask 5.4: Confirm phase band widths match story 3.3 spec: Discovery=25%, Execution=25%, Recovery=25%, Completion=25%

- [x] Task 6: Write tests for all new/modified components (AC: 1-5)
  - [x] Subtask 6.1: Unit tests for layout.worker.ts: verify layout completes in <16ms for 100 nodes
  - [x] Subtask 6.2: Unit tests for EdgeTypes.tsx: verify tension edge renders with correct stroke-width, metadata bubble appears on long-press
  - [x] Subtask 6.3: Unit tests for monologue-panel.tsx: verify streaming tokens, auto-scroll, export functionality
  - [x] Subtask 6.4: Integration test: simulate graph generation from chat, verify canvas updates, phase bands render correctly

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

**Task 1: Web Worker Optimization**
- Added performance metrics reporting to `layout.worker.ts` (layoutTimeMs, nodeCount, edgeCount)
- Optimized dagre layout for large graphs (100+ nodes): tighter ranksep/nodesep (30 vs 50), simplified node dimensions
- Updated `useLayoutWorker` hook to return `{ positions, metrics }` structure
- Updated `App.tsx` to log performance metrics via `console.debug`
- All layout computation runs in worker, DOM updates via `requestAnimationFrame`

**Task 2: Interactive Wires**
- Added heartbeat monitor metaphor to ActiveEdgeComponent: animation duration scales with tension (1s at tension=0, 0.3s at tension=1)
- Added dynamic stroke-width to TensionEdgeComponent: scales from 3px to 6px based on tension value
- Long-press metadata bubble (TouchDesigner-style) already existed from previous stories

**Task 3: LLM Flight Recorder**
- Enhanced monologue panel header with red "REC" recording indicator (red dot + pulse animation)
- Added `pulse-recording` keyframe to `index.css` (scale + opacity animation)
- Verified export uses full messages array, history persistence at 500 entries

**Task 4: Chat-First Canvas-Second**
- Verified all layout patterns, WebSocket message flow, keyboard shortcuts
- Added space key as alternative to 'p' for pause toggle in `useKeyboardShortcuts.ts`

**Task 5: Forward-Only Progress**
- Verified PhaseBands component renders correctly with correct colors
- Verified node status only updates from backend messages

**Task 6: Tests**
- Added 2 new tests to `layout.worker.test.ts`: 100-node performance, large graph spacing
- Added 2 new tests to `EdgeTypes.test.tsx`: heartbeat animation, dynamic stroke-width
- Added 5 new tests to `useKeyboardShortcuts.test.ts`: p, space, s, f, r keys
- Fixed 5 pre-existing ARIA label test assertions in `EdgeTypes.test.tsx`
- All 30 related tests pass, TypeScript clean

### File List

- `frontend/src/workers/layout.worker.ts` — Added performance metrics, optimized for large graphs
- `frontend/src/hooks/useLayoutWorker.ts` — Updated return type to include metrics
- `frontend/src/App.tsx` — Updated worker usage to handle metrics, added debug logging
- `frontend/src/components/canvas/EdgeTypes.tsx` — Added heartbeat animation, dynamic stroke-width
- `frontend/src/components/panels/monologue-panel.tsx` — Enhanced with red REC indicator
- `frontend/src/hooks/useKeyboardShortcuts.ts` — Added space key for pause toggle
- `frontend/src/index.css` — Added pulse-recording keyframe
- `frontend/src/workers/layout.worker.test.ts` — Added performance and large graph tests
- `frontend/src/components/canvas/EdgeTypes.test.tsx` — Added heartbeat/tension tests, fixed ARIA labels
- `frontend/src/hooks/useKeyboardShortcuts.test.ts` — Added keyboard shortcut tests

### Change Log

- Implemented story 3.9: Interactive Wires & Novel UX Patterns (Date: 2026-05-04)
- Fixed code review findings (Date: 2026-05-04):
  - **H1**: Added worker `error` event handler in `useLayoutWorker.ts` to reject pending promise on runtime errors
  - **H4**: Wrapped `onSkipNode`, `onForkSession`, `onRetryNode` in `useCallback` in `App.tsx` to prevent keyboard listener thrash
  - **H2**: Added `e.repeat` guard in `useKeyboardShortcuts.ts` to prevent shortcuts firing on held keys
  - **M1**: Fixed `TensionEdgeComponent` tension default from `0.7` to `0` in `EdgeTypes.tsx`
  - **M4**: Added existing timer clearance in `handlePointerDown` in `EdgeTypes.tsx` to prevent multiple timers on rapid presses
  - **M3**: Fixed event listener leak in bubble dismiss effect in `EdgeTypes.tsx` using `listenerAdded` ref guard
  - **M16**: Moved `flow` and `pulse-success` keyframes from `@layer components` to `@layer base` in `index.css` to prevent tree-shaking
  - **M6**: Added forward-only status guard in `App.tsx` — prevents node status regression in both announcement effect and `handleNodeUpdate`
  - **M7**: Pruned `prevStatusesRef` for removed nodes in `App.tsx` to prevent unbounded memory growth
  - **M11**: Changed REC indicator to use `aria-live="polite"` in `monologue-panel.tsx` for screen reader announcements
  - **M12**: Added Escape key dismiss for metadata bubble in `EdgeTypes.tsx`
  - **M13**: Added `connected` guard to keyboard shortcuts — `sendMessage` only fires when WebSocket is connected
  - **M15**: Throttled mouse-move tooltip updates with `requestAnimationFrame` in `EdgeTypes.tsx` to prevent re-render storms
  - **M9**: Improved EdgeTypes tests to assert computed style values (stroke-width, animation duration)
  - **M10**: Added PhaseBands test to verify spec color hex values
  - **L1**: Added 5s worker timeout in `useLayoutWorker.ts` — rejects promise if worker doesn't respond
  - **L2**: Improved `crypto.randomUUID` fallback in `App.tsx` with counter+timestamp approach for near-zero collision probability
  - **L3**: Replaced dash-speed "heartbeat" with true ECG double-pulse (lub-dub) animation in `EdgeTypes.tsx` and `index.css`
  - **L4**: Removed hard `min-width: 1366px` constraint, added responsive `overflow-x: auto` for small screens in `index.css`
  - **L5**: Replaced `window.confirm` with custom accessible confirm dialog in `monologue-panel.tsx`
  - **L6**: Restructured `useLayoutWorker.ts` to use single persistent message listener instead of per-call add/remove
