# Story 3.2: ChatPanel & MonologuePanel

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want a ChatPanel (320px) for goal input that generates graphs and a MonologuePanel streaming LLM thoughts,
So that I can describe goals and watch AI thinking in real-time.

## Acceptance Criteria

1. **Given** the ChatPanel and MonologuePanel components are implemented
   **When** the user types a goal in the ChatPanel and presses Enter
   **Then** the graph is generated from chat input and displayed on canvas (FR1, UX-DR29: "Chat-First, Canvas-Second" pattern)

2. **Given** the MonologuePanel is implemented
   **When** a node is executing and the LLM is processing
   **Then** the MonologuePanel slides in from right (400px wide) with streaming LLM Chain-of-Thought tokens (UX-DR3, FR13)

3. **Given** the MonologuePanel is open during node execution
   **When** new tokens arrive via WebSocket
   **Then** auto-scroll keeps latest token visible, monologue history is exportable as Markdown, and panel is toggleable via 'm' key (UX-DR3, UX-DR28: "flight recorder" pattern)

4. **Given** the ChatPanel is open and user submits a goal
   **When** the backend is processing the goal to generate a graph
   **Then** ChatPanel shows "Generating graph..." state with animated ellipsis and disabled input during generation (UX-DR4)

5. **Given** both panels are implemented
   **When** the user interacts with the system
   **Then** the "Chat-First, Canvas-Second" novel UX pattern is evident: chat generates graph, canvas becomes monitor/controller (not builder), familiar metaphor: "Chat is the spec, Canvas is the execution dashboard" (UX-DR29)

6. **Given** the MonologuePanel is implemented
   **When** LLM tokens are streaming
   **Then** the "flight recorder" pattern is implemented: side panel streams thinking in real-time, saves history for debugging (UX-DR28)

## Tasks / Subtasks

- [ ] Task 1: ChatPanel implementation (AC: #1, #4, #5)
  - [ ] Subtask 1.1: Implement goal input validation (min 10 chars, max 500 chars)
  - [ ] Subtask 1.2: Implement "Generating graph..." state with animated ellipsis and disabled input
  - [ ] Subtask 1.3: Wire ChatPanel to WebSocket goal submission (sendMessage({ type: 'goal', text }))
  - [ ] Subtask 1.4: Implement collapsible panel with toggle button
  - [ ] Subtask 1.5: Display chat message history (user + system messages)
  - [ ] Subtask 1.6: Ensure "Chat-First, Canvas-Second" pattern: chat generates graph, canvas visualizes execution

- [ ] Task 2: MonologuePanel implementation (AC: #2, #3, #6)
  - [ ] Subtask 2.1: Implement Radix Dialog slide-over from right (400px wide)
  - [ ] Subtask 2.2: Stream LLM Chain-of-Thought tokens via WebSocket (monologue messages)
  - [ ] Subtask 2.3: Implement auto-scroll with toggle switch (default: on)
  - [ ] Subtask 2.4: Implement export monologue history as Markdown file
  - [ ] Subtask 2.5: Implement clear history functionality
  - [ ] Subtask 2.6: Wire 'm' key toggle via useKeyboardShortcuts hook
  - [ ] Subtask 2.7: Show streaming indicator (pulsing dot) during LLM processing
  - [ ] Subtask 2.8: Implement empty state: "LLM thoughts will appear here during graph execution..."
  - [ ] Subtask 2.9: Show last 100 messages with timestamps, truncate older messages

- [ ] Task 3: Integration with App.tsx (AC: #1, #2, #3, #5, #6)
  - [ ] Subtask 3.1: Integrate ChatPanel with useWebSocket hook (sendMessage, graphUpdateQueue)
  - [ ] Subtask 3.2: Integrate MonologuePanel with useWebSocket hook (monologueMessages, isStreaming)
  - [ ] Subtask 3.3: Wire keyboard shortcuts: 'm' for MonologuePanel, Enter for ChatPanel submit
  - [ ] Subtask 3.4: Show connection status indicator (connected/disconnected)
  - [ ] Subtask 3.5: Implement notification system for success/error messages

- [ ] Task 4: Testing (All ACs)
  - [ ] Subtask 4.1: Unit tests for ChatPanel (validation, submit, generating state)
  - [ ] Subtask 4.2: Unit tests for MonologuePanel (export, clear, auto-scroll, keyboard toggle)
  - [ ] Subtask 4.3: Integration test for goal submission → graph generation flow
  - [ ] Subtask 4.4: Accessibility test: ARIA labels, keyboard navigation, screen reader announcements

## Dev Notes

- **Architecture patterns:** React + TypeScript with Radix UI Primitives for accessibility-first components
- **State management:** React useState/useEffect in components, useWebSocket hook for real-time updates
- **WebSocket message types:** `goal` (send), `monologue` (receive), `graph_update` (receive)
- **Styling:** Tailwind CSS + CSS variables (--bg-primary, --text-secondary, etc.)
- **Animation:** CSS @keyframes for pulsing dots and slide transitions (300ms ease)
- **Export utility:** exportMonologueAsMarkdown function in utils/monologue-export.ts

### Project Structure Notes

- **Frontend files to create/modify:**
  - `frontend/src/components/panels/ChatPanel.tsx` (exists, needs verification)
  - `frontend/src/components/panels/monologue-panel.tsx` (exists, needs verification)
  - `frontend/src/App.tsx` (exists, already integrated)
  - `frontend/src/hooks/useWebSocket.ts` (exists, provides monologueMessages, sendMessage)
  - `frontend/src/hooks/useKeyboardShortcuts.ts` (exists, handles 'm' key)
  - `frontend/src/utils/monologue-export.ts` (exists, export utility)

- **Alignment with unified project structure:**
  - Components in `frontend/src/components/panels/` (kebab-case.tsx files)
  - Hooks in `frontend/src/hooks/` (camelCase.ts files)
  - Utils in `frontend/src/utils/` (camelCase.ts files)
  - Radix UI Primitives for accessibility (WCAG 2.1 AA compliance)

- **Detected conflicts or variances:**
  - None detected: existing implementation aligns with architecture specs
  - ChatPanel already has: input validation, generating state, animated ellipsis, collapsible panel
  - MonologuePanel already has: Radix Dialog, auto-scroll, export, clear, 'm' key toggle, streaming indicator, empty state

### References

- [Story 3.2 in Epics](_bmad-output/planning-artifacts/epics.md#story-32-chatpanel--monologuepanel) — User story and acceptance criteria
- [UX-DR3: MonologuePanel](_bmad-output/planning-artifacts/ux-design-specification.md#3-monologuepaneltsx-custom-input--radix-components) — Panel specs: 400px wide, slide from right, streaming tokens
- [UX-DR4: ChatPanel](_bmad-output/planning-artifacts/ux-design-specification.md#5-chatpaneltsx-custom-input--radix-components) — Panel specs: 320px wide, goal input, generating state
- [UX-DR28: Flight Recorder Pattern](_bmad-output/planning-artifacts/ux-design-specification.md#4-llm-inner-monologue-panelnovel-for-spec-driven-tools) — Streaming thoughts, save history for debugging
- [UX-DR29: Chat-First, Canvas-Second](_bmad-output/planning-artifacts/ux-design-specification.md#1-chat-first-canvas-second-novel-combination) — Novel UX pattern: chat generates graph, canvas monitors
- [Architecture: Frontend](_bmad-output/planning-artifacts/architecture.md#frontend-architecture) — React + Vite + @xyflow/react, Radix UI, Tailwind CSS
- [Architecture: WebSocket API](_bmad-output/planning-artifacts/architecture.md#api--communication-patterns) — Message types: goal, monologue, graph_update, node_update, edge_update
- [PRD: FR1](_bmad-output/planning-artifacts/prd.md#fr1-user-can-create-a-new-session-with-a-goal-description) — Auto-generated node graph from goal
- [PRD: FR13](_bmad-output/planning-artifacts/prd.md#fr13-user-can-see-llm-inner-monologue-chain-of-thought-in-a-side-panel-with-streaming-tokens) — LLM Inner Monologue streaming

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
