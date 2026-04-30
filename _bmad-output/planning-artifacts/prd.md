---
stepsCompleted: ['step-01-init', 'step-02-discovery', 'step-02b-vision', 'step-02c-executive-summary', 'step-03-success', 'step-04-journeys', 'step-05-domain', 'step-06-innovation', 'step-07-project-type', 'step-08-scoping', 'step-09-functional', 'step-10-nonfunctional', 'step-11-polish']
inputDocuments: ['_bmad-output/brainstorming/brainstorming-session-2026-04-28-050000.md']
workflowType: 'prd'
classification:
  projectType: 'developer_tool'
  domain: 'scientific'
  complexity: 'high'
  projectContext: 'greenfield'
---

# Product Requirements Document - nfv2

**Author:** NLG
**Date:** 2026-04-28

## Executive Summary

NodeForge OS is a visual, spec-driven programming workbench that transforms software development into **deterministic, node-graph-based workflows**. Where current spec-driven frameworks (BMAD Method, Get-It-Done, GitHub Spec Kit) produce **static documents**, NodeForge OS produces an **executable visual graph** that plans, researches, writes code, verifies results, and documents autonomously.

**Problem:** Existing AI coding tools and spec-driven frameworks are either too verbose (endless conversations in Cursor, Claude Code) or too static (markdown PRDs that don't execute). Solo developers lack a tool that combines **spec-driven discipline** with **autonomous execution** — seeing the entire project as a living, trackable graph.

**Solution:** Developers describe their goal (e.g., "Convert this JS project to Go"). NodeForge's AI analyzes the input, creates **layered executable nodes** (Goal → Research → Plan → Implement → Test → Review), attaches **skill sub-nodes** with installation hooks and memory, and executes each node **deterministically** — pausing only when acceptance criteria aren't met or human approval is required. The graph **only moves forward** when the current node verifies its result. Meanwhile, the system **auto-generates specs**, adds system references, and narrows scope — delivering **agile, visual, spec-driven development**.

**Target Users:** Solo developers who want the structure of BMAD/Get-It-Done with the autonomy of AI agents — **manageable at scale**, not buried in verbose conversations.

**Why Now:** LLMs have reached the capability threshold to autonomously plan architectures, research patterns, write production code, verify results, AND generate documentation — all as an integrated, visual graph.

---

### What Makes This Special

Unlike conversational AI tools (Cursor, Claude Code) or static spec frameworks (BMAD, GitHub Spec Kit), NodeForge OS is the **anti-conversation**:

1. **Executable Spec Framework** — nodes don't just describe work, they DO the work (research, write, test, verify)
2. **Professional Node Structure** — n8n-style workflow canvas with clean connections, TouchDesigner-style interactive wires (pluck edges for info), DaVinci Resolve-style node trees with clear input/output pins
3. **Forward-Only Progress** — the graph state is the source of truth; nodes save memory and only advance when verified
4. **Visual by Default** — n8n-inspired canvas (animated pulses, color-coded phase bands, mini-map heat) shows the entire project as a living graph
5. **Provider-Agnostic with Race Mode** — Ollama (local, free) races OpenAI/Anthropic; fastest result wins

---

## Project Classification

- **Project Type:** `developer_tool` (with `cli_tool` and `visual_workflow` elements)
- **Domain:** `scientific` (AI/ML orchestration, graph algorithms)
- **Complexity:** `high` (10+ subsystems, real-time WebSocket, marketplace, multi-provider LLM, sandboxing)
- **Project Context:** `greenfield`

---

## Success Criteria

### User Success

**For the solo developer target user:**

1. **Deterministic Progress Visibility** — User can see the entire project state at a glance (green/red nodes). Within 30 seconds of opening NodeForge, they know exactly what's done and what's failing.
2. **Zero Verbose Conversations** — User describes goal ("Convert JS→Go project") → AI creates and executes nodes autonomously. No 50-turn conversations like Cursor/Claude Code.
3. **Autonomous Node Execution** — Each node (Research, Implement, Test) works until acceptance criteria are met. User only intervenes at "Human Approval" nodes. 95%+ of development happens without user input.
4. **Professional Node Structure** — n8n-style canvas with clean connections, TouchDesigner-style interactive wires (pluck edges for info), DaVinci Resolve-style node trees with clear input/output pins. Visual clarity like professional node-based tools.
5. **Forward-Only Progress** — Nodes save state/memory and only advance when verified. No backtracking, no lost context, no "what were we doing?" moments.

### Business Success

1. **Solo Developer Adoption** — 1,000+ active solo developers within 6 months of launch (measured by unique session creations). 10,000+ within 12 months.
2. **Project Completion Rate** — 80%+ of started graphs reach "all nodes green" within the estimated timeframe. Successfully converts projects (JS→Go, etc.) without user rewrites.
3. **Skill Ecosystem** — 50+ community-contributed skills in marketplace within 12 months. 10+ "verified" skills with high ratings.
4. **Time-to-Value** — New user can describe a goal → see first node executing within 5 minutes of installing `nforge`.

### Technical Success

1. **Executable Spec Framework** — NodeForge graphs ARE the spec, progress tracker, and execution engine. A graph with all green nodes = completed project (e.g., JS→Go migration complete with tests passing).
2. **Gin Backend** — REST API + WebSocket hub using Gin (not Chi). High-performance HTTP + WS in single framework. 5000+ concurrent WebSocket connections supported.
3. **Smart Context Engine** — Knowledge graph for token-efficient context assembly. Embeds summaries, reuses node memory. 30%+ token reduction vs naive prompt assembly. LLM prompts reference knowledge graph instead of full context dump.
4. **LLM Provider Agnostic** — Works with Ollama (local, free), OpenAI, Anthropic, DeepSeek, OpenRouter. Race mode selects fastest/cheapest. 99.9% uptime via fallback chains.
5. **CLI/UI Feature Parity** — `nforge run <spec>` executes same graphs as the browser UI. Headless CI/CD works identically to interactive mode.
6. **Skill Marketplace** — Installable skill manifests with dependency resolution. gRPC plugins enable third-party node types. MCP server allows AI-to-AI orchestration (Claude Desktop, Cursor).
7. **Performance** — 100+ node graphs render smoothly (Web Worker offloading). Merkle tree incremental execution skips unchanged nodes. Sub-200ms provider race wins.

### Measurable Outcomes

| Metric | Target (6mo) | Target (12mo) |
|---|---|---|
| Active solo developers | 1,000 | 10,000 |
| Project completion rate | 70% | 85% |
| Community skills | — | 50+ |
| Avg time-to-first-node | <5 min | <3 min |
| LLM cost per project | <$2.00 | <$1.00 (race mode) |
| CLI adoption vs UI | 50/50 | 60/40 (CLI preference) |
| Token reduction (Smart Context) | 20%+ | 30%+ |

---

## Product Scope

### MVP — Minimum Viable Product

**Essential for proving the concept:**

1. **Core Graph Engine** (`internal/engine/`) — Goal → Spec → Plan → Implement → Test → Review nodes with acceptance criteria
2. **Gin Backend** — REST API + WebSocket hub for real-time updates
3. **React Frontend (React Flow)** — Visual node editor with green/red status, basic layout
4. **LLM Integration** — OpenAI + Ollama support, basic prompt templates
5. **Smart Context Engine** — Knowledge graph for token-efficient context, embeds summaries, reuses node memory
6. **CLI** — `nforge serve`, `nforge run <spec>`, basic commands
7. **Session Management** — Create, save, resume sessions with workspace isolation
8. **Embedded Frontend** — `embed.FS` serves React build from Go binary
9. **Deterministic Execution** — Nodes run sequentially, retry on failure, only advance when green

### Growth Features (Post-MVP)

1. **Skill Marketplace** — Skill manifests, install/dependency resolution, community ratings
2. **Provider Race Mode** — Ollama vs OpenAI, fastest/cheapest wins
3. **Professional Node Structure** — n8n-style canvas with clean connections, TouchDesigner interactive wires, DaVinci Resolve node trees
4. **gRPC Plugin Interface** — Third-party node types, dynamic loading
5. **MCP Server** — Claude Desktop/Cursor orchestrates NodeForge via MCP
6. **Advanced Canvas** — Reactive edges, n8n features, touch gestures
7. **LLM Inner Monologue** — Transparency panel showing Chain-of-Thought
8. **Auto-Prompt Optimization** — Prompts that learn from execution feedback

### Vision (Future)

1. **Autonomous Development** — Human designs process, graph executes without intervention
2. **Universal Workflow Engine** — Beyond dev: "NodeForge for X" (Data Science, Content, Ops)
3. **Self-Improving Graph** — Graph optimizes its own topology based on execution history
4. **Professional Node Canvas** — n8n-style, TouchDesigner wires, DaVinci Resolve trees — the visual standard for spec-driven development
5. **Solo Developer Empowerment** — One developer, AI swarm within each node on an n8n/DaVinci/TouchDesigner-style canvas — entire project as a living visual graph

## Domain-Specific Requirements

### Compliance & Regulatory

- **Reproducibility Standards** — Graphs must produce identical results with same inputs (deterministic execution). Critical for scientific computing.
- **Validation Methodology** — Each node's acceptance criteria = scientific validation. Must be auditable, versioned, and peer-reviewable.
- **Computational Resources** — LLM calls are expensive. Token budgets, race mode, and smart context engine (knowledge graph) are NOT optional — they're core requirements.

### Technical Constraints

- **Performance** — Real-time WebSocket updates for 100+ node graphs. Web Worker offloading, incremental execution (Merkle trees).
- **Accuracy** — LLM outputs must be verified (hallucination detector, dual-LLM verification). Scientific computing demands correctness.
- **Data Privacy** — API keys encrypted at rest, workspace chroot jail, session secrets via Vault.

### Integration Requirements

- **LLM Provider Diversity** — Ollama (local), OpenAI, Anthropic, DeepSeek, OpenRouter. Scientific workflows need model comparison.
- **Benchmarking** — Algorithm Olympics — compare multiple approaches with benchmarks. Critical for scientific research.
- **Research Nodes** — Literature review personas, research documents as skill sub-nodes.

### Risk Mitigations

- **LLM Hallucination** — Dual-LLM verification, hallucination detector. In scientific computing, incorrect results are catastrophic.
- **Token Cost Runaway** — Token budget enforcer, rate limiting, race mode.
- **Reproducibility Failure** — Merkle tree hashing, Git auto-commit per node, session export as tarball.

## Project Scoping

### Strategy & Philosophy

**Single Release Approach:** All user-specified requirements ship together in v1.0. NodeForge OS is positioned as a complete, revolutionary tool — not a phased rollout that might feel "incomplete" compared to Cursor/Claude Code.

**Resource Requirements:** 1 developer (NLG), ~6 months to MVP based on 14 build milestones.

### Complete Feature Set

**Core User Journeys Supported:**
1. **JS→Go Migration** (Alex: Goal → Research → Plan → Implement → Test → Review, all green)
2. **Stuck Node Recovery** (Sam: Inner Monologue, node memory, forward-only progress)
3. **CI/CD Integration** (Jordan: Headless CLI, YAML graphs, webhooks, human approval)

**Must-Have Capabilities (v1.0 — All Ship):**
1. **Core Graph Engine** (`internal/engine/`) — Goal → Spec → Plan → Implement → Test → Review nodes with acceptance criteria
2. **Gin Backend** — REST API + WebSocket hub (5000+ concurrent connections)
3. **React Frontend** (React Flow) — n8n-style canvas, TouchDesigner wires, DaVinci Resolve trees
4. **LLM Integration** — OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama (local)
5. **Smart Context Engine** — Knowledge graph, 30%+ token reduction
6. **CLI** — `nforge serve`, `nforge run <spec>`, `nforge new`, `nforge config`, `nforge skill`, `nforge session`
7. **Session Management** — Create, save, resume, export, fork, auto-commit
8. **Embedded Frontend** — `embed.FS` serves React build from Go binary
9. **Deterministic Execution** — Sequential, retry on failure, only advance when green
10. **Skill System** — Manifests, sub-nodes, hooks, memory, dependency resolution
11. **Provider Race Mode** — Ollama vs OpenAI, fastest/cheapest wins
12. **LLM Inner Monologue** — Transparency panel, Chain-of-Thought streaming
13. **Advanced Canvas** — Reactive edges, n8n features, color-coded phase bands, mini-map heat
14. **Security** — eBPF filtering, API key encryption, chroot jail, rate limiting
15. **DevOps** — Multi-arch Docker, distroless, health checks, graceful shutdown

**Nice-to-Have Capabilities (Also in v1.0):**
1. **AI Swarm per Node** — Multiple LLM agents negotiate within a single node on the n8n/DaVinci/TouchDesigner-style canvas; node structure is visual, swarm is internal to node execution
2. **Auto-Prompt Optimization** — Prompts that learn from execution feedback
3. **Session Star/Fork** — GitHub-style social features
4. **Accessibility** — Screen readers, colorblind themes, RTL, 20+ languages
5. **gRPC Plugins** — Third-party node types, dynamic loading
6. **MCP Server** — Claude Desktop/Cursor orchestrates NodeForge via MCP

### Risk Mitigation Strategy

**Technical Risks:**
- **Complexity (HIGH):** 10+ subsystems, real-time WS, marketplace, multi-provider LLM → Mitigate: Follow 14 build milestones sequentially
- **Gin + WebSocket Expertise:** Not Chi as originally planned → Mitigate: Gin is simpler for unified HTTP+WS

**Market Risks:**
- **Cursor/Claude Code Dominance:** → Mitigate: "Anti-conversation" positioning, solo-dev focus, deterministic progress visibility

**Resource Risks:**
- **1 Developer (Solo):** 14 milestones is substantial → Mitigate: Focus on MVP essentials first (scaffold → config → CLI → engine → sessions → LLM → skills)

## Functional Requirements

### Core Graph Capabilities

- FR1: User can create a new session with a goal description ("Convert JS→Go") and AI auto-generates a complete node graph (Goal → Spec → Plan → Implement → Test → Review)
- FR2: User can see the entire project state at a glance with color-coded nodes (green=complete, red=failed, yellow=running)
- FR3: User can watch nodes execute deterministically — each node works until acceptance criteria are met, then advances forward
- FR4: User can interact with node connections n8n-style — pluck edges to see metadata, see data/control flow (TouchDesigner-style interactive wires)
- FR5: User can view DaVinci Resolve-style node trees with clear input/output pins and signal flow
- FR6: User can see animated pulses along edges showing real-time execution progress (n8n-inspired canvas)
- FR7: User can deactivate/activate individual nodes without deleting them (n8n feature)
- FR8: User can view a mini-map with execution heat — nodes glow based on recent activity
- FR9: User can add comments/notes to nodes (sticky notes like n8n)

### LLM Integration Capabilities

- FR10: User can configure multiple LLM providers (OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama) via `nforge config`
- FR11: System can race multiple LLM providers simultaneously — fastest token wins, slower is cancelled
- FR12: System can auto-fallback through providers when rate limits or errors occur (semantic matching: rate limit → cheaper model)
- FR13: User can see LLM "Inner Monologue" (Chain-of-Thought) in a side panel with streaming tokens
- FR14: System can optimize prompts automatically based on execution feedback (prompt learns over time)
- FR15: System can enforce token budgets — pre-flight estimation rejects expensive requests
- FR16: User can use local Ollama (free) racing against remote APIs (cost control)

### Smart Context Engine Capabilities

- FR17: System builds a knowledge graph for token-efficient context assembly (30%+ reduction vs naive prompts)
- FR18: System reuses node memory — each node's output becomes context for downstream nodes
- FR19: System auto-generates specs and adds system references as nodes execute
- FR20: System can handle context overflow by auto-splitting graphs into sub-graphs

### CLI Capabilities

- FR21: User can start the web UI + API with `nforge serve`
- FR22: User can run headless execution with `nforge run <spec-file>` (same as UI, but in terminal)
- FR23: User can create new projects with `nforge new <project-name>`
- FR24: User can configure settings with `nforge config set/get` (API keys, models, ports)
- FR25: User can manage skills with `nforge skill list/install <name>`
- FR26: User can resume/export sessions with `nforge session resume/export <id>`
- FR27: User can see ASCII art graph in terminal with `nforge graph viz`
- FR28: User can check system health with `nforge doctor`
- FR29: CLI has tab completion for all commands and node types
- FR30: CLI and UI have feature parity — same graphs execute identically

### Session Management Capabilities

- FR31: User can create isolated sessions with separate workspaces
- FR32: System auto-saves session state — graph JSON, chat log, workspace files
- FR33: User can resume sessions after restart (graceful shutdown with snapshot)
- FR34: User can fork sessions (like Git branches) — try different approaches, merge or discard
- FR35: System auto-commits workspace changes to Git after each node completion
- FR36: User can time-travel debug — checkout workspace state at any completed node
- FR37: User can export session as self-contained tarball (graph + source + README)
- FR38: System enforces session quotas (max sessions, max workspace size)
- FR39: System auto-cleans zombie sessions (heartbeat timeout detection)

### Skill System Capabilities

- FR40: User can browse and install skills from marketplace (community-contributed manifests)
- FR41: Skills can have dependencies — installing one skill auto-installs its dependency tree
- FR42: Skills are sandboxed before trust — time limits, no network, read-only filesystem
- FR43: System supports gRPC plugins — third-parties can define entirely new node types
- FR44: System exposes MCP server — Claude Desktop, Cursor can orchestrate NodeForge via MCP tools
- FR45: Skills can have sub-nodes — e.g., "JS-to-Go" skill has "patterns," "goroutines" sub-nodes
- FR46: User can A/B test skills — system routes to different versions, collects metrics

### Visual Canvas Capabilities

- FR47: User can drag-and-drop files onto canvas to auto-create nodes (e.g., `go.mod` → `Setup` node)
- FR48: User can see color-coded node bands by lifecycle phase (blue=Discovery, orange=Execution, red=Recovery, green=Completion)
- FR49: User can see reactive edge tension — edges visually tighten when upstream nodes fail
- FR50: User can use Vim/Emacs keybindings for canvas navigation (hjkl, Ctrl-f/b/n/p)
- FR51: System supports accessibility — screen reader announcements for node status changes, high-contrast themes, RTL canvas, 20+ language localization

### Execution & Performance Capabilities

- FR52: System executes nodes sequentially with retry loops — stays inside node until acceptance criteria met
- FR53: System can run multiple attempts in parallel within a node (speculative execution, best result wins)
- FR54: System supports incremental execution — Merkle tree hashing skips unchanged nodes
- FR55: System offloads graph layout to Web Worker — 100+ node graphs render smoothly
- FR56: System pre-fetches LLM provider status on session start — zero wait for connectivity checks
- FR57: System uses Gin backend — single framework for REST API + WebSocket hub (5000+ concurrent connections)

### Security Capabilities

- FR58: System encrypts API keys at rest using Argon2 key derivation + AES-256-GCM
- FR59: System sandboxes node execution in chroot jail — no escape to parent directories
- FR60: System applies eBPF syscall filtering — blocks dangerous system calls during execution
- FR61: System enforces rate limiting per API key using token bucket algorithm
- FR62: System signs graph snapshots with Ed25519 — detects tampering on session import

### DevOps Capabilities

- FR63: System builds as multi-arch Docker (amd64 + arm64) in single container
- FR64: System uses distroless Docker image — only Go binary, no shell, no OS utilities
- FR65: System includes Ollama sidecar option — auto-provisions local LLM on `docker compose up`
- FR66: System exposes health check with graph diagnostics (/healthz returns session stats, LLM connectivity)
- FR67: System can send webhook notifications — nodes reach out to Slack, GitHub, etc. when certain states reached
- FR68: System exports telemetry as Prometheus metrics (/metrics endpoint)

## Non-Functional Requirements

### Performance

| NFR ID | Requirement | Measurement |
|--------|-------------|-------------|
| NFR-01 | WebSocket state propagation | <50ms latency from node state change to browser UI (Gin WS hub, 5000+ concurrent connections) |
| NFR-02 | Graph rendering | 100+ node graphs render at 60fps with Web Worker offloading, zero main-thread blocking |
| NFR-03 | LLM Race Mode | Sub-200ms provider race wins — fastest first token wins, slower providers cancelled immediately |
| NFR-04 | Smart Context Assembly | Knowledge graph context assembled in <100ms, achieving 30%+ token reduction vs naive prompt assembly |
| NFR-05 | Token Budget Pre-flight | Budget estimation completes in <10ms before LLM call is dispatched |
| NFR-06 | Incremental Execution | Merkle tree hashing skips unchanged nodes — re-execution of 100-node graph completes in <2s when 95% unchanged |

### Security

| NFR ID | Requirement | Measurement |
|--------|-------------|-------------|
| NFR-07 | API Key Storage | Argon2 key derivation + AES-256-GCM encryption at rest; keys never in plaintext config files |
| NFR-08 | Workspace Isolation | Chroot jail per session — node execution cannot escape to parent directories; verified by integration tests |
| NFR-09 | Syscall Filtering | eBPF filters block dangerous syscalls (exec, mount, reboot) during node execution |
| NFR-10 | Session Secrets | Vault integration for secret management; API keys never logged, never in session export tarballs |
| NFR-11 | Rate Limiting | Token bucket algorithm, per-API-key limits — 100 req/min for REST, 10 msg/sec for WebSocket |
| NFR-12 | Graph Integrity | Ed25519 signing of graph snapshots — tampered session imports rejected with audit log entry |
| NFR-13 | Skill Sandboxing | Pre-trust execution: time limits (30s), no network, read-only filesystem; escalated after signature verification |

### Scalability

| NFR ID | Requirement | Measurement |
|--------|-------------|-------------|
| NFR-14 | Concurrent Connections | Single Gin instance supports 5000+ WebSocket connections with <5% memory growth beyond baseline |
| NFR-15 | User Growth Trajectory | Architecture supports growth from 1,000 to 10,000+ solo developers without structural changes (horizontal scaling ready via stateless Gin + external session store) |
| NFR-16 | Graph Complexity | Smooth interaction with 100+ node graphs; layout Web Worker prevents UI freeze; pan/zoom remains 60fps |
| NFR-17 | Session Density | 100+ concurrent sessions per instance, each with isolated workspace (max 500MB workspace size quota) |
| NFR-18 | Multi-Arch Distribution | Docker images built for amd64 + arm64 in single multi-arch manifest; Ollama sidecar auto-provisions on `docker compose up` |
| NFR-19 | LLM Provider Failover | 99.9% LLM uptime via fallback chains — Ollama → OpenAI → Anthropic → DeepSeek → OpenRouter |

### Accessibility

| NFR ID | Requirement | Measurement |
|--------|-------------|-------------|
| NFR-20 | Screen Reader Support | WCAG 2.1 AA compliance — node status changes announced via ARIA live regions; canvas keyboard-navigable |
| NFR-21 | Visual Accessibility | High-contrast themes + colorblind-friendly palette — nodes distinguished by shape + position, not just color (red/green alone insufficient) |
| NFR-22 | Internationalization | RTL canvas support, 20+ language localization (minimum: EN, ES, FR, DE, ZH, JA, KO, PT, RU, AR) |
| NFR-23 | Keyboard Navigation | Vim (hjkl) and Emacs (Ctrl-f/b/n/p) keybindings for canvas navigation; all interactions possible without mouse |
| NFR-24 | Motor Impairment Support | All node operations (create, connect, delete, configure) accessible via keyboard shortcuts and context menu |

### Integration

| NFR ID | Requirement | Measurement |
|--------|-------------|-------------|
| NFR-25 | LLM Provider Agnosticism | 5+ providers (Ollama, OpenAI, Anthropic, DeepSeek, OpenRouter) — pluggable provider interface, new providers added via config (no code changes) |
| NFR-26 | gRPC Plugin System | Third-party node types load dynamically via gRPC; plugins sandboxed (separate process, resource limits); crash of one plugin doesn't affect core engine |
| NFR-27 | MCP Server Compliance | Claude Desktop/Cursor orchestrates NodeForge via MCP tools — full session lifecycle exposed (create, resume, fork, export) |
| NFR-28 | Git Integration | Auto-commit per node completion with deterministic commit messages; time-travel debug via `git checkout` at any completed node |
| NFR-29 | Webhook Notifications | Outbound webhooks to Slack, GitHub, etc. when nodes reach configurable states (success, failure, approval needed); retries with exponential backoff |
| NFR-30 | Telemetry Export | Prometheus metrics (/metrics endpoint) — session count, LLM token usage, node execution duration, WebSocket connection count |

