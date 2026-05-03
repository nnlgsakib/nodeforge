# Story 3.6: Accessibility - High-Contrast, RTL & Screen Readers

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want high-contrast theme, RTL canvas support, screen reader announcements via ARIA live regions, and full keyboard navigation with one-key controls,
so that NodeForge is accessible to all users.

## Acceptance Criteria

1. **Given** accessibility features are implemented, **when** the user toggles high-contrast mode, **then** canvas shows black background (#000000), white text, bright node colors (Goal=#00ff00, Spec=#00aaff, etc.) (UX-DR14, NFR-21)
2. **And** RTL canvas support inverts coordinates horizontally, adapts text alignment, and mirrors mini-map to bottom-left (UX-DR15, NFR-22)
3. **And** screen reader announces: "Node Goal-1 changed to running" (polite), "Node failed" (assertive) via ARIA live regions (UX-DR16, NFR-20)
4. **And** Vim/Emacs navigation (hjkl, Ctrl-f/b/n/p) works for all interactions without mouse; one-key controls: p=pause/resume, r=retry, f=fork, s=skip (UX-DR17, UX-DR18, FR50, NFR-23, NFR-24)

## Tasks / Subtasks

- [ ] Task 1 (AC: 1): Implement high-contrast theme
  - [ ] Subtask 1.1: Extend Tailwind config with `highContrast` theme variant (bg: `#000000`, surface: `#0a0a0a`, text: `#ffffff`)
  - [ ] Subtask 1.2: Define bright node colors for high-contrast mode (Goal=`#00ff00`, Spec=`#00aaff`, Plan=`#ff00ff`, Implement=`#ff8800`, Test=`#ffff00`, Review=`#00ffff`)
  - [ ] Subtask 1.3: Define bright edge colors for high-contrast (default=`#cccccc`, active=`#00ffff`, tension=`#ff0000`, success=`#00ff00`)
  - [ ] Subtask 1.4: Implement theme toggle mechanism (React Context `ThemeContext` with `isHighContrast: boolean` state)
  - [ ] Subtask 1.5: Apply `body.high-contrast` CSS class when high-contrast mode is active
  - [ ] Subtask 1.6: Ensure all custom node components read theme context and apply bright colors in high-contrast mode
  - [ ] Subtask 1.7: Ensure all custom edge components read theme context and apply bright stroke colors in high-contrast mode
  - [ ] Subtask 1.8: Ensure all panels (ChatPanel, MonologuePanel, SessionExplorer, SkillMarketplace) adapt to high-contrast colors

- [ ] Task 2 (AC: 2): Implement RTL canvas support
  - [ ] Subtask 2.1: Add RTL state to ThemeContext (`isRTL: boolean`)
  - [ ] Subtask 2.2: Invert React Flow canvas horizontal coordinates when `isRTL` is true (transform: `scaleX(-1)` on canvas container, `scaleX(-1)` on each node to re-flip content)
  - [ ] Subtask 2.3: Mirror mini-map position from bottom-right to bottom-left when RTL active
  - [ ] Subtask 2.4: Adapt text alignment: `text-align: right` for all panel text content when RTL active
  - [ ] Subtask 2.5: Flip panel slide-in direction (right panel slides from left, left panel slides from right)
  - [ ] Subtask 2.6: Ensure edge rendering works correctly with RTL-transformed canvas coordinates

- [ ] Task 3 (AC: 3): Implement screen reader announcements via ARIA live regions
  - [ ] Subtask 3.1: Create `AriaAnnouncer` component (`frontend/src/components/ui/aria-announcer.tsx`) with `aria-live="polite"` and `aria-live="assertive"` regions
  - [ ] Subtask 3.2: Create `useAnnounce` hook (`frontend/src/hooks/use-announce.ts`) that queues announcements to the ARIA regions
  - [ ] Subtask 3.3: Announce node status changes via polite region: "Node {nodeId} changed to {status}"
  - [ ] Subtask 3.4: Announce node failures via assertive region: "Node {nodeId} failed: {error}"
  - [ ] Subtask 3.5: Announce graph completion: "All nodes complete! Session completed successfully"
  - [ ] Subtask 3.6: Announce panel open/close: "LLM Monologue Panel opened", "Chat panel closed"
  - [ ] Subtask 3.7: Add `aria-label` and `aria-labelledby` to all interactive elements (nodes, edges, buttons, panels)
  - [ ] Subtask 3.8: Add `role="status"` to node status indicators
  - [ ] Subtask 3.9: Ensure all dialogs use `aria-modal="true"` and proper focus trapping (Radix Dialog handles this)

- [ ] Task 4 (AC: 4): Implement full keyboard navigation with one-key controls
  - [ ] Subtask 4.1: Implement Vim keybindings for canvas navigation (h=left, j=down, k=up, l=right) using `useKeyPress` hook
  - [ ] Subtask 4.2: Implement Emacs keybindings for canvas navigation (Ctrl-f=forward, Ctrl-b=back, Ctrl-n=next, Ctrl-p=previous)
  - [ ] Subtask 4.3: Implement one-key node controls: p=pause/resume, r=retry, f=fork, s=skip, m=toggle monologue, space=pause
  - [ ] Subtask 4.4: Implement node selection via keyboard (Tab cycles through nodes, Enter opens config)
  - [ ] Subtask 4.5: Implement Escape key to close any open panel/dialog
  - [ ] Subtask 4.6: Display active keyboard shortcuts in CanvasControls help overlay
  - [ ] Subtask 4.7: Ensure all interactions are possible without mouse (create, connect, delete, configure nodes via keyboard)
  - [ ] Subtask 4.8: Add visible focus indicators (2px solid cyan `#06b6d4` outline on `:focus-visible`) for keyboard navigation

## Dev Notes

### Architecture Patterns and Constraints
- **Frontend Stack**: Vite + React + @xyflow/react + Tailwind CSS + Radix UI Primitives (Source: architecture.md#Frontend-Architecture)
- **Design System**: Tailwind config with dark theme (#1a1b1e bg) + high-contrast extension (Source: 3-5-design-system-foundation.md)
- **Component Strategy**: Custom nodes/edges extend React Flow; shared UI uses Radix (Source: ux-design-specification.md#Component-Strategy)
- **State Management**: React Context for graph state and theme state (Source: architecture.md#State-Management)
- **Naming**: TypeScript files `kebab-case.tsx`, components `PascalCase`, variables `camelCase` (Source: architecture.md#Naming-Patterns)

### Previous Story Intelligence (from Story 3.5)
- Story 3.5 (Design System Foundation) established: Tailwind config with color tokens, JetBrains Mono typography, Radix UI Primitives installation, button hierarchy, basic ARIA live regions stub, high-contrast theme toggle stub, RTL support stub, focus indicators
- **Files created/modified by 3.5**: `frontend/tailwind.config.js`, `frontend/src/components/ui/` (Button, Dialog, Tooltip, Toast, Switch, ScrollArea, Separator), `frontend/src/App.tsx` (theme provider), `frontend/package.json` (Radix dependencies)
- **Key patterns established**: Theme context provider in App.tsx, Tailwind design tokens, Radix component wrappers
- **Learnings**: 3.5 created the foundation - this story (3.6) builds on top of those stubs to fully implement the accessibility features

### Existing Files to UPDATE (read before implementing)
- `frontend/tailwind.config.js` - Add highContrast theme extension with full color overrides
- `frontend/src/App.tsx` - Add ARIA live regions, theme context provider enhancements
- `frontend/src/components/canvas/NodeTypes.tsx` (from 3-1) - Add high-contrast color support, aria-labels, keyboard selection
- `frontend/src/components/canvas/EdgeTypes.tsx` (from 3-1) - Add high-contrast stroke colors, aria-labels
- `frontend/src/components/canvas/CanvasControls.tsx` (from 3-3) - Add RTL mini-map repositioning, keyboard shortcut display
- `frontend/src/components/panels/ChatPanel.tsx` (from 3-2) - Add high-contrast colors, ARIA labels, RTL text alignment
- `frontend/src/components/panels/MonologuePanel.tsx` (from 3-2) - Add high-contrast colors, ARIA labels, RTL support
- `frontend/src/components/panels/SessionExplorer.tsx` (from 3-3) - Add high-contrast colors, ARIA labels
- `frontend/src/components/ui/Switch.tsx` (from 3-5) - Already has theme toggle support, verify RTL
- `frontend/src/hooks/use-graph-state.ts` (from 3-3) - May need keyboard navigation hooks
- `frontend/src/components/ui/AccessibilityToolbar.tsx` (from 3-4) - Already has high-contrast toggle and RTL switch stubs

### New Files to Create
- `frontend/src/components/ui/aria-announcer.tsx` - ARIA live region component with polite/assertive queues
- `frontend/src/hooks/use-announce.ts` - Hook for queueing screen reader announcements
- `frontend/src/hooks/use-keyboard-nav.ts` - Hook for Vim/Emacs keybindings and one-key controls
- `frontend/src/components/ui/themes/high-contrast.css` - CSS overrides for high-contrast mode (if Tailwind utilities insufficient)
- `frontend/src/components/ui/themes/rtl.css` - CSS overrides for RTL layout

### Accessibility Requirements Detail
- **WCAG 2.1 AA**: 4.5:1 minimum contrast ratio for all text/background (Source: epics.md#NFR-20, ux-design-specification.md#WCAG-2.1-AA-Compliance)
- **Colorblind-friendly**: Nodes distinguished by shape + label + position, not just color (Source: 3-5-design-system-foundation.md AC:4)
- **High-contrast colors** (Source: ux-design-specification.md#High-Contrast-Mode):
  - Canvas bg: `#000000`, surface: `#0a0a0a`, text: `#ffffff`
  - Goal: `#00ff00`, Spec: `#00aaff`, Plan: `#ff00ff`, Implement: `#ff8800`, Test: `#ffff00`, Review: `#00ffff`
  - Edge default: `#cccccc`, active: `#00ffff`, tension: `#ff0000`, success: `#00ff00`
- **RTL support** (Source: ux-design-specification.md#RTL-Support): Canvas scaleX(-1) transform, mini-map bottom-left, text-align right, panels flip slide direction
- **Keyboard shortcuts** (Source: ux-design-specification.md#Keyboard-Navigation):
  - Canvas: h=left, j=down, k=up, l=right, Ctrl-f/b/n/p
  - Node controls: p=pause/resume, r=retry, f=fork, s=skip, m=toggle monologue, space=pause
  - General: Tab=cycle focus, Enter=activate, Escape=close panel

### ARIA Label Patterns (from UX spec)
- Node: `aria-label="Node Goal-1, status: running"`
- Edge: `aria-label="Edge from Spec-1 to Plan-1, status: active"`
- Monologue Panel: `aria-label="LLM Monologue Panel, open"`
- Canvas Controls: `aria-label="Canvas controls, Vim keys: hjkl"`
- Chat Panel: `aria-label="Chat panel, type your goal"`
- Buttons: `aria-label="Retry node Implement-1 (r key)"`

### Testing Standards Summary
- **Frontend**: Vitest + React Testing Library (co-located `*.test.tsx` files)
- **Accessibility Testing**: axe DevTools (Chrome extension), Lighthouse accessibility audit (90%+ target), Pa11y CLI
- **Screen Reader Testing**: NVDA (Windows) - manual testing with actual screen reader
- **Keyboard Testing**: Unplug mouse, use only keyboard for full interaction flow
- **Key Tests**: High-contrast color application, RTL coordinate inversion, ARIA announcement firing, keyboard shortcut handling, focus trap in dialogs

### Project Structure Notes
- **Alignment**: Follows `frontend/src/components/ui/` for shared UI, `frontend/src/components/panels/` for side panels, `frontend/src/hooks/` for hooks (Source: architecture.md#React-Project-Organization)
- **Theme Context**: Already exists from Story 3.5 in `App.tsx` - extend it with `isHighContrast` and `isRTL` toggles
- **No Conflicts**: Stories 3.7, 3.8, 3.9 are backlog - this story only updates components from 3-1 through 3-5

### References
- [Source: epics.md#Story-3.6] Story 3.6: Accessibility - High-Contrast, RTL & Screen Readers, Acceptance Criteria
- [Source: epics.md#FR51] Accessibility support (screen readers, high-contrast, RTL, 20+ languages)
- [Source: epics.md#NFR-20] WCAG 2.1 AA compliance, ARIA live regions
- [Source: epics.md#NFR-21] High-contrast themes, colorblind-friendly palette
- [Source: epics.md#NFR-22] RTL canvas support, 20+ languages
- [Source: epics.md#NFR-23] Vim/Emacs keyboard navigation
- [Source: epics.md#NFR-24] Motor impairment support, all operations via keyboard
- [Source: epics.md#UX-DR14] High-contrast theme toggle
- [Source: epics.md#UX-DR15] RTL canvas support
- [Source: epics.md#UX-DR16] Screen reader announcements via ARIA live regions
- [Source: epics.md#UX-DR17] Vim/Emacs keyboard navigation
- [Source: epics.md#UX-DR18] One-key node controls
- [Source: architecture.md#Frontend-Architecture] React + @xyflow/react, state management via Context
- [Source: architecture.md#Accessibility] WCAG 2.1 AA, ARIA, Vim/Emacs keybindings, RTL
- [Source: ux-design-specification.md#Accessibility-Considerations] WCAG 2.1 AA, high-contrast, RTL, screen reader, keyboard nav
- [Source: ux-design-specification.md#Accessibility-Patterns] Focus indicators, ARIA live regions, keyboard navigation code patterns
- [Source: ux-design-specification.md#High-Contrast-Mode] Tailwind config highContrast extension, colors
- [Source: 3-5-design-system-foundation.md] Design system foundation: Tailwind config, Radix UI, theme provider, basic accessibility stubs

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
