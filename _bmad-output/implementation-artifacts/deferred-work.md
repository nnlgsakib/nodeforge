# Deferred Work

## Deferred from: code review of 2-1-chat-interface-and-auto-generated-node-graph (2026-05-02)

- Vim/Emacs canvas navigation keybindings are non-functional — AC3 violation: keybindings only log to console with no actual React Flow canvas panning — deferred to story 3-1/3-3 (UX: canvas navigation)
- Execution controls (pause/skip/fork/retry) lack backend support — AC3 violation: frontend sends messages but executor.go has no handlers — deferred to story 2-7 (incremental execution and web worker offloading)
- WebSocket <50ms latency guarantee not implemented — AC1/NFR-01 violation: no code enforces or measures <50ms broadcast latency — deferred to story 2-6 (performance optimization) or 6-6 (provider failover & performance)

## Deferred from: code review of 2-2-llm-provider-abstraction-and-race-mode (2026-05-02)

- WebSocket hub `run()` loop has no stop mechanism [serve.go:61-91,232] — pre-existing issue not caused by this story, would require shutdown signal handling in hub

## Deferred from: code review of story-3.1-custom-nodetypes-and-edgetypes (2026-05-03)

- `hideAttribution: true` license risk — Requires paid React Flow license [App.tsx:368] — deferred, pre-existing decision
- `as any[]` pervasive in App.tsx layout worker — Pre-existing pattern not introduced by this diff [App.tsx] — deferred, pre-existing technical debt
- `chatGenerating` flag only reset by layout effect — Pre-existing in App.tsx, not in this diff [App.tsx] — deferred, pre-existing

## Deferred from: code review of 2-4-prompt-optimization-and-token-budget (2026-05-03)

- Missing REST API endpoint for budget status reporting (AC4) — WebSocket implemented but REST API for BudgetStatus() not implemented, separate concern from story scope
- Context cancellation ignored in budget and optimizer methods — methods accept context.Context but never check for cancellation, long-running operations won't abort, not critical for initial implementation
