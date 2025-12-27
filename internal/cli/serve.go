package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
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
	serveCmd.Flags().StringVar(&serveFlags.repoURL, "repo", "", "Git repository URL (required)")
	serveCmd.Flags().StringVar(&serveFlags.branch, "branch", "main", "Git branch to watch")
	serveCmd.Flags().StringVar(&serveFlags.workDir, "workdir", "/var/lib/kedge/repo", "Working directory for git clone")
	serveCmd.Flags().StringVar(&serveFlags.composePath, "compose", "docker-compose.yaml", "Path to compose file relative to repo root")
	serveCmd.Flags().StringVar(&serveFlags.projectName, "project", "kedge", "Docker compose project name")
	serveCmd.Flags().StringVar(&serveFlags.statePath, "state", "/var/lib/kedge/state.db", "Path to state database")
	serveCmd.Flags().DurationVar(&serveFlags.pollInterval, "poll", time.Minute, "Git poll interval")
	serveCmd.Flags().StringVar(&serveFlags.mode, "mode", "auto", "Reconcile mode: auto, notify, manual")

	_ = serveCmd.MarkFlagRequired("repo")

	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	watcher := git.NewWatcher(serveFlags.repoURL, serveFlags.branch, serveFlags.workDir, serveFlags.pollInterval)

	cfg := controller.Config{
		ProjectName:  serveFlags.projectName,
		ComposePath:  serveFlags.composePath,
		StatePath:    serveFlags.statePath,
		ReconcileCfg: reconcile.Config{Mode: reconcile.Mode(serveFlags.mode)},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctrl, err := controller.New(ctx, watcher, cfg, logger)
	if err != nil {
		return err
	}
	defer ctrl.Close()

	slog.Info("starting kedge", "repo", serveFlags.repoURL, "branch", serveFlags.branch, "mode", serveFlags.mode)

	return ctrl.Run(ctx)
}
