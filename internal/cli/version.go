package cli

import (
	"fmt"

	"github.com/LoriKarikari/kedge/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kedge %s (%s)\n", version.Version(), version.Commit())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
