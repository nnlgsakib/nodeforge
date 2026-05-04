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
	Long:  "Create, list, resume, and manage NodeForge sessions",
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

var sessionResumeCmd = &cobra.Command{
	Use:   "resume <session-id>",
	Short: "Resume a session",
	Long:  "Restore a session from a previous shutdown and set its status back to running",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runResumeSession(sessionWorkspaceDir, args[0])
	},
}

var sessionForkCmd = &cobra.Command{
	Use:   "fork <session-id>",
	Short: "Fork a session",
	Long:  "Create a new session branch from an existing session, copying workspace state and Git history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForkSession(sessionWorkspaceDir, args[0])
	},
}

var sessionExportCmd = &cobra.Command{
	Use:   "export <session-id>",
	Short: "Export a session as a tarball",
	Long:  "Export a session as a self-contained tarball containing graph JSON, workspace files, and a README. API keys and secrets are excluded.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExportSession(sessionWorkspaceDir, args[0])
	},
}

var exportOutput string

func init() {
	sessionCmd.PersistentFlags().StringVar(&sessionWorkspaceDir, "workspace-dir", ".", "Workspace root directory")
	sessionExportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path for the tarball (default: <session-id>.tar.gz)")
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionResumeCmd)
	sessionCmd.AddCommand(sessionForkCmd)
	sessionCmd.AddCommand(sessionExportCmd)
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
	fmt.Fprintln(w, "SESSION ID\tPROJECT NAME\tSTATUS\tCREATED AT\tSIZE")
	fmt.Fprintln(w, "----------\t------------\t------\t----------\t----")

	for _, s := range sessions {
		created := s.CreatedAt.Format("2006-01-02 15:04:05")
		stats, err := mgr.GetSessionStatsFromSession(&s)
		size := "-"
		if err == nil && stats != nil {
			size = formatBytes(stats.WorkspaceSize)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.ID, s.Name, s.Status, created, size)
	}
	w.Flush()

	fmt.Printf("\nTotal: %d session(s)\n", len(sessions))
	return nil
}

// formatBytes returns a human-readable representation of bytes
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	if exp >= 6 {
		return fmt.Sprintf("%.1f EB", float64(b)/float64(div))
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func runResumeSession(workspaceDir string, id string) error {
	mgr, err := session.NewManager(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}
	defer mgr.Close()

	sess, err := mgr.ResumeSession(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to resume session: %w", err)
	}

	fmt.Printf("Session resumed successfully:\n")
	fmt.Printf("  ID:     %s\n", sess.ID)
	fmt.Printf("  Name:   %s\n", sess.Name)
	fmt.Printf("  Status: %s\n", sess.Status)
	fmt.Printf("  Goal:   %s\n", sess.Goal)
	return nil
}

func runForkSession(workspaceDir string, id string) error {
	mgr, err := session.NewManager(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}
	defer mgr.Close()

	sess, err := mgr.ForkSession(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to fork session: %w", err)
	}

	fmt.Printf("Session forked successfully:\n")
	fmt.Printf("  ID:     %s\n", sess.ID)
	fmt.Printf("  Name:   %s\n", sess.Name)
	fmt.Printf("  Parent: %s\n", id)
	fmt.Printf("  Status: %s\n", sess.Status)
	return nil
}

func runExportSession(workspaceDir string, id string) error {
	mgr, err := session.NewManager(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}
	defer mgr.Close()

	ctx := context.Background()

	// Warn if session is not complete
	sess, err := mgr.GetSession(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}
	if sess.Status != session.StatusComplete {
		fmt.Printf("Warning: session status is %q (expected complete). Exporting anyway.\n", sess.Status)
	}

	actualPath, err := session.ExportSession(ctx, mgr, id, exportOutput)
	if err != nil {
		return fmt.Errorf("failed to export session: %w", err)
	}

	fmt.Printf("Session exported successfully:\n")
	fmt.Printf("  Session: %s\n", id)
	fmt.Printf("  Output:  %s\n", actualPath)
	return nil
}
