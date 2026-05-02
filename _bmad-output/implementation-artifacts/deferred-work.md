# Deferred Work

## Deferred from: code review of 2-1-chat-interface-and-auto-generated-node-graph (2026-05-02)

- Vim/Emacs canvas navigation keybindings are non-functional — AC3 violation: keybindings only log to console with no actual React Flow canvas panning — deferred to story 3-1/3-3 (UX: canvas navigation)
- Execution controls (pause/skip/fork/retry) lack backend support — AC3 violation: frontend sends messages but executor.go has no handlers — deferred to story 2-7 (incremental execution and web worker offloading)
- WebSocket <50ms latency guarantee not implemented — AC1/NFR-01 violation: no code enforces or measures <50ms broadcast latency — deferred to story 2-6 (performance optimization) or 6-6 (provider failover & performance)
