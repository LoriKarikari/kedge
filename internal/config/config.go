package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Git            Git            `yaml:"git"`
	Docker         Docker         `yaml:"docker"`
	Reconciliation Reconciliation `yaml:"reconciliation"`
	State          State          `yaml:"state"`
	Logging        Logging        `yaml:"logging"`
	Server         Server         `yaml:"server"`
}

type Git struct {
	URL          string        `yaml:"url"`
	Branch       string        `yaml:"branch"`
	PollInterval time.Duration `yaml:"poll_interval"`
	WorkDir      string        `yaml:"work_dir"`
}

type Docker struct {
	ProjectName string `yaml:"project_name"`
	ComposeFile string `yaml:"compose_file"`
}

type Reconciliation struct {
	Mode     string        `yaml:"mode"`
	Interval time.Duration `yaml:"interval"`
}

type State struct {
	Path string `yaml:"path"`
}

type Logging struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type Server struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

func Default() *Config {
	return &Config{
		Git: Git{
			Branch:       "main",
			PollInterval: time.Minute,
			WorkDir:      ".kedge/repo",
		},
		Docker: Docker{
			ProjectName: "kedge",
			ComposeFile: "docker-compose.yaml",
		},
		Reconciliation: Reconciliation{
			Mode:     "auto",
			Interval: time.Minute,
		},
		State: State{
			Path: ".kedge/state.db",
		},
		Logging: Logging{
			Level:  "info",
			Format: "text",
		},
		Server: Server{
			Enabled: false,
			Port:    8080,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	expanded := expandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

var envPattern = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)

func expandEnv(s string) string {
	return envPattern.ReplaceAllStringFunc(s, func(match string) string {
		parts := envPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		name := parts[1]
		defaultVal := ""
		if len(parts) >= 3 {
			defaultVal = parts[2]
		}

		if val := os.Getenv(name); val != "" {
			return val
		}
		return defaultVal
	})
}
