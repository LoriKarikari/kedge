package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current deployment status",
	Long:  `Display the current state of deployed services and last deployment info.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if repo == nil {
		return fmt.Errorf("--repo is required")
	}

	ctx := context.Background()

	client, err := docker.NewClient(cfg.Docker.ProjectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	composePath := filepath.Join(repoWorkDir(repo.Name), cfg.Docker.ComposeFile)
	project, err := docker.LoadProject(ctx, composePath, cfg.Docker.ProjectName)
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
	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()
	deployment, err := store.GetLastDeployment(ctx, repo.Name)
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
