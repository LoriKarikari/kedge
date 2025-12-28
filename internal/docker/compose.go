package docker

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/samber/lo"
)

func LoadProject(ctx context.Context, composePath, projectName string) (*types.Project, error) {
	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return nil, fmt.Errorf("resolve compose path: %w", err)
	}

	opts, err := cli.NewProjectOptions(
		[]string{absPath},
		cli.WithWorkingDirectory(filepath.Dir(absPath)),
		cli.WithName(projectName),
		cli.WithResolvedPaths(true),
		cli.WithInterpolation(true),
	)
	if err != nil {
		return nil, fmt.Errorf("create project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("parse compose file: %w", err)
	}

	return project, nil
}

func ServiceNames(project *types.Project) []string {
	return lo.Keys(project.Services)
}
