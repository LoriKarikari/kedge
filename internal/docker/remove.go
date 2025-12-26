package docker

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/samber/lo"
)

func (c *Client) kedgeFilters() filters.Args {
	return filters.NewArgs(
		filters.Arg("label", LabelManaged+"=true"),
		filters.Arg("label", fmt.Sprintf("%s=%s", LabelProject, c.projectName)),
	)
}

func (c *Client) Remove(ctx context.Context) error {
	c.logger.Info("removing project resources")

	if err := c.removeContainers(ctx); err != nil {
		return err
	}

	return c.removeNetworks(ctx)
}

func (c *Client) removeContainers(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: c.kedgeFilters(),
	})
	if err != nil {
		return fmt.Errorf("list containers: %w", err)
	}

	var errs []error
	for i := range containers {
		if err := c.removeContainer(ctx, containers[i].ID); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (c *Client) removeNetworks(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	networks, err := c.cli.NetworkList(ctx, network.ListOptions{Filters: c.kedgeFilters()})
	if err != nil {
		return fmt.Errorf("list networks: %w", err)
	}

	var errs []error
	for i := range networks {
		c.logger.Info("removing network", "network", networks[i].Name)
		if err := c.cli.NetworkRemove(ctx, networks[i].ID); err != nil {
			errs = append(errs, fmt.Errorf("remove network %s: %w", networks[i].Name, err))
		}
	}

	return errors.Join(errs...)
}

func (c *Client) RemoveService(ctx context.Context, serviceName string) error {
	cont, err := c.findContainer(ctx, serviceName)
	if err != nil {
		return err
	}

	if cont == nil {
		return nil
	}

	return c.removeContainer(ctx, cont.ID)
}

func (c *Client) Prune(ctx context.Context, keepServices []string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: c.kedgeFilters(),
	})
	if err != nil {
		return fmt.Errorf("list containers: %w", err)
	}

	var errs []error
	for i := range containers {
		serviceName := containers[i].Labels[LabelService]
		if lo.Contains(keepServices, serviceName) {
			continue
		}
		c.logger.Info("pruning orphan container", "service", serviceName)
		if err := c.removeContainer(ctx, containers[i].ID); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
