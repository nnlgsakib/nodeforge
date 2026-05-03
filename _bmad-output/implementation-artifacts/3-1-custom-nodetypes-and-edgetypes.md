# Story 3.1: Custom NodeTypes & EdgeTypes

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want custom NodeTypes with n8n/TouchDesigner/DaVinci hybrid visuals and EdgeTypes with reactive tension,
so that I can see a professional canvas with interactive wires, input/output pins, color-coded phase bands, and pluckable edge metadata.

## Acceptance Criteria

1. **Given** React Flow is set up with custom node/edge components
   **When** the graph engine creates nodes (Goal, Spec, Plan, Implement, Test, Review)
   **Then** each node type renders with correct visuals: Goal (green rounded rect), Spec (blue diamond), Plan (purple rect), Implement (orange rect), Test (yellow rounded rect), Review (cyan rect) (UX-DR1)

2. **And** edges show reactive tension (stroke-width based on upstream health), animated pulses during execution, and pluckable metadata bubbles (TouchDesigner-style) (UX-DR2, FR4, FR49)

3. **And** nodes have clear input/output pins (DaVinci-style) and color-coded phase bands across canvas top: blue (Discovery), orange (Execution), red (Recovery), green (Completion) (FR5, FR48)

4. **And** user can drag-and-drop files onto canvas to auto-create nodes (e.g., `go.mod` → `Setup` node) (FR47)

## Tasks / Subtasks

- [ ] Create custom NodeTypes.tsx with Goal, Spec, Plan, Implement, Test, Review node components (AC: 1)
  - [ ] Define node type interfaces in frontend/src/types/nodes.ts
  - [ ] Implement GoalNode (green rounded rect, input pin) per architecture NodeTypes.tsx
  - [ ] Implement SpecNode (blue diamond, input/output pins) per architecture NodeTypes.tsx
  - [ ] Implement PlanNode (purple rect, input/output pins) per architecture NodeTypes.tsx
  - [ ] Implement ImplementNode (orange rect, input/output pins) per architecture NodeTypes.tsx
  - [ ] Implement TestNode (yellow rounded rect, input/output pins) per architecture NodeTypes.tsx
  - [ ] Implement ReviewNode (cyan rect, input/output pins) per architecture NodeTypes.tsx
  - [ ] Add ARIA labels, keyboard nav, screen reader support per accessibility requirements

- [ ] Create custom EdgeTypes.tsx with reactive tension and animated pulses (AC: 2)
  - [ ] Define edge type interface in frontend/src/types/edges.ts
  - [ ] Implement default edge (gray, 2px stroke)
  - [ ] Implement active edge (cyan, animated dash flow during execution)
  - [ ] Implement tension edge (red, 4px stroke, upstream failure)
  - [ ] Implement success edge (green, brief pulse on completion)
  - [ ] Add hover tooltip with edge metadata (latency, data flow rate)
  - [ ] Add long-press (TouchDesigner-style) for metadata bubble
  - [ ] Add ARIA labels, keyboard nav for accessibility

- [ ] Implement phase bands across canvas top (AC: 3)
  - [ ] Add color-coded phase bands: blue (Discovery), orange (Execution), red (Recovery), green (Completion)
  - [ ] Integrate with React Flow canvas layout
  - [ ] Ensure phase bands are visible in all zoom levels

- [ ] Implement drag-and-drop file to node creation (AC: 4)
  - [ ] Add HTML5 drag-and-drop event handlers to canvas
  - [ ] Map file types to node types (e.g., go.mod → Setup node)
  - [ ] Integrate with graph engine to add new nodes programmatically

## Dev Notes

### Relevant Architecture Patterns and Constraints

- **Frontend Stack**: React + Vite + @xyflow/react (^12.10.0) + TypeScript 5.3.3 + Tailwind CSS + Radix UI Primitives
- **Component Structure**: `frontend/src/components/canvas/NodeTypes.tsx`, `EdgeTypes.tsx`, `CanvasControls.tsx` (per architecture lines 276-281)
- **Node Types**: Goal (green rounded rect), Spec (blue diamond), Plan (purple rect), Implement (orange rect), Test (yellow rounded rect), Review (cyan rect) (per epics 3.1 AC)
- **Edge Types**: default (gray), active (cyan, animated), tension (red, thick), success (green) (per UX design section 934-967)
- **Accessibility**: WCAG 2.1 AA compliance, ARIA live regions, screen reader support, Vim/Emacs keybindings (hjkl, Ctrl-f/b/n/p) (per NFR-20 to 24)
- **Design System**: Tailwind config with custom colors (canvas bg: #1a1b1e, node colors by type, edge states) (per UX design section 276-347)
- **State Management**: React Context for graph state, WebSocket client for real-time updates (per architecture lines 270-273)
- **Web Worker**: Offload graph layout for 100+ node graphs at 60fps (FR55, NFR-02)

### Source Tree Components to Touch

- `frontend/src/components/canvas/NodeTypes.tsx` (NEW - main node components)
- `frontend/src/components/canvas/EdgeTypes.tsx` (NEW - edge components)
- `frontend/src/types/nodes.ts` (NEW - node type definitions)
- `frontend/src/types/edges.ts` (NEW - edge type definitions)
- `frontend/src/App.tsx` (UPDATE - register custom node/edge types with React Flow)
- `frontend/src/workers/layout.worker.ts` (UPDATE - ensure layout supports custom nodes/edges)

### Testing Standards Summary

- **Framework**: Vitest (not Jest!) + @testing-library/react (per project context)
- **File Pattern**: `*.test.tsx` co-located with components (e.g., `NodeTypes.test.tsx`)
- **Test Types**: Unit (individual components), Integration (React Flow canvas with custom nodes/edges)
- **Accessibility Tests**: ARIA labels, keyboard nav, screen reader compatibility (per NFR-20)
- **Performance Tests**: 100+ node graphs render at 60fps with Web Worker offload (NFR-02)

### Project Structure Notes

- **Alignment**: Matches architecture's frontend structure (frontend/src/components/canvas/, types/, workers/)
- **Naming Conventions**:
  - TypeScript files: `kebab-case.tsx` (e.g., `node-types.tsx`)
  - Components: `PascalCase` (e.g., `GoalNode`, `SpecNode`)
  - Variables/Functions: `camelCase` (e.g., `handleNodeClick`)
  - Interfaces: `PascalCase` (e.g., `NodeType`, `EdgeType`)
- **No Conflicts**: Architecture clearly separates `frontend/src/components/canvas/` for React Flow custom components, no variance detected

### References

- [Epic 3 Details: epics.md#Story3.1](_bmad-output/planning-artifacts/epics.md#Story3.1) (lines 532-546)
- [UX Design NodeTypes: ux-design-specification.md#ComponentStrategy](_bmad-output/planning-artifacts/ux-design-specification.md#CustomComponents) (lines 898-967)
- [Architecture Frontend: architecture.md#FrontendArchitecture](_bmad-output/planning-artifacts/architecture.md#FrontendArchitecture) (lines 260-308)
- [Architecture Components: architecture.md#ProjectStructure](_bmad-output/planning-artifacts/architecture.md#CompleteProjectDirectoryStructure) (lines 639-771)
- [Project Context: project-context.md#TechnologyStack](_bmad-output/project-context.md#TechnologyStack) (lines 30-46)
- [Acceptance Criteria Mapping: epics.md#Story3.1](_bmad-output/planning-artifacts/epics.md#Story3.1) (lines 538-546)

## Dev Agent Record

### Agent Model Used

Claude Code (tencent/hy3-preview:free)

### Debug Log References

None yet (pre-development)

### Completion Notes List

None yet (pre-development)

### File List

- `frontend/src/components/canvas/NodeTypes.tsx` (NEW)
- `frontend/src/components/canvas/EdgeTypes.tsx` (NEW)
- `frontend/src/types/nodes.ts` (NEW)
- `frontend/src/types/edges.ts` (NEW)
- `frontend/src/App.tsx` (UPDATE)
- `frontend/src/workers/layout.worker.ts` (UPDATE - if needed)
