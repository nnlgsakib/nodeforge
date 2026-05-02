# Story 2.3: LLM Inner Monologue Panel

Status: ready-for-dev

<!-- Validation: Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to see LLM "Inner Monologue" (Chain-of-Thought) streaming in a side panel,
So that I can understand why the AI made certain decisions.

## Acceptance Criteria

**Given** the MonologuePanel component is implemented in React
**When** a node is executing and the LLM is processing
**Then** the Chain-of-Thought tokens stream in real-time to the MonologuePanel via WebSocket (FR13)
**And** the panel is collapsible (400px wide slide-over from right), toggleable via 'm' key (UX-DR3)
**And** monologue history is saved and exportable for debugging
**And** auto-scroll keeps the latest token visible

## Tasks / Subtasks

- [ ] Task 1: Implement MonologuePanel React component (AC: 1, 2, 3, 4)
  - [ ] Subtask 1.1: Create `monologue-panel.tsx` with collapsible slide-over (400px right), Radix Dialog base, include [Clear] [Export] [Auto-scroll ✓] buttons (UX-DR3)
  - [ ] Subtask 1.2: Implement WebSocket listener for LLM Chain-of-Thought tokens (message types: `llm_chunk`, `monologue`) via existing Gin WS hub
  - [ ] Subtask 1.3: Add auto-scroll to keep latest token visible (enabled by default, UX-DR3)
  - [ ] Subtask 1.4: Implement toggle via 'm' key (UX-DR3, UX-DR18)

- [ ] Task 2: Save and export monologue history (AC: 3)
  - [ ] Subtask 2.1: Save monologue history to session state (BadgerDB via `internal/context/`)
  - [ ] Subtask 2.2: Implement export functionality (download as markdown file per UX-DR3)

- [ ] Task 3: Integrate with backend WebSocket hub (AC: 1)
  - [ ] Subtask 3.1: Add MonologuePanel to `frontend/src/App.tsx` layout (right side, collapsible)
  - [ ] Subtask 3.2: Ensure real-time streaming via existing WebSocket connection (`useWebSocket` hook)

## Dev Notes

### Architecture Patterns and Constraints

**Frontend Stack:** Vite + React + @xyflow/react (TypeScript) + Tailwind CSS + Radix UI Primitives.

**WebSocket:** Gin WebSocket hub (/ws) for real-time token streaming (FR13, NFR-01).

**Panel Pattern:** Collapsible slide-over from right (400px) — custom implementation or Radix Dialog (UX-DR3, UX-DR22).

**Keyboard Shortcut:** 'm' key toggles panel visibility (UX-DR3, UX-DR18: one-key controls).

**Flight Recorder Pattern:** Side panel streams thinking in real-time, saves history for debugging (UX-DR28).

### Source Tree Components to Touch

**New Files (CREATE):**
- `frontend/src/components/panels/monologue-panel.tsx` — Main panel component (kebab-case filename, PascalCase component: `MonologuePanel`)
- `frontend/src/hooks/use-monologue.ts` — WebSocket hook for monologue tokens (kebab-case filename)
- `frontend/src/utils/monologue-export.ts` — Export history as markdown (per UX-DR3)

**Modified Files (UPDATE):**
- `frontend/src/App.tsx` — Integrate MonologuePanel into layout
- `frontend/src/hooks/useKeyboardShortcuts.ts` — Add 'm' key handler

### Naming Conventions (CRITICAL — Must Follow):
- TypeScript files: `kebab-case.tsx` — `monologue-panel.tsx`
- TypeScript components: `PascalCase` — `MonologuePanel`
- TypeScript variables: `camelCase` — `monologueHistory`
- JSON fields: `camelCase` — `{"tokens": [...]}`

### Testing Standards
- **TypeScript**: Vitest + React Testing Library — co-located `*.test.tsx` files.

## References

- [Story 2.3 Definition: epics.md#Story-2.3](_bmad-output/planning-artifacts/epics.md#Story-2.3)
- [UX-DR3: MonologuePanel Spec](_bmad-output/planning-artifacts/ux-design-specification.md#UX-DR3)
- [UX-DR28: Flight Recorder Pattern](_bmad-output/planning-artifacts/ux-design-specification.md#UX-DR28)
- [FR13: LLM Inner Monologue Requirement](_bmad-output/planning-artifacts/epics.md#FR13)
- [UX-DR18: One-Key Controls](_bmad-output/planning-artifacts/epics.md#UX-DR18)

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List

## Change Log

## Review Findings
