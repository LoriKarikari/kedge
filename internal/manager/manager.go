package manager

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/samber/lo"

	"github.com/LoriKarikari/kedge/internal/config"
	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/LoriKarikari/kedge/internal/telemetry"
)

type Config struct {
	StatePath    string
	PollInterval string
}

type RepoStatus struct {
	Running bool
	Error   error
}

type Manager struct {
	store       *state.Store
	telemetry   *telemetry.Provider
	controllers map[string]*controller.Controller
	repoStatus  map[string]*RepoStatus
	logger      *slog.Logger
	mu          sync.RWMutex
}

func New(store *state.Store, tp *telemetry.Provider, logger *slog.Logger) *Manager {
	return &Manager{
		store:       store,
		telemetry:   tp,
		controllers: make(map[string]*controller.Controller),
		repoStatus:  make(map[string]*RepoStatus),
		logger:      logger.With(slog.String("component", "manager")),
	}
}

func (m *Manager) Start(ctx context.Context, cfg Config) error {
	repos, err := m.store.ListRepos(ctx)
	if err != nil {
		return fmt.Errorf("list repos: %w", err)
	}

	if len(repos) == 0 {
		m.logger.Info("no repositories registered, waiting for repos to be added")
		<-ctx.Done()
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(repos))

	for _, repo := range repos {
		wg.Add(1)
		go func(r *state.Repo) {
			defer wg.Done()
			if err := m.startRepo(ctx, r, cfg); err != nil {
				m.logger.Error("failed to start repo", slog.String("repo", r.Name), slog.Any("error", err))
				errCh <- fmt.Errorf("repo %s: %w", r.Name, err)
			}
		}(repo)
	}

	wg.Wait()
	close(errCh)

	var failedRepos []string
	for err := range errCh {
		failedRepos = append(failedRepos, err.Error())
	}

	m.mu.RLock()
	runningCount := len(m.controllers)
	m.mu.RUnlock()

	if runningCount == 0 && len(failedRepos) > 0 {
		return fmt.Errorf("all repos failed to start: %s", strings.Join(failedRepos, "; "))
	}

	if len(failedRepos) > 0 {
		m.logger.Warn("some repos failed to start", slog.Int("failed", len(failedRepos)), slog.Int("running", runningCount))
	}

	<-ctx.Done()
	return nil
}

func (m *Manager) startRepo(ctx context.Context, repo *state.Repo, mgrCfg Config) error {
	workDir := repoWorkDir(repo.Name)

	var watcherOpts []git.WatcherOption
	watcherOpts = append(watcherOpts, git.WithRepoName(repo.Name))
	if m.telemetry != nil {
		watcherOpts = append(watcherOpts, git.WithMetrics(m.telemetry.Metrics))
	}

	watcher := git.NewWatcher(repo.URL, repo.Branch, workDir, config.Default().Git.PollInterval, m.logger, watcherOpts...)

	if err := watcher.Clone(ctx); err != nil {
		m.mu.Lock()
		m.repoStatus[repo.Name] = &RepoStatus{Running: false, Error: fmt.Errorf("clone: %w", err)}
		m.mu.Unlock()
		return fmt.Errorf("clone: %w", err)
	}

	repoCfg, err := loadRepoConfig(repo.Name)
	if err != nil {
		m.mu.Lock()
		m.repoStatus[repo.Name] = &RepoStatus{Running: false, Error: fmt.Errorf("kedge.yaml not found")}
		m.mu.Unlock()
		return fmt.Errorf("kedge.yaml not found")
	}

	mode, err := reconcile.ParseMode(repoCfg.Reconciliation.Mode)
	if err != nil {
		mode = reconcile.ModeAuto
	}

	ctrlCfg := controller.Config{
		RepoName:     repo.Name,
		ProjectName:  repoCfg.Docker.ProjectName,
		ComposePath:  repoCfg.Docker.ComposeFile,
		StatePath:    mgrCfg.StatePath,
		ReconcileCfg: reconcile.Config{Mode: mode},
	}

	var metrics *telemetry.Metrics
	if m.telemetry != nil {
		metrics = m.telemetry.Metrics
	}
	ctrl, err := controller.New(ctx, watcher, ctrlCfg, metrics, m.logger)
	if err != nil {
		m.mu.Lock()
		m.repoStatus[repo.Name] = &RepoStatus{Running: false, Error: fmt.Errorf("create controller: %w", err)}
		m.mu.Unlock()
		return fmt.Errorf("create controller: %w", err)
	}

	m.mu.Lock()
	m.controllers[repo.Name] = ctrl
	m.repoStatus[repo.Name] = &RepoStatus{Running: true}
	m.mu.Unlock()

	m.logger.Info("starting repo", slog.String("repo", repo.Name), slog.String("url", repo.URL))

	go func() {
		if err := ctrl.Run(ctx); err != nil && ctx.Err() == nil {
			m.mu.Lock()
			m.repoStatus[repo.Name] = &RepoStatus{Running: false, Error: err}
			delete(m.controllers, repo.Name)
			m.mu.Unlock()
			m.logger.Error("controller stopped", slog.String("repo", repo.Name), slog.Any("error", err))
		}
	}()

	return nil
}

func (m *Manager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return lo.SomeBy(lo.Values(m.controllers), func(ctrl *controller.Controller) bool {
		return ctrl.IsReady()
	})
}

func (m *Manager) Status() map[string]*RepoStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*RepoStatus, len(m.repoStatus))
	for k, v := range m.repoStatus {
		result[k] = &RepoStatus{Running: v.Running, Error: v.Error}
	}
	return result
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, ctrl := range m.controllers {
		if err := ctrl.Close(); err != nil {
			m.logger.Error("failed to close controller", slog.String("repo", name), slog.Any("error", err))
			lastErr = err
		}
	}
	return lastErr
}

func repoWorkDir(name string) string {
	return filepath.Join(".kedge", "repos", name)
}

func loadRepoConfig(name string) (*config.Config, error) {
	workDir := repoWorkDir(name)
	configPath := filepath.Join(workDir, "kedge.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		configPath = filepath.Join(workDir, "kedge.yml")
		cfg, err = config.Load(configPath)
	}
	return cfg, err
}
