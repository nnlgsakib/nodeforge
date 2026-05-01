## Deferred from: code review of 1-3-gin-server-with-nforge-serve (2026-04-30)

- No HTTP server timeouts — no ReadTimeout/WriteTimeout/IdleTimeout, vulnerable to slowloris [serve.go:222-225] — deferred, pre-existing
- Unlimited WebSocket connections — no cap, no connection tracking. Spec requires 5000+ support [serve.go:102-115] — deferred, later story
- WebSocket messages ignored — spec defines message format but messages are discarded [serve.go:109-113] — deferred, later story
- Hardcoded 5s shutdown timeout — not configurable [serve.go:241] — deferred, pre-existing
- CORS no credentials support — `Access-Control-Allow-Credentials` not set [serve.go:76-85] — deferred, pre-existing
- No integration tests (Task 5) — no test files provided [N/A] — deferred, needs test framework decision
- Inconsistent error logging — mix of `fmt.Printf`, returned errors, and silent client messages [throughout serve.go] — deferred, pre-existing

## Deferred from: code review of 1-4-project-creation-via-cli-and-ui.md (2026-04-30)

- Large frontend files read entirely into memory [serve.go:NoRoute] — deferred, pre-existing
- Missing //go:embed directive for frontend assets [main.go] — `var distFS embed.FS` without `//go:embed` comment. Pre-existing, not introduced by Story 1.4.
