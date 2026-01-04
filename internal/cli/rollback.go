package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback <commit>",
	Short: "Rollback to a previous deployment",
	Long:  `Rollback to a previously deployed commit by redeploying its compose configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRollback,
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	if repo == nil {
		return fmt.Errorf("--repo is required")
	}

	ctx := context.Background()
	commitPrefix := args[0]

	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	deployment, err := findDeployment(ctx, store, repo.Name, commitPrefix)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "kedge-rollback-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	composePath := filepath.Join(tmpDir, "docker-compose.yaml")
	if err := os.WriteFile(composePath, []byte(deployment.ComposeContent), 0o600); err != nil {
		return fmt.Errorf("write compose file: %w", err)
	}

	project, err := docker.LoadProject(ctx, composePath, cfg.Docker.ProjectName)
	if err != nil {
		return fmt.Errorf("load compose: %w", err)
	}

	client, err := docker.NewClient(cfg.Docker.ProjectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Printf("Rolling back to commit %s...\n", lo.Substring(deployment.CommitHash, 0, 8))

	if err := client.Deploy(ctx, project, deployment.CommitHash); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	_, err = store.SaveDeployment(ctx, repo.Name, deployment.CommitHash, deployment.ComposeContent, state.StatusRolledBack, "rollback")
	if err != nil {
		logger.Warn("failed to record rollback", slog.Any("error", err))
	}

	fmt.Println("Rollback completed successfully")
	return nil
}

func findDeployment(ctx context.Context, store *state.Store, repoName, prefix string) (*state.Deployment, error) {
	deployment, err := store.GetDeploymentByCommit(ctx, repoName, prefix)
	if err == nil {
		return deployment, nil
	}
	if err != state.ErrNotFound {
		return nil, err
	}

	deployments, err := store.ListDeployments(ctx, repoName, 100)
	if err != nil {
		return nil, err
	}

	for _, d := range deployments {
		if strings.HasPrefix(d.CommitHash, prefix) {
			return d, nil
		}
	}

	return nil, fmt.Errorf("no deployment found for commit %s", prefix)
}
