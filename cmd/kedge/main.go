package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		repoURL      = flag.String("repo", "", "git repository URL")
		branch       = flag.String("branch", "main", "git branch")
		workDir      = flag.String("workdir", "/var/lib/kedge/repo", "working directory for git clone")
		composePath  = flag.String("compose", "docker-compose.yaml", "path to compose file relative to repo root")
		projectName  = flag.String("project", "kedge", "docker compose project name")
		statePath    = flag.String("state", "/var/lib/kedge/state.db", "path to state database")
		pollInterval = flag.Duration("poll", time.Minute, "git poll interval")
		mode         = flag.String("mode", "auto", "reconcile mode: auto, notify, manual")
	)
	flag.Parse()

	if *repoURL == "" {
		slog.Error("repo URL is required")
		return 1
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	watcher := git.NewWatcher(*repoURL, *branch, *workDir, *pollInterval)

	cfg := controller.Config{
		ProjectName:  *projectName,
		ComposePath:  *composePath,
		StatePath:    *statePath,
		ReconcileCfg: reconcile.Config{Mode: reconcile.Mode(*mode)},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ctrl, err := controller.New(ctx, watcher, cfg, logger)
	if err != nil {
		slog.Error("failed to create controller", "error", err)
		return 1
	}
	defer ctrl.Close()

	slog.Info("starting kedge", "repo", *repoURL, "branch", *branch, "mode", *mode)

	if err := ctrl.Run(ctx); err != nil {
		slog.Error("controller error", "error", err)
		return 1
	}

	return 0
}
