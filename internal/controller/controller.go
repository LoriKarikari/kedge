package controller

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/LoriKarikari/kedge/internal/telemetry"
	"github.com/samber/lo"
)

type Config struct {
	RepoName     string
	ProjectName  string
	ComposePath  string
	WorkDir      string
	StatePath    string
	ReconcileCfg reconcile.Config
}

type Controller struct {
	watcher    *git.Watcher
	client     *docker.Client
	reconciler *reconcile.Reconciler
	store      *state.Store
	metrics    *telemetry.Metrics
	config     Config
	workDir    string
	logger     *slog.Logger
	ready      atomic.Bool
}

func New(ctx context.Context, watcher *git.Watcher, cfg Config, metrics *telemetry.Metrics, logger *slog.Logger) (*Controller, error) {
	ctrl, err := newController(ctx, cfg, metrics, logger)
	if err != nil {
		return nil, err
	}
	ctrl.watcher = watcher
	ctrl.workDir = watcher.WorkDir()
	return ctrl, nil
}

func NewStandalone(ctx context.Context, cfg Config, metrics *telemetry.Metrics, logger *slog.Logger) (*Controller, error) {
	if cfg.WorkDir == "" {
		return nil, fmt.Errorf("workdir is required")
	}
	ctrl, err := newController(ctx, cfg, metrics, logger)
	if err != nil {
		return nil, err
	}
	ctrl.workDir = cfg.WorkDir
	return ctrl, nil
}

func newController(ctx context.Context, cfg Config, metrics *telemetry.Metrics, logger *slog.Logger) (*Controller, error) {
	if filepath.IsAbs(cfg.ComposePath) {
		return nil, fmt.Errorf("compose path must be relative: %s", cfg.ComposePath)
	}

	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With(slog.String("component", "controller"))

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
		client:     client,
		reconciler: reconciler,
		store:      store,
		metrics:    metrics,
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

	c.ready.Store(true)

	go c.watchDrift(ctx)

	c.watcher.Watch(ctx, func(event git.ChangeEvent) {
		c.handleChange(ctx, event)
	})

	return nil
}

func (c *Controller) IsReady() bool {
	return c.ready.Load()
}

func (c *Controller) watchDrift(ctx context.Context) {
	results := make(chan *reconcile.Result)
	go c.reconciler.Watch(ctx, results)

	for {
		select {
		case <-ctx.Done():
			return
		case result := <-results:
			c.handleDriftResult(ctx, result)
		}
	}
}

func (c *Controller) handleDriftResult(ctx context.Context, result *reconcile.Result) {
	if result.Error != nil {
		c.logger.Error("drift check failed", slog.Any("error", result.Error))
		return
	}
	if result.Reconciled {
		c.logger.Info("drift reconciled", slog.Int("changes", len(result.Changes)))
		if c.metrics != nil {
			for _, change := range result.Changes {
				c.metrics.RecordDrift(ctx, c.config.RepoName, change.Service)
			}
		}
	}
}

func (c *Controller) handleChange(ctx context.Context, event git.ChangeEvent) {
	c.logger.Info("git change detected", slog.String("commit", lo.Substring(event.Commit, 0, 8)), slog.String("message", event.Message))

	if err := c.loadAndReconcile(ctx, event.Commit); err != nil {
		c.logger.Error("reconcile failed", slog.Any("error", err))
	}
}

func (c *Controller) loadAndReconcile(ctx context.Context, commit string) error {
	if err := c.loadProject(ctx, commit); err != nil {
		return err
	}

	root, err := os.OpenRoot(c.workDir)
	if err != nil {
		return fmt.Errorf("open work directory: %w", err)
	}
	defer root.Close()

	composeContent, err := root.ReadFile(c.config.ComposePath)
	if err != nil {
		return fmt.Errorf("read compose file: %w", err)
	}

	deployment, err := c.store.SaveDeployment(ctx, c.config.RepoName, commit, string(composeContent), state.StatusPending, "")
	if err != nil {
		c.logger.Warn("failed to save deployment", slog.Any("error", err))
	}

	start := time.Now()
	result := c.reconciler.Reconcile(ctx)
	duration := time.Since(start)

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

	if c.metrics != nil {
		c.metrics.RecordDeployment(ctx, c.config.RepoName, string(status))
		c.metrics.RecordReconciliation(ctx, c.config.RepoName, duration, result.Error == nil)
	}

	if deployment != nil {
		if err := c.store.UpdateDeploymentStatus(ctx, deployment.ID, status, message); err != nil {
			c.logger.Warn("failed to update deployment status", slog.Any("error", err))
		}
	}

	return result.Error
}

func (c *Controller) loadProject(ctx context.Context, commit string) error {
	composePath := filepath.Join(c.workDir, c.config.ComposePath)
	project, err := docker.LoadProject(ctx, composePath, c.config.ProjectName)
	if err != nil {
		return err
	}
	c.reconciler.SetProject(project)
	c.reconciler.SetCommit(commit)
	return nil
}

func (c *Controller) PullAndReconcile(ctx context.Context) error {
	changed, hash, err := c.watcher.Pull(ctx)
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}
	if !changed {
		return nil
	}
	return c.loadAndReconcile(ctx, hash)
}

func (c *Controller) Sync(ctx context.Context) (*reconcile.Result, error) {
	if err := c.loadProject(ctx, ""); err != nil {
		return nil, err
	}
	return c.reconciler.Sync(ctx), nil
}

func (c *Controller) Reconcile(ctx context.Context) (*reconcile.Result, error) {
	if err := c.loadProject(ctx, ""); err != nil {
		return nil, err
	}
	return c.reconciler.Reconcile(ctx), nil
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
