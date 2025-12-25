package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Git     GitConfig     `yaml:"git"`
	Logging LoggingConfig `yaml:"logging"`
}

type GitConfig struct {
	RepoURL      string        `yaml:"repo_url"`
	Branch       string        `yaml:"branch"`
	PollInterval time.Duration `yaml:"poll_interval"`
	WorkDir      string        `yaml:"work_dir"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Default() *Config {
	return &Config{
		Git: GitConfig{
			Branch:       "main",
			PollInterval: 60 * time.Second,
			WorkDir:      "/var/lib/kedge/repo",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
