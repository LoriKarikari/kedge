package cli

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/LoriKarikari/kedge/internal/git/auth"
	"github.com/LoriKarikari/kedge/internal/state"
)

var repoAddFlags struct {
	name        string
	branch      string
	sshKeyPath  string
	username    string
	passwordEnv string
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
	repoAddCmd.Flags().StringVar(&repoAddFlags.sshKeyPath, "ssh-private-key-path", "", "Path to SSH private key for authentication")
	repoAddCmd.Flags().StringVar(&repoAddFlags.username, "username", "", "Username for HTTPS authentication (defaults to x-access-token)")
	repoAddCmd.Flags().StringVar(&repoAddFlags.passwordEnv, "password-env", "", "Environment variable name containing the password/token")
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

	repoAuth, err := buildAuthConfig(repoURL)
	if err != nil {
		return err
	}

	ctx := context.Background()
	store, err := state.New(ctx, cfg.State.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	repo, err := store.SaveRepo(ctx, name, repoURL, repoAddFlags.branch, repoAuth)
	if err != nil {
		return fmt.Errorf("save repo: %w", err)
	}

	fmt.Printf("Added repository %q (%s)\n", repo.Name, repo.URL)
	if repoAuth != nil {
		fmt.Printf("  Auth: %s\n", repoAuth.Type)
	}
	return nil
}

func buildAuthConfig(repoURL string) (*state.RepoAuth, error) {
	if repoAddFlags.sshKeyPath != "" {
		if _, err := os.Stat(repoAddFlags.sshKeyPath); err != nil {
			return nil, fmt.Errorf("SSH key not found: %s", repoAddFlags.sshKeyPath)
		}
		return &state.RepoAuth{
			Type:       string(auth.TypeSSHKey),
			SSHKeyPath: repoAddFlags.sshKeyPath,
		}, nil
	}

	if repoAddFlags.passwordEnv != "" {
		return &state.RepoAuth{
			Type:        string(auth.TypeToken),
			Username:    repoAddFlags.username,
			PasswordEnv: repoAddFlags.passwordEnv,
		}, nil
	}

	if isSSHURL(repoURL) {
		var sshKeyPath string
		err := huh.NewInput().
			Title("SSH private key path").
			Description("Path to the SSH private key for authentication").
			Placeholder("~/.ssh/id_ed25519").
			Value(&sshKeyPath).
			Validate(func(s string) error {
				if s == "" {
					return nil
				}
				expanded := expandPath(s)
				if _, err := os.Stat(expanded); err != nil {
					return fmt.Errorf("file not found: %s", expanded)
				}
				return nil
			}).
			Run()
		if err != nil {
			return nil, err
		}
		if sshKeyPath != "" {
			return &state.RepoAuth{
				Type:       string(auth.TypeSSHKey),
				SSHKeyPath: expandPath(sshKeyPath),
			}, nil
		}
	}

	return nil, nil
}

func isSSHURL(repoURL string) bool {
	return strings.HasPrefix(repoURL, "git@") ||
		strings.HasPrefix(repoURL, "ssh://")
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return strings.Replace(p, "~", home, 1)
		}
	}
	return p
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
