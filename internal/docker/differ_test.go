package docker

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testSummaryInSync     = "all services in sync"
	testReasonNotDeployed = "service not deployed"
)

func TestBuildSummary(t *testing.T) {
	tests := []struct {
		name    string
		changes []ServiceDiff
		want    string
	}{
		{
			name:    "no changes",
			changes: nil,
			want:    testSummaryInSync,
		},
		{
			name: "one create",
			changes: []ServiceDiff{
				{Service: "web", Action: ActionCreate},
			},
			want: "1 to create",
		},
		{
			name: "multiple actions",
			changes: []ServiceDiff{
				{Service: "web", Action: ActionCreate},
				{Service: "api", Action: ActionCreate},
				{Service: "db", Action: ActionUpdate},
				{Service: "cache", Action: ActionRemove},
			},
			want: "2 to create, 1 to update, 1 to remove",
		},
		{
			name: "only removes",
			changes: []ServiceDiff{
				{Service: "old1", Action: ActionRemove},
				{Service: "old2", Action: ActionRemove},
			},
			want: "2 to remove",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSummary(tt.changes)
			if got != tt.want {
				t.Errorf("buildSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIntegrationDiffNoContainers(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	client := skipIfNoDocker(t)
	ctx := t.Context()

	dir := t.TempDir()
	composePath := filepath.Join(dir, testComposeFile)

	content := `
services:
  web:
    image: nginx:alpine
  api:
    image: nginx:alpine
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(ctx, composePath, testProjectName)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Diff(ctx, project)
	if err != nil {
		t.Fatal(err)
	}

	if result.InSync {
		t.Error("expected not in sync when no containers exist")
	}

	if len(result.Changes) != 2 {
		t.Errorf("got %d changes, want 2", len(result.Changes))
	}

	for _, change := range result.Changes {
		if change.Action != ActionCreate {
			t.Errorf("got action %q, want %q", change.Action, ActionCreate)
		}
		if change.Reason != testReasonNotDeployed {
			t.Errorf("got reason %q, want %q", change.Reason, testReasonNotDeployed)
		}
	}
}

func TestIntegrationDiffInSync(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	client := skipIfNoDocker(t)
	ctx := t.Context()

	dir := t.TempDir()
	composePath := filepath.Join(dir, testComposeFile)

	content := `
services:
  web:
    image: nginx:alpine
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(ctx, composePath, testProjectName)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = client.Remove(ctx) })

	if err := client.Deploy(ctx, project, "test"); err != nil {
		t.Fatal(err)
	}

	result, err := client.Diff(ctx, project)
	if err != nil {
		t.Fatal(err)
	}

	if !result.InSync {
		t.Errorf("expected in sync after deploy, got changes: %v", result.Changes)
	}

	if result.Summary != testSummaryInSync {
		t.Errorf("got summary %q, want %q", result.Summary, testSummaryInSync)
	}
}

func TestIntegrationDiffOrphanService(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationMsg)
	}

	client := skipIfNoDocker(t)
	ctx := t.Context()

	dir := t.TempDir()
	composePath := filepath.Join(dir, testComposeFile)

	content := `
services:
  web:
    image: nginx:alpine
  api:
    image: nginx:alpine
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(ctx, composePath, testProjectName)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = client.Remove(ctx) })

	if err := client.Deploy(ctx, project, "test"); err != nil {
		t.Fatal(err)
	}

	reducedContent := `
services:
  web:
    image: nginx:alpine
`
	if err := os.WriteFile(composePath, []byte(reducedContent), 0o644); err != nil {
		t.Fatal(err)
	}

	reducedProject, err := LoadProject(ctx, composePath, testProjectName)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Diff(ctx, reducedProject)
	if err != nil {
		t.Fatal(err)
	}

	if result.InSync {
		t.Error("expected drift when service removed from compose")
	}

	var foundRemove bool
	for _, change := range result.Changes {
		if change.Service == "api" && change.Action == ActionRemove {
			foundRemove = true
			break
		}
	}

	if !foundRemove {
		t.Errorf("expected remove action for api service, got: %v", result.Changes)
	}
}
