package nforge

import (
	"fmt"

	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Visualize node graph",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("graph: not yet implemented (story 1.7)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
}
