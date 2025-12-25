package docker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/client"
)

type Client struct {
	cli         *client.Client
	logger      *slog.Logger
	projectName string
}

func NewClient(projectName string, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("component", "docker", "project", projectName)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := cli.Ping(ctx); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("ping docker daemon: %w", err)
	}

	logger.Info("docker client initialized")
	return &Client{
		cli:         cli,
		logger:      logger,
		projectName: projectName,
	}, nil
}

func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}
