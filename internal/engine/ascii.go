package engine

import (
	"fmt"
	"strings"
	"unicode"
)

// ASCIIRenderer renders a Graph as ASCII art for terminal output
type ASCIIRenderer struct {
	graph *Graph
}

// NewASCIIRenderer creates a new ASCII renderer for the given graph
func NewASCIIRenderer(graph *Graph) *ASCIIRenderer {
	return &ASCIIRenderer{graph: graph}
}

// Render produces a left-to-right ASCII representation of the graph
func (r *ASCIIRenderer) Render() string {
	if r.graph == nil || len(r.graph.Nodes) == 0 {
		return "(empty graph)"
	}

	// Render nodes with status suffix
	nodeLines := make([]string, len(r.graph.Nodes))
	for i, node := range r.graph.Nodes {
		statusSuffix := statusSuffix(node.Status)
		nodeLines[i] = fmt.Sprintf("[%s: %s] (%s)", titleCase(string(node.Type)), node.Label, statusSuffix)
	}

	// Build the output: nodes connected by arrows
	// Use topological order (nodes are already in order from graph generation)
	if len(nodeLines) == 1 {
		return nodeLines[0]
	}

	var parts []string
	for _, line := range nodeLines {
		parts = append(parts, line)
	}

	return strings.Join(parts, " \u2192 ")
}

// RenderVerbose produces a multi-line ASCII representation with edges
func (r *ASCIIRenderer) RenderVerbose() string {
	if r.graph == nil || len(r.graph.Nodes) == 0 {
		return "(empty graph)"
	}

	var sb strings.Builder

	// Render nodes
	for i, node := range r.graph.Nodes {
		statusSuffix := statusSuffix(node.Status)
		indent := "  "
		if i > 0 {
			sb.WriteString("       |\n")
			sb.WriteString("       v\n")
		}
		sb.WriteString(fmt.Sprintf("%s[%s: %s] (%s)\n", indent, titleCase(string(node.Type)), node.Label, statusSuffix))
	}

	// Render edges separately
	if len(r.graph.Edges) > 0 {
		sb.WriteString("\nEdges:\n")
		for _, edge := range r.graph.Edges {
			sb.WriteString(fmt.Sprintf("  %s \u2192 %s\n", edge.Source, edge.Target))
		}
	}

	return sb.String()
}

// statusSuffix returns a single-letter status suffix for ASCII rendering
// with ANSI color codes for terminal display
func statusSuffix(status NodeStatus) string {
	const (
		green  = "\033[32m"
		yellow = "\033[33m"
		red    = "\033[31m"
		gray   = "\033[90m"
		reset  = "\033[0m"
	)
	switch status {
	case NodeStatusComplete:
		return green + "G" + reset // green
	case NodeStatusRunning:
		return yellow + "Y" + reset // yellow
	case NodeStatusFailed:
		return red + "R" + reset // red
	case NodeStatusPending:
		return gray + "P" + reset // gray/pending
	default:
		return gray + "S" + reset // gray/skipped
	}
}

// titleCase capitalizes the first letter of a string
func titleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
