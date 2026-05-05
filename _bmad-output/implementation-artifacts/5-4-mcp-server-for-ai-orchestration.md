# Story 5.4: MCP Server for AI Orchestration

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want NodeForge to expose an MCP server so that Claude Desktop/Cursor can orchestrate my workflows via MCP tools,
so that I can integrate NodeForge into my existing AI-assisted development setup.

## Acceptance Criteria

1. **Given** the MCP server is implemented (`internal/skills/mcp.go`)
   **When** Claude Desktop or Cursor connects via MCP
   **Then** full session lifecycle tools are exposed: `create_node`, `run_node`, `get_status`, `fork_session`, `export_session` (FR44, NFR-27)

2. **Given** the MCP server is running
   **When** a client connects via MCP protocol
   **Then** the server uses JSON-RPC 2.0 format per MCP spec

3. **Given** the MCP tools are registered
   **When** an AI client queries available tools
   **Then** tools are documented with input/output schemas for AI consumption

4. **Given** a session exists
   **When** an MCP tool call modifies session state (e.g., `create_node`, `fork_session`)
   **Then** session state is accessible and modifiable via MCP tool calls

## Tasks / Subtasks

- [ ] Implement MCP server core (AC: 1,2)
  - [ ] Define MCP tool schemas (create_node, run_node, get_status, fork_session, export_session)
  - [ ] Implement JSON-RPC 2.0 handler per MCP spec
  - [ ] Create `internal/skills/mcp.go` with MCP server logic
- [ ] Expose session lifecycle tools (AC: 1,4)
  - [ ] Implement `create_node` tool (create node in existing session)
  - [ ] Implement `run_node` tool (trigger node execution)
  - [ ] Implement `get_status` tool (retrieve session/graph status)
  - [ ] Implement `fork_session` tool (fork existing session)
  - [ ] Implement `export_session` tool (export session as tarball)
- [ ] Document tools for AI consumption (AC: 3)
  - [ ] Add input/output schemas to each tool
  - [ ] Ensure schemas are discoverable via MCP tools/list endpoint
- [ ] Testing (AC: 1,2,3,4)
  - [ ] Unit tests for MCP server JSON-RPC handling
  - [ ] Integration tests for tool execution against live session
  - [ ] Test MCP compliance (JSON-RPC 2.0 format validation)

## Dev Notes

### Architecture Patterns & Constraints

- **MCP Server Location**: `internal/skills/mcp.go` (per architecture.md API & Communication Patterns)
- **MCP Tools**: `create_node`, `run_node`, `get_status`, `fork_session`, `export_session` (FR44, NFR-27)
- **Protocol**: JSON-RPC 2.0 format per MCP spec (epics.md 5.4 AC2)
- **Integration Points**:
  - Session manager (`internal/session/`) for session lifecycle operations
  - Graph engine (`internal/engine/`) for node creation/execution
  - Skill system (`internal/skills/`) where MCP server resides
- **Naming Conventions**:
  - Go: `camelCase` functions, `PascalCase` structs, `snake_case` package (project-context.md)
  - JSON: `camelCase` fields in request/response payloads
- **Security**: MCP server must validate session access; no unauthorized session modification
- **Dependencies**: MCP server depends on session management + LLM abstraction (architecture.md Integration Points)

### Source Tree Components to Touch

| File | Action | Purpose |
|-----|--------|---------|
| `internal/skills/mcp.go` | NEW | MCP server implementation with JSON-RPC 2.0 handler |
| `internal/skills/manager.go` | UPDATE | Register MCP server in skill system initialization |
| `cmd/nforge/serve.go` | UPDATE | Add MCP server startup (if part of main server) |
| `internal/session/manager.go` | READ | Called by MCP tools for session operations |
| `internal/engine/graph.go` | READ | Called by MCP tools for node operations |

### Testing Standards

- **Framework**: Go `testify` + `go test ./internal/skills/...`
- **Patterns**: Table-driven tests for JSON-RPC handling, mock session/engine for tool tests
- **Coverage**: All 5 MCP tools + JSON-RPC parsing edge cases
- **Boundaries**: Unit: MCP server logic; Integration: Tool execution with live session
- **Requirements**: CGO_ENABLED=1 for SQLite tests (project-context.md Testing Rules)

## Project Structure Notes

- **Alignment**: Follows `internal/skills/` package structure per architecture.md Project Structure
- **No Conflicts**: MCP server is a new component in existing `internal/skills/` package
- **Consistency**: Matches existing skill system patterns (gRPC plugins in same package)

## References

- [Source: epics.md#Story5.4] Epic 5, Story 5.4: MCP Server for AI Orchestration (Acceptance Criteria)
- [Source: architecture.md#API&CommunicationPatterns] MCP Server (FR44, NFR-27): Tools list, JSON-RPC 2.0 format
- [Source: architecture.md#ProjectStructure] `internal/skills/mcp.go` for MCP server implementation
- [Source: architecture.md#IntegrationPoints] MCP server depends on session management + LLM abstraction
- [Source: project-context.md#NamingConventions] Go naming: snake_case packages, camelCase functions, PascalCase structs
- [Source: project-context.md#JSON] JSON fields: camelCase (json:"sessionId")
- [Source: project-context.md#TestingRules] Go: `go test ./...` + testify, CGO for SQLite
- [Source: prd.md#GrowthFeatures] MCP Server listed as growth feature: Claude Desktop/Cursor orchestrates NodeForge via MCP
- [Source: prd.md#Nice-to-Have] MCP Server: Claude Desktop/Cursor orchestrates NodeForge via MCP
- [Source: prd.md#FR44] System exposes MCP server — Claude Desktop, Cursor can orchestrate NodeForge via MCP tools
- [Source: prd.md#NFR-27] MCP Server Compliance: full session lifecycle exposed

## Dev Agent Record

### Agent Model Used

tencent/hy3-preview:free

### Debug Log References

### Completion Notes List

### File List
