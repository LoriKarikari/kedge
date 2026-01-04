package manager

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/LoriKarikari/kedge/internal/config"
	"github.com/LoriKarikari/kedge/internal/controller"
	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
	"github.com/LoriKarikari/kedge/internal/state"
)

type Config struct {
	StatePath    string
	PollInterval string
}

type Manager struct {
	store       *state.Store
	controllers map[string]*controller.Controller
	logger      *slog.Logger
	mu          sync.RWMutex
}

func New(store *state.Store, logger *slog.Logger) *Manager {
	return &Manager{
		store:       store,
		controllers: make(map[string]*controller.Controller),
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

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (m *Manager) startRepo(ctx context.Context, repo *state.Repo, mgrCfg Config) error {
	workDir := repoWorkDir(repo.Name)

	watcher := git.NewWatcher(repo.URL, repo.Branch, workDir, config.Default().Git.PollInterval, m.logger)
	if err := watcher.Clone(ctx); err != nil {
		return fmt.Errorf("clone: %w", err)
	}

	repoCfg, err := loadRepoConfig(repo.Name)
	if err != nil {
		return fmt.Errorf("kedge.yaml not found")
	}

	mode, err := reconcile.ParseMode(repoCfg.Reconciliation.Mode)
	if err != nil {
		mode = reconcile.ModeAuto
	}

	watcher = git.NewWatcher(repo.URL, repo.Branch, workDir, repoCfg.Git.PollInterval, m.logger)

	ctrlCfg := controller.Config{
		RepoName:     repo.Name,
		ProjectName:  repoCfg.Docker.ProjectName,
		ComposePath:  repoCfg.Docker.ComposeFile,
		StatePath:    mgrCfg.StatePath,
		ReconcileCfg: reconcile.Config{Mode: mode},
	}

	ctrl, err := controller.New(ctx, watcher, ctrlCfg, m.logger)
	if err != nil {
		return fmt.Errorf("create controller: %w", err)
	}

	m.mu.Lock()
	m.controllers[repo.Name] = ctrl
	m.mu.Unlock()

	m.logger.Info("starting repo", slog.String("repo", repo.Name), slog.String("url", repo.URL))

	return ctrl.Run(ctx)
}

func (m *Manager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.controllers) == 0 {
		return false
	}

	for _, ctrl := range m.controllers {
		if !ctrl.IsReady() {
			return false
		}
	}
	return true
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
