package nforge

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("session: not yet implemented (story 4.5)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}
