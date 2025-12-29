package cli

import (
	"log/slog"
	"os"

	"github.com/LoriKarikari/kedge/internal/config"
	"github.com/LoriKarikari/kedge/internal/logging"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	logger *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "kedge",
	Short: "GitOps controller for Docker Compose",
	Long:  `Kedge watches a Git repository and automatically deploys Docker Compose applications when changes are detected.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if _, statErr := os.Stat("kedge.yaml"); statErr == nil {
			cfg, err = config.Load("kedge.yaml")
		} else {
			cfg = config.Default()
		}
		if err != nil {
			return err
		}
		logger = logging.New(logging.Config{
			Level:  cfg.Logging.Level,
			Format: cfg.Logging.Format,
		})
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
