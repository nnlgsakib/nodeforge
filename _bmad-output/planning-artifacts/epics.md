---
stepsCompleted: [1, 2, 3, 4]
inputDocuments: ['_bmad-output/planning-artifacts/prd.md', '_bmad-output/planning-artifacts/architecture.md', '_bmad-output/planning-artifacts/ux-design-specification.md']
---

# nfv2 - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for nfv2, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: User can create a new session with a goal description ("Convert JS→Go") and AI auto-generates a complete node graph (Goal → Spec → Plan → Implement → Test → Review)
FR2: User can see the entire project state at a glance with color-coded nodes (green=complete, red=failed, yellow=running)
FR3: User can watch nodes execute deterministically — each node works until acceptance criteria are met, then advances forward
FR4: User can interact with node connections n8n-style — pluck edges to see metadata, see data/control flow (TouchDesigner-style interactive wires)
FR5: User can view DaVinci Resolve-style node trees with clear input/output pins and signal flow
FR6: User can see animated pulses along edges showing real-time execution progress (n8n-inspired canvas)
FR7: User can deactivate/activate individual nodes without deleting them (n8n feature)
FR8: User can view a mini-map with execution heat — nodes glow based on recent activity
FR9: User can add comments/notes to nodes (sticky notes like n8n)
FR10: User can configure multiple LLM providers (OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama) via `nforge config`
FR11: System can race multiple LLM providers simultaneously — fastest token wins, slower is cancelled
FR12: System can auto-fallback through providers when rate limits or errors occur (semantic matching: rate limit → cheaper model)
FR13: User can see LLM "Inner Monologue" (Chain-of-Thought) in a side panel with streaming tokens
FR14: System can optimize prompts automatically based on execution feedback (prompt learns over time)
FR15: System can enforce token budgets — pre-flight estimation rejects expensive requests
FR16: User can use local Ollama (free) racing against remote APIs (cost control)
FR17: System builds a knowledge graph for token-efficient context assembly (30%+ reduction vs naive prompts)
FR18: System reuses node memory — each node's output becomes context for downstream nodes
FR19: System auto-generates specs and adds system references as nodes execute
FR20: System can handle context overflow by auto-splitting graphs into sub-graphs
FR21: User can start the web UI + API with `nforge serve`
FR22: User can run headless execution with `nforge run <spec-file>` (same as UI, but in terminal)
FR23: User can create new projects with `nforge new <project-name>`
FR24: User can configure settings with `nforge config set/get` (API keys, models, ports)
FR25: User can manage skills with `nforge skill list/install <name>`
FR26: User can resume/export sessions with `nforge session resume/export <id>`
FR27: User can see ASCII art graph in terminal with `nforge graph viz`
FR28: User can check system health with `nforge doctor`
FR29: CLI has tab completion for all commands and node types
FR30: CLI and UI have feature parity — same graphs execute identically
FR31: User can create isolated sessions with separate workspaces
FR32: System auto-saves session state — graph JSON, chat log, workspace files
FR33: User can resume sessions after restart (graceful shutdown with snapshot)
FR34: User can fork sessions (like Git branches) — try different approaches, merge or discard
FR35: System auto-commits workspace changes to Git after each node completion
FR36: User can time-travel debug — checkout workspace state at any completed node
FR37: User can export session as self-contained tarball (graph + source + README)
FR38: System enforces session quotas (max sessions, max workspace size)
FR39: System auto-cleans zombie sessions (heartbeat timeout detection)
FR40: User can browse and install skills from marketplace (community-contributed manifests)
FR41: Skills can have dependencies — installing one skill auto-installs its dependency tree
FR42: Skills are sandboxed before trust — time limits, no network, read-only filesystem
FR43: System supports gRPC plugins — third-parties can define entirely new node types
FR44: System exposes MCP server — Claude Desktop, Cursor can orchestrate NodeForge via MCP tools
FR45: Skills can have sub-nodes — e.g., "JS-to-Go" skill has "patterns," "goroutines" sub-nodes
FR46: User can A/B test skills — system routes to different versions, collects metrics
FR47: User can drag-and-drop files onto canvas to auto-create nodes (e.g., `go.mod` → `Setup` node)
FR48: User can see color-coded node bands by lifecycle phase (blue=Discovery, orange=Execution, red=Recovery, green=Completion)
FR49: User can see reactive edge tension — edges visually tighten when upstream nodes fail
FR50: User can use Vim/Emacs keybindings for canvas navigation (hjkl, Ctrl-f/b/n/p)
FR51: System supports accessibility — screen reader announcements for node status changes, high-contrast themes, RTL canvas, 20+ language localization
FR52: System executes nodes sequentially with retry loops — stays inside node until acceptance criteria met
FR53: System can run multiple attempts in parallel within a node (speculative execution, best result wins)
FR54: System supports incremental execution — Merkle tree hashing skips unchanged nodes
FR55: System offloads graph layout to Web Worker — 100+ node graphs render smoothly
FR56: System pre-fetches LLM provider status on session start — zero wait for connectivity checks
FR57: System uses Gin backend — single framework for REST API + WebSocket hub (5000+ concurrent connections)
FR58: System encrypts API keys at rest using Argon2 key derivation + AES-256-GCM
FR59: System sandboxes node execution in chroot jail — no escape to parent directories
FR60: System applies eBPF syscall filtering — blocks dangerous system calls during execution
FR61: System enforces rate limiting per API key using token bucket algorithm
FR62: System signs graph snapshots with Ed25519 — detects tampering on session import
FR63: System builds as multi-arch Docker (amd64 + arm64) in single container
FR64: System uses distroless Docker image — only Go binary, no shell, no OS utilities
FR65: System includes Ollama sidecar option — auto-provisions local LLM on `docker compose up`
FR66: System exposes health check with graph diagnostics (/healthz returns session stats, LLM connectivity)
FR67: System can send webhook notifications — nodes reach out to Slack, GitHub, etc. when certain states reached
FR68: System exports telemetry as Prometheus metrics (/metrics endpoint)

### NonFunctional Requirements

NFR-01: WebSocket state propagation <50ms latency from node state change to browser UI (Gin WS hub, 5000+ concurrent connections)
NFR-02: Graph rendering 100+ node graphs render at 60fps with Web Worker offloading, zero main-thread blocking
NFR-03: LLM Race Mode sub-200ms provider race wins — fastest first token wins, slower providers cancelled immediately
NFR-04: Smart Context Assembly knowledge graph context assembled in <100ms, achieving 30%+ token reduction vs naive prompt assembly
NFR-05: Token Budget Pre-flight budget estimation completes in <10ms before LLM call is dispatched
NFR-06: Incremental Execution Merkle tree hashing skips unchanged nodes — re-execution of 100-node graph completes in <2s when 95% unchanged
NFR-07: API Key Storage Argon2 key derivation + AES-256-GCM encryption at rest; keys never in plaintext config files
NFR-08: Workspace Isolation Chroot jail per session — node execution cannot escape to parent directories; verified by integration tests
NFR-09: Syscall Filtering eBPF filters block dangerous syscalls (exec, mount, reboot) during node execution
NFR-10: Session Secrets Vault integration for secret management; API keys never logged, never in session export tarballs
NFR-11: Rate Limiting Token bucket algorithm, per-API-key limits — 100 req/min for REST, 10 msg/sec for WebSocket
NFR-12: Graph Integrity Ed25519 signing of graph snapshots — tampered session imports rejected with audit log entry
NFR-13: Skill Sandboxing Pre-trust execution: time limits (30s), no network, read-only filesystem; escalated after signature verification
NFR-14: Concurrent Connections Single Gin instance supports 5000+ WebSocket connections with <5% memory growth beyond baseline
NFR-15: User Growth Trajectory Architecture supports growth from 1,000 to 10,000+ solo developers without structural changes (horizontal scaling ready via stateless Gin + external session store)
NFR-16: Graph Complexity Smooth interaction with 100+ node graphs; layout Web Worker prevents UI freeze; pan/zoom remains 60fps
NFR-17: Session Density 100+ concurrent sessions per instance, each with isolated workspace (max 500MB workspace size quota)
NFR-18: Multi-Arch Distribution Docker images built for amd64 + arm64 in single multi-arch manifest; Ollama sidecar auto-provisions on `docker compose up`
NFR-19: LLM Provider Failover 99.9% LLM uptime via fallback chains — Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter
NFR-20: Screen Reader Support WCAG 2.1 AA compliance — node status changes announced via ARIA live regions; canvas keyboard-navigable
NFR-21: Visual Accessibility High-contrast themes + colorblind-friendly palette — nodes distinguished by shape + position, not just color (red/green alone insufficient)
NFR-22: Internationalization RTL canvas support, 20+ language localization (minimum: EN, ES, FR, DE, ZH, JA, KO, PT, RU, AR)
NFR-23: Keyboard Navigation Vim (hjkl) and Emacs (Ctrl-f/b/n/p) keybindings for canvas navigation; all interactions possible without mouse
NFR-24: Motor Impairment Support All node operations (create, connect, delete, configure) accessible via keyboard shortcuts and context menu
NFR-25: LLM Provider Agnosticism 5+ providers (Ollama, OpenAI, Anthropic, DeepSeek, OpenRouter) — pluggable provider interface, new providers added via config (no code changes)
NFR-26: gRPC Plugin System Third-party node types load dynamically via gRPC; plugins sandboxed (separate process, resource limits); crash of one plugin doesn't affect core engine
NFR-27: MCP Server Compliance Claude Desktop/Cursor orchestrates NodeForge via MCP tools — full session lifecycle exposed (create, resume, fork, export)
NFR-28: Git Integration Auto-commit per node completion with deterministic commit messages; time-travel debug via `git checkout` at any completed node
NFR-29: Webhook Notifications Outbound webhooks to Slack, GitHub, etc. when nodes reach configurable states (success, failure, approval needed); retries with exponential backoff
NFR-30: Telemetry Export Prometheus metrics (/metrics endpoint) — session count, LLM token usage, node execution duration, WebSocket connection count

### Additional Requirements

- **Starter Template**: Custom setup (not standard starter) — Go module init with `go mod init github.com/nlg/nfv2`, npm install from `xyflow/vite-react-flow-template`
- **Go Version**: Go 1.24+ (62% GC pause reduction, improved reflect.Blueprint support)
- **Gin Framework**: Gin 1.10+ (radix tree router, 38% lower allocation overhead, HTTP/3 support) for REST API + WebSocket hub
- **Database**: SQLite (mattn/go-sqlite3) for sessions/skills + BadgerDB (dgraph-io/badger/v4) for knowledge graph
- **CLI Framework**: Cobra with subcommands: `serve`, `run <spec>`, `new <name>`, `config set/get`, `skill list/install`, `session resume/export`
- **Frontend Stack**: Vite + React + @xyflow/react (TypeScript) + Custom Canvas Extensions
- **Design System**: Tailwind CSS + Radix UI Primitives for accessibility-first components
- **LLM Provider Abstraction**: Race mode goroutines, fastest token wins, fallback chains
- **AI Swarm per Node**: Multiple LLM agents negotiate within single node (goroutines, channel for result collection)
- **Smart Context Engine**: Knowledge graph in BadgerDB, node memory reuse, auto-spec generation
- **Merkle Tree Hashing**: Incremental execution, skips unchanged nodes
- **gRPC Plugin Interface**: Proto definitions (PluginService.proto), Unix socket IPC, plugin sandboxing
- **MCP Server**: Tools for Claude Desktop/Cursor — `create_node`, `run_node`, `get_status`, `fork_session`, `export_session`
- **Security**: Chroot jail per session, eBPF syscall filtering, Argon2 + AES-256-GCM encryption, Ed25519 signing
- **Docker**: Multi-stage build, distroless runtime (gcr.io/distroless/static-debian12), Ollama sidecar option
- **Build Tooling**: Makefile for orchestration, protoc for gRPC code generation, Go embed.FS for serving React build
- **Testing**: Ginkgo + Testify for Go backend

### UX Design Requirements

UX-DR1: Implement custom NodeTypes (Goal, Spec, Plan, Implement, Test, Review) with n8n/TouchDesigner/DaVinci hybrid visuals — rounded rectangles, diamonds, Input/Output pins, color-coded by type
UX-DR2: Implement custom EdgeTypes with reactive tension (stroke-width based on upstream health), animated pulses during execution, pluckable metadata (TouchDesigner-style floating bubble with latency, data flow, status)
UX-DR3: Implement MonologuePanel with collapsible slide-over from right (400px), streaming LLM Chain-of-Thought tokens, auto-scroll, export history
UX-DR4: Implement ChatPanel (320px narrow) for goal input, generates graph from chat — collapsible, "Generating graph..." state with animated ellipsis
UX-DR5: Implement CanvasControls with mini-map heat (nodes glow based on recent activity), zoom/pan, Vim/Emacs keybindings display, keybinding hints
UX-DR6: Implement SessionExplorer panel for session management — resume, fork, export, search/filter by status/date
UX-DR7: Implement NodeConfig dialog (Radix Dialog) for node parameters — timeout, retry count, token budget with real-time validation
UX-DR8: Implement SkillMarketplace panel for browsing/installing skills — grid layout, rating stars, category filter, install button
UX-DR9: Implement AccessibilityToolbar for high-contrast toggle, RTL switch, font-size slider
UX-DR10: Implement design tokens (Tailwind config) — dark theme (#1a1b1e background), node colors by type, edge states, phase colors (discovery=blue, execution=orange, recovery=red, completion=green)
UX-DR11: Implement Typography System with JetBrains Mono throughout, compact type scale (h1=1.5rem, body=0.875rem), line heights for developer tools
UX-DR12: Implement WCAG 2.1 AA compliance — 4.5:1 minimum contrast ratio, ARIA live regions for node status changes, screen reader support
UX-DR13: Implement colorblind-friendly design — nodes distinguished by shape + label + position, not just color (red/green alone insufficient for colorblind users)
UX-DR14: Implement high-contrast theme toggle — black background (#000000), white text, bright node colors (Goal=#00ff00, Spec=#00aaff, etc.)
UX-DR15: Implement RTL canvas support — canvas coordinates invert horizontally, text alignment adapts, mini-map mirrors position (bottom-left instead of bottom-right)
UX-DR16: Implement screen reader announcements via ARIA live regions — "Node Goal-1 changed to running" (polite), "Node failed" (assertive)
UX-DR17: Implement Vim/Emacs keyboard navigation (h=left, j=down, k=up, l=right, Ctrl-f/b/n/p) for canvas, all interactions possible without mouse
UX-DR18: Implement one-key node controls — p=pause/resume, r=retry, f=fork, s=skip, m=toggle monologue, space=pause
UX-DR19: Implement button hierarchy — Primary=Cyan (#06b6d4), Secondary=Gray outline, Danger=Red (#ef4444), Icon-only=32x32px with Radix Tooltip
UX-DR20: Implement feedback patterns — Success (green toast 3s auto-dismiss), Error (red persistent toast), Warning (yellow pause label), Info (cyan edge pulse)
UX-DR21: Implement loading states — Node yellow border pulse (300ms), Edge animated dash flow (cyan), Panel skeleton lines (60% opacity pulse)
UX-DR22: Implement modal/overlay patterns — Radix AlertDialog for confirmations, Radix Dialog for config, custom slide-over for monologue panel
UX-DR23: Implement empty states — No Sessions (📭 + "Start Chat" button), No Skills (🔌 + "Browse Marketplace"), Waiting Monologue (🕭 + animated ellipsis)
UX-DR24: Implement search/filter patterns — Session search with status/date filters, Skill search with category sort (Name/Rating/Installs/Recent)
UX-DR25: Implement responsive strategy — desktop-first (1366px+ minimum), no mobile/tablet support, multi-column layouts, `min-width: 1366px` enforced
UX-DR26: Implement Web Worker offloading for 100+ node graphs at 60fps — layout.worker.ts, zero main-thread blocking
UX-DR27: Implement interactive wires as health indicators — TouchDesigner-style pluck for metadata bubble, edge tension visualization, heartbeat monitor metaphor
UX-DR28: Implement LLM Inner Monologue as "flight recorder" pattern — side panel streams thinking in real-time, saves history for debugging
UX-DR29: Implement "Chat-First, Canvas-Second" novel UX pattern — chat generates graph, canvas becomes monitor/controller (not builder), familiar metaphor: "Chat is the spec, Canvas is the execution dashboard"
UX-DR30: Implement forward-only progress with color-coded phase bands — nodes only advance when verified, blue/orange/red/green bands across canvas top, Git-branch mental model

### FR Coverage Map

FR1: Epic 2 - Create session with goal description and auto-generated node graph
FR2: Epic 3 - See entire project state at a glance with color-coded nodes
FR3: Epic 2 - Watch nodes execute deterministically until acceptance criteria met
FR4: Epic 3 - Interact with node connections n8n-style, pluck edges for metadata
FR5: Epic 3 - View DaVinci Resolve-style node trees with input/output pins
FR6: Epic 3 - See animated pulses along edges showing real-time execution progress
FR7: Epic 3 - Deactivate/activate individual nodes without deleting (n8n feature)
FR8: Epic 3 - View mini-map with execution heat
FR9: Epic 3 - Add comments/notes to nodes (sticky notes)
FR10: Epic 2 - Configure multiple LLM providers via nforge config
FR11: Epic 2 - Race multiple LLM providers simultaneously, fastest token wins
FR12: Epic 2 - Auto-fallback through providers when rate limits or errors occur
FR13: Epic 3 - See LLM Inner Monologue in side panel with streaming tokens
FR14: Epic 2 - Auto-optimize prompts based on execution feedback
FR15: Epic 2 - Enforce token budgets with pre-flight estimation
FR16: Epic 2 - Use local Ollama racing against remote APIs
FR17: Epic 2 - Build knowledge graph for token-efficient context assembly
FR18: Epic 2 - Reuse node memory, each node's output becomes context for downstream
FR19: Epic 2 - Auto-generate specs and add system references as nodes execute
FR20: Epic 2 - Handle context overflow by auto-splitting graphs into sub-graphs
FR21: Epic 1 - Start web UI + API with nforge serve
FR22: Epic 2 - Run headless execution with nforge run <spec-file>
FR23: Epic 1 - Create new projects with nforge new <project-name>
FR24: Epic 1 - Configure settings with nforge config set/get
FR25: Epic 5 - Manage skills with nforge skill list/install
FR26: Epic 4 - Resume/export sessions with nforge session resume/export
FR27: Epic 3 - See ASCII art graph in terminal with nforge graph viz
FR28: Epic 1 - Check system health with nforge doctor
FR29: Epic 1 - CLI tab completion for all commands and node types
FR30: Epic 1 - CLI and UI feature parity, same graphs execute identically
FR31: Epic 4 - Create isolated sessions with separate workspaces
FR32: Epic 4 - Auto-save session state (graph JSON, chat log, workspace files)
FR33: Epic 4 - Resume sessions after restart with graceful shutdown snapshot
FR34: Epic 4 - Fork sessions like Git branches, try different approaches
FR35: Epic 4 - Auto-commit workspace changes to Git after each node completion
FR36: Epic 4 - Time-travel debug, checkout workspace state at any completed node
FR37: Epic 4 - Export session as self-contained tarball
FR38: Epic 4 - Enforce session quotas (max sessions, max workspace size)
FR39: Epic 4 - Auto-clean zombie sessions (heartbeat timeout detection)
FR40: Epic 5 - Browse and install skills from marketplace
FR41: Epic 5 - Skills have dependencies, auto-install dependency tree
FR42: Epic 5 - Skills sandboxed before trust (time limits, no network, read-only)
FR43: Epic 5 - Support gRPC plugins for third-party node types
FR44: Epic 5 - Expose MCP server for Claude Desktop/Cursor orchestration
FR45: Epic 5 - Skills have sub-nodes (e.g., JS-to-Go patterns, goroutines)
FR46: Epic 5 - A/B test skills, route to different versions, collect metrics
FR47: Epic 3 - Drag-and-drop files onto canvas to auto-create nodes
FR48: Epic 3 - Color-coded node bands by lifecycle phase
FR49: Epic 3 - Reactive edge tension, edges tighten when upstream nodes fail
FR50: Epic 3 - Vim/Emacs keybindings for canvas navigation
FR51: Epic 3 - Accessibility support (screen readers, high-contrast, RTL, 20+ languages)
FR52: Epic 2 - Execute nodes sequentially with retry loops
FR53: Epic 2 - Run multiple attempts in parallel within a node (speculative execution)
FR54: Epic 2 - Support incremental execution with Merkle tree hashing
FR55: Epic 3 - Offload graph layout to Web Worker for 100+ node graphs
FR56: Epic 2 - Pre-fetch LLM provider status on session start
FR57: Epic 2 - Use Gin backend for REST API + WebSocket hub
FR58: Epic 6 - Encrypt API keys at rest using Argon2 + AES-256-GCM
FR59: Epic 6 - Sandbox node execution in chroot jail
FR60: Epic 6 - Apply eBPF syscall filtering to block dangerous syscalls
FR61: Epic 6 - Enforce rate limiting per API key using token bucket
FR62: Epic 6 - Sign graph snapshots with Ed25519
FR63: Epic 6 - Build as multi-arch Docker (amd64 + arm64)
FR64: Epic 6 - Use distroless Docker image (only Go binary)
FR65: Epic 6 - Include Ollama sidecar option in docker compose
FR66: Epic 6 - Expose health check with graph diagnostics
FR67: Epic 6 - Send webhook notifications to Slack, GitHub, etc.
FR68: Epic 6 - Export telemetry as Prometheus metrics

## Epic List

### Epic 1: Project Setup & CLI Foundation
User can install, configure, and start NodeForge OS with a working CLI and server.
**User Value:** "I can get NodeForge running and ready to accept my goals in under 5 minutes."
**FRs covered:** FR21, FR23, FR24, FR28, FR29, FR30
**Dependencies:** None (foundational epic)
**Implementation Notes:** Go module init, Gin server scaffold, Cobra CLI with `serve`, `new`, `config`, `doctor` commands, SQLite setup, embed.FS foundation.

### Epic 2: Graph Execution Engine & LLM Integration
User can create a session with a goal description and watch nodes execute deterministically with multi-provider LLM support and smart context.
**User Value:** "I can describe my goal and watch the system execute it autonomously with the best available LLM."
**FRs covered:** FR1, FR2, FR3, FR10, FR11, FR12, FR13, FR14, FR15, FR16, FR17, FR18, FR19, FR20, FR22, FR52, FR53, FR54, FR55, FR56, FR57
**Dependencies:** Epic 1 (needs running server)
**Implementation Notes:** Graph engine (`internal/engine/`), LLM abstraction (`internal/llm/`), Smart Context (`internal/context/`), headless CLI (`nforge run`), WebSocket hub.

### Epic 3: Visual Canvas & User Experience
User can watch and interact with graph execution on a professional n8n/TouchDesigner/DaVinci-style canvas with full accessibility.
**User Value:** "I can see my entire project state at a glance, pluck edges for info, and control execution visually."
**FRs covered:** FR4, FR5, FR6, FR7, FR8, FR9, FR27, FR47, FR48, FR49, FR50, FR51
**UX-DRs covered:** UX-DR1 through UX-DR30 (all 30 UX design requirements)
**Dependencies:** Epic 2 (needs graph engine)
**Implementation Notes:** React + Vite + React Flow, custom NodeTypes/EdgeTypes, MonologuePanel, ChatPanel, CanvasControls, SessionExplorer, Tailwind + Radix UI, WCAG 2.1 AA compliance.

### Epic 4: Session Management & Recovery
User can manage sessions, fork for experimentation, resume after restart, export results, and recover from failures.
**User Value:** "I'm in control — I can fork, pause, retry, and never lose my work."
**FRs covered:** FR26, FR31, FR32, FR33, FR34, FR35, FR36, FR37, FR38, FR39
**Dependencies:** Epic 2 (needs sessions)
**Implementation Notes:** Session manager (`internal/session/`), workspace isolation (chroot), Git auto-commit, fork/merge, time-travel debug, export tarball, zombie cleanup.

### Epic 5: Skill System & Extensibility
User can browse, install, and create custom skills with gRPC plugins and MCP server for AI orchestration.
**User Value:** "I can extend NodeForge with community skills and let Claude Desktop orchestrate my workflows."
**FRs covered:** FR25, FR40, FR41, FR42, FR43, FR44, FR45, FR46
**Dependencies:** Epic 2 (needs engine)
**Implementation Notes:** Skill system (`internal/skills/`), marketplace, gRPC plugins (proto/), MCP server, sub-nodes, A/B testing, dependency resolution.

### Epic 6: Security, Performance & DevOps
NodeForge is secure, high-performance, and production-ready with Docker deployment and observability.
**User Value:** "It's secure, fast, and ready for CI/CD pipelines and production use."
**FRs covered:** FR58, FR59, FR60, FR61, FR62, FR63, FR64, FR65, FR66, FR67, FR68
**NFRs covered:** NFR-01 through NFR-30 (all 30 non-functional requirements)
**Dependencies:** Epics 2, 3, 4, 5 (needs working system to secure and optimize)
**Implementation Notes:** Security (`internal/security/`), DevOps (`internal/devops/`), chroot+jail+ebpf, Argon2 encryption, multi-arch Docker, distroless, Ollama sidecar, Prometheus metrics, webhooks.

## Epic 1: Project Setup & CLI Foundation

User can install, configure, and start NodeForge OS with a working CLI and server.

### Story 1.1: Project Scaffolding & Module Init

As a developer,
I want to initialize the Go module and set up the custom project structure,
So that development can begin with the correct foundation.

**Acceptance Criteria:**

**Given** an empty project directory
**When** the developer runs `go mod init github.com/nlg/nfv2` and installs Gin, Cobra, and protobuf dependencies
**Then** the `go.mod` file is created with Go 1.24+ and all required dependencies (gin-gonic/gin, spf13/cobra, google.golang.org/protobuf)
**And** the directory structure is created: `cmd/nforge/`, `internal/` (with engine, llm, context, session, skills, canvas, security, devops subdirectories), `frontend/`, `proto/`, `main.go`, `Makefile`, `Dockerfile`, `docker-compose.yml`

### Story 1.2: CLI Root Command with Cobra Framework

As a user,
I want a working `nforge` CLI with root command and persistent flags,
So that I can access all subcommands consistently.

**Acceptance Criteria:**

**Given** the CLI scaffold is set up with Cobra
**When** the user runs `nforge --help`
**Then** the root command displays usage information with available subcommands: `serve`, `run`, `new`, `config`, `skill`, `session`, `doctor`, `graph`
**And** persistent flags (e.g., `--verbose`, `--config-path`) are available across all subcommands
**And** the CLI displays version information with `nforge --version`

### Story 1.3: Gin Server with `nforge serve`

As a user,
I want to start the web UI + API with `nforge serve`,
So that I can access the React Flow canvas in my browser.

**Acceptance Criteria:**

**Given** the Gin framework is installed and `main.go` is set up
**When** the user runs `nforge serve`
**Then** the Gin server starts on the configured port (default :8080) with REST API (`/api/v1/*`) and WebSocket hub (`/ws`)
**And** the React build (from `frontend/dist/`) is served via `embed.FS` at the root path
**And** health check is available at `/healthz` and metrics at `/metrics`

### Story 1.4: Project Creation via CLI & UI

As a user,
I want to create new projects with `nforge new <project-name>` via CLI and from the UI,
So that I can quickly scaffold a workspace for my goals from either interface.

**Acceptance Criteria:**

**Given** the CLI and UI are functional
**When** the user runs `nforge new <project-name>` via CLI
**Then** a new session workspace is created with the specified project name, initialized with a `.nforge/` directory structure
**And** from the UI, the user can click "New Project" button (or type a project name in the chat panel) to create a new project workspace
**And** FR30 (CLI/UI parity) is satisfied: both interfaces create identical project structures

### Story 1.5: Configuration Management with `nforge config`

As a user,
I want to configure settings with `nforge config set/get`,
So that I can manage API keys, models, and ports.

**Acceptance Criteria:**

**Given** the configuration system is set up with Cobra `config` subcommand
**When** the user runs `nforge config set <key> <value>`
**Then** the configuration is saved to the config file (e.g., `~/.nforge/config.yaml`) with the specified key-value pair
**And** `nforge config get <key>` retrieves and displays the value
**And** supported keys include: `llm.openai-key`, `llm.anthropic-key`, `llm.ollama-url`, `server.port`, `llm.default-model`

### Story 1.6: System Health Check with `nforge doctor`

As a user,
I want to check system health with `nforge doctor`,
So that I can verify all dependencies and connectivity are working.

**Acceptance Criteria:**

**Given** the CLI includes the `doctor` subcommand
**When** the user runs `nforge doctor`
**Then** the system checks: Go version (1.24+), Gin framework availability, frontend build status, SQLite/BadgerDB connectivity, LLM provider connectivity (Ollama, OpenAI, Anthropic)
**And** results are displayed with clear pass/fail indicators for each check
**And** exit code is 0 if all checks pass, non-zero if any fail

### Story 1.7: CLI Tab Completion

As a user,
I want tab completion for all commands and node types,
So that I can navigate the CLI faster without remembering exact syntax.

**Acceptance Criteria:**

**Given** the CLI uses Cobra framework
**When** the user presses Tab after typing `nforge `
**Then** available subcommands are displayed: `serve`, `run`, `new`, `config`, `skill`, `session`, `doctor`, `graph`
**And** Tab completion works for subcommand flags and node types (Goal, Spec, Plan, Implement, Test, Review)
**And** completion works in bash, zsh, and PowerShell shells

### Story 1.8: Frontend Scaffolding with Vite + React Flow

As a developer,
I want the React + Vite + React Flow frontend scaffolded and served via `embed.FS`,
So that the UI is ready for canvas development.

**Acceptance Criteria:**

**Given** the `frontend/` directory is set up
**When** the developer runs `npx degit xyflow/vite-react-flow-template frontend` and `npm install`
**Then** the React Flow starter template is cloned with TypeScript support and Vite build system
**And** the built output (`frontend/dist/`) is embeddable via Go's `embed.FS` in `main.go`
**And** the frontend serves a basic React Flow canvas at `http://localhost:8080` when `nforge serve` is running

## Epic 2: Graph Execution Engine & LLM Integration

User can create a session with a goal description and watch nodes execute deterministically with multi-provider LLM support and smart context.

### Story 2.1: Chat Interface & Auto-Generated Node Graph

As a user,
I want to describe my goal in a chat interface and have AI automatically create and execute a node graph, with the ability to interact (pause, skip, fork, retry),
So that I can start execution within 5 minutes without manual graph construction.

**Acceptance Criteria:**

**Given** the Gin server is running and WebSocket hub is active
**When** the user types a goal (e.g., "Convert JS→Go project") in the chat panel
**Then** the AI analyzes the input and auto-generates a complete node graph (Goal → Spec → Plan → Implement → Test → Review)
**And** the graph is displayed on the React Flow canvas with color-coded nodes (green=complete, red=failed, yellow=running) (FR2)
**And** the user can interact with execution: pause (spacebar/p), skip node (s), fork session (f), retry failed node (r)
**And** nodes execute deterministically, each working until acceptance criteria are met before advancing (FR3)
**And** the graph state is the source of truth — it only moves forward when verified (FR1, FR52)

### Story 2.2: LLM Provider Abstraction & Race Mode

As a user,
I want to configure multiple LLM providers with race mode and auto-fallback,
So that I get the fastest/cheapest response with reliable connectivity.

**Acceptance Criteria:**

**Given** the LLM provider interface is defined (`type LLMProvider interface`)
**When** the user configures providers via `nforge config set llm.openai-key <key>` (FR10)
**Then** the system supports: OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama (local) (FR16)
**And** race mode runs multiple providers simultaneously — fastest token wins, slower providers are cancelled (FR11, NFR-03)
**And** auto-fallback triggers when rate limits or errors occur (e.g., rate limit → cheaper model) (FR12, NFR-19)
**And** provider status is pre-fetched on session start for zero-wait connectivity checks (FR56)

### Story 2.3: LLM Inner Monologue Panel

As a user,
I want to see LLM "Inner Monologue" (Chain-of-Thought) streaming in a side panel,
So that I can understand why the AI made certain decisions.

**Acceptance Criteria:**

**Given** the MonologuePanel component is implemented in React
**When** a node is executing and the LLM is processing
**Then** the Chain-of-Thought tokens stream in real-time to the MonologuePanel via WebSocket (FR13)
**And** the panel is collapsible (400px wide slide-over from right), toggleable via 'm' key (UX-DR3)
**And** monologue history is saved and exportable for debugging
**And** auto-scroll keeps the latest token visible

### Story 2.4: Prompt Optimization & Token Budget

As a system,
I want to auto-optimize prompts based on execution feedback and enforce token budgets,
So that costs are controlled and prompts improve over time.

**Acceptance Criteria:**

**Given** the prompt optimization system and token budget enforcer are implemented
**When** a node prepares to call an LLM provider
**Then** the system optimizes prompts automatically based on past execution feedback (FR14)
**And** token budget pre-flight estimation completes in <10ms before the LLM call is dispatched (FR15, NFR-05)
**And** expensive requests are rejected if they exceed the configured token budget
**And** budget usage is tracked and reported back to the user

### Story 2.5: Smart Context Engine (Knowledge Graph)

As a system,
I want to build a knowledge graph for token-efficient context assembly and reuse node memory,
So that token usage is reduced by 30%+ and context overflow is handled gracefully.

**Acceptance Criteria:**

**Given** BadgerDB is set up for the knowledge graph (internal/context/)
**When** nodes execute and produce outputs
**Then** a knowledge graph is built for token-efficient context assembly, achieving 30%+ reduction vs naive prompts (FR17, NFR-04)
**And** each node's output becomes context for downstream nodes (node memory reuse) (FR18)
**And** the system auto-generates specs and adds system references as nodes execute (FR19)
**And** context overflow is handled by auto-splitting graphs into sub-graphs (FR20, NFR-04 <100ms assembly)

### Story 2.6: Speculative Execution within Nodes

As a system,
I want to run multiple attempts in parallel within a node and select the best result,
So that execution quality improves without user intervention.

**Acceptance Criteria:**

**Given** the node executor supports parallel execution paths
**When** a node starts executing with speculative execution enabled (FR53)
**Then** multiple LLM agents negotiate within the node (AI Swarm per Node)
**And** parallel attempts run via goroutines with results collected via channels
**And** the best result (based on acceptance criteria verification) wins and proceeds
**And** failed attempts are logged for debugging but don't block progress

### Story 2.7: Incremental Execution & Web Worker Offloading

As a system,
I want to support incremental execution with Merkle tree hashing and offload graph layout to Web Worker,
So that 100+ node graphs render smoothly at 60fps with zero main-thread blocking.

**Acceptance Criteria:**

**Given** the Merkle tree hashing and Web Worker offloading are implemented
**When** a graph re-execution is triggered with 95% unchanged nodes
**Then** Merkle tree hashing skips unchanged nodes, completing in <2s for 100-node graph (FR54, NFR-06)
**And** graph layout is offloaded to Web Worker (layout.worker.ts) for smooth 60fps rendering (FR55, NFR-02)
**And** React Flow canvas handles 100+ node graphs without UI freeze (NFR-16)
**And** the system supports 5000+ WebSocket connections with <50ms state propagation (FR57, NFR-01, NFR-14)

### Story 2.8: Headless CLI Execution

As a user,
I want to run headless execution with `nforge run <spec-file>`,
So that I can execute graphs identically in terminal without a browser.

**Acceptance Criteria:**

**Given** the CLI includes the `run` subcommand
**When** the user runs `nforge run <spec-file>` in terminal
**Then** the same graph execution logic runs as the browser UI — identical nodes, identical results (FR22, FR30)
**And** ASCII art graph is displayed in terminal with `nforge graph viz` (FR27)
**And** headless mode exits with code 0 when all nodes are green, code 1 on red node (CI/CD integration)
**And** CLI and UI have feature parity — same graphs execute identically (FR30)

## Epic 3: Visual Canvas & User Experience

User can watch and interact with graph execution on a professional n8n/TouchDesigner/DaVinci-style canvas with full accessibility.

### Story 3.1: Custom NodeTypes & EdgeTypes

As a user,
I want custom NodeTypes with n8n/TouchDesigner/DaVinci hybrid visuals and EdgeTypes with reactive tension,
So that I can see a professional canvas with interactive wires, input/output pins, color-coded phase bands, and pluckable edge metadata.

**Acceptance Criteria:**

**Given** React Flow is set up with custom node/edge components
**When** the graph engine creates nodes (Goal, Spec, Plan, Implement, Test, Review)
**Then** each node type renders with correct visuals: Goal (green rounded rect), Spec (blue diamond), Plan (purple rect), Implement (orange rect), Test (yellow rounded rect), Review (cyan rect) (UX-DR1)
**And** edges show reactive tension (stroke-width based on upstream health), animated pulses during execution, and pluckable metadata bubbles (TouchDesigner-style) (UX-DR2, FR4, FR49)
**And** nodes have clear input/output pins (DaVinci-style) and color-coded phase bands across canvas top: blue (Discovery), orange (Execution), red (Recovery), green (Completion) (FR5, FR48)
**And** user can drag-and-drop files onto canvas to auto-create nodes (e.g., `go.mod` → `Setup` node) (FR47)

### Story 3.2: ChatPanel & MonologuePanel

As a user,
I want a ChatPanel (320px) for goal input that generates graphs and a MonologuePanel streaming LLM thoughts,
So that I can describe goals and watch AI thinking in real-time.

**Acceptance Criteria:**

**Given** the ChatPanel and MonologuePanel components are implemented
**When** the user types a goal in the ChatPanel and presses Enter
**Then** the graph is generated from chat input and displayed on canvas (FR1, UX-DR29: "Chat-First, Canvas-Second" pattern)
**And** MonologuePanel slides in from right (400px wide) with streaming LLM Chain-of-Thought tokens (UX-DR3, FR13)
**And** auto-scroll keeps latest token visible, monologue history is exportable, and panel is toggleable via 'm' key (UX-DR3, UX-DR28: "flight recorder" pattern)
**And** ChatPanel shows "Generating graph..." state with animated ellipsis and disabled input during generation (UX-DR4)

### Story 3.3: CanvasControls & SessionExplorer

As a user,
I want mini-map heat visualization, Vim/Emacs keyboard navigation, session management with search/filter, and node configuration dialogs,
So that I can navigate the canvas efficiently and manage sessions visually.

**Acceptance Criteria:**

**Given** CanvasControls and SessionExplorer components are implemented
**When** the user views the canvas
**Then** mini-map shows execution heat (nodes glow based on recent activity), zoom/pan controls, and Vim/Emacs keybinding hints (UX-DR5, FR8, FR50)
**And** Vim (h=left, j=down, k=up, l=right) and Emacs (Ctrl-f/b/n/p) keys work for canvas navigation (NFR-23, UX-DR17)
**And** SessionExplorer allows resume, fork, export with search/filter by status/date (UX-DR6, FR26)
**And** NodeConfig dialog (Radix Dialog) allows setting timeout, retry count, token budget with real-time validation (UX-DR7)

### Story 3.4: SkillMarketplace & AccessibilityToolbar

As a user,
I want to browse/install skills from a marketplace and toggle accessibility features (high-contrast, RTL, font-size),
So that I can extend NodeForge and adapt it to my needs.

**Acceptance Criteria:**

**Given** SkillMarketplace and AccessibilityToolbar components are implemented
**When** the user opens the SkillMarketplace panel
**Then** skills are displayed in grid layout with rating stars, category filter, and install button (UX-DR8, FR25, FR40)
**And** AccessibilityToolbar provides high-contrast toggle, RTL switch, and font-size slider (UX-DR9, FR51, NFR-21, NFR-22)
**And** skill dependencies auto-install when installing (FR41), and A/B testing routes to different versions with metrics collection (FR46)

### Story 3.5: Design System Foundation

As a user,
I want a complete design system with Tailwind dark theme, JetBrains Mono typography, WCAG 2.1 AA compliance, and colorblind-friendly design,
So that the UI is professional, accessible, and readable.

**Acceptance Criteria:**

**Given** Tailwind CSS + Radix UI Primitives are configured
**When** the app renders any UI component
**Then** design tokens are applied: dark theme (#1a1b1e background), node colors by type, edge states, phase colors (UX-DR10)
**And** typography uses JetBrains Mono throughout with compact type scale (h1=1.5rem, body=0.875rem) (UX-DR11)
**And** WCAG 2.1 AA compliance is met: 4.5:1 minimum contrast ratio, ARIA live regions for node status changes (UX-DR12, NFR-20)
**And** nodes are distinguished by shape + label + position, not just color (red/green alone insufficient for colorblind users) (UX-DR13)

### Story 3.6: Accessibility - High-Contrast, RTL & Screen Readers

As a user,
I want high-contrast theme, RTL canvas support, screen reader announcements via ARIA live regions, and full keyboard navigation with one-key controls,
So that NodeForge is accessible to all users.

**Acceptance Criteria:**

**Given** accessibility features are implemented
**When** the user toggles high-contrast mode
**Then** canvas shows black background (#000000), white text, bright node colors (Goal=#00ff00, Spec=#00aaff, etc.) (UX-DR14, NFR-21)
**And** RTL canvas support inverts coordinates horizontally, adapts text alignment, and mirrors mini-map to bottom-left (UX-DR15, NFR-22)
**And** screen reader announces: "Node Goal-1 changed to running" (polite), "Node failed" (assertive) via ARIA live regions (UX-DR16, NFR-20)
**And** Vim/Emacs navigation (hjkl, Ctrl-f/b/n/p) works for all interactions without mouse; one-key controls: p=pause/resume, r=retry, f=fork, s=skip (UX-DR17, UX-DR18, FR50, NFR-23, NFR-24)

### Story 3.7: UI Patterns - Buttons, Feedback, Loading, Modals

As a user,
I want consistent button hierarchy, clear feedback patterns, proper loading states, and modal/overlay patterns,
So that the UI communicates effectively and follows accessibility standards.

**Acceptance Criteria:**

**Given** UI pattern components are implemented with Radix UI Primitives
**When** any action button is displayed
**Then** button hierarchy is applied: Primary=Cyan (#06b6d4), Secondary=Gray outline, Danger=Red (#ef4444), Icon-only=32x32px with Radix Tooltip (UX-DR19)
**And** feedback patterns work: Success (green toast 3s auto-dismiss), Error (red persistent toast), Warning (yellow pause label), Info (cyan edge pulse) (UX-DR20)
**And** loading states show: Node yellow border pulse (300ms), Edge animated dash flow (cyan), Panel skeleton lines (60% opacity pulse) (UX-DR21)
**And** modal/overlay patterns use: Radix AlertDialog for confirmations, Radix Dialog for config, custom slide-over for monologue panel (UX-DR22)

### Story 3.8: Empty States, Search/Filter & Responsive Strategy

As a user,
I want helpful empty states, powerful search/filter capabilities, and a desktop-optimized responsive strategy,
So that I can find content quickly and work comfortably on my development machine.

**Acceptance Criteria:**

**Given** empty states, search/filter, and responsive layout are implemented
**When** there are no sessions or skills
**Then** empty states display: 📭 + "Start Chat" button (No Sessions), 🔌 + "Browse Marketplace" (No Skills), 💭 + animated ellipsis (Waiting Monologue) (UX-DR23)
**And** search/filter works: Session search with status/date filters, Skill search with category sort (Name/Rating/Installs/Recent) (UX-DR24)
**And** responsive strategy is desktop-first (1366px+ minimum), no mobile/tablet support, multi-column layouts, `min-width: 1366px` enforced (UX-DR25)

### Story 3.9: Interactive Wires & Novel UX Patterns

As a user,
I want Web Worker offloading for 60fps rendering, interactive wires as health indicators, LLM "flight recorder" pattern, and forward-only progress with color-coded phase bands,
So that the canvas is fast, informative, and revolutionary.

**Acceptance Criteria:**

**Given** Web Worker and interactive wire components are implemented
**When** 100+ node graphs render on canvas
**Then** layout is offloaded to Web Worker (layout.worker.ts) for smooth 60fps rendering with zero main-thread blocking (UX-DR26, FR55, NFR-02, NFR-16)
**And** interactive wires show: TouchDesigner-style pluck for metadata bubble, edge tension visualization, heartbeat monitor metaphor (UX-DR27, FR4, FR49)
**And** LLM Inner Monologue acts as "flight recorder": side panel streams thinking in real-time, saves history for debugging (UX-DR28)
**And** "Chat-First, Canvas-Second" novel UX pattern is implemented: chat generates graph, canvas becomes monitor/controller (not builder) (UX-DR29)
**And** forward-only progress shows color-coded phase bands: nodes only advance when verified, blue/orange/red/green bands across canvas top, Git-branch mental model (UX-DR30, FR48, FR3)

## Epic 4: Session Management & Recovery

User can manage sessions, fork for experimentation, resume after restart, export results, and recover from failures.

### Story 4.1: Session Creation & Isolation.

As a user,
I want to create isolated sessions with separate workspaces that auto-save state,
So that my work is protected and organized from the start.

**Acceptance Criteria:**

**Given** the session manager (`internal/session/`) is implemented with SQLite
**When** the user creates a new session via CLI (`nforge new`) or UI ("New Project" button)
**Then** an isolated session is created with separate workspace directory
**And** session state is auto-saved: graph JSON, chat log, workspace files (FR31, FR32)
**And** sessions are listed with unique IDs and creation timestamps

### Story 4.2: Session Resume & Graceful Shutdown.

As a user,
I want to resume sessions after restart with graceful shutdown snapshots and auto-cleaned zombie sessions,
So that I never lose my work and stale sessions are removed.

**Acceptance Criteria:**

**Given** the session resume and heartbeat systems are implemented
**When** the user runs `nforge session resume <id>` or clicks "Resume" in UI
**Then** the session is restored with snapshot from graceful shutdown (FR33)
**And** zombie sessions are auto-cleaned via heartbeat timeout detection (FR39)
**And** `nforge doctor` verifies session health and connectivity (FR28)

### Story 4.3: Fork, Git Auto-Commit & Time-Travel Debug.

As a user,
I want to fork sessions like Git branches, have workspace changes auto-committed to Git after each node, and time-travel debug by checking out workspace state at any completed node,
So that I can experiment safely and debug historically.

**Acceptance Criteria:**

**Given** the fork, Git auto-commit, and time-travel systems are implemented
**When** the user presses 'f' (fork) or clicks "Fork Session"
**Then** a new session branch is created with the current state, allowing different approaches to be tried (FR34)
**And** workspace changes are auto-committed to Git after each node completion with deterministic commit messages (FR35, NFR-28)
**And** the user can time-travel debug by checking out workspace state at any completed node (FR36)

### Story 4.4: Session Export & Quotas.

As a user,
I want to export sessions as self-contained tarballs and have session quotas enforced (max sessions, max workspace size),
So that I can share results and the system stays within resource limits.

**Acceptance Criteria:**

**Given** the export and quota enforcement systems are implemented
**When** the user runs `nforge session export <id>` or clicks "Export" in UI
**Then** a self-contained tarball is generated (graph + source + README) (FR37)
**And** session quotas are enforced: max sessions limit, max workspace size (500MB per session) (FR38, NFR-17)
**And** export includes only necessary files — API keys and secrets are excluded (NFR-10)

### Story 4.5: Session Resume/Export CLI.

As a user,
I want to resume/export sessions with `nforge session resume/export <id>` via CLI,
So that I can manage sessions headlessly in terminal.

**Acceptance Criteria:**

**Given** the CLI includes the `session` subcommand
**When** the user runs `nforge session resume <id>`
**Then** the session is restored and the user can continue working via CLI or UI (FR26, FR30 CLI/UI parity)
**And** `nforge session export <id>` generates the tarball in the current directory
**And** `nforge session list` shows all sessions with status, creation date, and workspace size

## Epic 5: Skill System & Extensibility.

User can browse, install, and create custom skills with gRPC plugins and MCP server for AI orchestration.

### Story 5.1: Skill Marketplace & Dynamic Fetch.

As a user,
I want to browse, search, and install skills dynamically from a third-party marketplace (e.g., skillsmp.com) via CLI (`nforge skill list/install`) and UI SkillMarketplace,
So that I can discover and install community skills on-demand without manual downloads.

**Acceptance Criteria:**

**Given** the SkillMarketplace backend integration with a third-party registry (e.g., https://skillsmp.com/)
**When** the user runs `nforge skill list` or browses the SkillMarketplace in UI
**Then** skills are fetched dynamically from the remote registry with search/filter (by name, rating, installs, category) (FR25, FR40, UX-DR8)
**And** `nforge skill install <name>` fetches and installs the skill manifest + dependencies from the registry (FR41)
**And** the UI SkillMarketplace displays grid layout with rating stars, category filter, install button, and dynamically loads skill data via backend API (`/api/v1/skills`) (UX-DR8)
**And** CLI and UI have feature parity — same skills available via both interfaces (FR30)

### Story 5.2: Skill Dependencies & Sandboxing.

As a system,
I want skills to have dependencies that auto-install and run sandboxed (time limits, no network, read-only) before trust,
So that skills are safe and self-contained.

**Acceptance Criteria:**

**Given** the skill dependency resolver and sandbox are implemented (`internal/skills/`)
**When** a skill is installed
**Then** its dependency tree is auto-installed recursively (FR41)
**And** skills run in pre-trust sandbox: time limits (30s), no network, read-only filesystem (FR42, NFR-13)
**And** after signature verification, skills are escalated to full trust with expanded permissions
**And** sandbox violations are logged with audit entries (NFR-12, Ed25519 signing)

### Story 5.3: gRPC Plugins & Sub-Nodes.

As a third-party developer,
I want to define entirely new node types via gRPC plugins with sub-nodes support,
So that I can extend NodeForge beyond built-in functionality.

**Acceptance Criteria:**

**Given** the gRPC plugin interface is defined (`proto/plugin.proto`) with Unix socket IPC
**When** a third-party plugin is loaded via `internal/skills/grpc.go`
**Then** it can define entirely new node types (FR43, NFR-26)
**And** skills can have sub-nodes (e.g., "JS-to-Go" skill has "patterns," "goroutines" sub-nodes) (FR45)
**And** plugins run in separate processes with resource limits; crash of one plugin doesn't affect core engine (NFR-26)
**And** plugin communication uses protobuf-defined messages for type safety

### Story 5.4: MCP Server for AI Orchestration.

As a user,
I want NodeForge to expose an MCP server so that Claude Desktop/Cursor can orchestrate my workflows via MCP tools,
So that I can integrate NodeForge into my existing AI-assisted development setup.

**Acceptance Criteria:**

**Given** the MCP server is implemented (`internal/skills/mcp.go`)
**When** Claude Desktop or Cursor connects via MCP
**Then** full session lifecycle tools are exposed: `create_node`, `run_node`, `get_status`, `fork_session`, `export_session` (FR44, NFR-27)
**And** MCP server uses JSON-RPC 2.0 format per MCP spec
**And** tools are documented with input/output schemas for AI consumption
**And** session state is accessible and modifiable via MCP tool calls

### Story 5.5: A/B Testing Skills.

As a user,
I want to A/B test skills where the system routes to different versions and collects metrics,
So that I can objectively compare skill performance.

**Acceptance Criteria:**

**Given** the A/B testing framework is implemented (`internal/skills/abtest.go`)
**When** A/B testing is enabled for a skill
**Then** the system routes requests to different skill versions and collects metrics (execution time, success rate, token usage) (FR46)
**And** metrics are reported back to the marketplace and displayed in SkillMarketplace
**And** the user can view A/B test results and choose the preferred version
**And** Prometheus metrics (`/metrics`) include A/B test results (NFR-30)

## Epic 6: Security, Performance & DevOps.

NodeForge is secure, high-performance, and production-ready with Docker deployment and observability.

### Story 6.1: API Key Encryption & Secret Management.

As a system,
I want to encrypt API keys at rest using Argon2 + AES-256-GCM and integrate Vault for secrets,
So that keys are never in plaintext configs or session exports.

**Acceptance Criteria:**

**Given** the encryption and Vault systems are implemented (`internal/security/`)
**When** an API key is saved via `nforge config set llm.openai-key <key>`
**Then** the key is encrypted at rest using Argon2 key derivation + AES-256-GCM (FR58, NFR-07)
**And** secrets are managed via Vault integration for advanced users (NFR-10)
**And** API keys are never logged and never included in session export tarballs (NFR-10)
**And** encrypted keys are stored in config file, not plaintext (NFR-07)

### Story 6.2: Workspace Isolation & Syscall Filtering.

As a system,
I want to sandbox node execution in chroot jail with eBPF syscall filtering and Ed25519 graph signing,
So that nodes can't escape or tamper with data.

**Acceptance Criteria:**

**Given** the chroot, eBPF, and signing systems are implemented
**When** a node starts executing in its session workspace
**Then** it runs inside a chroot jail — cannot escape to parent directories (FR59, NFR-08)
**And** eBPF filters block dangerous syscalls (exec, mount, reboot) during execution (FR60, NFR-09)
**And** graph snapshots are signed with Ed25519 — tampered session imports are rejected with audit log entry (FR62, NFR-12)
**And** these protections are verified by integration tests (NFR-08)

### Story 6.3: Rate Limiting & Session Quotas.

As a system,
I want to enforce rate limiting per API key (token bucket) and session quotas (max sessions, max workspace size),
So that resources are protected from abuse.

**Acceptance Criteria:**

**Given** the rate limiting and quota systems are implemented
**When** API requests exceed 100 req/min for REST or 10 msg/sec for WebSocket per API key
**Then** requests are rejected with appropriate error responses (FR61, NFR-11)
**And** session quotas are enforced: max sessions per instance, max 500MB workspace size (FR38, NFR-17)
**And** rate limiting uses token bucket algorithm with per-API-key limits (NFR-11)
**And** zombie sessions are auto-cleaned via heartbeat timeout detection (FR39)

### Story 6.4: Multi-Arch Docker & Ollama Sidecar.

As a user,
I want NodeForge built as multi-arch Docker (amd64+arm64) with distroless image and Ollama sidecar option,
So that it deploys anywhere securely with local LLM support.

**Acceptance Criteria:**

**Given** the Dockerfile and docker-compose.yml are set up
**When** building the Docker image
**Then** it supports multi-arch manifest for amd64 + arm64 (FR63, NFR-18)
**And** runtime uses distroless image (gcr.io/distroless/static-debian12) — only Go binary, no shell, no OS utilities (FR64)
**And** `docker compose up` auto-provisions Ollama sidecar for local LLM (FR65, NFR-18)
**And** multi-stage build: golang:1.24 builder → distroless runtime (NFR-18)

### Story 6.5: Health Checks, Webhooks & Prometheus Metrics.

As a user,
I want health diagnostics (/healthz), webhook notifications to Slack/GitHub, and Prometheus metrics (/metrics),
So that I can monitor and integrate NodeForge into CI/CD.

**Acceptance Criteria:**

**Given** the health, webhook, and metrics systems are implemented (`internal/devops/`)
**When** a user or monitoring system accesses `/healthz`
**Then** it returns session stats, LLM connectivity, and workspace quota info (FR66, NFR-28)
**And** outbound webhooks fire to Slack, GitHub, etc. when nodes reach configurable states (success/failure/approval needed) with exponential backoff retries (FR67, NFR-29)
**And** `/metrics` exports Prometheus metrics: session count, LLM token usage, node execution duration, WebSocket connection count (FR68, NFR-30)
**And** health checks and metrics are accessible in both CLI (`nforge doctor`) and Docker deployments (FR28)

### Story 6.6: Performance Optimization - WebSocket & Rendering.

As a system,
I want WebSocket state propagation <50ms, 100+ node graphs at 60fps, LLM race mode <200ms, smart context <100ms, token budget <10ms, Merkle tree <2s, 5000+ WS connections,
So that performance exceeds all targets.

**Acceptance Criteria:**

**Given** performance optimizations are implemented across the system
**When** a node state changes
**Then** WebSocket state propagates to browser UI in <50ms via Gin WS hub with 5000+ concurrent connections and <5% memory growth (FR57, NFR-01, NFR-14)
**And** 100+ node graphs render at 60fps with Web Worker offloading, zero main-thread blocking (FR55, NFR-02, NFR-16)
**And** LLM race mode fastest token wins in sub-200ms, slower providers cancelled immediately (FR11, NFR-03)
**And** smart context assembly completes in <100ms achieving 30%+ token reduction (FR17, NFR-04)
**And** token budget pre-flight estimation completes in <10ms before LLM call (FR15, NFR-05)
**And** Merkle tree hashing skips unchanged nodes — 100-node graph re-execution completes in <2s when 95% unchanged (FR54, NFR-06)

### Story 6.7: Provider Failover, Accessibility & i18n.

As a system,
I want 99.9% LLM uptime via fallback chains, horizontal scaling, WCAG 2.1 AA compliance, screen readers, keyboard navigation, RTL canvas, 20+ languages, gRPC plugins, and MCP server compliance,
So that NodeForge is reliable, accessible, and extensible.

**Acceptance Criteria:**

**Given** failover, accessibility, and extensibility systems are implemented
**When** an LLM provider fails or hits rate limit
**Then** the system fails over with 99.9% uptime: Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter (NFR-19)
**And** architecture supports growth from 1,000 to 10,000+ solo developers without structural changes (horizontal scaling ready) (NFR-15)
**And** WCAG 2.1 AA compliance is met: 4.5:1 contrast, ARIA live regions, screen reader support (NFR-20, UX-DR12)
**And** nodes are distinguished by shape + label + position, not just color — high-contrast theme available (NFR-21, UX-DR13, UX-DR14)
**And** RTL canvas support for 20+ languages (EN, ES, FR, DE, ZH, JA, KO, PT, RU, AR) with inverted coordinates (NFR-22, UX-DR15)
**And** Vim/Emacs keyboard navigation works for all interactions without mouse (NFR-23, UX-DR17)
**And** all node operations accessible via keyboard shortcuts and context menu (NFR-24)
**And** 5+ LLM providers supported via pluggable interface — new providers added via config, no code changes (NFR-25)
**And** gRPC plugins load dynamically, sandboxed in separate processes with resource limits (NFR-26)
**And** MCP server exposes full session lifecycle for Claude Desktop/Cursor orchestration (NFR-27, FR44)

## Epic 7: Documentation & Developer Experience

User can access comprehensive documentation, README, API docs, and developer guides.

**User Value:** "I can understand and contribute to NodeForge with clear, comprehensive documentation."

**FRs covered:** N/A (documentation epic)
**Dependencies:** Epics 1-6 (needs working system to document)
**Implementation Notes:** README.md, API documentation (Swagger/OpenAPI), architecture docs, developer guides, Storybook for UI components, CLI help system.

### Story 7.1: Comprehensive README & Project Documentation

As a developer or user,
I want a comprehensive README.md with project overview, quick start guide, feature list, and contribution guidelines,
So that I can quickly understand and start using NodeForge.

**Acceptance Criteria:**

**Given** the project has reached a stable state
**When** a user visits the repository or reads README.md
**Then** they see: project title, tagline, badges (build status, Go version, license), feature list with emojis
**And** quick start section: `go install`, `nforge serve`, `nforge new <project>`, open browser
**And** complete feature matrix table linking to epic/story documentation
**And** architecture overview diagram (text-based or link to detailed docs)
**And** contribution guidelines: how to run tests, code conventions, PR process

### Story 7.2: API Documentation (Swagger/OpenAPI)

As a developer,
I want auto-generated Swagger/OpenAPI documentation for the REST API,
So that I can understand and integrate with the NodeForge API.

**Acceptance Criteria:**

**Given** the REST API is implemented with Gin
**When** a developer visits `/swagger/index.html` or `/docs/api`
**Then** they see interactive Swagger UI with all endpoints documented: request/response schemas, auth requirements, example requests
**And** OpenAPI spec is auto-generated from code comments or annotations
**And** WebSocket API documented separately with message type definitions (graph_update, node_update, llm_chunk, monologue, connected)
**And** Postman collection export available at `/docs/postman.json`

### Story 7.3: Architecture & Developer Guides

As a contributor,
I want detailed architecture documentation and developer guides,
So that I can understand the codebase and contribute effectively.

**Acceptance Criteria:**

**Given** the codebase has stable architecture
**When** a developer reads `docs/architecture.md`
**Then** they see: system overview, backend (Go) architecture, frontend (React) architecture, data flow diagrams
**And** `docs/developer-guide.md` covers: local development setup, testing strategy, debugging tips, adding new LLM providers, creating custom skills
**And** `docs/context-engine.md` explains the Smart Context Engine (BadgerDB knowledge graph, token reduction, context assembly)
**And** `docs/session-management.md` covers session lifecycle, workspace isolation, forking, time-travel debug

### Story 7.4: UI Component Storybook & Design System Docs

As a frontend developer,
I want Storybook or equivalent for UI components with design system documentation,
So that I can build consistent UI and understand available components.

**Acceptance Criteria:**

**Given** the React frontend has stable component library
**When** a developer runs `cd frontend && npm run storybook`
**Then** they see all UI components documented: NodeTypes, EdgeTypes, Panels (ChatPanel, MonologuePanel, SessionExplorer), UI patterns (buttons, modals, toasts)
**And** design system docs: Tailwind config, color palette, typography (JetBrains Mono), spacing, accessibility notes
**And** each component shows: props API, usage examples, accessibility features (ARIA attributes, keyboard nav)
**And** visual regression tests for components (optional, if Storybook addon available)

### Story 7.5: CLI Documentation & Help System

As a user,
I want comprehensive CLI documentation with examples for every command and flag,
So that I can use the CLI effectively without guessing.

**Acceptance Criteria:**

**Given** the CLI uses Cobra framework
**When** the user runs `nforge --help` or `nforge <command> --help`
**Then** they see: command description, usage examples, available flags with defaults, related commands
**And** `docs/cli-reference.md` provides: full command tree, exit codes, environment variables (NFORGE_VERBOSE, NFORGE_CONFIG), configuration file format
**And** interactive examples: `nforge config set llm.openai-key <key>`, `nforge session resume <id>`, `nforge skill install <name>`
**And** man pages or equivalent offline docs generated for power users

### Story 7.6: Video Tutorials & Interactive Demos

As a new user,
I want video tutorials and interactive demos showing key workflows,
So that I can learn NodeForge visually and try it live.

**Acceptance Criteria:**

**Given** the system has stable core features
**When** a new user visits the documentation site or README
**Then** they see embedded video tutorials: "5-minute quickstart", "Creating your first goal", "Forking sessions", "Using the Skill Marketplace"
**And** interactive demo (if feasible): live sandbox at `demo.nodeforge.io` where users can try NodeForge in browser (read-only mode)
**And** GIF/screencast recordings in documentation: showing chat-to-graph flow, canvas interactions, monologue panel, session forking
**And** links to community resources: Discord/Slack, GitHub Discussions, Stack Overflow tag
