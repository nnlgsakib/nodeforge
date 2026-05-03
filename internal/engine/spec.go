package engine

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// SpecMode represents the parsing mode for a spec file
type SpecMode string

const (
	SpecModeGoal  SpecMode = "goal"
	SpecModeGraph SpecMode = "graph"
)

// SpecFile represents a parsed YAML spec file
type SpecFile struct {
	Mode   SpecMode
	Goal   string     // Used in goal mode
	Nodes  []*SpecNode // Used in graph mode
	Edges  []*SpecEdge // Used in graph mode
}

// SpecNode represents a node in graph-mode spec
type SpecNode struct {
	ID     string `yaml:"id"`
	Type   string `yaml:"type"`
	Label  string `yaml:"label"`
}

// SpecEdge represents an edge in graph-mode spec
type SpecEdge struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// rawSpecFile is the intermediate YAML structure
type rawSpecFile struct {
	Goal  string               `yaml:"goal,omitempty"`
	Nodes []map[string]string  `yaml:"nodes,omitempty"`
	Edges []map[string]string  `yaml:"edges,omitempty"`
}

// ParseSpecFile reads and parses a YAML spec file from the given path
func ParseSpecFile(path string) (*SpecFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}
	return ParseSpec(data)
}

// ParseSpec parses raw YAML bytes into a SpecFile
func ParseSpec(data []byte) (*SpecFile, error) {
	var raw rawSpecFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	hasGoal := strings.TrimSpace(raw.Goal) != ""
	hasNodes := len(raw.Nodes) > 0

	spec := &SpecFile{}

	if hasGoal && !hasNodes {
		// Goal-only mode
		spec.Mode = SpecModeGoal
		spec.Goal = raw.Goal
	} else if hasNodes {
		// Graph mode
		spec.Mode = SpecModeGraph
		spec.Nodes = make([]*SpecNode, 0, len(raw.Nodes))
		seenIDs := make(map[string]bool, len(raw.Nodes))
		for _, m := range raw.Nodes {
			node := &SpecNode{
				ID:    m["id"],
				Type:  m["type"],
				Label: m["label"],
			}
			if node.ID == "" {
				return nil, fmt.Errorf("node missing required field 'id'")
			}
			if seenIDs[node.ID] {
				return nil, fmt.Errorf("duplicate node ID %q", node.ID)
			}
			seenIDs[node.ID] = true
			if node.Type == "" {
				return nil, fmt.Errorf("node %q missing required field 'type'", node.ID)
			}
			if node.Label == "" {
				return nil, fmt.Errorf("node %q missing required field 'label'", node.ID)
			}
			spec.Nodes = append(spec.Nodes, node)
		}

		spec.Edges = make([]*SpecEdge, 0, len(raw.Edges))
		for _, m := range raw.Edges {
			edge := &SpecEdge{
				Source: m["source"],
				Target: m["target"],
			}
			if edge.Source == "" {
				return nil, fmt.Errorf("edge missing required field 'source'")
			}
			if edge.Target == "" {
				return nil, fmt.Errorf("edge missing required field 'target'")
			}
			spec.Edges = append(spec.Edges, edge)
		}
	} else {
		return nil, fmt.Errorf("spec file must contain either a 'goal' field (goal mode) or 'nodes'/'edges' fields (graph mode)")
	}

	return spec, nil
}

// ToGraph converts a SpecFile to a Graph
func (s *SpecFile) ToGraph() (*Graph, error) {
	switch s.Mode {
	case SpecModeGoal:
		return nil, fmt.Errorf("goal-mode spec cannot be converted directly; use Generator.Generate()")
	case SpecModeGraph:
		graph := &Graph{
			ID:     generateID(),
			Goal:   "",
			Status: "pending",
		}
		nodeIDs := make(map[string]bool, len(s.Nodes))
		for _, sn := range s.Nodes {
			nodeIDs[sn.ID] = true
			graph.Nodes = append(graph.Nodes, &Node{
				ID:     sn.ID,
				Type:   NodeType(sn.Type),
				Label:  sn.Label,
				Status: NodeStatusPending,
			})
		}
		for i, se := range s.Edges {
			if !nodeIDs[se.Source] {
				return nil, fmt.Errorf("edge %d references non-existent node %q as source", i+1, se.Source)
			}
			if !nodeIDs[se.Target] {
				return nil, fmt.Errorf("edge %d references non-existent node %q as target", i+1, se.Target)
			}
			graph.Edges = append(graph.Edges, &Edge{
				ID:     fmt.Sprintf("edge-%d", i+1),
				Source: se.Source,
				Target: se.Target,
				Type:   "default",
			})
		}
		return graph, nil
	default:
		return nil, fmt.Errorf("unknown spec mode: %q", s.Mode)
	}
}
