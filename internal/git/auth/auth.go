package auth

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type Type string

const (
	TypeNone   Type = ""
	TypeSSHKey Type = "ssh-key"
	TypeToken  Type = "token"
)

type Config struct {
	Type        Type
	SSHKeyPath  string
	Username    string
	PasswordEnv string
}

func (c *Config) Resolve(logger *slog.Logger) (transport.AuthMethod, error) {
	if c == nil {
		return nil, nil
	}

	switch c.Type {
	case TypeNone:
		return nil, nil

	case TypeSSHKey:
		return c.resolveSSH(logger)

	case TypeToken:
		return c.resolveToken()

	default:
		return nil, fmt.Errorf("unknown auth type: %s", c.Type)
	}
}

func (c *Config) resolveSSH(logger *slog.Logger) (transport.AuthMethod, error) {
	if c.SSHKeyPath == "" {
		return nil, fmt.Errorf("SSH key path is required for ssh-key auth")
	}

	c.warnInsecureKeyPermissions(logger)

	auth, err := ssh.NewPublicKeysFromFile("git", c.SSHKeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("loading SSH key from %s: %w", c.SSHKeyPath, err)
	}

	return auth, nil
}

func (c *Config) warnInsecureKeyPermissions(logger *slog.Logger) {
	if logger == nil {
		return
	}

	info, err := os.Stat(c.SSHKeyPath)
	if err != nil {
		return
	}

	mode := info.Mode().Perm()
	if mode > 0o600 {
		logger.Warn("SSH key has insecure permissions",
			slog.String("path", c.SSHKeyPath),
			slog.String("mode", fmt.Sprintf("%04o", mode)),
			slog.String("recommended", "0600"),
		)
	}
}

func (c *Config) resolveToken() (transport.AuthMethod, error) {
	if c.PasswordEnv == "" {
		return nil, fmt.Errorf("password env var name is required for token auth")
	}

	password := os.Getenv(c.PasswordEnv)
	if password == "" {
		return nil, fmt.Errorf("environment variable %s is not set or empty", c.PasswordEnv)
	}

	username := c.Username
	if username == "" {
		username = "x-access-token"
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

func (c *Config) IsEmpty() bool {
	return c == nil || c.Type == TypeNone
}
