package cli

import (
	"context"
	"fmt"

	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var historyFlags struct {
	limit int
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show deployment history",
	Long:  `Display a list of past deployments with their status and commit info.`,
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVar(&historyFlags.limit, "limit", 10, "Maximum number of entries to show")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	if repo == nil {
		return fmt.Errorf("--repo is required")
	}

	ctx := context.Background()

	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	deployments, err := store.ListDeployments(ctx, repo.Name, historyFlags.limit)
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
