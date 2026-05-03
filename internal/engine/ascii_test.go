package engine

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stripAnsi removes ANSI escape codes for test assertions
var ansiRe = regexp.MustCompile("\033\\[[0-9;]*m")

func stripAnsi(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func TestASCIIRenderer_Render_Single(t *testing.T) {
	graph := &Graph{
		Nodes: []*Node{
			{ID: "n1", Type: NodeTypeGoal, Label: "Convert JS->Go", Status: NodeStatusComplete},
		},
	}

	renderer := NewASCIIRenderer(graph)
	output := stripAnsi(renderer.Render())

	assert.Contains(t, output, "[Goal: Convert JS->Go]")
	assert.Contains(t, output, "(G)")
}

func TestASCIIRenderer_Render_Multiple(t *testing.T) {
	graph := &Graph{
		Nodes: []*Node{
			{ID: "n1", Type: NodeTypeGoal, Label: "Convert", Status: NodeStatusComplete},
			{ID: "n2", Type: NodeTypeSpec, Label: "Spec", Status: NodeStatusRunning},
			{ID: "n3", Type: NodeTypePlan, Label: "Plan", Status: NodeStatusComplete},
		},
	}

	renderer := NewASCIIRenderer(graph)
	output := stripAnsi(renderer.Render())

	// All nodes should appear in output
	assert.Contains(t, output, "[Goal: Convert] (G)")
	assert.Contains(t, output, "[Spec: Spec] (Y)")
	assert.Contains(t, output, "[Plan: Plan] (G)")

	// Nodes connected by unicode arrows
	assert.Contains(t, output, "\u2192")
}

func TestASCIIRenderer_Render_Empty(t *testing.T) {
	graph := &Graph{}

	renderer := NewASCIIRenderer(graph)
	output := renderer.Render()

	assert.Equal(t, "(empty graph)", output)
}

func TestASCIIRenderer_Render_Nil(t *testing.T) {
	renderer := NewASCIIRenderer(nil)
	output := renderer.Render()

	assert.Equal(t, "(empty graph)", output)
}

func TestASCIIRenderer_RenderVerbose(t *testing.T) {
	graph := &Graph{
		Nodes: []*Node{
			{ID: "n1", Type: NodeTypeGoal, Label: "Convert", Status: NodeStatusComplete},
			{ID: "n2", Type: NodeTypeSpec, Label: "Spec", Status: NodeStatusFailed},
		},
		Edges: []*Edge{
			{ID: "e1", Source: "n1", Target: "n2"},
		},
	}

	renderer := NewASCIIRenderer(graph)
	output := stripAnsi(renderer.RenderVerbose())

	assert.Contains(t, output, "[Goal: Convert] (G)")
	assert.Contains(t, output, "[Spec: Spec] (R)")
	assert.Contains(t, output, "n1 \u2192 n2")
	assert.Contains(t, output, "Edges:")
}

func TestStatusSuffix(t *testing.T) {
	// Status suffixes now include ANSI color codes — verify they contain the right letter
	assert.Contains(t, stripAnsi(statusSuffix(NodeStatusComplete)), "G")
	assert.Contains(t, stripAnsi(statusSuffix(NodeStatusRunning)), "Y")
	assert.Contains(t, stripAnsi(statusSuffix(NodeStatusFailed)), "R")
	assert.Contains(t, stripAnsi(statusSuffix(NodeStatusPending)), "P")
	assert.Contains(t, stripAnsi(statusSuffix(NodeStatusSkipped)), "S")
}

func TestStatusSuffix_HasAnsiColors(t *testing.T) {
	// Verify ANSI codes are actually present
	g := statusSuffix(NodeStatusComplete)
	y := statusSuffix(NodeStatusRunning)
	r := statusSuffix(NodeStatusFailed)

	assert.Contains(t, g, "\033[32m") // green
	assert.Contains(t, y, "\033[33m") // yellow
	assert.Contains(t, r, "\033[31m") // red
}

func TestASCIIRenderer_OutputFormat(t *testing.T) {
	graph := &Graph{
		Nodes: []*Node{
			{ID: "goal-1", Type: NodeTypeGoal, Label: "Convert JS->Go", Status: NodeStatusComplete},
			{ID: "spec-1", Type: NodeTypeSpec, Label: "Generate Spec", Status: NodeStatusRunning},
			{ID: "plan-1", Type: NodeTypePlan, Label: "Create Plan", Status: NodeStatusComplete},
		},
	}

	renderer := NewASCIIRenderer(graph)
	output := stripAnsi(renderer.Render())

	// Verify the format matches the expected example from the story:
	// [Goal: Convert JS->Go] (G) \u2192 [Spec: Generate Spec] (Y) \u2192 [Plan: Create Plan] (G)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Len(t, lines, 1)

	parts := strings.Split(output, " \u2192 ")
	assert.Len(t, parts, 3)
	assert.Equal(t, "[Goal: Convert JS->Go] (G)", parts[0])
	assert.Equal(t, "[Spec: Generate Spec] (Y)", parts[1])
	assert.Equal(t, "[Plan: Create Plan] (G)", parts[2])
}
