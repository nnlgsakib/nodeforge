# Story 3.4: SkillMarketplace & AccessibilityToolbar

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to browse/install skills from a marketplace and toggle accessibility features (high-contrast, RTL, font-size),
so that I can extend NodeForge and adapt it to my needs.

## Acceptance Criteria

1. **Given** SkillMarketplace and AccessibilityToolbar components are implemented, **When** the user opens the SkillMarketplace panel, **Then** skills are displayed in grid layout with rating stars, category filter, and install button (UX-DR8, FR25, FR40)

2. **And** AccessibilityToolbar provides high-contrast toggle, RTL switch, and font-size slider (UX-DR9, FR51, NFR-21, NFR-22)

3. **And** skill dependencies auto-install when installing (FR41), and A/B testing routes to different versions with metrics collection (FR46)

## Tasks / Subtasks

- [ ] Task 1: Implement SkillMarketplace component (AC: 1, 3)
  - [ ] Subtask 1.1: Create `frontend/src/components/panels/SkillMarketplace.tsx` with grid layout
  - [ ] Subtask 1.2: Implement skill card with rating stars, category filter, install button
  - [ ] Subtask 1.3: Backend API endpoint `GET /api/v1/skills` to fetch skills from registry
  - [ ] Subtask 1.4: Backend API endpoint `POST /api/v1/skills/install` to install skill + auto-install dependencies (FR41)
  - [ ] Subtask 1.5: Implement A/B testing routing for skills with metrics collection (FR46)
  - [ ] Subtask 1.6: Integrate with `internal/skills/` package (manifest.go, resolver.go)

- [ ] Task 2: Implement AccessibilityToolbar component (AC: 2)
  - [ ] Subtask 2.1: Create `frontend/src/components/ui/AccessibilityToolbar.tsx` with high-contrast toggle
  - [ ] Subtask 2.2: Implement RTL switch with canvas coordinate inversion
  - [ ] Subtask 2.3: Add font-size slider with JetBrains Mono scaling
  - [ ] Subtask 2.4: Implement high-contrast theme in Tailwind config (`highContrast` section)
  - [ ] Subtask 2.5: Add WCAG 2.1 AA compliance checks (4.5:1 contrast ratio)
  - [ ] Subtask 2.6: Integrate with `frontend/src/components/ui/themes/` and `i18n/` directories

- [ ] Task 3: Integration & Testing (AC: 1, 2, 3)
  - [ ] Subtask 3.1: Connect SkillMarketplace to backend WebSocket for install status updates
  - [ ] Subtask 3.2: Add Prometheus metrics for A/B testing (`/metrics` endpoint, NFR-30)
  - [ ] Subtask 3.3: Write Vitest tests for SkillMarketplace component
  - [ ] Subtask 3.4: Write Vitest tests for AccessibilityToolbar component
  - [ ] Subtask 3.5: Verify WCAG 2.1 AA compliance with axe DevTools

## Dev Notes

### Relevant Architecture Patterns and Constraints

- **Frontend Stack:** React + Vite + @xyflow/react (TypeScript) with Tailwind CSS + Radix UI Primitives
- **Backend Skills Package:** `internal/skills/` with manifest.go, resolver.go, sandbox.go, grpc.go, mcp.go, subnodes.go, abtest.go
- **API Endpoints (Gin):** `GET /api/v1/skills`, `POST /api/v1/skills/install`, `/api/v1/skills/abtest`
- **WebSocket Messages:** `skill_installed`, `skill_install_failed`, `abtest_metrics` message types
- **State Management:** React Context for accessibility settings, session storage for preferences
- **Prometheus Metrics:** A/B test results in `/metrics` endpoint (NFR-30)

### Source Tree Components to Touch

**Frontend (new files):**
- `frontend/src/components/panels/SkillMarketplace.tsx` — Main marketplace panel (Radix Dialog + Tailwind grid)
- `frontend/src/components/ui/AccessibilityToolbar.tsx` — Toolbar with toggles/slider (Radix Switch + Slider)
- `frontend/src/components/ui/themes/` — High-contrast theme definition
- `frontend/src/components/ui/i18n/` — RTL support and 20+ languages

**Frontend (modify existing):**
- `frontend/src/App.tsx` — Add AccessibilityToolbar, wire to WebSocket
- `frontend/src/hooks/useWebSocket.ts` — Add skill install message handlers

**Backend (new files):**
- `internal/skills/abtest.go` — A/B testing framework (FR46)
- `internal/skills/manifest.go` — Skill manifest parsing (if not exists)
- `internal/skills/resolver.go` — Dependency resolution (if not exists)

**Backend (modify existing):**
- `cmd/nforge/skill.go` — Add install endpoint, A/B test routes
- `internal/devops/metrics.go` — Add A/B test Prometheus metrics

### Testing Standards Summary

**Go (Backend):**
- `go test ./...` + testify assertions, `*_test.go` co-located
- Table-driven tests preferred, CGO required for SQLite
- Test skill installation with dependency tree resolution
- Test A/B testing routing and metrics collection

**TypeScript (Frontend):**
- Vitest (not Jest!) + @testing-library/react, `*.test.tsx` co-located
- `npx vitest run` (CI) or `npx vitest` (watch)
- Test SkillMarketplace grid rendering, category filter, install button
- Test AccessibilityToolbar toggles, RTL inversion, font-size slider
- Verify WCAG 2.1 AA with axe DevTools (Chrome extension)

### Key Technical Constraints

- **Skill Dependencies:** Auto-install dependency tree recursively (FR41) — use `internal/skills/resolver.go`
- **A/B Testing:** Route to different skill versions, collect metrics (execution time, success rate, token usage) (FR46)
- **High-Contrast Mode:** Toggle switches canvas to `#000000` background, bright node colors (Goal=`#00ff00`, Spec=`#00aaff`) (NFR-21)
- **RTL Support:** Canvas coordinates invert horizontally, text alignment adapts, mini-map mirrors to bottom-left (NFR-22)
- **Colorblind-Friendly:** Nodes distinguished by shape + label + position, not just color (UX-DR13)
- **WCAG 2.1 AA:** 4.5:1 minimum contrast ratio, ARIA live regions for status changes (NFR-20)

## Project Structure Notes

### Alignment with Unified Project Structure

- **Frontend components:** `kebab-case.tsx` files (SkillMarketplace.tsx, AccessibilityToolbar.tsx)
- **Frontend components:** `PascalCase` component names (SkillMarketplace, AccessibilityToolbar)
- **Frontend variables:** `camelCase` (skillList, isHighContrast, fontSize)
- **Backend packages:** `snake_case` (internal/skills/)
- **Backend functions:** `camelCase` (installSkill, resolveDependencies)
- **Backend structs:** `PascalCase` (SkillManifest, ABTestConfig)
- **API endpoints:** `snake_case` (`/api/v1/skills`, `/api/v1/skills/install`)
- **JSON fields:** `camelCase` (`{"skillId": "...", "dependencyTree": [...]})

### Detected Conflicts or Variances

- **None detected** — this story follows established patterns from Epic 2 (skill system foundation) and Epic 3 (UI patterns)

## References

- [Source: epics.md#Story3.4](_bmad-output/planning-artifacts/epics.md#Story-3.4-SkillMarketplace-&-AccessibilityToolbar) — Story definition, user story, acceptance criteria
- [Source: architecture.md#SkillSystem](_bmad-output/planning-artifacts/architecture.md#Skill-System-Capabilities) — `internal/skills/` package structure, API endpoints
- [Source: architecture.md#FrontendArchitecture](_bmad-output/planning-artifacts/architecture.md#Frontend-Architecture) — Component structure, SkillMarketplace.tsx, themes/, i18n/
- [Source: architecture.md#API&CommunicationPatterns](_bmad-output/planning-artifacts/architecture.md#API-&-Communication-Patterns) — REST endpoints, WebSocket message types
- [Source: ux-design-specification.md#ComponentStrategy](_bmad-output/planning-artifacts/ux-design-specification.md#Component-Strategy) — SkillMarketplace (Phase 3, step 9), AccessibilityToolbar (Phase 3, step 9)
- [Source: ux-design-specification.md#DesignSystemComponents](_bmad-output/planning-artifacts/ux-design-specification.md#Design-System-Components) — Radix UI primitives, Tailwind CSS customization
- [Source: ux-design-specification.md#AccessibilityPatterns](_bmad-output/planning-artifacts/ux-design-specification.md#Accessibility-Patterns) — High-contrast mode, RTL support, WCAG 2.1 AA
- [Source: ux-design-specification.md#UXPatternAnalysis](_bmad-output/planning-artifacts/ux-design-specification.md#UX-Pattern-Analysis-&-Inspiration) — UX-DR8 (SkillMarketplace), UX-DR9 (AccessibilityToolbar)
- [Source: project-context.md#TechnologyStack](_bmad-output/project-context.md#Technology-Stack-&-Versions) — React + Vite + @xyflow/react, Tailwind + Radix UI, Vitest testing
- [Source: project-context.md#CriticalImplementationRules](_bmad-output/project-context.md#Critical-Implementation-Rules) — Naming conventions, framework-specific rules
- [Source: prd.md#FunctionalRequirements](_bmad-output/planning-artifacts/prd.md#Functional-Requirements) — FR25, FR40, FR41, FR46, FR51
- [Source: prd.md#NonFunctionalRequirements](_bmad-output/planning-artifacts/prd.md#Non-Functional-Requirements) — NFR-20, NFR-21, NFR-22, NFR-30

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

## Developer Context Section

### Developer Guardrails

**MUST READ BEFORE IMPLEMENTING:**

1. **Read this file completely** — contains all context needed for flawless implementation
2. **Follow naming conventions exactly:**
   - Go: `snake_case` packages, `camelCase` functions, `PascalCase` structs
   - TypeScript: `kebab-case.tsx` files, `PascalCase` components, `camelCase` variables
   - API: `snake_case` endpoints (`/api/v1/skills`), `camelCase` JSON fields
3. **Use correct technologies:**
   - Frontend: React + Vite + @xyflow/react, Tailwind CSS + Radix UI
   - Backend: Go + Gin, `internal/skills/` package
   - Testing: Vitest (frontend), testify (backend) — NO Jest!
4. **Avoid anti-patterns:**
   - [AVOID] TypeScript: `export const skill_marketplace: Component` (snake_case)
   - [AVOID] Go: `func Install_Skill()` (snake_case in Go functions)
   - [AVOID] Frontend: Serve from filesystem — use `embed.FS` via Gin
   - [AVOID] Testing: Jest for frontend — use Vitest
5. **Critical requirements:**
   - Skill dependencies auto-install recursively (FR41)
   - A/B testing with metrics collection (FR46)
   - WCAG 2.1 AA compliance (NFR-20)
   - High-contrast toggle, RTL switch, font-size slider (UX-DR9)

### Technical Requirements

1. **SkillMarketplace Panel (UX-DR8, FR25, FR40):**
   - Grid layout with rating stars, category filter, install button
   - Fetches skills from registry via `GET /api/v1/skills` (backend: `internal/skills/`)
   - Installs skill + auto-installs dependencies via `POST /api/v1/skills/install` (FR41)
   - A/B testing routes to different versions, collects metrics (FR46)
   - Prometheus metrics in `/metrics` endpoint (NFR-30)

2. **AccessibilityToolbar (UX-DR9, FR51, NFR-21, NFR-22):**
   - High-contrast toggle: Switches canvas to `#000000` background, bright node colors
   - RTL switch: Inverts canvas coordinates horizontally, mirrors mini-map
   - Font-size slider: Scales JetBrains Mono (body: 0.875rem base)
   - WCAG 2.1 AA: 4.5:1 contrast ratio, ARIA live regions

3. **Integration Points:**
   - WebSocket: `skill_installed`, `abtest_metrics` messages via Gin WS hub
   - React Context: Accessibility settings state management
   - Prometheus: A/B test metrics (execution time, success rate, token usage)

### Architecture Compliance

**Architecture Decision Records:**

1. **Frontend Architecture (from architecture.md):**
   - `frontend/src/components/panels/SkillMarketplace.tsx` — Radix Dialog + Tailwind grid
   - `frontend/src/components/ui/AccessibilityToolbar.tsx` — Radix Switch + Slider
   - `frontend/src/components/ui/themes/` — High-contrast theme
   - `frontend/src/components/ui/i18n/` — RTL + 20+ languages

2. **API & Communication (from architecture.md):**
   - REST: `GET /api/v1/skills`, `POST /api/v1/skills/install` (Gin, `snake_case`)
   - WebSocket: `skill_installed`, `abtest_metrics` (JSON, `camelCase` fields)
   - Prometheus: `/metrics` endpoint with A/B test results (NFR-30)

3. **Skill System (from architecture.md):**
   - `internal/skills/manifest.go` — Skill manifest parsing
   - `internal/skills/resolver.go` — Dependency resolution (FR41)
   - `internal/skills/abtest.go` — A/B testing framework (FR46)
   - SQLite for skill manifests (mattn/go-sqlite3)

4. **Accessibility (from architecture.md):**
   - WCAG 2.1 AA compliance (NFR-20)
   - High-contrast theme (NFR-21)
   - RTL canvas support (NFR-22)
   - Radix UI primitives for keyboard navigation

### Library/Framework Requirements

**Frontend (from project-context.md):**
- React 18.2.0 + TypeScript 5.3.3
- @xyflow/react ^12.10.0 (React Flow base)
- Vite ^5.0.12 (build system)
- Tailwind CSS (utility-first styling)
- Radix UI Primitives (accessibility-first components)
- Vitest ^3.1.1 (testing, NOT Jest!)
- @testing-library/react ^15.0.7
- Lucide React (icons)

**Backend (from project-context.md):**
- Go 1.26.2
- github.com/gin-gonic/gin v1.11.0 (REST + WebSocket)
- github.com/spf13/cobra v1.10.2 (CLI)
- github.com/mattn/go-sqlite3 v1.14.44 (SQLite for skills)
- github.com/stretchr/testify v1.11.1 (testing)

**Critical Version Constraints:**
- Go 1.26.2 (NOT 1.24+ as originally planned — check go.mod)
- Vite `base: './'` required for Go `embed.FS` serving
- TypeScript `"strict": true`

### File Structure Requirements

**Complete File List (from architecture.md):**

```
frontend/src/
├── components/
│   ├── panels/
│   │   ├── SkillMarketplace.tsx      # NEW - Grid layout, rating stars, install button
│   │   └── AccessibilityToolbar.tsx  # NEW - Toggles, slider, theme switch
│   └── ui/
│       ├── themes/                    # NEW - High-contrast theme definition
│       └── i18n/                      # NEW - RTL support, 20+ languages
├── hooks/
│   └── useWebSocket.ts             # MODIFY - Add skill message handlers
└── App.tsx                        # MODIFY - Add AccessibilityToolbar
```

```
internal/skills/
├── manifest.go                      # MAYBE NEW - Skill manifest (if not exists)
├── resolver.go                      # MAYBE NEW - Dependency resolution (if not exists)
├── abtest.go                       # NEW - A/B testing framework (FR46)
└── skills_test.go                  # MODIFY - Add A/B test tests
```

```
cmd/nforge/
└── skill.go                        # MODIFY - Add install, A/B test routes
```

```
internal/devops/
└── metrics.go                      # MODIFY - Add A/B test Prometheus metrics
```

### Testing Requirements

**Go Testing (from project-context.md):**
- `go test ./internal/skills/...` — Skill installation, dependency resolution, A/B testing
- Table-driven tests preferred
- CGO required for SQLite (mattn/go-sqlite3)
- Testify assertions: `assert.Equal(t, expected, actual)`

**TypeScript Testing (from project-context.md):**
- Vitest (NOT Jest!) + @testing-library/react
- `*.test.tsx` co-located with components
- `npx vitest run` (CI) or `npx vitest` (watch)
- Test SkillMarketplace: grid render, category filter, install button click
- Test AccessibilityToolbar: toggle high-contrast, switch RTL, adjust font-size
- axe DevTools for WCAG 2.1 AA compliance (4.5:1 contrast ratio)

**Integration Testing:**
- WebSocket: Skill install status updates via Gin WS hub
- Prometheus: Verify A/B test metrics in `/metrics` endpoint
- RTL: Test canvas coordinate inversion for Arabic/Hebrew

## Previous Story Intelligence

No previous Epic 3 story implementations available (Epic 3 is in backlog, no stories completed yet). This is the fourth story in Epic 3; stories 3.1-3.3 are still in backlog.

**Lessons from Epic 2 (completed stories):**
- Story 2.8 (Headless CLI) established CLI patterns: `nforge skill list/install` commands in `cmd/nforge/skill.go`
- Story 2.5 (Smart Context Engine) established BadgerDB + SQLite patterns
- Story 2.7 (Incremental Execution) established Web Worker patterns for canvas

## Git Intelligence Summary

**Recent Commits (last 5):**
1. `9c4d032` — Implement story 2.8 Headless CLI Execution with code review fixes
2. `10af6c1` — Apply code review fixes for story 2.7 Incremental Execution & Web Worker Offloading
3. `54fad67` — Update BMAD configs, add LLM swarm support, update stories and epics
4. `3a35a19` — Mark story 2.5 Smart Context Engine as done
5. `d417cd0` — Apply code review patches for story 2.5 Smart Context Engine

**Patterns Observed:**
- Code review fixes applied after each story completion
- BMAD configs updated with new story status
- Go backend changes accompany frontend changes
- Test files co-located with source (`*_test.go`, `*.test.tsx`)

## Latest Tech Information

**Technology Versions (from project-context.md, verified 2026-05-02):**
- Go 1.26.2 (updated from 1.24+ in architecture.md)
- github.com/gin-gonic/gin v1.11.0
- github.com/spf13/cobra v1.10.2
- github.com/mattn/go-sqlite3 v1.14.44
- React 18.2.0 + TypeScript 5.3.3
- @xyflow/react ^12.10.0
- Vite ^5.0.12 + Vitest ^3.1.1

**No web research performed** — all tech versions from project-context.md (verified 2026-05-02).

## Project Context Reference

**Full project context available at:** `_bmad-output/project-context.md`

**Key Rules Summary:**
1. Go: `snake_case` packages, `camelCase` functions, `PascalCase` structs
2. TypeScript: `kebab-case.tsx` files, `PascalCase` components, `camelCase` variables
3. API: `snake_case` endpoints, `camelCase` JSON fields
4. Testing: Vitest (frontend), testify (backend)
5. Security: chroot + eBPF + Argon2 (not relevant for this story)
6. Performance: WebSocket <50ms latency, 5000+ connections (Gin)

## Story Completion Status

**Status:** ready-for-dev

**Completion Note:** Ultimate context engine analysis completed — comprehensive developer guide created with all architecture patterns, technical requirements, file structure, testing standards, and relevant references from epics.md, architecture.md, ux-design-specification.md, prd.md, and project-context.md.

**Next Steps:**
1. Review the comprehensive story in `_bmad-output/implementation-artifacts/3-4-skillmarketplace-and-accessibilitytoolbar.md`
2. Run dev agents `bmad-dev-story` for optimized implementation
3. Run `bmad-code-review` when complete (auto-marks done)
4. Optional: If Test Architect module installed, run `/bmad:tea:automate` after `dev-story` to generate guardrail tests
