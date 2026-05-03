# Story 3.8: Empty States, Search/Filter & Responsive Strategy

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want helpful empty states, powerful search/filter capabilities, and a desktop-optimized responsive strategy,
so that I can find content quickly and work comfortably on my development machine.

## Acceptance Criteria

1. **Given** empty states, search/filter, responsive layout are implemented
   **When** there are no sessions or skills
   **Then** empty states display:
   - 📭 + "Start Chat" button (No Sessions) (UX-DR23)
   - 🔌 + "Browse Marketplace" (No Skills) (UX-DR23)
   - 🕭 + animated ellipsis (Waiting Monologue) (UX-DR23)

2. **Given** search/filter components are implemented
   **When** the user interacts with session or skill lists
   **Then** search/filter works:
   - Session search with status/date filters (UX-DR24)
   - Skill search with category sort (Name/Rating/Installs/Recent) (UX-DR24)

3. **Given** responsive layout is implemented
   **When** the app renders on any screen size ≥1366px
   **Then** responsive strategy is enforced:
   - Desktop-first (1366px+ minimum) (UX-DR25)
   - No mobile/tablet support (UX-DR25)
   - Multi-column layouts (UX-DR25)
   - `min-width: 1366px` enforced via CSS (UX-DR25)

## Tasks / Subtasks

- [ ] Task 1: Implement empty states for all panels (AC: 1)
  - [ ] Subtask 1.1: Create empty state component for SessionExplorer (📭 icon, "No sessions yet", "Start Chat" button)
  - [ ] Subtask 1.2: Create empty state component for SkillMarketplace (🔌 icon, "No Skills Installed", "Browse Marketplace" button)
  - [ ] Subtask 1.3: Create empty state for MonologuePanel (🕭 icon, "Waiting...", animated ellipsis)

- [ ] Task 2: Implement search and filter functionality (AC: 2)
  - [ ] Subtask 2.1: Add search box (Radix Input) to SessionExplorer with real-time filtering
  - [ ] Subtask 2.2: Add status filter (All/Running/Complete/Failed) and date filter (All/Today/Week/Month) to SessionExplorer
  - [ ] Subtask 2.3: Add search box + category filter (All/Installed/Available/Featured) to SkillMarketplace
  - [ ] Subtask 2.4: Add sort options (Name/Rating/Installs/Recent) to SkillMarketplace

- [ ] Task 3: Implement desktop-first responsive strategy (AC: 3)
  - [ ] Subtask 3.1: Enforce `min-width: 1366px` on app container via CSS
  - [ ] Subtask 3.2: Configure Tailwind breakpoints (laptop: 1366px, desktop: 1920px, no mobile/tablet)
  - [ ] Subtask 3.3: Adjust panel widths for laptop (1366px: Chat 280px, Monologue 350px) vs desktop (1920px+: Chat 320px, Monologue 400px)
  - [ ] Subtask 3.4: Reposition mini-map to bottom-left for laptop screens to avoid panel overlap

- [ ] Task 4: Write tests for all new components (AC: 1,2,3)
  - [ ] Subtask 4.1: Unit tests for empty state components
  - [ ] Subtask 4.2: Unit tests for search/filter functionality
  - [ ] Subtask 4.3: Responsive behavior tests (if applicable)

## Dev Notes

- Relevant architecture patterns and constraints:
  - Frontend uses Tailwind CSS + Radix UI Primitives (UX Design Spec Section 276-296)
  - Design tokens defined in Tailwind config (colors, typography, spacing) (UX Design Spec Section 304-316)
  - Desktop-first approach with `min-width: 1366px` (UX Design Spec Section 1322-1356)
  - No mobile/tablet support per UX-DR25

- Source tree components to touch:
  - **UPDATE**: `frontend/src/components/panels/SessionExplorer.tsx` — Add empty state, search/filter
  - **CREATE**: `frontend/src/components/panels/SkillMarketplace.tsx` — New component with empty state, search/filter
  - **UPDATE**: `frontend/src/App.tsx` — Add min-width enforcement, responsive panel widths
  - **UPDATE**: `frontend/src/index.css` — Add min-width: 1366px, Tailwind breakpoint config
  - **CREATE**: `frontend/src/components/panels/SkillMarketplace.test.tsx` — Co-located tests
  - **UPDATE**: `frontend/src/components/panels/SessionExplorer.test.tsx` — Add tests for search/filter

- Testing standards summary:
  - Framework: Jest + React Testing Library (per Architecture.md Section 131)
  - Co-located test files: `*.test.tsx` next to component files
  - Test patterns: Render component, simulate user interactions, assert empty states/search/filter behavior

### Project Structure Notes

- Alignment with unified project structure (paths, modules, naming):
  - Frontend files follow `kebab-case.tsx` naming (TypeScript convention from Architecture.md Section 430-446)
  - Components use `PascalCase` naming (SessionExplorer, SkillMarketplace)
  - Variables/functions use `camelCase` (TypeScript convention)
  - Matches architecture: `frontend/src/components/panels/` for all panel components

- Detected conflicts or variances (with rationale):
  - No conflicts detected — follows established patterns from existing components (SessionExplorer.tsx, ChatPanel.tsx, MonologuePanel.tsx)

### References

- Epic 3 overview: [Source: epics.md#Epic3](_bmad-output/planning-artifacts/epics.md#Epic3) (Lines 528-665)
- Story 3.8 details: [Source: epics.md#Story3.8](_bmad-output/planning-artifacts/epics.md#Story3.8) (Lines 636-648)
- UX Design Spec (Empty States): [Source: ux-design-specification.md#EmptyStates](_bmad-output/planning-artifacts/ux-design-specification.md#EmptyStates) (Lines 1194-1233)
- UX Design Spec (Search/Filter): [Source: ux-design-specification.md#SearchFilter](_bmad-output/planning-artifacts/ux-design-specification.md#SearchFilter) (Lines 1286-1296)
- UX Design Spec (Responsive Strategy): [Source: ux-design-specification.md#ResponsiveDesign](_bmad-output/planning-artifacts/ux-design-specification.md#ResponsiveDesign) (Lines 1322-1356)
- Architecture (Frontend): [Source: architecture.md#FrontendArchitecture](_bmad-output/planning-artifacts/architecture.md#FrontendArchitecture) (Lines 260-307)
- Architecture (Naming Conventions): [Source: architecture.md#NamingPatterns](_bmad-output/planning-artifacts/architecture.md#NamingPatterns) (Lines 429-446)
- UX-DR23: Empty states requirement (epics.md Line 646)
- UX-DR24: Search/filter requirement (epics.md Line 647)
- UX-DR25: Responsive strategy requirement (epics.md Line 648)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
