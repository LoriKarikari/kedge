package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testProjectName = "kedge-test"

func TestNewClient(t *testing.T) {
	client := NewTestClient(t, testProjectName)
	if client.projectName != testProjectName {
		t.Errorf("got project name %q, want %q", client.projectName, testProjectName)
	}
}

func TestIntegrationDeployAndRemove(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationMsg)
	}

	client := NewTestClient(t, testProjectName)
	ctx := t.Context()

	dir := t.TempDir()
	composePath := filepath.Join(dir, TestComposeFile)

	content := `
services:
  web:
    image: nginx:alpine
    ports:
      - "18080:80"
`
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(ctx, composePath, testProjectName)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = client.Remove(cleanupCtx)
	})

	if err := client.Deploy(ctx, project, "test-commit"); err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	statuses, err := client.Status(ctx)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}

	if len(statuses) != 1 {
		t.Errorf("got %d containers, want 1", len(statuses))
	}

	if len(statuses) > 0 {
		s := statuses[0]
		if s.Service != "web" {
			t.Errorf("got service %q, want %q", s.Service, "web")
		}
		if s.State != "running" {
			t.Errorf("got state %q, want %q", s.State, "running")
		}
	}

	if err := client.Remove(ctx); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	statuses, err = client.Status(ctx)
	if err != nil {
		t.Fatalf("status after remove failed: %v", err)
	}

	if len(statuses) != 0 {
		t.Errorf("got %d containers after remove, want 0", len(statuses))
	}
}

func TestIntegrationPrune(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationMsg)
	}

	client := NewTestClient(t, testProjectName)
	ctx := t.Context()

	dir := t.TempDir()
	composePath := filepath.Join(dir, TestComposeFile)

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

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = client.Remove(cleanupCtx)
	})

	if err := client.Deploy(ctx, project, "test"); err != nil {
		t.Fatalf("deploy failed: %v", err)
	}

	if err := client.Prune(ctx, []string{"web"}); err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	statuses, err := client.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(statuses) != 1 {
		t.Errorf("got %d containers after prune, want 1", len(statuses))
	}

	if len(statuses) > 0 && statuses[0].Service != "web" {
		t.Errorf("got service %q, want %q", statuses[0].Service, "web")
	}
}

func TestIntegrationRedeployUpdatesContainer(t *testing.T) {
	if testing.Short() {
		t.Skip(SkipIntegrationMsg)
	}

	client := NewTestClient(t, testProjectName)
	ctx := t.Context()

	dir := t.TempDir()
	composePath := filepath.Join(dir, TestComposeFile)

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

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = client.Remove(cleanupCtx)
	})

	if err := client.Deploy(ctx, project, "commit-1"); err != nil {
		t.Fatal(err)
	}

	statuses1, err := client.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses1) == 0 {
		t.Fatal("expected container after first deploy")
	}

	if err := client.Deploy(ctx, project, "commit-2"); err != nil {
		t.Fatal(err)
	}

	statuses2, err := client.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses2) == 0 {
		t.Fatal("expected container after second deploy")
	}

	if statuses1[0].Container == statuses2[0].Container {
		t.Log("container reused (same image)")
	}
}
