package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/LoriKarikari/kedge/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GitOps controller",
	Long:  `Start watching the Git repository and automatically deploy changes.`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	if repo == nil {
		return fmt.Errorf("--repo is required")
	}

	repoURL := repo.URL
	branch := repo.Branch
	workDir := repoWorkDir(repo.Name)
	composePath := cfg.Docker.ComposeFile
	projectName := cfg.Docker.ProjectName
	statePath := cfg.State.Path
	pollInterval := cfg.Git.PollInterval
	modeStr := cfg.Reconciliation.Mode

	if err := os.MkdirAll(filepath.Dir(statePath), 0o750); err != nil {
		return err
	}

	mode, err := reconcile.ParseMode(modeStr)
	if err != nil {
		return fmt.Errorf("invalid mode %q: must be one of auto, notify, manual", modeStr)
	}

	watcher := git.NewWatcher(repoURL, branch, workDir, pollInterval, logger)

	ctrlCfg := controller.Config{
		RepoName:     repo.Name,
		ProjectName:  projectName,
		ComposePath:  composePath,
		StatePath:    statePath,
		ReconcileCfg: reconcile.Config{Mode: mode},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctrl, err := controller.New(ctx, watcher, ctrlCfg, logger)
	if err != nil {
		return err
	}
	defer ctrl.Close()

	srv := server.New(cfg.Server.Port, ctrl, logger)
	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown error", slog.Any("error", err))
		}
	}()
	logger.Info("server started", slog.Int("port", cfg.Server.Port))

	logger.Info("starting kedge", slog.String("repo", repoURL), slog.String("branch", branch), slog.String("mode", modeStr))

	return ctrl.Run(ctx)
}
