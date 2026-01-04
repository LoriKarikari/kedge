package cli

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/LoriKarikari/kedge/internal/docker"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show drift between desired and actual state",
	Long:  `Compare the compose file with running containers and show differences.`,
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	if repo == nil {
		return fmt.Errorf("--repo is required")
	}

	ctx := context.Background()

	client, err := docker.NewClient(cfg.Docker.ProjectName, logger)
	if err != nil {
		return err
	}
	defer client.Close()

	composePath := filepath.Join(repoWorkDir(repo.Name), cfg.Docker.ComposeFile)
	project, err := docker.LoadProject(ctx, composePath, cfg.Docker.ProjectName)
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
