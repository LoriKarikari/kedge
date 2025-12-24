package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging LoggingConfig `yaml:"logging"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Default() *Config {
	return &Config{
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
