package docker

import (
	"context"
	"fmt"
	"io"
	"maps"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/samber/lo"
)

const pullTimeout = 5 * time.Minute

func (c *Client) Deploy(ctx context.Context, project *types.Project, commit string) error {
	c.logger.Info("deploying project", "services", len(project.Services))

	if err := c.ensureNetworks(ctx, project); err != nil {
		return err
	}

	for name := range project.Services {
		svc := project.Services[name]
		if err := c.deployService(ctx, project.Name, name, svc, commit); err != nil {
			return fmt.Errorf("deploy service %s: %w", name, err)
		}
	}

	return nil
}

func (c *Client) ensureNetworks(ctx context.Context, project *types.Project) error {
	for name := range project.Networks {
		networkName := fmt.Sprintf("%s_%s", project.Name, name)
		if err := c.ensureNetwork(ctx, networkName); err != nil {
			return fmt.Errorf("ensure network %s: %w", name, err)
		}
	}
	return nil
}

func (c *Client) ensureNetwork(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	networks, err := c.cli.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
	if err != nil {
		return err
	}

	if lo.ContainsBy(networks, func(n network.Summary) bool { return n.Name == name }) {
		return nil
	}

	_, err = c.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Labels: kedgeLabels(c.projectName, "", ""),
	})
	return err
}

func (c *Client) deployService(ctx context.Context, projectName, serviceName string, svc types.ServiceConfig, commit string) error {
	c.logger.Info("deploying service", "service", serviceName, "image", svc.Image)

	if err := c.pullImage(ctx, svc.Image); err != nil {
		return fmt.Errorf("pull image: %w", err)
	}

	existing, err := c.findContainer(ctx, serviceName)
	if err != nil {
		return err
	}

	if existing != nil {
		if existing.Image == svc.Image && existing.State == "running" {
			c.logger.Info("service already running with correct image", "service", serviceName)
			return nil
		}
		if err := c.removeContainer(ctx, existing.ID); err != nil {
			return fmt.Errorf("remove existing container: %w", err)
		}
	}

	return c.createAndStartContainer(ctx, projectName, serviceName, svc, commit)
}

func (c *Client) pullImage(ctx context.Context, imageName string) error {
	ctx, cancel := context.WithTimeout(ctx, pullTimeout)
	defer cancel()

	c.logger.Info("pulling image", "image", imageName)

	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	return err
}

func (c *Client) findContainer(ctx context.Context, serviceName string) (*container.Summary, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	args := filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("%s=true", LabelManaged)),
		filters.Arg("label", fmt.Sprintf("%s=%s", LabelProject, c.projectName)),
		filters.Arg("label", fmt.Sprintf("%s=%s", LabelService, serviceName)),
	)

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, nil
	}
	return &containers[0], nil
}

func (c *Client) removeContainer(ctx context.Context, containerID string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	c.logger.Info("removing container", "container", containerID[:12])
	return c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

func (c *Client) createAndStartContainer(ctx context.Context, projectName, serviceName string, svc types.ServiceConfig, commit string) error {
	labels := kedgeLabels(projectName, serviceName, commit)
	maps.Copy(labels, svc.Labels)

	exposedPorts, portBindings := buildPortMappings(svc.Ports)

	env := lo.MapToSlice(svc.Environment, func(k string, v *string) string {
		return lo.Ternary(v != nil, fmt.Sprintf("%s=%s", k, *v), k)
	})

	config := &container.Config{
		Image:        svc.Image,
		Env:          env,
		Labels:       labels,
		ExposedPorts: exposedPorts,
		Cmd:          []string(svc.Command),
		Entrypoint:   []string(svc.Entrypoint),
		WorkingDir:   svc.WorkingDir,
	}

	hostConfig := &container.HostConfig{
		PortBindings:  portBindings,
		RestartPolicy: buildRestartPolicy(svc),
	}

	contName := containerName(projectName, serviceName)

	createCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	resp, err := c.cli.ContainerCreate(createCtx, config, hostConfig, nil, nil, contName)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	c.logger.Info("created container", "container", resp.ID[:12], "service", serviceName)

	if err := c.connectToNetworks(ctx, resp.ID, svc, projectName); err != nil {
		return err
	}

	startCtx, startCancel := context.WithTimeout(ctx, defaultTimeout)
	defer startCancel()

	if err := c.cli.ContainerStart(startCtx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	c.logger.Info("started container", "container", resp.ID[:12], "service", serviceName)
	return nil
}

func (c *Client) connectToNetworks(ctx context.Context, containerID string, svc types.ServiceConfig, projectName string) error {
	for netName := range svc.Networks {
		networkName := fmt.Sprintf("%s_%s", projectName, netName)

		connectCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
		err := c.cli.NetworkConnect(connectCtx, networkName, containerID, nil)
		cancel()

		if err != nil {
			return fmt.Errorf("connect to network %s: %w", netName, err)
		}
	}
	return nil
}

func buildPortMappings(ports []types.ServicePortConfig) (nat.PortSet, nat.PortMap) {
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	for _, p := range ports {
		port, err := nat.NewPort(p.Protocol, fmt.Sprintf("%d", p.Target))
		if err != nil {
			continue
		}
		exposedPorts[port] = struct{}{}

		if p.Published != "" {
			portBindings[port] = []nat.PortBinding{{
				HostIP:   p.HostIP,
				HostPort: p.Published,
			}}
		}
	}

	return exposedPorts, portBindings
}

func buildRestartPolicy(svc types.ServiceConfig) container.RestartPolicy {
	policy := svc.Restart
	if svc.Deploy != nil && svc.Deploy.RestartPolicy != nil {
		policy = svc.Deploy.RestartPolicy.Condition
	}

	switch policy {
	case "always":
		return container.RestartPolicy{Name: container.RestartPolicyAlways}
	case "on-failure":
		return container.RestartPolicy{Name: container.RestartPolicyOnFailure}
	case "unless-stopped":
		return container.RestartPolicy{Name: container.RestartPolicyUnlessStopped}
	default:
		return container.RestartPolicy{Name: container.RestartPolicyDisabled}
	}
}

func containerName(projectName, serviceName string) string {
	return fmt.Sprintf("%s-%s-1", projectName, serviceName)
}

func kedgeLabels(projectName, serviceName, commit string) map[string]string {
	labels := map[string]string{
		LabelManaged: "true",
		LabelProject: projectName,
	}
	if serviceName != "" {
		labels[LabelService] = serviceName
	}
	if commit != "" {
		labels[LabelCommit] = commit
	}
	return labels
}
