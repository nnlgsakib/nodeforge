package nforge

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a node type or spec file",
	Args:  cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Return node types with descriptions
		completions := make([]string, 0, len(NodeTypes))
		for _, nt := range NodeTypes {
			if desc, ok := NodeTypeDescriptions[nt]; ok {
				completions = append(completions, cobra.CompletionWithDesc(nt, desc))
			} else {
				completions = append(completions, nt)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("run: not yet implemented (story 1.8)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
