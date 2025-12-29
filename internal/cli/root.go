package cli

import (
	"os"

	"github.com/LoriKarikari/kedge/internal/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

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
		return err
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
