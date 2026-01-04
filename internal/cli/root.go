package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/LoriKarikari/kedge/internal/config"
	"github.com/LoriKarikari/kedge/internal/logging"
	"github.com/LoriKarikari/kedge/internal/state"
)

var (
	cfg      *config.Config
	logger   *slog.Logger
	repoFlag string
	repo     *state.Repo
)

var rootCmd = &cobra.Command{
	Use:   "kedge",
	Short: "GitOps controller for Docker Compose",
	Long:  `Kedge watches a Git repository and automatically deploys Docker Compose applications when changes are detected.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "version" {
			return nil
		}

		cfg = config.Default()

		if repoFlag != "" {
			ctx := context.Background()
			store, err := state.New(ctx, cfg.State.Path)
			if err != nil {
				return fmt.Errorf("open state: %w", err)
			}
			defer store.Close()

			repo, err = store.GetRepo(ctx, repoFlag)
			if err != nil {
				if err == state.ErrNotFound {
					return fmt.Errorf("repository %q not found", repoFlag)
				}
				return err
			}

			cfg, err = loadRepoConfig(repo.Name)
			if err != nil {
				return fmt.Errorf("load repo config: %w", err)
			}
		}

		logger = logging.New(logging.Config{
			Level:  cfg.Logging.Level,
			Format: cfg.Logging.Format,
		})
		slog.SetDefault(logger)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository name to operate on")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
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
