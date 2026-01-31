package cli

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/LoriKarikari/kedge/internal/state"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var repoAddFlags struct {
	name   string
	branch string
}

var repoAddCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a repository",
	Long:  `Add a Git repository to be watched by kedge.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRepoAdd,
}

func init() {
	repoAddCmd.Flags().StringVar(&repoAddFlags.name, "name", "", "Repository name (defaults to repo name from URL)")
	repoAddCmd.Flags().StringVar(&repoAddFlags.branch, "branch", "main", "Branch to watch")
	repoCmd.AddCommand(repoAddCmd)
}

func runRepoAdd(cmd *cobra.Command, args []string) error {
	repoURL := args[0]

	name := repoAddFlags.name
	if name == "" {
		name = repoNameFromURL(repoURL)
	}

	if name == "" {
		err := huh.NewInput().
			Title("Repository name").
			Value(&name).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("name is required")
				}
				return nil
			}).
			Run()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	repo, err := store.SaveRepo(ctx, name, repoURL, repoAddFlags.branch, nil)
	if err != nil {
		return fmt.Errorf("save repo: %w", err)
	}

	fmt.Printf("Added repository %q (%s)\n", repo.Name, repo.URL)
	return nil
}

func repoNameFromURL(repoURL string) string {
	u, err := url.Parse(repoURL)
	if err != nil {
		return ""
	}

	name := path.Base(u.Path)
	name = strings.TrimSuffix(name, ".git")
	return name
}
