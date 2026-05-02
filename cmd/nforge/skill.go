package nforge

import (
	"fmt"

	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("skill: not yet implemented (story 1.5/skill system)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
}
