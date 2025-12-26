package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/samber/lo"
)

const defaultTimeout = 30 * time.Second

func (c *Client) Status(ctx context.Context) ([]ServiceStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	args := filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("%s=true", LabelManaged)),
		filters.Arg("label", fmt.Sprintf("%s=%s", LabelProject, c.projectName)),
	)

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	return lo.Map(containers, func(cont container.Summary, _ int) ServiceStatus {
		return ServiceStatus{
			Service:   cont.Labels[LabelService],
			Container: shortContainerID(cont),
			Image:     cont.Image,
			State:     cont.State,
			Health:    extractHealth(cont),
			CreatedAt: time.Unix(cont.Created, 0),
		}
	}), nil
}

func shortContainerID(cont container.Summary) string {
	return lo.CoalesceOrEmpty(
		lo.FirstOrEmpty(cont.Names),
		lo.Substring(cont.ID, 0, 12),
	)
}

func extractHealth(cont container.Summary) string {
	return lo.Switch[string, string](cont.State).
		Case("running", cont.Status).
		Default(cont.State)
}
