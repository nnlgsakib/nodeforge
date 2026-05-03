package nforge

import (
	"errors"
	"fmt"
	"time"

	"github.com/nnlgsakib/nodeforge/internal/engine"
	"github.com/nnlgsakib/nodeforge/internal/llm"
	"github.com/spf13/cobra"
)

// exitCodeError carries an exit code for the headless CLI
type exitCodeError struct {
	Code int
	Err  error
}

func (e *exitCodeError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

func (e *exitCodeError) ExitCode() int {
	return e.Code
}

// ExitCodeForError returns the exit code for an error, defaulting to 1
func ExitCodeForError(err error) int {
	if err == nil {
		return 0
	}
	var ec *exitCodeError
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

var runCmd = &cobra.Command{
	Use:   "run [spec-file]",
	Short: "Run a spec file in headless mode",
	Long: `Execute a graph from a YAML spec file without the browser UI.

Two spec file modes:
  Goal mode:  spec file contains only 'goal: <string>'
  Graph mode: spec file contains 'nodes:' and 'edges:' lists

Exit codes:
  0 - All nodes completed successfully (green)
  1 - One or more nodes failed (red)
  2 - Usage error (missing file, invalid format)`,
	Args: cobra.ExactArgs(1),
	RunE: runSpecFile,
}

func init() {
	runCmd.Flags().Bool("ascii", false, "Display ASCII graph during execution")
	runCmd.Flags().Bool("no-llm", false, "Run in simulation mode without LLM provider (for testing)")
	rootCmd.AddCommand(runCmd)
}

func runSpecFile(cmd *cobra.Command, args []string) error {
	specPath := args[0]

	// Parse spec file
	spec, err := engine.ParseSpecFile(specPath)
	if err != nil {
		return &exitCodeError{Code: 2, Err: fmt.Errorf("failed to parse spec file: %w", err)}
	}

	var graph *engine.Graph

	switch spec.Mode {
	case engine.SpecModeGoal:
		// Goal mode: use Generator to auto-generate graph
		noLLM, _ := cmd.Flags().GetBool("no-llm")
		var gen *engine.Generator
		if noLLM {
			gen = engine.NewGenerator(nil, nil)
		} else {
			provider := initLLMProvider()
			gen = engine.NewGenerator(provider, nil)
		}
		graph, err = gen.Generate(cmd.Context(), spec.Goal)
		if err != nil {
			return &exitCodeError{Code: 2, Err: fmt.Errorf("failed to generate graph from goal: %w", err)}
		}
	case engine.SpecModeGraph:
		// Graph mode: convert spec to graph
		graph, err = spec.ToGraph()
		if err != nil {
			return &exitCodeError{Code: 2, Err: fmt.Errorf("failed to convert spec to graph: %w", err)}
		}
	default:
		return &exitCodeError{Code: 2, Err: fmt.Errorf("unknown spec mode: %q", spec.Mode)}
	}

	// Show ASCII graph before execution if flag is set
	showAscii, _ := cmd.Flags().GetBool("ascii")
	if showAscii {
		renderer := engine.NewASCIIRenderer(graph)
		fmt.Println(renderer.RenderVerbose())
		fmt.Println("---")
	}

	// Execute the graph (headless: no WebSocket hub)
	noLLM, _ := cmd.Flags().GetBool("no-llm")
	var provider llm.LLMProvider
	if !noLLM {
		provider = initLLMProvider()
	}
	exec := engine.NewExecutor(graph, provider, nil, nil, nil)

	fmt.Printf("Executing graph: %s (%d nodes)\n", graph.ID, len(graph.Nodes))

	err = exec.Run(cmd.Context())
	if err != nil {
		return &exitCodeError{Code: 1, Err: fmt.Errorf("execution failed: %w", err)}
	}

	// Show final ASCII graph after execution if flag is set
	if showAscii {
		fmt.Println()
		renderer := engine.NewASCIIRenderer(graph)
		fmt.Println(renderer.Render())
	}

	// Check if any node failed
	for _, node := range graph.Nodes {
		if node.Status == engine.NodeStatusFailed {
			return &exitCodeError{Code: 1, Err: fmt.Errorf("node %q (%s) failed", node.Label, node.ID)}
		}
	}

	fmt.Println("All nodes completed successfully.")
	return nil
}

// initLLMProvider initializes the first available LLM provider for headless execution
func initLLMProvider() llm.LLMProvider {
	ollamaCfg := &llm.ProviderConfig{
		Type:     llm.ProviderOllama,
		BaseURL:  "http://localhost:11434",
		Model:    "llama3",
		Timeout:  30 * time.Second,
	}
	if provider, err := llm.NewProvider(ollamaCfg); err == nil {
		return provider
	}
	// Fall back to nil (simulation mode) if no provider is available
	return nil
}
