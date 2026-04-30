# Story 1.7: CLI Tab Completion

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want tab completion for all commands and node types,
So that I can navigate the CLI faster without remembering exact syntax.

## Acceptance Criteria

**Given** the CLI uses Cobra framework (story 1.2 completed)
**When** the user presses Tab after typing `nforge `
**Then** available subcommands are displayed: `serve`, `run`, `new`, `config`, `skill`, `session`, `doctor`, `graph`
**And** Tab completion works for subcommand flags and node types (Goal, Spec, Plan, Implement, Test, Review)
**And** completion works in bash, zsh, and PowerShell shells

## Tasks / Subtasks

- [ ] Task 1: Enable Cobra shell completion (AC: 1, 3)
  - [ ] Subtask 1.1: Add `completion` subcommand in `cmd/nforge/root.go` using `cobra.OnInitialize` or direct registration
  - [ ] Subtask 1.2: Generate bash completion script: `nforge completion bash > /etc/bash_completion.d/nforge` (or user-local path)
  - [ ] Subtask 1.3: Generate zsh completion script: `nforge completion zsh > "${fpath[#fpath[@]}/_nforge"` (or user-local path)
  - [ ] Subtask 1.4: Generate PowerShell completion script: `nforge completion powershell > nforge.ps1`

- [ ] Task 2: Register all subcommands for completion (AC: 1, 2)
  - [ ] Subtask 2.1: Verify all 8 subcommands are registered in `root.go`: `serve`, `run`, `new`, `config`, `skill`, `session`, `doctor`, `graph`
  - [ ] Subtask 2.2: Ensure each subcommand has `ValidArgs` or `ValidArgsFunction` set for argument completion
  - [ ] Subtask 2.3: Add `node type` completion for commands that accept node types (Goal, Spec, Plan, Implement, Test, Review)

- [ ] Task 3: Implement custom completions for flags (AC: 2)
  - [ ] Subtask 3.1: Add `--port` completion values (common ports: 8080, 9090, 3000) in `serve.go`
  - [ ] Subtask 3.2: Add `--config-path` file/directory completion in `root.go`
  - [ ] Subtask 3.3: Add `config set <key>` completion with supported keys: `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `server.port`, `llm.default-model`
  - [ ] Subtask 3.4: Add `config get <key>` completion with same key list

- [ ] Task 4: Implement node type completion (AC: 2)
  - [ ] Subtask 4.1: Define node type list as constant: `Goal`, `Spec`, `Plan`, `Implement`, `Test`, `Review`
  - [ ] Subtask 4.2: Register `ValidArgsFunction` on commands that accept node types (e.g., `nforge run <spec>` could accept node type hints)
  - [ ] Subtask 4.3: Add node type descriptions in completion output (e.g., "Goal - Top-level goal node")

- [ ] Task 5: Shell integration and documentation (AC: 3)
  - [ ] Subtask 5.1: Add `nforge completion --help` with installation instructions for each shell
  - [ ] Subtask 5.2: Document bash install: `echo 'source <(nforge completion bash)' >> ~/.bashrc`
  - [ ] Subtask 5.3: Document zsh install: `echo 'source <(nforge completion zsh)' >> ~/.zshrc`
  - [ ] Subtask 5.4: Document PowerShell install: `. (nforge completion powershell)` in `$PROFILE`

- [ ] Task 6: Verify end-to-end (AC: 1, 2, 3)
  - [ ] Subtask 6.1: `nforge completion bash` outputs valid bash completion script
  - [ ] Subtask 6.2: `nforge completion zsh` outputs valid zsh completion script
  - [ ] Subtask 6.3: `nforge completion powershell` outputs valid PowerShell completion script
  - [ ] Subtask 6.4: Tab after `nforge ` shows all 8 subcommands
  - [ ] Subtask 6.5: Tab after `nforge config set ` shows supported keys
  - [ ] Subtask 6.6: Tab after `nforge graph ` shows subcommand completions

## Dev Notes

### Architecture Patterns and Constraints

**Cobra Version:** v1.10.2 (latest stable, Dec 2025) — [Source: story-1.2, architecture.md#Starter-Template-Evaluation]

**Cobra Built-in Completion:**
- Cobra has native `completion` command supporting bash, zsh, fish, and PowerShell
- Use `cobra.Command.CompletionOptions` for customizing completion behavior
- `ValidArgs` / `ValidArgsFunction` for positional argument completion
- `ValidFlags` / `RegisterFlagCompletionFunc` for flag value completion

**Shell Completion Architecture:**
```go
// root.go - Cobra root command
var rootCmd = &cobra.Command{
    Use:   "nforge",
    Short: "NodeForge OS - Spec-driven development platform",
    CompleteOptions: cobra.CompleteOptions{
        DisableDefaultCmd: true, // Hides __complete internal command from help
    },
}
// Completion subcommand is auto-added by Cobra when completion is registered
```

**Naming Conventions (CRITICAL):**
- Go functions: `camelCase` — `registerCompletions()`
- Go structs: `PascalCase` — `type CompletionConfig struct`
- CLI flags: `kebab-case` — `--config-path`
- Subcommand names: `kebab-case` — `nforge skill list`
— [Source: architecture.md#Naming-Patterns]

**Command Registration (Prerequisite Check):**
All 8 subcommands MUST be registered in `root.go` for tab completion to work:
1. `serve` — `nforge serve` (story 1.2 stub, 1.3 implementation)
2. `run` — `nforge run <spec-file>` (story 1.2 stub, 1.8 implementation)
3. `new` — `nforge new <project-name>` (story 1.2 stub, 1.4 implementation)
4. `config` — `nforge config set/get` (story 1.2 stub, 1.5 implementation)
5. `skill` — `nforge skill list/install` (story 1.2 stub, 5.1 implementation)
6. `session` — `nforge session resume/export` (story 1.2 stub, 4.5 implementation)
7. `doctor` — `nforge doctor` (story 1.2 stub, 1.6 implementation)
8. `graph` — `nforge graph viz` (story 1.2 stub, story TBD)

**IMPORTANT:** If any subcommand is NOT registered (still a stub or missing), tab completion will not show it. The developer must verify all 8 are registered before completing this story.

### Source Tree Components to Touch

**Files to Modify (UPDATE):**
- `cmd/nforge/root.go` — Add completion setup, register all subcommands, add flag completions
  - Current state: Has all 8 subcommand stubs registered (from story 1.2)
  - What this story changes: Adds completion subcommand, enhances each subcommand with ValidArgs/ValidArgsFunction
  - What must be preserved: All existing subcommand registrations, persistent flags (`--verbose`, `--config-path`)

**New Files (CREATE):**
- None required — all changes are in existing `cmd/nforge/` files

**Files That Must NOT Be Modified:**
- `main.go` — calls `root.go Execute()` only
- `go.mod` — Cobra v1.10.2 already present
- `internal/` — no changes needed for CLI completion

### Testing Standards

**Go Testing (Ginkgo + Testify):**
- Co-located `*_test.go` files
- Test that completion script generation works for all 3 shells
- Test that `ValidArgsFunction` returns correct node types
- Test that flag completion returns supported keys for `config set/get`

**Test Pattern:**
```go
// cmd/nforge/completion_test.go
func TestCompletionBash(t *testing.T) {
    // Test: nforge completion bash outputs valid script
    output, err := executeCommand("completion", "bash")
    assert.NoError(t, err)
    assert.Contains(t, output, "complete -o default -F")
}

func TestSubcommandCompletion(t *testing.T) {
    // Test: Tab after 'nforge ' shows all subcommands
    cmd := rootCmd
    args := []string{""} // Simulating Tab after 'nforge '
    completions := cmd.ValidArgs(nil, args)
    assert.Contains(t, completions, "serve", "run", "new", "config")
}
```

## Previous Story Intelligence (from Story 1.5)

**What Story 1.5 Established:**
- Cobra v1.10.2 is already in `go.mod` — do NOT run `go get cobra` again
- `cmd/nforge/config.go` implements full `config set/get` with Viper
- `cmd/nforge/root.go` has all 8 subcommands registered as stubs (from story 1.2)
- Persistent flags defined: `--verbose` (bool), `--config-path` (string, default: `~/.nforge/config.yaml`)

**Key Learnings from Story 1.5:**
- Cobra v1.10.2 uses `go.yaml.in/yaml/v3` (not `gopkg.in/yaml.v3`)
- Naming conventions: Go functions `camelCase`, CLI flags `kebab-case`
- All subcommands must be registered in root.go first as stubs, then fully implemented in their own files
- The `config` command was stubbed in 1.2, fully implemented in 1.5 — same pattern applies to completion

**Integration Point:**
- Story 1.7 enhances ALL subcommands with completion support
- If a subcommand is missing or not registered, completion won't work for it
- Developer must verify `root.go` has all 8 subcommands before adding completion logic

## Project Structure Notes

### Alignment with Unified Project Structure

The `cmd/nforge/` directory must contain:

```
nfv2/
├── cmd/nforge/
│   ├── root.go           # ← UPDATE: add completion, verify all 8 subcommands
│   ├── serve.go         # ← EXISTS: serve subcommand (enhance with flag completion)
│   ├── run.go           # ← EXISTS: run subcommand (enhance with arg completion)
│   ├── new.go           # ← EXISTS: new subcommand
│   ├── config.go        # ← EXISTS: config set/get (add key completion in 3.3, 3.4)
│   ├── skill.go         # ← EXISTS: skill list/install
│   ├── session.go       # ← EXISTS: session resume/export
│   ├── doctor.go        # ← EXISTS: doctor (health check)
│   └── graph.go         # ← EXISTS: graph viz (ASCII art)
```

— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**Subcommand Registration Dependency:**
- Story 1.2 created stubs for all 8 subcommands in `root.go`
- If story 1.2 is not fully implemented (missing subcommands), this story cannot complete
- Developer should verify: `grep -r "AddCommand" cmd/nforge/root.go` shows all 8 subcommands

**No Direct Conflicts:** Tab completion is additive — it enhances existing commands without changing their behavior.

**Critical Reminder:** Cobra's `completion` command is auto-registered when `cobra.Command` is used properly. Do NOT manually create a `completion` subcommand — use Cobra's built-in support.

## References

- [Story 1.7 Definition: epics.md#Story-1.7](_bmad-output/planning-artifacts/epics.md#Story-1.7)
- [Story 1.2 (Prerequisite): epics.md#Story-1.2](_bmad-output/planning-artifacts/epics.md#Story-1.2)
- [Story 1.5 (Previous): 1-5-configuration-management-with-nforge-config.md](_bmad-output/implementation-artifacts/1-5-configuration-management-with-nforge-config.md)
- [Cobra v1.10.2 Completion Docs](https://github.com/spf13/cobra/blob/main/shell_completions.md) — Native bash/zsh/fish/PowerShell support
- [Cobra ValidArgsFunction Examples](https://github.com/spf13/cobra/blob/main/user_guide.md#custom-completions) — Dynamic completion for arguments
- [PRD - CLI Capabilities: prd.md#CLI-Capabilities](_bmad-output/planning-artifacts/prd.md#CLI-Capabilities)
- [Architecture - CLI Structure: architecture.md#Starter-Template-Evaluation](_bmad-output/planning-artifacts/architecture.md#Starter-Template-Evaluation)
- [Node Types: epics.md#Story-1.1](_bmad-output/planning-artifacts/epics.md#Story-1.1) — Goal, Spec, Plan, Implement, Test, Review

## Dev Agent Record

### Agent Model Used

tencent/hy3-review:free

### Debug Log References

### Completion Notes List

### File List

- `cmd/nforge/root.go` (UPDATE — add completion setup, verify all subcommands)
- `cmd/nforge/config.go` (UPDATE — add key completion for `config set/get`)
- `cmd/nforge/completion_test.go` (NEW — tests for completion)
- Shell completion scripts (GENERATED — not stored in repo, generated per-install)
