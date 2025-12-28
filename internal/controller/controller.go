package controller

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/samber/lo"
)

type Config struct {
	ProjectName  string
	ComposePath  string
	StatePath    string
	ReconcileCfg reconcile.Config
}

type Controller struct {
	watcher    *git.Watcher
	client     *docker.Client
	reconciler *reconcile.Reconciler
	store      *state.Store
	config     Config
	logger     *slog.Logger
}

func New(ctx context.Context, watcher *git.Watcher, cfg Config, logger *slog.Logger) (*Controller, error) {
	if filepath.IsAbs(cfg.ComposePath) {
		return nil, fmt.Errorf("compose path must be relative: %s", cfg.ComposePath)
	}
	if strings.Contains(cfg.ComposePath, "..") {
		return nil, fmt.Errorf("compose path must not contain '..': %s", cfg.ComposePath)
	}

	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("component", "controller")

	client, err := docker.NewClient(cfg.ProjectName, logger)
	if err != nil {
		return nil, err
	}

	store, err := state.New(ctx, cfg.StatePath)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	reconciler := reconcile.New(client, nil, cfg.ReconcileCfg, logger)

	return &Controller{
		watcher:    watcher,
		client:     client,
		reconciler: reconciler,
		store:      store,
		config:     cfg,
		logger:     logger,
	}, nil
}

func (c *Controller) Run(ctx context.Context) error {
	if err := c.watcher.Clone(ctx); err != nil {
		return err
	}

	if err := c.loadAndReconcile(ctx, c.watcher.LastCommit()); err != nil {
		return fmt.Errorf("initial reconcile: %w", err)
	}

	c.watcher.Watch(ctx, func(event git.ChangeEvent) {
		c.handleChange(ctx, event)
	})

	return nil
}

func (c *Controller) handleChange(ctx context.Context, event git.ChangeEvent) {
	c.logger.Info("git change detected", "commit", lo.Substring(event.Commit, 0, 8), "message", event.Message)

	if err := c.loadAndReconcile(ctx, event.Commit); err != nil {
		c.logger.Error("reconcile failed", "error", err)
	}
}

func (c *Controller) loadAndReconcile(ctx context.Context, commit string) error {
	composePath := filepath.Join(c.watcher.WorkDir(), c.config.ComposePath)

	root, err := os.OpenRoot(c.watcher.WorkDir())
	if err != nil {
		return fmt.Errorf("open work directory: %w", err)
	}
	defer root.Close()

	composeContent, err := root.ReadFile(c.config.ComposePath)
	if err != nil {
		return fmt.Errorf("read compose file: %w", err)
	}

	project, err := docker.LoadProject(ctx, composePath, c.config.ProjectName)
	if err != nil {
		return err
	}

	c.reconciler.SetProject(project)
	c.reconciler.SetCommit(commit)

	deployment, err := c.store.SaveDeployment(ctx, commit, string(composeContent), state.StatusPending, "")
	if err != nil {
		c.logger.Warn("failed to save deployment", "error", err)
	}

	result := c.reconciler.Reconcile(ctx)

	if deployment != nil {
		var status state.DeploymentStatus
		var message string
		switch {
		case result.Error != nil:
			status, message = state.StatusFailed, result.Error.Error()
		case result.Reconciled:
			status = state.StatusSuccess
		default:
			status, message = state.StatusSkipped, "no changes applied"
		}
		if err := c.store.UpdateDeploymentStatus(ctx, deployment.ID, status, message); err != nil {
			c.logger.Warn("failed to update deployment status", "error", err)
		}
	}

	return result.Error
}

func (c *Controller) Sync(ctx context.Context) error {
	result := c.reconciler.Sync(ctx)
	return result.Error
}

func (c *Controller) Close() error {
	var err error
	if c.store != nil {
		err = c.store.Close()
	}
	if c.client != nil {
		_ = c.client.Close()
	}
	return err
}
