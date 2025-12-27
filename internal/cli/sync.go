package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/spf13/cobra"
)

var syncFlags struct {
	projectName string
	composePath string
	force       bool
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Trigger immediate reconciliation",
	Long:  `Force an immediate sync of the compose file to running containers.`,
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().StringVar(&syncFlags.projectName, "project", "kedge", "Docker compose project name")
	syncCmd.Flags().StringVar(&syncFlags.composePath, "compose", "docker-compose.yaml", "Path to compose file")
	syncCmd.Flags().BoolVar(&syncFlags.force, "force", false, "Force sync even if no drift detected")

	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := docker.NewClient(syncFlags.projectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	project, err := docker.LoadProject(ctx, syncFlags.composePath, syncFlags.projectName)
	if err != nil {
		return fmt.Errorf("load compose: %w", err)
	}

	reconciler := reconcile.New(client, project, reconcile.Config{Mode: reconcile.ModeAuto}, logger)

	var result *reconcile.Result
	if syncFlags.force {
		result = reconciler.Sync(ctx)
	} else {
		result = reconciler.Reconcile(ctx)
	}

	if result.Error != nil {
		return result.Error
	}

	if result.Reconciled {
		fmt.Println("Sync completed successfully")
	} else {
		fmt.Println("No changes needed - already in sync")
	}

	return nil
}
