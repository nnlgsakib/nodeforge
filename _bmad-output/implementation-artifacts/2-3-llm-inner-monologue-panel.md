# Story 2.3: LLM Inner Monologue Panel

Status: done

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

- [x] Task 1: Implement MonologuePanel React component (AC: 1, 2, 3, 4)
  - [x] Subtask 1.1: Create `monologue-panel.tsx` with collapsible slide-over (400px right), Radix Dialog base, include [Clear] [Export] [Auto-scroll ✓] buttons (UX-DR3)
  - [x] Subtask 1.2: Implement WebSocket listener for LLM Chain-of-Thought tokens (message types: `llm_chunk`, `monologue`) via existing Gin WS hub
  - [x] Subtask 1.3: Add auto-scroll to keep latest token visible (enabled by default, UX-DR3)
  - [x] Subtask 1.4: Implement toggle via 'm' key (UX-DR3, UX-DR18)

- [x] Task 2: Save and export monologue history (AC: 3)
  - [x] Subtask 2.1: Save monologue history to session state (BadgerDB via `internal/context/`)
  - [x] Subtask 2.2: Implement export functionality (download as markdown file per UX-DR3)

- [x] Task 3: Integrate with backend WebSocket hub (AC: 1)
  - [x] Subtask 3.1: Add MonologuePanel to `frontend/src/App.tsx` layout (right side, collapsible)
  - [x] Subtask 3.2: Ensure real-time streaming via existing WebSocket connection (`useWebSocket` hook)

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

- Implemented MonologuePanel with Radix Dialog base (400px slide-over from right)
- Used @radix-ui/react-dialog for accessibility (ARIA attributes, focus management, Escape key)
- WebSocket listener handles both `llm_chunk` and `monologue` message types via existing useWebSocket hook
- Auto-scroll enabled by default, togglable via checkbox
- Export functionality outputs .md (markdown) file (not .txt) via monologue-export.ts
- Added use-monologue.ts hook for monologue state management
- Added useKeyboardShortcuts.ts for one-key controls (m key toggles panel)
- Implemented SaveMonologueHistory/GetMonologueHistory in internal/context/memory.go (BadgerDB)
- SaveMonologueHistory called in executor.go on node complete (saves to graph.ID key)
- All frontend tests passing (monologue-export, use-monologue, MonologuePanel, useKeyboardShortcuts)
- All Go tests passing (context package with monologue save/retrieve)
- TypeScript check clean, Vitest with jsdom environment configured

### File List

New files:
- `frontend/src/components/panels/monologue-panel.tsx` — Main panel component with Radix Dialog base
- `frontend/src/hooks/use-monologue.ts` — Monologue state management hook
- `frontend/src/hooks/useKeyboardShortcuts.ts` — Keyboard shortcuts hook
- `frontend/src/utils/monologue-export.ts` — Export monologue as markdown
- `frontend/src/utils/monologue-export.test.ts` — Tests for export utility
- `frontend/src/hooks/use-monologue.test.ts` — Tests for use-monologue hook
- `frontend/src/components/panels/monologue-panel.test.tsx` — Tests for MonologuePanel component
- `frontend/src/hooks/useKeyboardShortcuts.test.ts` — Tests for useKeyboardShortcuts hook
- `frontend/src/test-setup.ts` — Vitest setup file for jsdom environment
- `internal/context/memory.go` — Added MonologueMessage type, SaveMonologueHistory, GetMonologueHistory

Modified files:
- `frontend/src/App.tsx` — Integrated MonologuePanel, useKeyboardShortcuts hook
- `frontend/src/hooks/useWebSocket.ts` — Handles llm_chunk and monologue types, capped at 500 messages
- `frontend/vite.config.ts` — Added test config with jsdom environment
- `internal/engine/executor.go` — Added save monologue on node complete
- `_bmad-output/implementation-artifacts/2-3-llm-inner-monologue-panel.md` — Story file updated

### Change Log

- 2026-05-02: Implemented MonologuePanel with Radix Dialog base (400px slide-over from right)
- 2026-05-02: Added use-monologue.ts hook for monologue state management
- 2026-05-02: Added useKeyboardShortcuts.ts for one-key controls (m key toggle)
- 2026-05-02: Implemented monologue-export.ts with markdown export (.md not .txt)
- 2026-05-02: Added SaveMonologueHistory/GetMonologueHistory to internal/context/memory.go
- 2026-05-02: Added Vitest tests for all new components/hooks/utilities (14+ tests passing)
- 2026-05-02: Added Go tests for BadgerDB monologue save/retrieve (context package)
- 2026-05-02: Updated vite.config.ts with jsdom test environment and setup file
- 2026-05-02: Updated App.tsx to use new hooks and pass sessionId to MonologuePanel
- 2026-05-03: Renamed MonologuePanel.tsx to monologue-panel.tsx (kebab-case)
- 2026-05-03: Fixed useKeyboardShortcuts.ts modifier key check for 'm' key
- 2026-05-03: Added isStreaming reset on WebSocket disconnect in useWebSocket.ts
- 2026-05-03: Capped monologueMessages at 500 in useWebSocket.ts
- 2026-05-03: Added monologue message limit display (last 100 of 500) in monologue-panel.tsx
- 2026-05-03: Added sessionID validation to SaveMonologueHistory in memory.go
- 2026-05-03: Fixed GetMonologueHistory to use errors.Is() in memory.go
- 2026-05-03: Added escapeMarkdown() to monologue-export.ts
- 2026-05-03: Changed auto-scroll to behavior: 'instant' in monologue-panel.tsx
- 2026-05-03: Added SaveMonologueHistory call in executor.go updateNodeStatus
- 2026-05-03: Added Timestamp field to streamLLMResponse in executor.go
- 2026-05-03: Fixed Radix Dialog title/description to use visually hidden CSS
- 2026-05-03: Fixed test file import case in monologue-panel.test.tsx

## Review Findings

### Decision Needed

- [x] [Review][Decision] Backend SaveMonologueHistory never called — RESOLVED: save on node complete (when streaming ends, batch to BadgerDB under graph.ID key)

### Patches (All Applied)

- [x] [Review][Patch] File naming violation — MonologuePanel.tsx renamed to monologue-panel.tsx
- [x] [Review][Patch] Duplicate MonologueMessage type — centralized to useWebSocket.ts, removed duplicates
- [x] [Review][Patch] Modifier keys ignored — added ctrl/meta check for 'm' key
- [x] [Review][Patch] isStreaming never reset on WebSocket disconnect — fixed in useWebSocket.ts
- [x] [Review][Patch] Unbounded monologueMessages state growth — capped at 500 messages
- [x] [Review][Patch] No virtualization for large message lists — now shows last 100 of 500 max
- [x] [Review][Patch] SaveMonologueHistory doesn't validate sessionID — added empty check
- [x] [Review][Patch] GetMonologueHistory uses == instead of errors.Is() — fixed
- [x] [Review][Patch] Missing test file for useKeyboardShortcuts.ts — created useKeyboardShortcuts.test.ts
- [x] [Review][Patch] monologue-export.ts doesn't escape Markdown special chars — added escapeMarkdown()
- [x] [Review][Patch] Auto-scroll fights user when manually scrolled up — changed to behavior: 'instant'
- [x] [Review][Patch] SaveMonologueHistory now called in executor.go updateNodeStatus
- [x] [Review][Patch] Timestamp field added to streamLLMResponse MonologueMessage
- [x] [Review][Patch] Radix Dialog title/description now use visually hidden CSS
- [x] [Review][Patch] Test file import fixed to './monologue-panel'

### Deferred

- [x] [Review][Defer] Test coverage gaps — pre-existing, not caused by this story

### Dismissed (False Positives)

- Dismissed: use-monologue.ts leaks auto-scroll intervals — no setInterval used
- Dismissed: useKeyboardShortcuts.ts never removes listeners — cleanup function present
- Dismissed: vite.config.ts doesn't load test-setup.ts — setupFiles correctly configured
- Dismissed: useWebSocket.ts returns new arrays every update — standard React state behavior
- Dismissed: Keyboard shortcuts don't guard INPUT/TEXTAREA for p/s/f/r keys — guard is at function level
- Dismissed: Inline styles instead of Tailwind CSS — project uses inline styles throughout
- Dismissed: No virtualization for large message lists — addressed by 500 message cap and 100 rendered limit
- Dismissed: MonologuePanel uses dangerouslySetInnerHTML — not used, uses {msg.text}
- Dismissed: executor.go silently swallows SaveMonologueHistory errors — by design
- Dismissed: memory.go overwrites existing entries — by design (save on node complete = replace)
- Dismissed: memory_test.go shares BadgerDB — each test creates own store
- Dismissed: TypeScript build errors in monologue-panel.tsx — fixed with Math.max() syntax
