package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nnlgsakib/nodeforge/internal/llm"
	nfContext "github.com/nnlgsakib/nodeforge/internal/context"
)

// NodeUpdateBroadcaster abstracts the WebSocket hub for broadcasting node updates
type NodeUpdateBroadcaster interface {
	BroadcastNodeUpdate(nodeID, status string, progress float64)
	BroadcastEdgeUpdate(source, target string, tension float64)
	BroadcastRaw(data []byte)
}

// Executor runs nodes sequentially with retry until acceptance criteria are met
type Executor struct {
	graph             *Graph
	llmProv           llm.LLMProvider
	store             *nfContext.Store
	hub               NodeUpdateBroadcaster // WebSocket hub for real-time updates
	monologueMessages  []nfContext.MonologueMessage
	currentNodeID     string
	contextAssembler  *nfContext.ContextAssembler // Context assembly for LLM calls (Task 5)
	specGen           *nfContext.SpecGenerator    // Auto-spec generation (AC3)
	swarm             *llm.Swarm                  // Speculative execution within nodes (Story 2.6)
	swarmConfig       *llm.SwarmConfig            // Speculative execution configuration
}

// NewExecutor creates a new executor for the given graph
func NewExecutor(graph *Graph, llmProv llm.LLMProvider, store *nfContext.Store, contextAssembler *nfContext.ContextAssembler, specGen *nfContext.SpecGenerator) *Executor {
	return &Executor{
		graph:            graph,
		llmProv:          llmProv,
		store:            store,
		contextAssembler: contextAssembler,
		specGen:          specGen,
		swarmConfig:      llm.DefaultSwarmConfig(),
	}
}

// SetHub sets the WebSocket hub for real-time updates (Task 5.4)
func (e *Executor) SetHub(hub NodeUpdateBroadcaster) {
	e.hub = hub
}

// SetSwarm configures speculative execution within nodes (Story 2.6)
func (e *Executor) SetSwarm(swarm *llm.Swarm, config *llm.SwarmConfig) {
	e.swarm = swarm
	if config != nil {
		e.swarmConfig = config
	}
}

// Run executes the graph sequentially, node by node
// Each node runs until its acceptance criteria are met (FR3)
// Forward-only progress: graph state is source of truth (FR1, FR52)
func (e *Executor) Run(ctx context.Context) error {
	for i := range e.graph.Nodes {
		node := e.graph.Nodes[i]

		// Skip non-pending nodes (idempotent)
		if node.Status != NodeStatusPending && node.Status != NodeStatusFailed {
			continue
		}

		// Mark node as running
		e.updateNodeStatus(ctx, node.ID, NodeStatusRunning, 0.0)

		// Build context from upstream node outputs (FR18)
		// Use ContextAssembler if available, otherwise fall back to buildContext
		var contextStr string
		if e.contextAssembler != nil {
			assembled, err := e.contextAssembler.AssembleContext(node.ID, 2000)
			if err == nil && assembled != "" {
				contextStr = assembled
			} else {
				contextStr = e.buildContext(ctx, i)
			}
		} else {
			contextStr = e.buildContext(ctx, i)
		}

		// Retry loop until acceptance criteria are met
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			// Check for context cancellation
			if ctx.Err() != nil {
				e.updateNodeStatus(ctx, node.ID, NodeStatusFailed, 0.0)
				return fmt.Errorf("node %s cancelled: %w", node.ID, ctx.Err())
			}

			// Execute node via LLM (speculative or sequential)
			output, err := e.executeNode(ctx, node, contextStr)
			if err != nil {
				if attempt == maxRetries-1 {
					e.updateNodeStatus(ctx, node.ID, NodeStatusFailed, 0.0)
					return fmt.Errorf("node %s failed after %d attempts: %w", node.ID, maxRetries, err)
				}
				continue
			}

			// Check acceptance criteria
			if e.checkAcceptanceCriteria(node, output) {
				// Node complete - store output for downstream nodes (FR18)
				if e.store != nil {
					e.store.SaveNodeOutput(ctx, e.graph.ID, node.ID, output)
				}

				e.updateNodeStatus(ctx, node.ID, NodeStatusComplete, 1.0)

				// Auto-spec generation (AC3 - FR19)
				if e.specGen != nil {
					_, specErr := e.specGen.GenerateSpec(node.ID, output)
					if specErr == nil {
						e.specGen.AddSystemReferences(node.ID, []string{string(node.Type)})
					}
				}
				break
			}

			// Criteria not met, retry
			if attempt == maxRetries-1 {
				e.updateNodeStatus(ctx, node.ID, NodeStatusFailed, 0.0)
				return fmt.Errorf("node %s failed: acceptance criteria not met after %d attempts", node.ID, maxRetries)
			}
		}

		// Forward-only: graph state is source of truth (FR1, FR52)
		// Only advance when verified - no backwards jumps
	}

	return nil
}

// executeNode runs a single node via LLM (with optional speculative execution)
func (e *Executor) executeNode(ctx context.Context, node *Node, contextStr string) (string, error) {
	// Reset monologue buffer for this node
	e.currentNodeID = node.ID
	e.monologueMessages = nil

	if e.llmProv == nil {
		// Simulate execution for testing
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(500 * time.Millisecond):
			return fmt.Sprintf("Simulated output for node %s", node.ID), nil
		}
	}

	prompt := fmt.Sprintf(`Execute node of type "%s" with label "%s".

Context from upstream nodes:
%s

Acceptance Criteria:
%v

Provide the output.`, node.Type, node.Label, contextStr, node.AcceptanceCriteria)

	messages := []llm.Message{
		{Role: "system", Content: "You are a node executor. Execute the node and provide output."},
		{Role: "user", Content: prompt},
	}

	// Check if speculative execution is enabled for this node type
	useSpeculative := e.swarmConfig != nil && e.swarmConfig.Enabled && e.swarm != nil

	if useSpeculative {
		// Broadcast speculative execution start
		e.broadcastSpeculativeStart(node.ID)

		// Run speculative execution: multiple parallel attempts, best result wins
		result, err := e.swarm.Execute(ctx, node.ID, messages, node.AcceptanceCriteria)
		if err != nil {
			return "", fmt.Errorf("speculative execution failed: %w", err)
		}

		// Stream the winning result to WebSocket
		e.streamLLMResponse(ctx, result.Output)

		return result.Output, nil
	}

	// Fall back to single execution (original behavior)
	ch, err := e.llmProv.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM chat failed: %w", err)
	}

	var output strings.Builder
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case token, ok := <-ch:
			if !ok {
				if output.Len() > 0 {
					return output.String(), nil
				}
				return "", fmt.Errorf("no response from LLM")
			}
			output.WriteString(token)
			// Stream to WebSocket
			e.streamLLMResponse(ctx, token)
		}
	}
}

// broadcastSpeculativeStart sends WebSocket update for speculative execution start
func (e *Executor) broadcastSpeculativeStart(nodeID string) {
	if e.hub == nil {
		return
	}
	maxAttempts := 1
	if e.swarmConfig != nil {
		maxAttempts = e.swarmConfig.MaxAttempts
	}
	data, _ := json.Marshal(map[string]interface{}{
		"type":        "node_update",
		"nodeId":      nodeID,
		"status":      "running",
		"speculative": true,
		"attempts":    maxAttempts,
	})
	e.hub.BroadcastRaw(data)
}

// buildContext builds context string from upstream node outputs (FR18)
func (e *Executor) buildContext(ctx context.Context, currentIdx int) string {
	var contextParts []string
	for i := 0; i < currentIdx && i < len(e.graph.Nodes); i++ {
		node := e.graph.Nodes[i]
		if node.Status == NodeStatusComplete && e.store != nil {
			output, err := e.store.GetNodeOutput(ctx, e.graph.ID, node.ID)
			if err == nil && output != "" {
				contextParts = append(contextParts, fmt.Sprintf("Node %s (%s): %s", node.ID, node.Type, output))
			}
		}
	}
	return strings.Join(contextParts, "\n")
}

// checkAcceptanceCriteria verifies if node output meets acceptance criteria
func (e *Executor) checkAcceptanceCriteria(node *Node, output string) bool {
	// Basic validation: output must be non-empty and longer than 10 chars
	if len(output) < 10 {
		return false
	}

	// If no acceptance criteria defined, accept any reasonable output
	if len(node.AcceptanceCriteria) == 0 {
		return true
	}

	// Check if output appears to address the acceptance criteria
	for _, criterion := range node.AcceptanceCriteria {
		if len(criterion) > 0 && !containsIgnoreCase(output, criterion) {
			// Output does not address this criterion
			return false
		}
	}

	return true
}

// containsIgnoreCase checks if substr appears in s (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	if substr == "" {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if stringEqualIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func stringEqualIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if lower(a[i]) != lower(b[i]) {
			return false
		}
	}
	return true
}

func lower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// streamLLMResponse broadcasts LLM token via WebSocket hub
func (e *Executor) streamLLMResponse(ctx context.Context, token string) {
	if e.hub == nil || token == "" {
		return
	}
	// Accumulate for persistence
	e.monologueMessages = append(e.monologueMessages, nfContext.MonologueMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Text:       token,
		Timestamp:  time.Now().UnixMilli(),
	})
	data, _ := json.Marshal(map[string]interface{}{
		"type": "llm_chunk",
		"text": token,
	})
	e.hub.BroadcastRaw(data)
}

// updateNodeStatus updates node status and broadcasts via WebSocket (Task 5.4)
func (e *Executor) updateNodeStatus(ctx context.Context, nodeID string, status NodeStatus, progress float64) {
	// Update graph nodes
	for _, node := range e.graph.Nodes {
		if node.ID == nodeID {
			node.Status = status
			node.Progress = progress
			break
		}
	}

	// Broadcast via WebSocket
	if e.hub != nil {
		e.hub.BroadcastNodeUpdate(nodeID, string(status), progress)

		// Also broadcast edge updates for connected edges
		for _, edge := range e.graph.Edges {
			if edge.Target == nodeID {
				tension := 0.0
				switch status {
				case NodeStatusRunning:
					tension = 0.5
				case NodeStatusComplete:
					tension = 0.0
				case NodeStatusFailed:
					tension = 1.0
				}
				e.hub.BroadcastEdgeUpdate(edge.Source, edge.Target, tension)
			}
		}
	}
	// Save monologue on node complete
	if status == NodeStatusComplete && e.store != nil && len(e.monologueMessages) > 0 {
		if err := e.store.SaveMonologueHistory(ctx, e.graph.ID, e.monologueMessages); err != nil {
			// Log but don't fail
			_ = err
		}
	}
}
