package cli

import (
	"context"
	"fmt"

	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/spf13/cobra"
)

var repoRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a repository",
	Long:  `Remove a repository from kedge.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoRemove,
}

func init() {
	repoCmd.AddCommand(repoRemoveCmd)
}

func runRepoRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	ctx := context.Background()
	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.DeleteRepo(ctx, name); err != nil {
		if err == state.ErrNotFound {
			return fmt.Errorf("repository %q not found", name)
		}
		return err
	}

	fmt.Printf("Removed repository %q\n", name)
	return nil
}
