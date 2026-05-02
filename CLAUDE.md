# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NodeForge (nfv2) is a spec-driven development workbench that uses LLMs to auto-generate and execute node graphs from natural language goals. Users describe a goal in a chat interface, and the system creates a node graph (Goal → Spec → Plan → Implement → Test → Review) that executes autonomously.

## Build & Development Commands

```bash
# Go backend
go build -o nforge ./cmd/nforge/     # Build binary
go test ./...                          # Run all Go tests
go test ./internal/engine/...             # Run specific package tests
go test -run TestGenerator ./...           # Run single test
go build -ldflags "-X main.version=v1.0" ./cmd/nforge/  # Build with version

# Frontend
cd frontend && npm run dev        # Dev server on :5173
cd frontend && npm run build      # Production build to dist/
cd frontend && npx tsc --noEmit  # TypeScript check
cd frontend && npm run lint      # ESLint
```

## Architecture

### Backend (Go)

```
cmd/nforge/          # CLI entry point (Cobra commands)
  ├── root.go           # Root command, version, completion
  ├── serve.go          # Gin server, WebSocket hub, API routes
  ├── config.go         # Configuration management
  ├── doctor.go         # Health checks
  ├── session.go        # Session management commands
  └── ...

internal/
  ├── engine/           # Graph execution engine
  │   ├── graph.go        # Graph/Node/Edge types, Generator (LLM→graph)
  │   └── executor.go    # Sequential executor, retry logic, acceptance criteria
  ├── llm/              # LLM provider abstraction
  │   ├── provider.go     # Provider interface, Race() for multi-provider
  │   ├── ollama.go      # Ollama local provider
  │   ├── openai.go      # OpenAI provider
  │   └── ...
  ├── context/          # BadgerDB storage
  │   └── memory.go       # SaveGraph, GetGraph, SaveNodeOutput, GetNodeOutput
  ├── canvas/           # React Flow canvas API routes
  └── session/          # Session manager, workspace
```

Key patterns:
- **Gin + WebSocket hub** at `cmd/nforge/serve.go` — REST API on `/api/v1/*`, WebSocket on `/ws`
- **Backend-first**: All graph logic/execution in Go; frontend only visualizes
- **NodeUpdateBroadcaster interface** (`executor.go`) abstracts WebSocket hub — executor calls `BroadcastNodeUpdate`, `BroadcastEdgeUpdate`, `BroadcastRaw`
- **LLM Provider interface** (`llm/provider.go`): `Chat()`, `ChatStream()`, `Name()` — supports OpenAI, Anthropic, DeepSeek, OpenRouter, Ollama
- **Race() function**: runs multiple LLM providers concurrently, returns fastest response, cancels the rest

### Frontend (TypeScript + React)

```
frontend/src/
  ├── App.tsx                    # Main app, WebSocket integration, keybindings
  ├── main.tsx                   # Entry point
  ├── index.css                  # Global styles, CSS variables
  ├── nodes.ts / edges.ts         # Initial React Flow data
  ├── components/
  │   ├── canvas/               # React Flow node/edge types
  │   │   ├── NodeTypes.tsx      # Goal, Spec, Plan, Implement, Test, Review
  │   │   ├── EdgeTypes.tsx      # default, active, tension, success
  │   │   └── CanvasControls.tsx
  │   ├── panels/               # UI panels
  │   │   ├── ChatPanel.tsx      # Goal input, collapsible
  │   │   ├── MonologuePanel.tsx  # LLM streaming thoughts
  │   │   └── SessionExplorer.tsx
  │   └── ui/                    # Shared UI components
  ├── hooks/
  │   └── useWebSocket.ts        # WebSocket hook, queue-based state
  └── workers/                   # Web Workers (performance)
```

Key patterns:
- **Vite + React 18 + TypeScript** with `@xyflow/react` for the canvas
- **useWebSocket hook** returns queue-based state (`graphUpdateQueue`, `nodeUpdateQueue`, `edgeUpdateQueue`) to prevent message overwrite
- **WebSocket message types**: `graph_update`, `node_update`, `edge_update`, `llm_chunk`, `monologue`, `connected`

### BMAD Project Management

```
_bmad-output/
  ├── planning-artifacts/       # PRD, architecture, epics, UX design
  │   ├── prd.md
  │   ├── architecture.md
  │   ├── epics.md
  │   └── ux-design-specification.md
  └── implementation-artifacts/
      ├── sprint-status.yaml    # Story status tracking
      ├── 2-1-chat-interface-...md
      └── deferred-work.md

.claude/skills/bmad-*/      # BMAD skill definitions
```

Stories follow naming: `{epic#}-{story#}-{slug}`. Status: `backlog` → `ready-for-dev` → `in-progress` → `review` → `done`.

## Code Conventions

- **Go**: `snake_case` packages (`internal/engine/`), errors wrapped with `%w`
- **TypeScript**: `kebab-case.tsx` files, `PascalCase` components, `camelCase` JSON fields
- **API**: `snake_case` endpoints (`/api/v1/sessions`), `camelCase` JSON fields
- **React Flow**: Nodes/edges as `any[]` (cast from `unknown`), status-driven styling
- **WebSocket**: JSON messages with `type` field routing in `App.tsx` effect hooks

## Environment Variables

- `NFORGE_VERBOSE=true` — Enable debug logging
- `NFORGE_CONFIG=/path` — Config file path
- Ollama default: `http://localhost:11434`
