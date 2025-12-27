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

var rollbackFlags struct {
	projectName string
	statePath   string
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback <commit>",
	Short: "Rollback to a previous deployment",
	Long:  `Rollback to a previously deployed commit by redeploying its compose configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRollback,
}

func init() {
	rollbackCmd.Flags().StringVar(&rollbackFlags.projectName, "project", "kedge", "Docker compose project name")
	rollbackCmd.Flags().StringVar(&rollbackFlags.statePath, "state", "/var/lib/kedge/state.db", "Path to state database")

	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commitPrefix := args[0]

	store, err := state.New(ctx, rollbackFlags.statePath)
	if err != nil {
		return err
	}
	defer store.Close()

	deployment, err := findDeployment(ctx, store, commitPrefix)
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

	project, err := docker.LoadProject(ctx, composePath, rollbackFlags.projectName)
	if err != nil {
		return fmt.Errorf("load compose: %w", err)
	}

	client, err := docker.NewClient(rollbackFlags.projectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	fmt.Printf("Rolling back to commit %s...\n", lo.Substring(deployment.CommitHash, 0, 8))

	if err := client.Deploy(ctx, project, deployment.CommitHash); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	_, err = store.SaveDeployment(ctx, deployment.CommitHash, deployment.ComposeContent, state.StatusRolledBack, "rollback")
	if err != nil {
		logger.Warn("failed to record rollback", "error", err)
	}

	fmt.Println("Rollback completed successfully")
	return nil
}

func findDeployment(ctx context.Context, store *state.Store, prefix string) (*state.Deployment, error) {
	deployment, err := store.GetDeploymentByCommit(ctx, prefix)
	if err == nil {
		return deployment, nil
	}
	if err != state.ErrNotFound {
		return nil, err
	}

	deployments, err := store.ListDeployments(ctx, 100)
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
