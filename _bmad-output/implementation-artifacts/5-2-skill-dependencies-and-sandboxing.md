# Story 5.2: Skill Dependencies & Sandboxing

Status: ready-for-dev

## Story

As a system,
I want skills to have dependencies that auto-install and run sandboxed (time limits, no network, read-only) before trust,
So that skills are safe and self-contained.

## Acceptance Criteria

1. **Given** the skill dependency resolver exists in `internal/skills/resolver.go`
   **When** a skill is installed via `nforge skill install <name>`
   **Then** its dependency tree is resolved recursively using depth-first search (FR41)
   **And** dependencies are installed in order before the requesting skill (already implemented in `ResolveDependencies()`)

2. **Given** the skill sandbox module (`internal/skills/sandbox.go`) is implemented
   **When** a skill runs before trust verification
   **Then** it is confined to a pre-trust sandbox with: time limit (30s max), no network access, read-only filesystem (FR42, NFR-13)
   **And** violations (timeout, network access, write attempts) are caught and logged

3. **Given** Ed25519 signature verification is implemented in `internal/security/signing.go`
   **When** a skill's signature is verified successfully
   **Then** the skill is escalated to full trust with expanded permissions (file write, network access)
   **And** unverified skills remain in sandbox mode permanently

4. **Given** sandbox violation detection is active
   **When** a skill attempts to exceed its sandbox constraints (time, network, filesystem)
   **Then** the violation is logged with an audit entry including: timestamp, skill ID, violation type, and session context (NFR-12)
   **And** graph snapshots signed with Ed25519 can detect tampering on session import

## Tasks / Subtasks

- [ ] Task 1: Create sandbox module (AC: #2)
  - [ ] Subtask 1.1: Define `SandboxConfig` struct with TimeLimit, NoNetwork, ReadOnlyFS fields
  - [ ] Subtask 1.2: Implement `RunInSandbox(ctx, skillPath, config) (result, error)` function
  - [ ] Subtask 1.3: Implement time limit enforcement (30s max, interrupt execution)
  - [ ] Subtask 1.4: Implement network access restriction (block network syscalls)
  - [ ] Subtask 1.5: Implement read-only filesystem restriction (chroot to skill dir, read-only mount)

- [ ] Task 2: Implement trust escalation (AC: #3)
  - [ ] Subtask 2.1: Integrate Ed25519 signature verification from `internal/security/signing.go`
  - [ ] Subtask 2.2: Implement `EscalateTrust(skillID, signature) error` function
  - [ ] Subtask 2.3: Create trust level enum: `Sandboxed | Verified | FullTrust`
  - [ ] Subtask 2.4: Store trust level in `SkillManifest` and persist to skill.json

- [ ] Task 3: Add audit logging for violations (AC: #4)
  - [ ] Subtask 3.1: Define `AuditEntry` struct with Timestamp, SkillID, ViolationType, SessionContext
  - [ ] Subtask 3.2: Implement `LogViolation(entry AuditEntry) error` function
  - [ ] Subtask 3.3: Write violations to BadgerDB (use `internal/context/` patterns)
  - [ ] Subtask 3.4: Add Ed25519 graph snapshot signing integration from `internal/security/signing.go`

- [ ] Task 4: Wire into skill install flow (AC: #1, #2, #3)
  - [ ] Subtask 4.1: Modify `nforge skill install` to resolve dependencies first (use existing `ResolveDependencies()`)
  - [ ] Subtask 4.2: Run newly installed skills in sandbox by default
  - [ ] Subtask 4.3: Add `--trust` flag to `nforge skill install` for pre-verified skills
  - [ ] Subtask 4.4: Add REST API endpoint `POST /api/v1/skills/:id/verify` for signature verification

- [ ] Task 5: Unit tests
  - [ ] Subtask 5.1: Test dependency resolution with circular dependency detection
  - [ ] Subtask 5.2: Test sandbox time limit enforcement (mock long-running skill)
  - [ ] Subtask 5.3: Test sandbox network restriction (attempt network access, expect block)
  - [ ] Subtask 5.4: Test trust escalation flow (sandbox → verified → full trust)
  - [ ] Subtask 5.5: Test audit logging (verify violation entries written)

## Dev Notes

### Project Structure Notes

- **Alignment:** Follows `internal/skills/` package pattern established by `manifest.go`, `resolver.go`, `abtest.go`
- **Conflicts:** None detected — `sandbox.go` is a new file; `resolver.go` already implements dependency resolution
- **Files to create:** `internal/skills/sandbox.go`, `internal/skills/trust.go`, `internal/skills/audit.go`
- **Files to modify:** `internal/skills/manifest.go` (add trust level field), `cmd/nforge/skill.go` (wire into CLI)

### References

- [Epic 5 in epics.md](_bmad-output/planning-artifacts/epics.md#epic-5-skill-system--extensibility) — Epic 5 overview, FR41-FR46 coverage
- [Story 5.1: Skill Marketplace](_bmad-output/planning-artifacts/epics.md#story-51-skill-marketplace--dynamic-fetch) — Marketplace backend provides `GET /api/v1/skills` that 5.2 depends on
- [Architecture: Skill System](_bmad-output/planning-artifacts/architecture.md#decision-priority-analysis) — `internal/skills/` package structure
- [Architecture: Security](_bmad-output/planning-artifacts/architecture.md#security-architecture) — chroot, eBPF, Ed25519 signing
- [PRD: Skill System Capabilities](_bmad-output/planning-artifacts/prd.md#skill-system-capabilities) — FR41, FR42 requirements
- [PRD: Security Capabilities](_bmad-output/planning-artifacts/prd.md#security-capabilities) — NFR-12, NFR-13
- [Existing resolver.go](internal/skills/resolver.go) — Dependency resolution already implemented, reuse `ResolveDependencies()`
- [Existing manifest.go](internal/skills/manifest.go) — `SkillManifest` struct with `Dependencies []string` field

### Technical Requirements

**Go Naming Conventions (from project-context.md):**
- Package: `skills` (snake_case)
- Functions: `camelCase` — `runInSandbox()`, `escalateTrust()`
- Structs: `PascalCase` — `SandboxConfig`, `AuditEntry`
- JSON tags: `camelCase` — `` `json:"timeLimit"` ``

**Architecture Patterns (from architecture.md):**
- Use `internal/security/signing.go` for Ed25519 operations (don't reimplement)
- Use `internal/security/chroot.go` for filesystem isolation
- Use BadgerDB (via `internal/context/`) for audit log storage
- Follow existing `manifest.go` struct patterns for new types

**Security Requirements (NFR-13):**
- Pre-trust sandbox: 30s time limit, no network, read-only FS
- eBPF syscall filtering (from `internal/security/ebpf.go`) blocks dangerous calls
- Chroot jail per skill (from `internal/security/chroot.go`)

**API Patterns (from architecture.md):**
- REST: `POST /api/v1/skills/:id/verify` — verify signature, escalate trust
- REST: `GET /api/v1/skills/:id/audit` — retrieve audit log for skill
- WebSocket: Reuse hub from `cmd/nforge/serve.go`

### Previous Story Intelligence

No previous story in Epic 5 has been implemented yet (5.1 is still in backlog). However:

**Existing Code Patterns in `internal/skills/`:**
- `SkillManifest` struct uses both `json` and `yaml` tags (see `manifest.go:11-27`)
- `ResolveDependencies()` uses depth-first search with cycle detection (see `resolver.go:25-44`)
- Error pattern: `fmt.Errorf("skills: ...")` with `%w` verb for wrapping (see `manifest.go:34`)

**Key Learning:** Follow the exact error wrapping pattern and struct tag conventions already established in this package.

### Git Intelligence Summary

**Recent commits show:**
- Session management stories (4.1-4.5) recently completed
- Pattern: stories committed with imperative mood ("Fix...", "Mark story...", "Implement...")
- Security primitives (`internal/security/`) being built out — check `signing.go`, `chroot.go`, `ebpf.go` existence before implementing sandbox

### Latest Tech Information

**Go 1.26.2** (from project-context.md):
- `crypto/ed25519` package for signature verification (standard library, no external dep)
- `context.WithTimeout()` for 30s time limit enforcement
- `os/exec` package for running skill processes in sandbox

**Dependencies (from go.mod via project-context.md):**
- `github.com/dgraph-io/badger/v4 v4.9.1` — Use for audit log storage
- `github.com/mattn/go-sqlite3 v1.14.44` — Alternative storage if needed

### Project Context Reference

| Area | Key Rule |
|------|----------|
| Go naming | `snake_case` packages, `camelCase` functions, `PascalCase` structs |
| JSON | `camelCase` fields in Go struct tags: `` `json:"skillId"` `` |
| Security | chroot + eBPF + Ed25519 signing (reuse `internal/security/`) |
| Testing Go | `go test ./...` + testify, `*_test.go` co-located |
| Errors | Go: `%w` verb for wrapping, never return directly from Gin handlers |

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
