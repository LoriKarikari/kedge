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

var serveFlags struct {
	repoURL      string
	branch       string
	workDir      string
	composePath  string
	projectName  string
	statePath    string
	pollInterval time.Duration
	mode         string
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GitOps controller",
	Long:  `Start watching the Git repository and automatically deploy changes.`,
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVar(&serveFlags.repoURL, "repo", "", "Git repository URL")
	serveCmd.Flags().StringVar(&serveFlags.branch, "branch", "", "Git branch to watch")
	serveCmd.Flags().StringVar(&serveFlags.workDir, "workdir", "", "Working directory for git clone")
	serveCmd.Flags().StringVar(&serveFlags.composePath, "compose", "", "Path to compose file relative to repo root")
	serveCmd.Flags().StringVar(&serveFlags.projectName, "project", "", "Docker compose project name")
	serveCmd.Flags().StringVar(&serveFlags.statePath, "state", "", "Path to state database")
	serveCmd.Flags().DurationVar(&serveFlags.pollInterval, "poll", 0, "Git poll interval")
	serveCmd.Flags().StringVar(&serveFlags.mode, "mode", "", "Reconcile mode: auto, notify, manual")

	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	repoURL := coalesce(serveFlags.repoURL, cfg.Git.URL)
	branch := coalesce(serveFlags.branch, cfg.Git.Branch)
	workDir := coalesce(serveFlags.workDir, cfg.Git.WorkDir)
	composePath := coalesce(serveFlags.composePath, cfg.Docker.ComposeFile)
	projectName := coalesce(serveFlags.projectName, cfg.Docker.ProjectName)
	statePath := coalesce(serveFlags.statePath, cfg.State.Path)
	pollInterval := coalesceDuration(serveFlags.pollInterval, cfg.Git.PollInterval)
	modeStr := coalesce(serveFlags.mode, cfg.Reconciliation.Mode)

	if repoURL == "" {
		return fmt.Errorf("repo URL required: use --repo flag or set git.url in kedge.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(statePath), 0o750); err != nil {
		return err
	}

	mode, err := reconcile.ParseMode(modeStr)
	if err != nil {
		return fmt.Errorf("invalid mode %q: must be one of auto, notify, manual", modeStr)
	}

	watcher := git.NewWatcher(repoURL, branch, workDir, pollInterval, logger)

	ctrlCfg := controller.Config{
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

func coalesce(flag, config string) string {
	if flag != "" {
		return flag
	}
	return config
}

func coalesceDuration(flag, config time.Duration) time.Duration {
	if flag != 0 {
		return flag
	}
	return config
}
