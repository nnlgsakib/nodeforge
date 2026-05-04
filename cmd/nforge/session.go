package nforge

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/nnlgsakib/nodeforge/internal/session"
	"github.com/spf13/cobra"
)

var sessionWorkspaceDir string

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long:  "Create, list, and manage NodeForge sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Usage()
		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	Long:  "Display all sessions with their IDs, names, status, and creation timestamps",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runListSessions(sessionWorkspaceDir)
	},
}

func init() {
	sessionCmd.PersistentFlags().StringVar(&sessionWorkspaceDir, "workspace-dir", ".", "Workspace root directory")
	sessionCmd.AddCommand(sessionListCmd)
	rootCmd.AddCommand(sessionCmd)
}

func runListSessions(workspaceDir string) error {
	mgr, err := session.NewManager(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}
	defer mgr.Close()

	sessions, err := mgr.ListSessions(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SESSION ID\tPROJECT NAME\tSTATUS\tCREATED AT\tWORKSPACE")
	fmt.Fprintln(w, "----------\t------------\t------\t----------\t---------")

	for _, s := range sessions {
		created := s.CreatedAt.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.ID, s.Name, s.Status, created, s.Workspace)
	}
	w.Flush()

	fmt.Printf("\nTotal: %d session(s)\n", len(sessions))
	return nil
}
