package cli

import (
	"context"
	"fmt"

	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var historyFlags struct {
	statePath string
	limit     int
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show deployment history",
	Long:  `Display a list of past deployments with their status and commit info.`,
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().StringVar(&historyFlags.statePath, "state", ".kedge/state.db", "Path to state database")
	historyCmd.Flags().IntVar(&historyFlags.limit, "limit", 10, "Maximum number of entries to show")

	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	store, err := state.New(ctx, historyFlags.statePath)
	if err != nil {
		return err
	}
	defer store.Close()

	deployments, err := store.ListDeployments(ctx, historyFlags.limit)
	if err != nil {
		return err
	}

	if len(deployments) == 0 {
		fmt.Println("No deployments yet")
		return nil
	}

	fmt.Printf("%-8s  %-10s  %-20s  %s\n", "COMMIT", "STATUS", "TIME", "MESSAGE")
	fmt.Println("--------  ----------  --------------------  -------")

	for _, d := range deployments {
		msg := d.Message
		if len(msg) > 40 {
			msg = msg[:37] + "..."
		}
		fmt.Printf("%-8s  %-10s  %-20s  %s\n",
			lo.Substring(d.CommitHash, 0, 8),
			d.Status,
			d.DeployedAt.Format("2006-01-02 15:04:05"),
			msg,
		)
	}

	return nil
}
