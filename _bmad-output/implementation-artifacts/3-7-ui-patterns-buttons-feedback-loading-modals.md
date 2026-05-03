# Story 3.7: UI Patterns - Buttons, Feedback, Loading, Modals

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want consistent button hierarchy, clear feedback patterns, proper loading states, and modal/overlay patterns,
so that the UI communicates effectively and follows accessibility standards.

## Acceptance Criteria

1. **Given** UI pattern components are implemented with Radix UI Primitives
   **When** any action button is displayed
   **Then** button hierarchy is applied: Primary=Cyan (`#06b6d4`), Secondary=Gray outline, Danger=Red (`#ef4444`), Icon-only=32x32px with Radix Tooltip (UX-DR19)
   **And** button variants match: `variant="default"` (Primary), `variant="outline"` (Secondary), `variant="destructive"` (Danger), Icon-only (32px + Tooltip)

2. **Given** feedback patterns are implemented
   **When** the system needs to communicate status
   **Then** Success feedback shows green toast (3s auto-dismiss), Error feedback shows red persistent toast, Warning shows yellow pause label, Info shows cyan edge pulse (UX-DR20)
   **And** toasts use Radix Toast with auto-dismiss (success) or persistent (error) behavior, edge pulses use animated dash flow (cyan `#06b6d4`)

3. **Given** loading states are implemented
   **When** any component enters a loading state
   **Then** Node shows yellow border pulse (300ms animation, `#FFC107`), Edge shows animated dash flow (cyan, 2px stroke), Panel shows skeleton lines (60% opacity pulse) (UX-DR21)
   **And** skeleton uses Radix Skeleton with 60% opacity pulse animation, node border pulse uses CSS `animation: pulse 300ms`

4. **Given** modal/overlay patterns are implemented
   **When** the system needs to show modal content
   **Then** use Radix AlertDialog for confirmations, Radix Dialog for config panels, custom slide-over for monologue panel (UX-DR22)
   **And** AlertDialog for destructive actions (fork, delete), Dialog for node config/settings, custom slide-over (400px right) for monologue panel with `m` key toggle

## Tasks / Subtasks

- [ ] Task 1 (AC: 1) — Button hierarchy component
  - [ ] Subtask 1.1: Install Radix UI Button + Slot dependencies: `@radix-ui/react-button@^1.1.15`, `@radix-ui/react-slot@^1.1.0`, `clsx@^2.1.1` (for className merging)
  - [ ] Subtask 1.2: Create `frontend/src/components/ui/button.tsx` with variants: `default` (Primary, Cyan `#06b6d4`), `outline` (Secondary, gray border), `destructive` (Danger, `#ef4444`), `ghost` (minimal)
  - [ ] Subtask 1.3: Implement icon-only button (32x32px) with Radix Tooltip (`@radix-ui/react-tooltip@^1.1.12`) — `aria-label` required
  - [ ] Subtask 1.4: Apply Tailwind classes: `font-medium`, `rounded-lg`, `focus-visible:outline-2`, `disabled:opacity-50` — follow TypeScript `kebab-case.tsx` naming

- [ ] Task 2 (AC: 2) — Feedback patterns (Toast + Edge pulse)
  - [ ] Subtask 2.1: Install Radix UI Toast: `@radix-ui/react-toast@^1.2.15` — wrap `App.tsx` with `<ToastProvider>`
  - [ ] Subtask 2.2: Create `frontend/src/components/ui/toast.tsx` with variants: `success` (green, 3s auto-dismiss), `destructive` (red, persistent), `warning` (yellow, pause label), `info` (cyan, edge pulse)
  - [ ] Subtask 2.3: Implement `useToast()` hook for programmatic toast triggering (success/error/warning/info)
  - [ ] Subtask 2.4: Edge animated dash flow: CSS `@keyframes dash { to { stroke-dashoffset: -20; } }` on cyan (`#06b6d4`) 2px stroke

- [ ] Task 3 (AC: 3) — Loading states (Skeleton + Node pulse)
  - [ ] Subtask 3.1: Install Radix UI Skeleton: `@radix-ui/react-skeleton@^1.1.10` (if available) or use Tailwind `animate-pulse` with 60% opacity
  - [ ] Subtask 3.2: Create `frontend/src/components/ui/skeleton.tsx` — 60% opacity pulse animation, used for panel loading states (MonologuePanel, SessionExplorer)
  - [ ] Subtask 3.3: Node yellow border pulse: CSS `animation: pulse 300ms infinite alternate` on `#FFC107` border, triggered by `status: 'running'`
  - [ ] Subtask 3.4: Integrate with existing `NodeTypes.tsx` (from story 3.1) — add `running` state with yellow pulse

- [ ] Task 4 (AC: 4) — Modal/Overlay patterns (Dialog + AlertDialog + Slide-over)
  - [ ] Subtask 4.1: Install Radix UI Dialog + AlertDialog: `@radix-ui/react-dialog@^1.1.15`, `@radix-ui/react-alert-dialog@^1.1.15`
  - [ ] Subtask 4.2: Create `frontend/src/components/ui/dialog.tsx` — Radix Dialog for NodeConfig, Settings; AlertDialog for confirmations (fork, delete)
  - [ ] Subtask 4.3: Implement custom slide-over for MonologuePanel (`frontend/src/components/panels/MonologuePanel.tsx` from story 3.2) — 400px wide, right slide, toggle via `m` key, `Escape` to close
  - [ ] Subtask 4.4: Apply ARIA: `aria-label` on dialog triggers, `aria-live` regions for toast announcements, focus trap inside dialogs (Radix built-in)

- [ ] Task 5 — Accessibility & Testing
  - [ ] Subtask 5.1: WCAG 2.1 AA compliance: 4.5:1 contrast ratio (verify with Tailwind `text-gray-*` classes), keyboard navigation (Tab, Enter, Escape, `m` key)
  - [ ] Subtask 5.2: Screen reader support: ARIA live regions for toasts, node status changes; test with NVDA/VoiceOver
  - [ ] Subtask 5.3: Write tests: `button.test.tsx`, `toast.test.tsx`, `dialog.test.tsx` using Vitest + @testing-library/react (not Jest!)

## Dev Notes

### Project Structure Notes

- **Alignment with unified project structure**: Components go in `frontend/src/components/ui/` (shared UI), panels in `frontend/src/components/panels/`
- **Naming conventions** (from `project-context.md`):
  - Files: `kebab-case.tsx` (e.g., `button.tsx`, `toast.tsx`)
  - Components: `PascalCase` (e.g., `Button`, `ToastProvider`)
  - Variables/functions: `camelCase` (e.g., `showToast`, `handleClick`)
  - CSS classes: `kebab-case` (Tailwind utility classes)
- **No conflicts detected**: Story 3.5 (Design System Foundation) defines Tailwind config tokens; this story consumes them. Story 3.6 (Accessibility) defines high-contrast/RTL; this story follows WCAG 2.1 AA patterns.

### Architecture Constraints

- **Technology stack** (from `architecture.md`):
  - Radix UI Primitives (unstyled, accessible) + Tailwind CSS (utility-first)
  - Versions: `@radix-ui/react-button@^1.1.15`, `@radix-ui/react-toast@^1.2.15`, `@radix-ui/react-dialog@^1.1.15`, `tailwindcss@^3.0`
  - React 18+ with TypeScript 5.3.3, Vite 5.0.12 build
- **Code organization** (from `project-context.md`):
  - Co-locate tests: `button.tsx` + `button.test.tsx` in same directory
  - Use `clsx` or `tailwind-merge` for conditional className merging
  - NEVER use Jest! Use Vitest (`npx vitest run`) + @testing-library/react

### Source Tree Components to Touch

| File | Action | Description |
|------|--------|-------------|
| `frontend/src/components/ui/button.tsx` | NEW | Button component with hierarchy variants |
| `frontend/src/components/ui/button.test.tsx` | NEW | Vitest tests for Button |
| `frontend/src/components/ui/toast.tsx` | NEW | Toast feedback component + useToast hook |
| `frontend/src/components/ui/toast.test.tsx` | NEW | Vitest tests for Toast |
| `frontend/src/components/ui/skeleton.tsx` | NEW | Skeleton loading component |
| `frontend/src/components/ui/dialog.tsx` | NEW | Dialog + AlertDialog components |
| `frontend/src/components/panels/MonologuePanel.tsx` | UPDATE | Add custom slide-over (from story 3.2) |
| `frontend/src/components/canvas/NodeTypes.tsx` | UPDATE | Add yellow pulse for running state (from story 3.1) |
| `frontend/src/App.tsx` | UPDATE | Wrap with `<ToastProvider>` from Radix |
| `tailwind.config.js` | REFERENCE | Design tokens from story 3.5 (colors.semantic.primary = '#06b6d4') |

### Testing Standards Summary

- **Framework**: Vitest (not Jest!) + @testing-library/react
- **File pattern**: `*.test.tsx` co-located with component
- **Commands**: `npx vitest run` (CI) or `npx vitest` (watch)
- **TypeScript checks**: `npx tsc --noEmit` (strict mode, no implicit any)
- **Accessibility tests**: Use `axe-core` or manual NVDA/VoiceOver testing per WCAG 2.1 AA

## References

- [Epics Story 3.7](_bmad-output/planning-artifacts/epics.md#Story-3.7) — Original story definition with ACs
- [UX Design UX-DR19-22](_bmad-output/planning-artifacts/ux-design-specification.md#UX-DR19) — Button hierarchy, feedback, loading, modal patterns
- [Architecture UI Patterns](_bmad-output/planning-artifacts/architecture.md#Frontend-Architecture) — Radix + Tailwind stack, component structure
- [Project Context](_bmad-output/project-context.md#Technology-Stack) — Naming conventions, testing rules (Vitest not Jest!)
- [Radix Button Docs](https://www.radix-ui.com/docs/primitives/components/button) — Accessible button primitives
- [Radix Toast Docs](https://www.radix-ui.com/docs/primitives/components/toast) — Toast notification system
- [Radix Dialog Docs](https://www.radix-ui.com/docs/primitives/components/dialog) — Modal dialog patterns
- [Tailwind CSS](https://tailwindcss.com/docs) — Utility-first styling for all components

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

