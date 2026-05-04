package nforge

import (
	"context"
	"fmt"

	"github.com/nnlgsakib/nodeforge/internal/session"
	"github.com/spf13/cobra"
)

var newWorkspaceDir string
var newGoal string

var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Create a new NodeForge project",
	Long:  "Creates a new project workspace with the specified name, initializing the .nforge/ directory structure",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNewProject(args[0], newWorkspaceDir, newGoal)
	},
}

func init() {
	newCmd.Flags().StringVar(&newWorkspaceDir, "workspace-dir", ".", "Parent directory to create the project in (default: current directory)")
	newCmd.Flags().StringVar(&newGoal, "goal", "", "Session goal description")
	rootCmd.AddCommand(newCmd)
}

func runNewProject(projectName, workspaceDir, goal string) error {
	fmt.Printf("Creating project %q in %s\n", projectName, workspaceDir)

	mgr, err := session.NewManager(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}
	defer mgr.Close()
	ctx := context.Background()
	sess, err := mgr.CreateSessionWithName(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Set goal if provided
	if goal != "" {
		sess.Goal = goal
		if err := mgr.UpdateSession(ctx, sess); err != nil {
			fmt.Printf("Warning: failed to save goal: %v\n", err)
		}
	}

	fmt.Printf("Project created successfully! Session ID: %s\n", sess.ID)
	return nil
}
