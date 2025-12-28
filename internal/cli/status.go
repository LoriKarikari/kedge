package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/spf13/cobra"
)

var statusFlags struct {
	projectName string
	composePath string
	statePath   string
	workdir     string
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current deployment status",
	Long:  `Display the current state of deployed services and last deployment info.`,
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&statusFlags.projectName, "project", "kedge", "Docker compose project name")
	statusCmd.Flags().StringVar(&statusFlags.composePath, "compose", "docker-compose.yaml", "Path to compose file relative to workdir")
	statusCmd.Flags().StringVar(&statusFlags.statePath, "state", ".kedge/state.db", "Path to state database")
	statusCmd.Flags().StringVar(&statusFlags.workdir, "workdir", ".kedge/repo", "Working directory containing the compose file")

	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	client, err := docker.NewClient(statusFlags.projectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	composePath := filepath.Join(statusFlags.workdir, statusFlags.composePath)
	project, err := docker.LoadProject(ctx, composePath, statusFlags.projectName)
	if err != nil {
		return fmt.Errorf("load compose: %w", err)
	}

	diff, err := client.Diff(ctx, project)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	fmt.Println("=== Service Status ===")
	if diff.InSync {
		fmt.Println("All services in sync ✓")
	} else {
		fmt.Printf("Drift detected: %s\n", diff.Summary)
		for _, change := range diff.Changes {
			fmt.Printf("  %s: %s (%s)\n", change.Service, change.Action, change.Reason)
		}
	}

	fmt.Println("\n=== Last Deployment ===")
	store, err := state.New(ctx, statusFlags.statePath)
	if err != nil {
		return err
	}
	defer store.Close()
	deployment, err := store.GetLastDeployment(ctx)
	switch {
	case err == state.ErrNotFound:
		fmt.Println("No deployments yet")
	case err != nil:
		return err
	default:
		fmt.Printf("Commit:  %s\n", deployment.CommitHash)
		fmt.Printf("Status:  %s\n", deployment.Status)
		fmt.Printf("Time:    %s\n", deployment.DeployedAt.Format("2006-01-02 15:04:05"))
		if deployment.Message != "" {
			fmt.Printf("Message: %s\n", deployment.Message)
		}
	}

	return nil
}
