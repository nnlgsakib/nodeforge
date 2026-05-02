package nforge

import (
	"context"
	"fmt"

	"github.com/nnlgsakib/nodeforge/internal/session"
	"github.com/spf13/cobra"
)

var newWorkspaceDir string

var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Create a new NodeForge project",
	Long:  "Creates a new project workspace with the specified name, initializing the .nforge/ directory structure",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNewProject(args[0], newWorkspaceDir)
	},
}

func init() {
	newCmd.Flags().StringVar(&newWorkspaceDir, "workspace-dir", ".", "Parent directory to create the project in (default: current directory)")
	rootCmd.AddCommand(newCmd)
}

func runNewProject(projectName, workspaceDir string) error {
	fmt.Printf("Creating project %q in %s\n", projectName, workspaceDir)

	mgr := session.NewManager(workspaceDir)
	ctx := context.Background()
	sess, err := mgr.CreateSessionWithName(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Project created successfully! Session ID: %s\n", sess.ID)
	return nil
}
