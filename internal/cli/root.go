package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kedge",
	Short: "GitOps controller for Docker Compose",
	Long:  `Kedge watches a Git repository and automatically deploys Docker Compose applications when changes are detected.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
