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

## Deferred from: code review of story-3.3-canvascontrols-and-sessionexplorer (2026-05-03)

- `statusColors` uses hardcoded hex instead of CSS variables — inconsistent with design system but not a bug [`SessionExplorer.tsx:22-27`]
- Web Worker never terminated on unmount — pre-existing issue in `useLayoutWorker.ts`
- `useKeyboardShortcuts` global 'p' key conflicts with CanvasControls Ctrl+p — pre-existing, not introduced by this diff
- Notification timeout not cleared on unmount — pre-existing pattern in App.tsx

## Deferred from: code review of story-3.6-accessibility (2026-05-04)

- Keyboard handler attaches to `document` as fallback — pre-existing, not introduced by this diff [CanvasControls.tsx:150-158]
- `handleClear` confirms even when messages empty — pre-existing logic issue [monologue-panel.tsx:57-60]
- `displayMessages` slice creates stale key references — pre-existing, minor DOM churn [monologue-panel.tsx:39-41]
- Subtask 4.7 node creation/deletion/connection via keyboard — substantial UX feature, better scoped as its own story. Mark 4.7 as partially done [NodeTypes.tsx, CanvasControls.tsx]

## Deferred from: code review of story-3.7-ui-patterns-buttons-feedback-loading-modals (2026-05-04)

- Edge dash flow CSS (`edge-dash` keyframe) defined but never applied by any edge component — pre-existing, CSS defined but not consumed
- Empty string toast title renders invisible title — cosmetic edge case, unlikely in practice
- Toast stacking on rapid success→error — low-impact UX, no dedup logic
- No test for `m` key MonologuePanel toggle required by AC4 — test coverage gap

## Deferred from: code review of 4-5-session-resume-export-cli (2026-05-04)

- `formatBytes` produces nonsense for negative values — `formatBytes(-1)` returns `"-1 B"`. Workspace size cannot be negative; defensive but not required for this change.
- `sanitizeMap` misses double-encoded JSON strings — If a sensitive value is nested inside a JSON string, it is not parsed or sanitized. Pre-existing sanitization limitation.
- Symlink check in `addWorkspaceToTar` may not work on Windows — `filepath.Walk` symlink detection behavior differs on Windows. Pre-existing.
- Export API uses POST instead of spec's GET — `serve.go` registers POST, spec task says GET. Dev intentionally changed to POST; pre-existing decision.

## Deferred from: code review of 5-1-skill-marketplace-and-dynamic-fetch (2026-05-05)

- close(hub.done) panics if runServer called twice [cmd/nforge/serve.go:719] — pre-existing
- Shutdown sleep 200ms may be insufficient under load [cmd/nforge/serve.go:722] — pre-existing
- Hub drain loop busy-spins without yielding [cmd/nforge/serve.go:112-119] — pre-existing
- sanitizeGraphJSON returns "{}" for non-JSON data, potentially losing valid content [internal/session/export.go:56-85] — pre-existing
- addWorkspaceToTar symlink handling inconsistent on Windows [internal/session/export.go:259-333] — pre-existing
- ExportSession cleanup doesn't close file before remove on error [internal/session/export.go:181-189] — pre-existing
- formatBytes produces nonsense for negative values [cmd/nforge/session.go:113-126] — pre-existing
