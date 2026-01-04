package cli

import (
	"context"
	"fmt"

	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/spf13/cobra"
)

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repositories",
	Long:  `List all registered repositories.`,
	RunE:  runRepoList,
}

func init() {
	repoCmd.AddCommand(repoListCmd)
}

func runRepoList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	repos, err := store.ListRepos(ctx)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Println("No repositories registered")
		return nil
	}

	fmt.Printf("%-20s  %-10s  %s\n", "NAME", "BRANCH", "URL")
	fmt.Println("--------------------  ----------  ---")
	for _, r := range repos {
		fmt.Printf("%-20s  %-10s  %s\n", r.Name, r.Branch, r.URL)
	}

	return nil
}
