# Story 5.3: gRPC Plugins & Sub-Nodes

Status: ready-for-dev

## Story

As a third-party developer,
I want to define entirely new node types via gRPC plugins with sub-nodes support,
so that I can extend NodeForge beyond built-in functionality.

## Acceptance Criteria

1. **Given** the gRPC plugin interface is defined (`proto/plugin.proto`) with Unix socket IPC
   **When** the proto file is compiled and ready
   **Then** it defines `PluginService` with `Execute(PluginRequest) returns (PluginResponse)` and `Health(HealthRequest) returns (HealthResponse)` RPCs
   **And** message types use `session_id`, `node_type`, `payload` (bytes), `result` (bytes), `success`, `error` fields
   **And** the proto package is `nodeforge` with `go_package = "github.com/nlg/nfv2/proto"`

2. **Given** the gRPC plugin loader is implemented in `internal/skills/grpc.go`
   **When** a third-party plugin is configured in a skill manifest
   **Then** it can define entirely new node types that the core engine can execute (FR43, NFR-26)
   **And** plugins communicate via Unix socket IPC (path: `<skill-dir>/plugin.sock`)
   **And** plugin communication uses protobuf-defined messages for type safety

3. **Given** sub-node support is implemented in `internal/skills/subnodes.go`
   **When** a skill manifest declares sub-nodes (e.g., "JS-to-Go" skill has "patterns," "goroutines" sub-nodes)
   **Then** the sub-nodes are available as child nodes under the parent skill node (FR45)
   **And** sub-nodes inherit the parent skill's workspace context
   **And** sub-nodes can be individually executed, retried, or skipped

4. **Given** plugins run as separate processes
   **When** a plugin is started for a skill
   **Then** it runs in a separate process with resource limits (CPU, memory constraints)
   **And** crash of one plugin doesn't affect the core engine or other plugins (NFR-26)
   **And** the plugin process is sandboxed (chroot jail per session, no network access for untrusted plugins)

5. **Given** the skill manifest schema is extended
   **When** a skill includes plugin or sub-node configuration
   **Then** the `SkillManifest` struct supports `Plugin` field (socket path, binary path) and `SubNodes` field (list of sub-node definitions)
   **And** manifest validation rejects invalid plugin configurations

## Tasks / Subtasks

- [ ] Task 1: Compile proto and generate gRPC code (AC: #1)
  - [ ] Subtask 1.1: Verify `proto/plugin.proto` is complete with all required message types
  - [ ] Subtask 1.2: Run `protoc --go_out=. --go-grpc_out=. proto/plugin.proto` to generate Go code
  - [ ] Subtask 1.3: Add generated files to `proto/` directory

- [ ] Task 2: Implement gRPC plugin loader in `internal/skills/grpc.go` (AC: #2)
  - [ ] Subtask 2.1: Create `PluginLoader` struct with methods to start, stop, and communicate with plugins
  - [ ] Subtask 2.2: Implement Unix socket connection (use `net.Dial("unix", path)`)
  - [ ] Subtask 2.3: Implement `ExecuteNode` method that sends `PluginRequest` and receives `PluginResponse`
  - [ ] Subtask 2.4: Implement health check via `Health` RPC
  - [ ] Subtask 2.5: Add plugin process management (start plugin binary, track PID, handle cleanup)

- [ ] Task 3: Implement sub-node support in `internal/skills/subnodes.go` (AC: #3)
  - [ ] Subtask 3.1: Extend `SkillManifest` struct with `SubNodes` field (in `manifest.go`)
  - [ ] Subtask 3.2: Create `SubNode` type with fields: ID, Name, Description, InputSchema, OutputSchema
  - [ ] Subtask 3.3: Implement sub-node registration and lookup functions
  - [ ] Subtask 3.4: Integrate sub-nodes with engine so they appear as child nodes in graph

- [ ] Task 4: Add plugin sandboxing and resource limits (AC: #4)
  - [ ] Subtask 4.1: Implement plugin process sandboxing (chroot to skill directory)
  - [ ] Subtask 4.2: Add resource limits (CPU/memory) using OS-level controls
  - [ ] Subtask 4.3: Implement plugin crash isolation (goroutine with recovery, don't kill core engine)
  - [ ] Subtask 4.4: Add plugin cleanup on session shutdown

- [ ] Task 5: Extend skill manifest validation (AC: #5)
  - [ ] Subtask 5.1: Add `Plugin` field to `SkillManifest` struct (`PluginConfig` with `SocketPath`, `BinaryPath`)
  - [ ] Subtask 5.2: Update `LoadManifest` to parse new fields
  - [ ] Subtask 5.3: Add manifest validation for plugin configs (valid paths, required fields)
  - [ ] Subtask 5.4: Add sub-nodes to manifest validation

- [ ] Task 6: Testing (All ACs)
  - [ ] Subtask 6.1: Unit tests for gRPC plugin loader (mock plugin server)
  - [ ] Subtask 6.2: Unit tests for sub-node registration and lookup
  - [ ] Subtask 6.3: Integration test: compile proto, start mock plugin, execute node
  - [ ] Subtask 6.4: Test plugin crash isolation (verify core engine survives)

## Dev Notes

### Architecture Patterns and Constraints

**gRPC Plugin System (from architecture.md):**
- Proto definitions in `proto/plugin.proto` — `PluginService` with `Execute` and `Health` RPCs
- Unix socket IPC for plugin communication (not HTTP)
- Plugins are separate processes with resource limits
- Crash of one plugin must not affect core engine (NFR-26)

**Skill System Integration:**
- Plugin loader lives in `internal/skills/grpc.go`
- Sub-node support in `internal/skills/subnodes.go`
- Manifest extended in `internal/skills/manifest.go` (existing file)
- Dependency resolution in `internal/skills/resolver.go` (existing, handles `Dependencies` field)

**Security Requirements (NFR-26, from architecture.md):**
- Plugins run in separate processes (not goroutines in main process)
- Chroot jail per session (reuse `internal/security/chroot.go`)
- Resource limits: CPU and memory constraints
- eBPF syscall filtering for untrusted plugins (from `internal/security/ebpf.go`)
- No network access for untrusted plugins unless explicitly configured

**Protobuf Code Generation:**
- Use `protoc` with Go gRPC plugin: `protoc --go_out=. --go-grpc_out=. proto/plugin.proto`
- Requires `google.golang.org/protobuf` and `google.golang.org/grpc` dependencies
- Generated files go in `proto/` directory

### Source Tree Components to Touch

| File | Action | Purpose |
|------|--------|---------|
| `proto/plugin.proto` | READ (exists) | Verify proto definition is complete |
| `proto/plugin.pb.go` | CREATE | Generated gRPC Go code |
| `proto/plugin_grpc.pb.go` | CREATE | Generated gRPC service code |
| `internal/skills/grpc.go` | CREATE | gRPC plugin loader |
| `internal/skills/subnodes.go` | CREATE | Sub-node support |
| `internal/skills/manifest.go` | UPDATE | Add `Plugin`, `SubNodes` fields to `SkillManifest` |
| `internal/skills/resolver.go` | READ | Understand dependency resolution for sub-node integration |
| `internal/security/chroot.go` | READ | Reuse chroot jail for plugin sandboxing |
| `internal/security/ebpf.go` | READ | Reuse eBPF filtering for untrusted plugins |
| `internal/engine/graph.go` | UPDATE (if needed) | Register new node types from plugins |
| `internal/engine/executor.go` | UPDATE (if needed) | Execute plugin nodes and sub-nodes |

### Testing Standards Summary

**Go Testing (from project-context.md):**
- Framework: `go test ./...` + `testify` assertions
- Pattern: Table-driven tests preferred
- Mock: `go.uber.org/mock` for gRPC client mocks
- Co-location: `*_test.go` files in same package
- CGO: Not required for gRPC (pure Go)

**Test Coverage Requirements:**
- gRPC plugin loader: Mock plugin server, test Execute, Health, crash handling
- Sub-node registration: Test registration, lookup, parent-child relationships
- Manifest validation: Test new fields parse correctly, invalid configs rejected
- Integration: Full flow from skill install → plugin start → node execution

## Project Structure Notes

### Alignment with Unified Project Structure

**Files created in this story align with architecture.md:**
```
proto/
└── plugin.proto              # EXISTS - verify completeness

internal/skills/
├── manifest.go              # EXISTS - UPDATE: add Plugin, SubNodes fields
├── resolver.go              # EXISTS - READ for dependency patterns
├── grpc.go                  # CREATE - gRPC plugin loader
├── subnodes.go             # CREATE - Sub-node support
└── skills_test.go          # EXISTS - ADD tests for new functionality
```

### Detected Conflicts or Variances

- **None detected** — This is the first story in Epic 5, creating new functionality that doesn't conflict with existing code.
- **Proto compilation**: Ensure `protoc` is available in build environment or commit generated files.
- **Plugin binary path**: Must be relative to skill directory for portability.

## References

- [Source: architecture.md#API & Communication Patterns] gRPC Plugin Interface: `PluginService.proto` — `ExecuteNode`, `GetNodeSchema` RPCs, Unix socket IPC
- [Source: architecture.md#Core Architectural Decisions] Plugin System: gRPC interface, MCP server, plugins sandboxed (separate process, resource limits)
- [Source: architecture.md#Project Structure & Boundaries] `internal/skills/grpc.go` for gRPC plugin loader, `proto/plugin.proto` for definitions
- [Source: epics.md#Story 5.3] Acceptance criteria for gRPC plugins and sub-nodes
- [Source: prd.md#Growth Features] gRPC Plugin Interface — Third-party node types, dynamic loading
- [Source: prd.md#Nice-to-Have] gRPC Plugins — Third-party node types, dynamic loading
- [Source: project-context.md#Critical Don't-Miss Rules] Plugins: gRPC + chroot + resource limits; Avoid direct function calls
- [Source: project-context.md#Security (MUST)] Plugins: gRPC + chroot + resource limits
- [Source: architecture.md#Decision Impact Analysis] gRPC plugins depend on security layer (sandboxing)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
