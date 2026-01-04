package cli

import (
	"context"
	"fmt"

	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/spf13/cobra"
)

var syncFlags struct {
	force bool
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Trigger immediate reconciliation",
	Long:  `Force an immediate sync of the compose file to running containers.`,
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncFlags.force, "force", false, "Force sync even if no drift detected")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	if repo == nil {
		return fmt.Errorf("--repo is required")
	}

	ctx := context.Background()

	ctrlCfg := controller.Config{
		RepoName:     repo.Name,
		ProjectName:  cfg.Docker.ProjectName,
		ComposePath:  cfg.Docker.ComposeFile,
		WorkDir:      repoWorkDir(repo.Name),
		StatePath:    cfg.State.Path,
		ReconcileCfg: reconcile.Config{Mode: reconcile.ModeAuto},
	}

	ctrl, err := controller.NewStandalone(ctx, ctrlCfg, logger)
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
