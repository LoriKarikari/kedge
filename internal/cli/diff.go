package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/spf13/cobra"
)

var diffFlags struct {
	projectName string
	composePath string
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show drift between desired and actual state",
	Long:  `Compare the compose file with running containers and show differences.`,
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().StringVar(&diffFlags.projectName, "project", "kedge", "Docker compose project name")
	diffCmd.Flags().StringVar(&diffFlags.composePath, "compose", "docker-compose.yaml", "Path to compose file")

	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	client, err := docker.NewClient(diffFlags.projectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	project, err := docker.LoadProject(ctx, diffFlags.composePath, diffFlags.projectName)
	if err != nil {
		return fmt.Errorf("load compose: %w", err)
	}

	diff, err := client.Diff(ctx, project)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	if diff.InSync {
		fmt.Println("No drift detected - all services in sync")
		return nil
	}

	fmt.Printf("Drift detected: %s\n\n", diff.Summary)

	for _, change := range diff.Changes {
		fmt.Printf("%s %s\n", actionSymbol(change.Action), change.Service)
		fmt.Printf("  Action: %s\n", change.Action)
		fmt.Printf("  Reason: %s\n", change.Reason)
		if change.DesiredImage != "" {
			fmt.Printf("  Desired: %s\n", change.DesiredImage)
		}
		if change.CurrentImage != "" {
			fmt.Printf("  Current: %s\n", change.CurrentImage)
		}
		fmt.Println()
	}

	return nil
}

func actionSymbol(action docker.DiffAction) string {
	switch action {
	case docker.ActionCreate:
		return "+"
	case docker.ActionUpdate:
		return "~"
	case docker.ActionRemove:
		return "-"
	default:
		return "?"
	}
}
