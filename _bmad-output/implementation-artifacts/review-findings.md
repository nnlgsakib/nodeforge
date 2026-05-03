# Code Review Findings - Story 2.7

## Triaged Findings

### CRITICAL (Patch)

**1. [blind] Merkle hash includes node.Output causing false change detection**
- **Detail**: `hashNode()` in `merkle.go` includes `node.Output` in the hash. Since Output is the execution result (not a config parameter), any executed node will always produce a new hash, making incremental execution via Merkle detection completely non-functional.
- **Location**: `internal/engine/merkle.go` — `parts = append(parts, node.Output)`
- **Source**: blind
- **Classification**: patch
- **Fix**: Remove `node.Output` from hashNode. Only hash config/input fields (Type, Label, AcceptanceCriteria, Metadata, and per spec, inputs).

**2. [blind] App.tsx discards all but the last graph update in the queue**
- **Detail**: Original code looped over all `graphUpdateQueue` items. New code uses `graphUpdateQueue[graphUpdateQueue.length - 1]` (only last update). This drops state for prior queued updates, breaking the project's queue-based WebSocket pattern.
- **Location**: `frontend/src/App.tsx` — `const lastUpdate = graphUpdateQueue[graphUpdateQueue.length - 1]`
- **Source**: blind+edge
- **Classification**: patch
- **Fix**: Process all items in the queue or accumulate node/edge updates properly.

**3. [blind] App.tsx fails to add new nodes from graph updates**
- **Detail**: `setNodes((nds) => nds.map(...))` only updates positions of existing nodes. Original code replaced the entire nodes array via `setNodes(data.nodes as any[])`. New nodes from graph updates are silently dropped.
- **Location**: `frontend/src/App.tsx` — `setNodes((nds) => nds.map(...))`
- **Source**: blind
- **Classification**: patch
- **Fix**: Merge new nodes into existing array instead of only mapping existing ones.

### HIGH (Patch)

**4. [blind] Changed complete nodes are not re-executed**
- **Detail**: Merkle skip logic skips unchanged nodes with `if !changedNodes[node.ID]` then `if node.Status == NodeStatusComplete { continue }`. But changed nodes with `NodeStatusComplete` are still skipped by the subsequent `node.Status != NodeStatusPending && node.Status != NodeStatusFailed` check.
- **Location**: `internal/engine/executor.go` — lines 216-224
- **Source**: blind
- **Classification**: patch
- **Fix**: Reorder checks so changed nodes are always executed regardless of status.

**5. [blind] Edge-only changes mark all nodes as changed unnecessarily**
- **Detail**: `detectChangedNodes()` adds ALL nodes to changed list if graph hash mismatches but no individual node hashes changed (only when edges change). This forces full re-execution for minor edge changes.
- **Location**: `internal/engine/merkle.go` — `detectChangedNodes()` edge change handling
- **Source**: blind
- **Classification**: patch
- **Fix**: Only mark nodes as changed if their edges actually changed (trace edge changes to affected nodes).

**6. [blind+auditor] Layout worker error fallback is non-functional**
- **Detail**: Spec requires fallback to React Flow built-in layout if worker fails. Code only logs `console.warn('Layout worker failed, falling back:', err)` with no actual fallback layout logic.
- **Location**: `frontend/src/App.tsx` — catch block
- **Source**: blind+auditor
- **Classification**: patch
- **Fix**: Implement React Flow built-in layout as fallback when worker fails.

**7. [blind+edge] Stale layout results overwrite newer state**
- **Detail**: `useLayoutWorker.ts` `runLayout` adds a new message event listener per call with no cancellation of prior requests. Pending layouts for old graph updates resolve and overwrite state with stale positions when new updates arrive.
- **Location**: `frontend/src/hooks/useLayoutWorker.ts` — `runLayout()`
- **Source**: blind+edge
- **Classification**: patch
- **Fix**: Track request IDs or implement cancellation of pending layout requests.

### MEDIUM (Patch/Decision)

**8. [edge] WebSocket client count can go negative or inflate**
- **Detail**: `clientCount.Add(1)` is called without checking if client is already registered. `clientCount.Add(-1)` is called without checking if client was actually registered. This can lead to incorrect metrics.
- **Location**: `cmd/nforge/serve.go` — register/unregister handlers
- **Source**: edge
- **Classification**: patch
- **Fix**: Check `if !h.clients[client] { h.clientCount.Add(1) }` and `if _, ok := h.clients[client]; ok { h.clientCount.Add(-1) }`.

**9. [auditor] Modified App.tsx instead of specified WorkflowCanvas.tsx**
- **Detail**: Spec's "Files to MODIFY" explicitly requires changes to `frontend/src/components/canvas/WorkflowCanvas.tsx` to offload layout to Web Worker, but diff modifies `frontend/src/App.tsx` instead.
- **Location**: Spec vs actual diff
- **Source**: auditor
- **Classification**: decision_needed
- **Fix**: Confirm if App.tsx was intentionally used instead of WorkflowCanvas.tsx, or refactor to match spec.

**10. [auditor+blind] TypeScript file naming violates kebab-case convention**
- **Detail**: Project context requires `kebab-case` for all TS files. New files use dots instead of hyphens: `layout.worker.ts` (should be `layout-worker.ts`), `layout.worker.test.ts` (should be `layout-worker.test.ts`), `useLayoutWorker.ts` (should be `use-layout-worker.ts`).
- **Location**: `frontend/src/workers/layout.worker.ts`, `frontend/src/workers/layout.worker.test.ts`, `frontend/src/hooks/useLayoutWorker.ts`
- **Source**: auditor+blind
- **Classification**: patch
- **Fix**: Rename files to kebab-case per project convention.

**11. [auditor] Merkle per-node hashes stored in GraphMetadata instead of NodeMetadata**
- **Detail**: Spec requires per-node hashes in `NodeMetadata.Hash` (SHA-256), but code uses `GraphMetadata.NodeHashes` map. No `NodeMetadata` struct is added to `Node`.
- **Location**: `internal/engine/graph.go`, `internal/engine/executor.go`
- **Source**: auditor
- **Classification**: decision_needed
- **Fix**: Confirm if GraphMetadata.NodeHashes map is acceptable, or implement NodeMetadata.Hash per spec.

**12. [blind] dagre import in layout worker likely fails due to CJS/ESM mismatch**
- **Detail**: `layout.worker.ts` uses `import dagre from 'dagre'` (default import). `dagre@^0.8.5` is a CommonJS-only package. Worker is created with `{ type: 'module' }`, where the default export may not exist.
- **Location**: `frontend/src/workers/layout.worker.ts` — `import dagre from 'dagre'`
- **Source**: blind
- **Classification**: patch
- **Fix**: Use `import * as dagre from 'dagre'` or adjust worker creation to not use module type.

**13. [blind] GraphMetadata has unused dead fields with invalid JSON tags**
- **Detail**: `GraphMetadata` includes `ID` and `CreatedAt` fields with no `json` tags (serializing as PascalCase, violating the project's camelCase JSON field convention). These fields are never populated anywhere in the codebase.
- **Location**: `internal/engine/graph.go` — `GraphMetadata` struct
- **Source**: blind
- **Classification**: patch
- **Fix**: Remove unused fields or add proper `json:"id"` and `json:"createdAt"` tags.

### LOW (Patch/Defer)

**14. [edge] resolveChangedNodes crashes if e.graph is nil**
- **Detail**: `resolveChangedNodes()` accesses `e.graph.Metadata` without checking if `e.graph` is nil first.
- **Location**: `internal/engine/executor.go` — `resolveChangedNodes()`
- **Source**: edge
- **Classification**: patch
- **Fix**: Add nil check: `if e.graph == nil { return nil, fmt.Errorf("graph is nil") }`.

**15. [edge] Nil node passed to hashNode function**
- **Detail**: `hashNode()` doesn't check if `node` is nil before accessing `node.Type`, `node.Label`, etc.
- **Location**: `internal/engine/merkle.go` — `hashNode()`
- **Source**: edge
- **Classification**: patch
- **Fix**: Add nil check at start of `hashNode()`.

**16. [blind] Layout worker ignores node-specific width/height properties**
- **Detail**: `layout.worker.ts` uses hardcoded defaults `nodeWidth`/`nodeHeight` instead of the `width`/`height` optional fields defined in the `LayoutNode` interface.
- **Location**: `frontend/src/workers/layout.worker.ts` — `g.setNode(node.id, { width: nodeWidth, height: nodeHeight })`
- **Source**: blind
- **Classification**: patch
- **Fix**: Use `node.width ?? nodeWidth` and `node.height ?? nodeHeight`.

**17. [blind] hashNode uses unstable %v verb for metadata hashing**
- **Detail**: `fmt.Sprintf("%s=%v", k, node.Metadata[k])` uses `%v` to hash metadata values, which may produce inconsistent strings for equivalent logical values.
- **Location**: `internal/engine/merkle.go` — metadata hashing
- **Source**: blind
- **Classification**: patch
- **Fix**: Use deterministic serialization (e.g., JSON encode metadata values).

**18. [blind] App.tsx skips edge-only graph updates**
- **Detail**: The `hasNewNodes` check returns early if no queue items contain nodes, ignoring valid graph updates that only include edge changes.
- **Location**: `frontend/src/App.tsx` — `if (!hasNewNodes) return;`
- **Source**: blind
- **Classification**: patch
- **Fix**: Also process edge-only updates.

**19. [auditor] Merkle node hash omits required inputs field**
- **Detail**: Spec requires `hashNode` to include node inputs, but implementation does not. Spec Task 1.2: "node type, config, inputs, outputs, acceptance criteria".
- **Location**: `internal/engine/merkle.go` — `hashNode()`
- **Source**: auditor
- **Classification**: decision_needed
- **Fix**: Confirm if Node struct has inputs field, or if this is out of scope.

**20. [edge] Worker sends unexpected message type**
- **Detail**: `useLayoutWorker.ts` only handles `layout-done` and `layout-error` message types. If worker sends an unexpected type, the promise never settles.
- **Location**: `frontend/src/hooks/useLayoutWorker.ts` — message handler
- **Source**: edge
- **Classification**: patch
- **Fix**: Add else branch: `else { reject(new Error('unexpected worker message')); }`.

**21. [edge] Input nodes have duplicate IDs in layout worker**
- **Detail**: `layout.worker.ts` doesn't check for duplicate node IDs before calling `g.setNode()`. dagre may silently overwrite or error.
- **Location**: `frontend/src/workers/layout.worker.ts`
- **Source**: edge
- **Classification**: patch
- **Fix**: Add duplicate check before `g.setNode()`.

### DEFER (Pre-existing or Out of Scope)

**22. [auditor] Missing WebSocket load testing and propagation benchmarking (AC4)**
- **Detail**: Spec requires load tests for 5000+ concurrent connections and <50ms propagation benchmarks. No such test files in diff.
- **Location**: N/A (missing files)
- **Source**: auditor
- **Classification**: defer
- **Reason**: Load testing may be scoped to a later story (Story 6.6 per spec references).

---

## Summary

| Classification | Count |
|-----------------|-------|
| Critical (patch) | 3 |
| High (patch) | 4 |
| Medium (patch/decision) | 6 |
| Low (patch/defer) | 8 |
| Defer | 1 |
| **Total findings** | **22** |

| By Source | Count |
|-----------|-------|
| blind | 13 |
| edge | 8 |
| auditor | 8 |
| (merged) | 7 |

**Failed layers**: None (all 3 layers completed successfully)
