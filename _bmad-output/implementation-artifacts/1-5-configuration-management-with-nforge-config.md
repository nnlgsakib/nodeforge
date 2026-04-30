# Story 1.5: Configuration Management with `nforge config`

Status: ready-for-dev

<!-- Validation: Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to configure settings with `nforge config set/get`,
so that I can manage API keys, models, and ports.

## Acceptance Criteria

1. **Given** the configuration system is set up with Cobra `config` subcommand (story 1.2 completed)
   **When** the user runs `nforge config set <key> <value>`
   **Then** the configuration is saved to the config file (e.g., `~/.nforge/config.yaml`) with the specified key-value pair
   **And** `nforge config get <key>` retrieves and displays the value
   **And** supported keys include: `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `server.port`, `llm.default-model`

## Tasks / Subtasks

- [ ] Task 1: Create config.go with set and get subcommands (AC: 1)
  - [ ] Subtask 1.1: Create `cmd/nforge/config.go` with Cobra config parent command
  - [ ] Subtask 1.2: Implement `config set <key> <value>` subcommand using Viper
  - [ ] Subtask 1.3: Implement `config get <key>` subcommand using Viper
  - [ ] Subtask 1.4: Validate supported keys: `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `server.port`, `llm.default-model`

- [ ] Task 2: Configure Viper for config persistence (AC: 1)
  - [ ] Subtask 2.1: Initialize Viper with config file path (default: `~/.nforge/config.yaml`)
  - [ ] Subtask 2.2: Use `viper.Set()` and `viper.WriteConfig()` for saving values
  - [ ] Subtask 2.3: Handle config file auto-creation if it doesn't exist (`viper.SafeWriteConfig()`)

- [ ] Task 3: Integrate with persistent `--config-path` flag (AC: 1)
  - [ ] Subtask 3.1: Use `--config-path` flag (defined in root.go, story 1.2) to override config location
  - [ ] Subtask 3.2: Fall back to default `~/.nforge/config.yaml` if `--config-path` not provided

- [ ] Task 4: Replace stub config command in root.go (AC: 1)
  - [ ] Subtask 4.1: Update `cmd/nforge/root.go` to use full config command from config.go instead of stub
  - [ ] Subtask 4.2: Verify `nforge config --help` shows set/get subcommands

- [ ] Task 5: Verify end-to-end (AC: 1)
  - [ ] Subtask 5.1: Run `nforge config set llm.openai-key sk-123` → verify saved to config file
  - [ ] Subtask 5.2: Run `nforge config get llm.openai-key` → verify returns `sk-123`
  - [ ] Subtask 5.3: Run `nforge config set server.port 8080` → verify saved as integer
  - [ ] Subtask 5.4: Run `nforge config get server.port` → verify returns `8080`
  - [ ] Subtask 5.5: Run `nforge config set invalid-key value` → verify error message

## Dev Notes

### Architecture Patterns and Constraints

**Go Version:** Go 1.24+ (62% GC pause reduction, improved reflect.Blueprint support) — [Source: architecture.md#Project-Context-Analysis]

**Framework Choices (Non-Negotiable):**
- **Cobra v1.10.2** for CLI with subcommands — [Source: epics.md#Epic-1, architecture.md#Starter-Template-Evaluation]
- **Viper** for configuration management (from web research: Cobra + Viper integration pattern) — Supports dot-separated keys (llm.openai-key), automatic config file creation, JSON/YAML serialization.

**Config File Specification:**
- **Default Location:** `~/.nforge/config.yaml` (matches `--config-path` default from story 1.2)
- **Format:** YAML (Viper default, matches `go.yaml.in/yaml/v3` dependency from Cobra v1.10.2)
- **Supported Keys:**
  - `llm.openai-key` (string) — OpenAI API key
  - `llm.anthropic-key` (string) — Anthropic API key
  - `llm.ollama-url` (string) — Ollama local URL (default: `http://localhost:11434`)
  - `server.port` (int) — Gin server port (default: `8080`)
  - `llm.default-model` (string) — Default LLM model (e.g., `gpt-4o`)

### Source Tree Components to Touch

**New Files (CREATE):**
- `cmd/nforge/config.go` — Full config set/get subcommands using Viper

**UPDATE Files (MODIFY):**
- `cmd/nforge/root.go` — Replace stub config command with full implementation from config.go
  - Current state: root.go has stub config command registered (from story 1.2)
  - What this story changes: Stub replaced with full config command from config.go
  - What must be preserved: All other subcommand stubs (serve, run, new, skill, session, doctor, graph) must remain unchanged

**Files That Must NOT Be Modified:**
- `main.go` — already wired to Cobra root command (from story 1.2)
- `go.mod` — Cobra v1.10.2 and Viper dependencies already present (add Viper if missing: `go get github.com/spf13/viper@v1.19.0`)
- `internal/` — no changes needed yet

### Testing Standards

- **Go**: Ginkgo + Testify (from cli-go-project-template pattern) — co-located `*_test.go` files
- **Test for this story:**
  - `cmd/nforge/config_test.go` — Test config set/get operations
  - Verify `nforge config set key value` writes to correct config file
  - Verify `nforge config get key` returns correct value
  - Verify unsupported keys are rejected with error message
  - Verify `--config-path` flag overrides default location
— [Source: architecture.md#Starter-Template-Evaluation]

### Previous Story Intelligence (from Story 1.2)

**What Story 1.2 Established:**
- Cobra v1.10.2 is already in `go.mod` — do NOT run `go get cobra` again
- `cmd/nforge/` directory exists with `root.go` (Cobra root command with persistent flags)
- Config command registered as **stub** in story 1.2 (Task 2.4: "Register `config` subcommand (stub for now, full impl in story 1.5)")
- Persistent flags defined: `--verbose` (bool), `--config-path` (string, default: `~/.nforge/config.yaml`)
- `main.go` calls `cmd/nforge/root.go` Execute()

**Key Learnings from Story 1.2:**
- Cobra v1.10.2 uses `go.yaml.in/yaml/v3` (not `gopkg.in/yaml.v3`) — Viper v1.19.0+ is compatible
- Naming conventions: Go functions `camelCase`, structs `PascalCase`, CLI flags `kebab-case` (e.g., `--config-path`)
- All subcommands must be registered in root.go first as stubs, then fully implemented in their own files

**Integration Point:**
- Story 1.2 created the stub; Story 1.5 delivers the full implementation
- Story 1.3 (`nforge serve`) will read config values (e.g., `server.port`, `llm.openai-key`) from Viper
- Config values become the source of truth for all LLM and server settings

## Project Structure Notes

### Alignment with Unified Project Structure

The `cmd/nforge/` directory must contain:
- `root.go` — UPDATE: replace stub config with full command
- `config.go` — NEW: full config set/get implementation
- `serve.go`, `run.go`, `new.go`, `skill.go`, `session.go`, `doctor.go`, `graph.go` — stubs (unchanged)

```
nfv2/
├── cmd/nforge/          # CLI entrypoint (Cobra commands)
│   ├── root.go           # ← UPDATE: replace stub config command
│   ├── config.go        # ← NEW: full config set/get (THIS STORY)
│   ├── serve.go         # nforge serve (starts Gin + WebSocket) — story 1.3
│   ├── run.go           # nforge run <spec-file> — story 1.8
│   ├── new.go           # nforge new <project-name> — story 1.4
│   ├── skill.go         # nforge skill list/install — story 5.1
│   ├── session.go       # nforge session resume/export — story 4.5
│   ├── doctor.go        # nforge doctor (health check) — story 1.6
│   └── graph.go         # nforge graph viz (ASCII art) — story 1.7
│
├── main.go              # ← UNCHANGED: calls root.go Execute()
...
```
— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**None** — this story delivers the full config implementation that was stubbed in story 1.2.

**Critical Reminder:** Viper must be initialized with the `--config-path` flag value (from `root.go` persistent flags) before any config operations. Use `viper.SetConfigFile()` if `--config-path` is provided, otherwise default to `~/.nforge/config.yaml`.

## References

- [Story 1.5 Definition: epics.md#Story-1.5](_bmad-output/planning-artifacts/epics.md#Story-1.5)
- [Story 1.2 (Previous): 1-2-cli-root-command-with-cobra-framework.md](_bmad-output/implementation-artifacts/1-2-cli-root-command-with-cobra-framework.md)
- [Architecture - CLI Structure: architecture.md#Starter-Template-Evaluation](_bmad-output/planning-artifacts/architecture.md#Starter-Template-Evaluation)
- [Architecture - Naming Patterns: architecture.md#Naming-Patterns](_bmad-output/planning-artifacts/architecture.md#Naming-Patterns)
- [Cobra v1.10.2 Documentation](https://github.com/spf13/cobra/releases/tag/v1.10.2)
- [Cobra + Viper Config Patterns (2025-2026)](https://cobra.dev/)
- [Viper v1.19.0 Documentation](https://github.com/spf13/viper/releases/tag/v1.19.0)
- [PRD - CLI Capabilities: prd.md#CLI-Capabilities](_bmad-output/planning-artifacts/prd.md#CLI-Capabilities)

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List

- `cmd/nforge/config.go` (NEW)
- `cmd/nforge/root.go` (UPDATE — replace stub config command)
- `cmd/nforge/config_test.go` (NEW — tests for config set/get)
- `go.mod` (UPDATE — add Viper dependency if missing: `go get github.com/spf13/viper@v1.19.0`)
