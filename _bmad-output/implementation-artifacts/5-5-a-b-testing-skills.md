# Story 5.5: A/B Testing Skills

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to A/B test skills where the system routes to different versions and collects metrics,
so that I can objectively compare skill performance.

## Acceptance Criteria

1. **Given** the A/B testing framework is implemented (`internal/skills/abtest.go`)
   **When** A/B testing is enabled for a skill
   **Then** the system routes requests to different skill versions and collects metrics (execution time, success rate, token usage) (FR46)

2. **Given** A/B testing is active
   **When** metrics are collected during skill execution
   **Then** metrics are reported back to the marketplace and displayed in SkillMarketplace

3. **Given** A/B test results are available
   **When** the user views the test results
   **Then** the user can view A/B test results and choose the preferred version

4. **Given** Prometheus metrics are enabled
   **When** A/B tests are executed
   **Then** Prometheus metrics (`/metrics`) include A/B test results (NFR-30)

## Tasks / Subtasks

- [ ] Task 1: Implement A/B testing framework (AC: 1, 2, 3, 4)
  - [ ] Subtask 1.1: Create `internal/skills/abtest.go` with version routing logic and metrics collection (execution time, success rate, token usage)
  - [ ] Subtask 1.2: Implement skill version routing (weighted random, round-robin, or performance-based routing)
  - [ ] Subtask 1.3: Add Prometheus metrics export for A/B test results (`/metrics` endpoint)
  - [ ] Subtask 1.4: Store A/B test configuration in skill manifest or SQLite

- [ ] Task 2: Integrate with Marketplace and UI (AC: 2, 3)
  - [ ] Subtask 2.1: Add A/B test metrics reporting to marketplace backend API (`/api/v1/skills`)
  - [ ] Subtask 2.2: Display A/B test results in SkillMarketplace panel (frontend)
  - [ ] Subtask 2.3: Add UI controls for enabling/disabling A/B tests and selecting preferred version

- [ ] Task 3: Testing and validation (AC: 1, 2, 3, 4)
  - [ ] Subtask 3.1: Unit tests for `internal/skills/abtest.go` (routing logic, metrics collection)
  - [ ] Subtask 3.2: Integration tests for A/B routing with multiple skill versions
  - [ ] Subtask 3.3: Validate Prometheus metrics export for A/B test results
  - [ ] Subtask 3.4: End-to-end test: enable A/B test, execute skills, verify metrics and UI display

## Dev Notes

- Relevant architecture patterns and constraints:
  - Skill system lives in `internal/skills/` (architecture.md#Skill-System)
  - A/B testing framework file: `internal/skills/abtest.go` (epics.md#Story-5.5)
  - Prometheus metrics endpoint: `/metrics` (architecture.md#API-&-Communication-Patterns)
  - Skill manifests stored in SQLite (`internal/skills/manifest.go`) (architecture.md#Data-Architecture)
  - Marketplace API: `/api/v1/skills` (epics.md#Story-5.1)
  - Frontend SkillMarketplace component: `frontend/src/components/panels/SkillMarketplace.tsx` (architecture.md#Frontend-Architecture)

- Source tree components to touch:
  - `internal/skills/abtest.go` (NEW) - Core A/B testing framework
  - `internal/skills/metrics.go` (UPDATE if exists) - Prometheus metrics for skills
  - `internal/devops/metrics.go` (UPDATE) - Add A/B test metrics to `/metrics`
  - `frontend/src/components/panels/SkillMarketplace.tsx` (UPDATE) - Display A/B test results
  - `frontend/src/components/panels/SkillMarketplace.css` or Tailwind classes (UPDATE) - Style A/B test UI
  - `cmd/nforge/skill.go` (UPDATE) - CLI commands for A/B test management (optional)

- Testing standards summary:
  - Go: `go test ./internal/skills/...` with testify assertions, CGO for SQLite if needed (project-context.md#Testing-Rules)
  - TypeScript: Vitest + @testing-library/react, `*.test.tsx` co-located (project-context.md#Testing-Rules)
  - Prometheus metrics validation: Use `curl http://localhost:8080/metrics` and verify A/B test metrics presence

### Project Structure Notes

- Alignment with unified project structure (paths, modules, naming):
  - Go: `internal/skills/abtest.go` follows `snake_case` package naming (project-context.md#Go-naming)
  - TypeScript: `SkillMarketplace.tsx` follows `PascalCase` component naming (project-context.md#TypeScript-naming)
  - API: `/api/v1/skills` uses `snake_case` endpoints (project-context.md#API)
  - JSON: `camelCase` fields in Go struct tags (project-context.md#JSON)

- Detected conflicts or variances (with rationale):
  - None detected. A/B testing framework is a new addition to `internal/skills/`, aligning with existing skill system patterns.

### References

- [Source: epics.md#Story-5.5](_bmad-output/planning-artifacts/epics.md#story-55-a-b-testing-skills) - Story definition and acceptance criteria
- [Source: architecture.md#Skill-System](_bmad-output/planning-artifacts/architecture.md#skill-system) - Skill system architecture
- [Source: architecture.md#API-&-Communication-Patterns](_bmad-output/planning-artifacts/architecture.md#api--communication-patterns) - Prometheus metrics and API patterns
- [Source: architecture.md#Data-Architecture](_bmad-output/planning-artifacts/architecture.md#data-architecture) - SQLite for skill manifests
- [Source: project-context.md#Testing-Rules](_bmad-output/project-context.md#testing-rules) - Go and TypeScript testing standards
- [Source: ux-design-specification.md#SkillMarketplace](_bmad-output/planning-artifacts/ux-design-specification.md) - UX design for SkillMarketplace (if applicable)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
