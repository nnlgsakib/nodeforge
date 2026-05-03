package nforge

import (
	"fmt"

	"github.com/nnlgsakib/nodeforge/internal/engine"
	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Visualize node graph",
}

var graphVizCmd = &cobra.Command{
	Use:   "viz [spec-file]",
	Short: "Display ASCII art representation of a graph",
	Long: `Render a graph as ASCII art in the terminal.

If a spec file is provided, it will be parsed and rendered.
Otherwise, displays an empty graph placeholder.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGraphViz,
}

func init() {
	graphVizCmd.Flags().Bool("verbose", false, "Show multi-line layout with edge details")
	graphCmd.AddCommand(graphVizCmd)
	rootCmd.AddCommand(graphCmd)
}

func runGraphViz(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	if len(args) == 0 {
		fmt.Println("(no spec file provided — usage: nforge graph viz <spec-file>)")
		return nil
	}

	specPath := args[0]

	// Parse spec file
	spec, err := engine.ParseSpecFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to parse spec file: %w", err)
	}

	// Convert spec to graph
	graph, err := spec.ToGraph()
	if err != nil {
		if spec.Mode == engine.SpecModeGoal {
			return fmt.Errorf("goal-mode spec requires LLM to generate graph; run 'nforge run %s' instead", specPath)
		}
		return fmt.Errorf("failed to convert spec to graph: %w", err)
	}

	// Render as ASCII
	renderer := engine.NewASCIIRenderer(graph)
	if verbose {
		fmt.Println(renderer.RenderVerbose())
	} else {
		fmt.Println(renderer.Render())
	}

	return nil
}
