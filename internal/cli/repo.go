package cli

import "github.com/spf13/cobra"

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories",
	Long:  `Commands for adding, listing, and removing repositories.`,
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
