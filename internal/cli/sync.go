package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/spf13/cobra"
)

var syncFlags struct {
	projectName string
	composePath string
	workdir     string
	statePath   string
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
	syncCmd.Flags().StringVar(&syncFlags.composePath, "compose", "docker-compose.yaml", "Path to compose file relative to workdir")
	syncCmd.Flags().StringVar(&syncFlags.workdir, "workdir", ".kedge/repo", "Working directory containing the compose file")
	syncCmd.Flags().StringVar(&syncFlags.statePath, "state", ".kedge/state.db", "Path to state database")
	syncCmd.Flags().BoolVar(&syncFlags.force, "force", false, "Force sync even if no drift detected")

	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := controller.Config{
		ProjectName:  syncFlags.projectName,
		ComposePath:  syncFlags.composePath,
		WorkDir:      syncFlags.workdir,
		StatePath:    syncFlags.statePath,
		ReconcileCfg: reconcile.Config{Mode: reconcile.ModeAuto},
	}

	ctrl, err := controller.NewStandalone(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer ctrl.Close()

	var result *reconcile.Result
	if syncFlags.force {
		result, err = ctrl.Sync(ctx)
	} else {
		result, err = ctrl.Reconcile(ctx)
	}
	if err != nil {
		return err
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
