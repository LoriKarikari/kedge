package docker

import (
	"context"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/samber/lo"
)

type DiffAction string

const (
	ActionCreate DiffAction = "create"
	ActionUpdate DiffAction = "update"
	ActionRemove DiffAction = "remove"
)

var actionSchema = z.String().OneOf([]string{
	string(ActionCreate),
	string(ActionUpdate),
	string(ActionRemove),
})

func (a DiffAction) IsValid() bool {
	s := string(a)
	return actionSchema.Validate(&s) == nil
}

type ServiceDiff struct {
	Service      string
	Action       DiffAction
	DesiredImage string
	CurrentImage string
	Reason       string
}

type DiffResult struct {
	Changes []ServiceDiff
	InSync  bool
	Summary string
}

func (c *Client) Diff(ctx context.Context, project *types.Project) (*DiffResult, error) {
	listCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	containers, err := c.listManagedContainers(listCtx)
	if err != nil {
		return nil, err
	}

	actual := lo.SliceToMap(containers, func(cont container.Summary) (string, container.Summary) {
		return cont.Labels[LabelService], cont
	})

	var changes []ServiceDiff

	for name := range project.Services {
		svc := project.Services[name]
		diff, err := c.diffService(ctx, name, svc, actual[name])
		if err != nil {
			return nil, fmt.Errorf("diff service %s: %w", name, err)
		}
		if diff != nil {
			changes = append(changes, *diff)
		}
		delete(actual, name)
	}

	for name := range actual {
		changes = append(changes, ServiceDiff{
			Service:      name,
			Action:       ActionRemove,
			CurrentImage: actual[name].Image,
			Reason:       "service removed from compose file",
		})
	}

	return &DiffResult{
		Changes: changes,
		InSync:  len(changes) == 0,
		Summary: buildSummary(changes),
	}, nil
}

func (c *Client) listManagedContainers(ctx context.Context) ([]container.Summary, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: c.kedgeFilters(),
	})
}

func (c *Client) diffService(ctx context.Context, name string, desired types.ServiceConfig, actual container.Summary) (*ServiceDiff, error) {
	if actual.ID == "" {
		return &ServiceDiff{
			Service:      name,
			Action:       ActionCreate,
			DesiredImage: desired.Image,
			Reason:       "service not deployed",
		}, nil
	}

	if actual.State != "running" {
		return &ServiceDiff{
			Service:      name,
			Action:       ActionUpdate,
			DesiredImage: desired.Image,
			CurrentImage: actual.Image,
			Reason:       fmt.Sprintf("container not running (state: %s)", actual.State),
		}, nil
	}

	imageChanged, err := c.isImageChanged(ctx, desired.Image, actual.ImageID)
	if err != nil {
		return nil, err
	}
	if imageChanged {
		return &ServiceDiff{
			Service:      name,
			Action:       ActionUpdate,
			DesiredImage: desired.Image,
			CurrentImage: actual.Image,
			Reason:       "image updated",
		}, nil
	}

	storedHash := actual.Labels[LabelConfigHash]
	currentHash := ConfigHash(desired)
	if storedHash != currentHash {
		return &ServiceDiff{
			Service:      name,
			Action:       ActionUpdate,
			DesiredImage: desired.Image,
			CurrentImage: actual.Image,
			Reason:       "config changed",
		}, nil
	}

	return nil, nil
}

func (c *Client) isImageChanged(ctx context.Context, desiredImage, actualImageID string) (bool, error) {
	inspectCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	inspect, err := c.cli.ImageInspect(inspectCtx, desiredImage)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf("inspect image %s: %w", desiredImage, err)
	}
	return inspect.ID != actualImageID, nil
}

func buildSummary(changes []ServiceDiff) string {
	if len(changes) == 0 {
		return "all services in sync"
	}

	counts := lo.CountValuesBy(changes, func(d ServiceDiff) DiffAction { return d.Action })
	parts := lo.FilterMap([]DiffAction{ActionCreate, ActionUpdate, ActionRemove}, func(action DiffAction, _ int) (string, bool) {
		count := counts[action]
		return fmt.Sprintf("%d to %s", count, action), count > 0
	})

	return strings.Join(parts, ", ")
}
