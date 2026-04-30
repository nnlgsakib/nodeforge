# Story 1.2: CLI Root Command with Cobra Framework

Status: ready-for-dev

<!-- Validation: Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want a working `nforge` CLI with root command and persistent flags,
so that I can access all subcommands consistently.

## Acceptance Criteria

**Given** the CLI scaffold is set up with Cobra (story 1.1 completed)
**When** the user runs `nforge --help`
**Then** the root command displays usage information with available subcommands: `serve`, `run`, `new`, `config`, `skill`, `session`, `doctor`, `graph`
**And** persistent flags (`--verbose`, `--config-path`) are available across all subcommands
**And** the CLI displays version information with `nforge --version`

## Tasks / Subtasks

- [ ] Task 1: Create Cobra root command with persistent flags (AC: 2, 3)
  - [ ] Subtask 1.1: Create `cmd/nforge/root.go` with Cobra root command
    ```go
    var version = "dev" // Set via ldflags at build time

    var rootCmd = &cobra.Command{
        Use:     "nforge",
        Short:   "NodeForge OS - Spec-driven development workbench",
        Version: version,
    }
    func Execute() error { return rootCmd.Execute() }
    ```
  - [ ] Subtask 1.2: Add persistent flags via `rootCmd.PersistentFlags()`:
    - `--verbose` (`bool`, env: `NFORGE_VERBOSE`) — enables debug logging
    - `--config-path` (`string`, default: `~/.nforge/config.yaml`, env: `NFORGE_CONFIG`)
  - [ ] Subtask 1.3: Cobra's built-in `--version` flag automatically uses `Version` field — verify `nforge --version` prints `&version` value

- [ ] Task 2: Register all subcommands in root (AC: 1)
  - [ ] Subtask 2.1: Register `serve` subcommand (stub for now, full impl in story 1.3)
    ```go
    // cmd/nforge/serve.go stub
    var serveCmd = &cobra.Command{
        Use:   "serve",
        Short: "Start web UI + API server",
        RunE:  func(cmd *cobra.Command, args []string) error {
            fmt.Println("serve: not yet implemented (story 1.3)")
            return nil
        },
    }
    func init() { rootCmd.AddCommand(serveCmd) }
    ```
  - [ ] Subtask 2.2: Register `run` subcommand (stub for now, full impl in story 1.8)
  - [ ] Subtask 2.3: Register `new` subcommand (stub for now, full impl in story 1.4)
  - [ ] Subtask 2.4: Register `config` subcommand (stub for now, full impl in story 1.5)
  - [ ] Subtask 2.5: Register `skill` subcommand (stub for now, full impl in story 1.5/skill system)
  - [ ] Subtask 2.6: Register `session` subcommand (stub for now, full impl in story 4.5)
  - [ ] Subtask 2.7: Register `doctor` subcommand (stub for now, full impl in story 1.6)
  - [ ] Subtask 2.8: Register `graph` subcommand (stub for now, full impl in story 1.7)

- [ ] Task 3: Wire root command into main.go (AC: 1, 2, 3)
  - [ ] Subtask 3.1: Update `main.go` to call `cmd/nforge/root.go` Execute()
  - [ ] Subtask 3.2: Add to `Makefile`:
    ```makefile
    VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
    LDFLAGS = -X github.com/nlg/nfv2/cmd/nforge.version=$(VERSION)
    build:
    	go build -ldflags "$(LDFLAGS)" -o nforge main.go
    ```
  - [ ] Subtask 3.3: Verify `nforge --help` shows all 8 subcommands
  - [ ] Subtask 3.4: Verify `nforge --version` displays version string

- [ ] Task 4: Implement persistent flag behavior (AC: 2)
  - [ ] Subtask 4.1: `--verbose` flag enables debug logging across all subcommands
  - [ ] Subtask 4.2: `--config-path` flag overrides default config location (`~/.nforge/config.yaml`)
  - [ ] Subtask 4.3: Verify flags work on subcommands: `nforge serve --verbose --config-path /tmp/cfg.yaml`

- [ ] Task 5: Verify end-to-end (AC: 1, 2, 3)
  - [ ] Subtask 5.1: Run `go build -o nforge main.go` — binary compiles
  - [ ] Subtask 5.2: Run `./nforge --help` — shows all 8 subcommands
  - [ ] Subtask 5.3: Run `./nforge --version` — displays version
  - [ ] Subtask 5.4: Run `./nforge serve --verbose` — accepted without error (stub executes)

- [ ] Task 6: Update `.gitignore` (AC: 2)
  - [ ] Subtask 6.1: Add `nforge` (built binary) to `.gitignore`
  - [ ] Subtask 6.2: Verify `git status` does not show binary

## Dev Notes

### Architecture Patterns and Constraints

**Go Version:** Go 1.24+ (62% GC pause reduction, improved reflect.Blueprint support) — [Source: architecture.md#Project-Context-Analysis]

**Framework Choices (Non-Negotiable):**
- **Cobra v1.10.2** for CLI with subcommands: `serve`, `run`, `new`, `config`, `skill`, `session`, `doctor`, `graph` — [Source: epics.md#Epic-1, architecture.md#Starter-Template-Evaluation]
- **Gin 1.11.0** (radix tree router, 38% lower allocation overhead, HTTP/3 support) — NOT Chi. Single framework for REST API + WebSocket hub — [Source: architecture.md#API-Communication-Patterns]

**Project Structure (MUST Match):**
```
cmd/nforge/
├── root.go           # Cobra root command, persistent flags (THIS STORY)
├── serve.go         # nforge serve (starts Gin + WebSocket) — story 1.3
├── run.go           # nforge run <spec-file> — story 1.8
├── new.go           # nforge new <project-name> — story 1.4
├── config.go        # nforge config set/get — story 1.5
├── skill.go         # nforge skill list/install — story 1.5/skill system
├── session.go       # nforge session resume/export — story 4.5
├── doctor.go        # nforge doctor (health check) — story 1.6
└── graph.go         # nforge graph viz (ASCII art) — story 1.7
```
— [Source: architecture.md#Complete-Project-Directory-Structure]

**Naming Conventions (CRITICAL — Must Follow):**
- Go packages: `snake_case` — `cmd/nforge/`, `internal/engine/`
- Go functions: `camelCase` — `executeNode(ctx)`
- Go structs: `PascalCase` — `type Session struct`, `type RootCommand struct`
- Go variables: `camelCase` — `sessionID`, `configPath`
- Go constants: `PascalCase` or `UPPER_SNAKE` — `MaxSessions`, `DefaultTimeout`
- CLI flags: `kebab-case` — `--config-path`, `--verbose`
— [Source: architecture.md#Naming-Patterns]

### Source Tree Components to Touch

**New Files (CREATE):**
- `cmd/nforge/root.go` — Cobra root command with persistent flags `--verbose`, `--config-path`, version flag

**UPDATE Files (MODIFY):**
- `main.go` — Wire Cobra root command Execute() instead of (or in addition to) Gin server startup
  - Current state: `main.go` has Gin server setup and embed.FS foundation (from story 1.1)
  - What this story changes: Add Cobra command execution; Gin server will be started via `nforge serve` subcommand
  - What must be preserved: Gin imports and setup code must NOT be removed — they'll be moved to `serve.go` in story 1.3

**Files That Must NOT Be Modified:**
- `go.mod` — already has cobra dependency from story 1.1
- `internal/` — no changes needed yet
- `frontend/` — no changes needed yet

### Testing Standards

- **Go**: Ginkgo + Testify (from cli-go-project-template pattern) — co-located `*_test.go` files
- **Test for this story:**
  - `cmd/nforge/root_test.go` — Test root command help output, version flag, persistent flags
  - Verify `nforge --help` returns expected subcommands
  - Verify `nforge --version` returns non-empty version string
  - Verify `--verbose` and `--config-path` flags are accepted on all subcommands
— [Source: architecture.md#Starter-Template-Evaluation]

### Previous Story Intelligence (from Story 1.1)

**What Story 1.1 Established:**
- Go module initialized: `github.com/nlg/nfv2` with Go 1.24+
- Dependencies installed: `gin-gonic/gin v1.11.0`, `spf13/cobra v1.10.2`, `google.golang.org/protobuf`
- Directory structure created: `cmd/nforge/`, `internal/` (with engine, llm, context, session, skills, canvas, security, devops subdirectories), `frontend/`, `proto/`
- `main.go` created with Gin server setup and embed.FS foundation
- `Makefile`, `Dockerfile`, `docker-compose.yml` created

**Key Learnings from Story 1.1:**
- Cobra v1.10.2 is already in `go.mod` — do NOT run `go get cobra` again
- The `cmd/nforge/` directory already exists — just add `root.go` file
- `main.go` currently has Gin server setup — we need to integrate Cobra WITHOUT removing Gin setup (it moves to `serve.go` in 1.3)
- Naming conventions established: `camelCase` functions, `PascalCase` structs, `snake_case` packages

**Integration Point:**
- Story 1.1 created the scaffold; Story 1.2 adds the CLI structure
- Story 1.3 will implement `serve.go` (move Gin setup there)
- All 8 subcommands should be registered as stubs now; full implementation per subcommand happens in later stories

## Project Structure Notes

### Alignment with Unified Project Structure

The `cmd/nforge/` directory must contain:
- `root.go` — THIS STORY (Cobra root with persistent flags)
- `serve.go`, `run.go`, `new.go`, `config.go`, `skill.go`, `session.go`, `doctor.go`, `graph.go` — stubs registered in this story, implemented in later stories

```
nfv2/
├── cmd/nforge/          # CLI entrypoint (Cobra commands)
│   ├── root.go           # ← THIS STORY: Cobra root command, persistent flags
│   ├── serve.go         # nforge serve (starts Gin + WebSocket) — story 1.3
│   ├── run.go           # nforge run <spec-file> — story 1.8
│   ├── new.go           # nforge new <project-name> — story 1.4
│   ├── config.go        # nforge config set/get — story 1.5
│   ├── skill.go         # nforge skill list/install — story 1.5
│   ├── session.go       # nforge session resume/export — story 4.5
│   ├── doctor.go        # nforge doctor (health check) — story 1.6
│   └── graph.go         # nforge graph viz (ASCII art) — story 1.7
│
├── main.go              # ← UPDATE: Wire Cobra root command Execute()
...
```
— [Source: architecture.md#Complete-Project-Directory-Structure]

### Detected Conflicts or Variances

**Conflict: main.go currently has Gin server in `main()` function**
- **Resolution:** This story should refactor `main.go` to use Cobra's `root.go Execute()` as the entry point
- The Gin server setup code stays in `main.go` for now BUT gets moved to `serve.go` in story 1.3
- For this story: `main.go` calls `cmd/nforge/root.go` Execute(), and root command's `serve` subcommand will later call the Gin setup

**Important:** Do NOT delete Gin setup from `main.go` in this story — just add the Cobra wiring. Story 1.3 will properly separate it.

## References

- [Story 1.2 Definition: epics.md#Story-1.2](_bmad-output/planning-artifacts/epics.md#Story-1.2)
- [Story 1.1 (Previous): epics.md#Story-1.1](_bmad-output/planning-artifacts/epics.md#Story-1.1)
- [Architecture - CLI Structure: architecture.md#Starter-Template-Evaluation](_bmad-output/planning-artifacts/architecture.md#Starter-Template-Evaluation)
- [Architecture - Project Structure: architecture.md#Complete-Project-Directory-Structure](_bmad-output/planning-artifacts/architecture.md#Complete-Project-Directory-Structure)
- [Architecture - Naming Patterns: architecture.md#Naming-Patterns](_bmad-output/planning-artifacts/architecture.md#Naming-Patterns)
- [Cobra v1.10.2 Documentation](https://github.com/spf13/cobra/releases/tag/v1.10.2)
- [Cobra User Guide](https://github.com/spf13/cobra/blob/main/user_guide.md)

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List

- `cmd/nforge/root.go` (NEW)
- `main.go` (UPDATE — wire Cobra root command)
- `cmd/nforge/root_test.go` (NEW — tests for root command)
- `Makefile` (UPDATE — add `-ldflags` for version variable)

