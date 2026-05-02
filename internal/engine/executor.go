package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/nnlgsakib/nodeforge/internal/llm"
	nfContext "github.com/nnlgsakib/nodeforge/internal/context"
)

// NodeUpdateBroadcaster abstracts the WebSocket hub for broadcasting node updates
type NodeUpdateBroadcaster interface {
	BroadcastNodeUpdate(nodeID, status string, progress float64)
}

// Executor runs nodes sequentially with retry until acceptance criteria are met
type Executor struct {
	graph     *Graph
	llmProv   llm.Provider
	store     *nfContext.Store
	hub       NodeUpdateBroadcaster // WebSocket hub for real-time updates
}

// NewExecutor creates a new executor for the given graph
func NewExecutor(graph *Graph, llmProv llm.Provider, store *nfContext.Store) *Executor {
	return &Executor{
		graph:   graph,
		llmProv: llmProv,
		store:   store,
	}
}

// SetHub sets the WebSocket hub for real-time updates (Task 5.4)
func (e *Executor) SetHub(hub NodeUpdateBroadcaster) {
	e.hub = hub
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
		e.updateNodeStatus(node.ID, NodeStatusRunning, 0.0)

		// Build context from upstream node outputs (FR18)
		contextStr := e.buildContext(i)

		// Retry loop until acceptance criteria are met
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			// Execute node via LLM
			output, err := e.executeNode(ctx, node, contextStr)
			if err != nil {
				if attempt == maxRetries-1 {
					e.updateNodeStatus(node.ID, NodeStatusFailed, 0.0)
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

				e.updateNodeStatus(node.ID, NodeStatusComplete, 1.0)
				break
			}

			// Criteria not met, retry
			if attempt == maxRetries-1 {
				e.updateNodeStatus(node.ID, NodeStatusFailed, 0.0)
				return fmt.Errorf("node %s failed: acceptance criteria not met after %d attempts", node.ID, maxRetries)
			}
		}

		// Forward-only: graph state is source of truth (FR1, FR52)
		// Only advance when verified - no backwards jumps
	}

	return nil
}

// executeNode runs a single node via LLM
func (e *Executor) executeNode(ctx context.Context, node *Node, contextStr string) (string, error) {
	if e.llmProv == nil {
		// Simulate execution for testing
		time.Sleep(500 * time.Millisecond)
		return fmt.Sprintf("Simulated output for node %s", node.ID), nil
	}

	prompt := fmt.Sprintf(`Execute node of type "%s" with label "%s".

Context from upstream nodes:
%s

Acceptance Criteria:
%v

Provide the output.`, node.Type, node.Label, contextStr, node.AcceptanceCriteria)

	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a node executor. Execute the node and provide output."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:  2000,
	}

	resp, err := e.llmProv.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}

// buildContext builds context string from upstream node outputs (FR18)
func (e *Executor) buildContext(currentIdx int) string {
	var contextParts []string
	for i := 0; i < currentIdx && i < len(e.graph.Nodes); i++ {
		node := e.graph.Nodes[i]
		if node.Status == NodeStatusComplete && e.store != nil {
			output, err := e.store.GetNodeOutput(context.Background(), e.graph.ID, node.ID)
			if err == nil && output != "" {
				contextParts = append(contextParts, fmt.Sprintf("Node %s (%s): %s", node.ID, node.Type, output))
			}
		}
	}
	return fmt.Sprintf("%v", contextParts)
}

// checkAcceptanceCriteria verifies if node output meets acceptance criteria
func (e *Executor) checkAcceptanceCriteria(node *Node, output string) bool {
	// Simplified: check if output is non-empty and longer than 10 chars
	if len(output) < 10 {
		return false
	}

	// In production, this would use LLM to verify against acceptance criteria
	// For now, non-empty output with reasonable length is considered success
	return true
}

// updateNodeStatus updates node status and broadcasts via WebSocket (Task 5.4)
func (e *Executor) updateNodeStatus(nodeID string, status NodeStatus, progress float64) {
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
	}
}
