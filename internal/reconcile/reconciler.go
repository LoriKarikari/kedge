package reconcile

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/compose-spec/compose-go/v2/types"
)

type Mode string

const (
	ModeAuto   Mode = "auto"
	ModeNotify Mode = "notify"
	ModeManual Mode = "manual"
)

type Config struct {
	Mode     Mode
	Interval time.Duration
}

type Result struct {
	Reconciled bool
	Changes    []docker.ServiceDiff
	Error      error
}

type Reconciler struct {
	client *docker.Client
	config Config
	logger *slog.Logger

	mu      sync.RWMutex
	project *types.Project
	commit  string
}

func New(client *docker.Client, project *types.Project, cfg Config, logger *slog.Logger) *Reconciler {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeAuto
	}
	if cfg.Interval == 0 {
		cfg.Interval = 30 * time.Second
	}

	return &Reconciler{
		client:  client,
		project: project,
		config:  cfg,
		logger:  logger.With("component", "reconciler"),
	}
}

func (r *Reconciler) SetCommit(commit string) {
	r.mu.Lock()
	r.commit = commit
	r.mu.Unlock()
}

func (r *Reconciler) SetProject(project *types.Project) {
	r.mu.Lock()
	r.project = project
	r.mu.Unlock()
}

func (r *Reconciler) getProjectAndCommit() (*types.Project, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.project, r.commit
}

func (r *Reconciler) Reconcile(ctx context.Context) *Result {
	project, _ := r.getProjectAndCommit()

	diff, err := r.client.Diff(ctx, project)
	if err != nil {
		return &Result{Error: err}
	}

	if diff.InSync {
		r.logger.Debug("no drift detected")
		return &Result{Reconciled: false}
	}

	r.logger.Info("drift detected", "summary", diff.Summary)

	if r.config.Mode == ModeNotify {
		r.logger.Info("notify mode: skipping remediation")
		return &Result{Reconciled: false, Changes: diff.Changes}
	}

	if r.config.Mode == ModeManual {
		r.logger.Info("manual mode: waiting for sync command")
		return &Result{Reconciled: false, Changes: diff.Changes}
	}

	return r.apply(ctx, diff.Changes)
}

func (r *Reconciler) Sync(ctx context.Context) *Result {
	r.logger.Info("force sync requested")

	project, commit := r.getProjectAndCommit()

	if err := r.client.Deploy(ctx, project, commit); err != nil {
		return &Result{Error: err}
	}

	serviceNames := docker.ServiceNames(project)
	if err := r.client.Prune(ctx, serviceNames); err != nil {
		r.logger.Warn("prune failed", "error", err)
	}

	return &Result{Reconciled: true}
}

func (r *Reconciler) apply(ctx context.Context, changes []docker.ServiceDiff) *Result {
	r.logger.Info("applying changes", "count", len(changes))

	project, commit := r.getProjectAndCommit()

	if err := r.client.Deploy(ctx, project, commit); err != nil {
		return &Result{Error: err, Changes: changes}
	}

	serviceNames := docker.ServiceNames(project)
	if err := r.client.Prune(ctx, serviceNames); err != nil {
		r.logger.Warn("prune failed", "error", err)
	}

	r.logger.Info("reconciliation complete")
	return &Result{Reconciled: true, Changes: changes}
}

func (r *Reconciler) Watch(ctx context.Context, results chan<- *Result) {
	ticker := time.NewTicker(r.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result := r.Reconcile(ctx)
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
