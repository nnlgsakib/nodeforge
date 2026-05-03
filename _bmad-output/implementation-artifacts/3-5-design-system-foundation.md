# Story 3.5: Design System Foundation

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want a complete design system with Tailwind dark theme, JetBrains Mono typography, WCAG 2.1 AA compliance, and colorblind-friendly design,
so that the UI is professional, accessible, and readable.

## Acceptance Criteria

1. **Given** Tailwind CSS + Radix UI Primitives are configured, **when** the app renders any UI component, **then** design tokens are applied: dark theme (#1a1b1e background), node colors by type, edge states, phase colors (UX-DR10)
2. **And** typography uses JetBrains Mono throughout with compact type scale (h1=1.5rem, body=0.875rem) (UX-DR11)
3. **And** WCAG 2.1 AA compliance is met: 4.5:1 minimum contrast ratio, ARIA live regions for node status changes (UX-DR12, NFR-20)
4. **And** nodes are distinguished by shape + label + position, not just color (red/green alone insufficient for colorblind users) (UX-DR13)

## Tasks / Subtasks

- [ ] Task 1 (AC: 1, 2): Configure Tailwind CSS design tokens
  - [ ] Subtask 1.1: Create/modify `frontend/tailwind.config.js` with color system (canvas bg `#1a1b1e`, node colors by type: goal `#4CAF50`, spec `#2196F3`, plan `#9C27B0`, implement `#FF9800`, test `#FFC107`, review `#00BCD4`; edge states: default `#94a3b8`, active `#06b6d4`, tension `#ef4444`, success `#22c55e`; phase colors: discovery `#3b82f6`, execution `#f97316`, recovery `#ef4444`, completion `#22c55e`)
  - [ ] Subtask 1.2: Configure typography (fontFamily.mono: `['JetBrains Mono', 'Fira Code', 'monospace']`; type scale: h1=1.5rem, h2=1.25rem, h3=1rem, body=0.875rem, small=0.75rem, tiny=0.625rem)
  - [ ] Subtask 1.3: Configure spacing (4px base unit) and desktop-first breakpoints (laptop: 1366px, desktop: 1920px; no mobile/tablet support)
  - [ ] Subtask 1.4: Set up dark theme (default) and high-contrast theme extension (bg: `#000000`, surface: `#1a1a1a`, bright node colors)

- [ ] Task 2 (AC: 3): Implement WCAG 2.1 AA compliance
  - [ ] Subtask 2.1: Add ARIA live regions for node status changes (`<div aria-live="polite">Node X changed to running</div>`, `<div aria-live="assertive">Node failed</div>`)
  - [ ] Subtask 2.2: Ensure 4.5:1 minimum contrast ratio for all text/background combinations (use Tailwind `text-gray-*` classes which meet WCAG standards)
  - [ ] Subtask 2.3: Implement colorblind-friendly design: nodes distinguished by shape (Goal=rounded rect, Spec=diamond, Plan=rect, Implement=rect, Test=rounded rect, Review=rect) + label + position, not just color

- [ ] Task 3 (AC: 1, 3): Set up Radix UI Primitives and shared components
  - [ ] Subtask 3.1: Install Radix UI Primitives (`@radix-ui/react-*`) and Lucide React icons (`lucide-react`)
  - [ ] Subtask 3.2: Create shared UI components in `frontend/src/components/ui/` (Button, Dialog, Tooltip, Toast, Switch, ScrollArea, Separator)
  - [ ] Subtask 3.3: Implement button hierarchy: Primary=Cyan `#06b6d4` (white text), Secondary=Gray outline (`#3a3b40` border, cyan text), Danger=Red `#ef4444` (white text), Icon-only=32x32px with Radix Tooltip

- [ ] Task 4 (AC: 3): Add accessibility features (high-contrast, RTL)
  - [ ] Subtask 4.1: Implement high-contrast theme toggle (Radix Switch in settings, CSS class `.high-contrast` on `<body>`)
  - [ ] Subtask 4.2: Add RTL canvas support (mirror coordinates horizontally, invert text alignment, reposition mini-map to bottom-left)
  - [ ] Subtask 4.3: Set up focus indicators (2px solid cyan `#06b6d4` outline, `:focus-visible` pseudo-class only)

## Dev Notes

### Architecture Patterns and Constraints
- **Frontend Stack**: Vite + React + @xyflow/react + Tailwind CSS + Radix UI Primitives (Source: architecture.md#Frontend-Architecture)
- **Design System Choice**: Custom design system with Tailwind + Radix UI (Source: ux-design-specification.md#Design-System-Choice)
- **Component Strategy**: Custom nodes/edges extend React Flow; shared UI uses Radix (Source: ux-design-specification.md#Component-Strategy)
- **Button Hierarchy**: Primary=Cyan `#06b6d4`, Secondary=Gray outline, Danger=Red `#ef4444` (Source: ux-design-specification.md#Button-Hierarchy)
- **Typography**: JetBrains Mono throughout, compact type scale (Source: ux-design-specification.md#Typography-System)
- **Accessibility**: WCAG 2.1 AA, ARIA live regions, colorblind-friendly (Source: ux-design-specification.md#Accessibility-Considerations)

### Source Tree Components to Touch
- `frontend/tailwind.config.js` (NEW - design tokens configuration)
- `frontend/src/components/ui/` (NEW - shared UI components: Button.tsx, Dialog.tsx, Tooltip.tsx, Toast.tsx, Switch.tsx)
- `frontend/src/components/ui/themes/` (NEW - high-contrast theme CSS, RTL support)
- `frontend/src/App.tsx` (UPDATE - add theme provider, ARIA live regions)
- `frontend/package.json` (UPDATE - add Radix UI and Lucide React dependencies)

### Testing Standards Summary
- **Frontend**: Vitest + React Testing Library (co-located `*.test.tsx` files)
- **Key Tests**: Button rendering/hierarchy, theme toggle, ARIA attributes, contrast compliance (axe DevTools)
- **Accessibility Testing**: NVDA/VoiceOver screen reader testing, keyboard-only navigation (hjkl, Ctrl-f/b/n/p)

### Project Structure Notes
- **Alignment**: Follows unified project structure: `frontend/src/components/ui/` for shared UI (kebab-case files, PascalCase components, camelCase variables) (Source: architecture.md#Naming-Patterns)
- **No Conflicts**: Epic 3 stories are all backlog; no existing UI components to modify yet
- **Co-location**: Tests co-located with components (e.g., `Button.tsx` + `Button.test.tsx`)

### References
- [Source: epics.md#Story-3.5] Story 3.5: Design System Foundation, Acceptance Criteria
- [Source: architecture.md#Frontend-Architecture] Frontend stack, Tailwind + Radix
- [Source: architecture.md#Design-System-Choice] Custom design system with Tailwind + Radix
- [Source: ux-design-specification.md#Design-System-Foundation] Design tokens, typography, spacing
- [Source: ux-design-specification.md#Component-Strategy] Radix components, button hierarchy
- [Source: ux-design-specification.md#Button-Hierarchy] Button types and colors
- [Source: ux-design-specification.md#Typography-System] JetBrains Mono, type scale
- [Source: ux-design-specification.md#Accessibility-Considerations] WCAG 2.1 AA, ARIA, high-contrast, RTL
- [Source: architecture.md#Naming-Patterns] TypeScript naming conventions (kebab-case files, PascalCase components)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
