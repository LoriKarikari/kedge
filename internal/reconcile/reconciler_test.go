package reconcile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/kedge/internal/docker"
)

const (
	testComposeFile = "docker-compose.yaml"
	skipMsg         = "skipping integration test"
)

func skipIfNoDocker(t *testing.T, projectName string) *docker.Client {
	t.Helper()
	client, err := docker.NewClient(projectName, nil)
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	_ = client.Remove(t.Context())
	t.Cleanup(func() {
		_ = client.Remove(t.Context())
		client.Close()
	})
	return client
}

func TestNewReconciler(t *testing.T) {
	client := skipIfNoDocker(t, "kedge-test-new")

	r := New(client, nil, Config{}, nil)

	if r.config.Mode != ModeAuto {
		t.Errorf("got mode %q, want %q", r.config.Mode, ModeAuto)
	}

	if r.config.Interval != 30*time.Second {
		t.Errorf("got interval %v, want %v", r.config.Interval, 30*time.Second)
	}
}

func TestReconcilerModes(t *testing.T) {
	tests := []struct {
		mode Mode
		want Mode
	}{
		{ModeAuto, ModeAuto},
		{ModeNotify, ModeNotify},
		{ModeManual, ModeManual},
	}

	client := skipIfNoDocker(t, "kedge-test-modes")

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			r := New(client, nil, Config{Mode: tt.mode}, nil)
			if r.config.Mode != tt.want {
				t.Errorf("got mode %q, want %q", r.config.Mode, tt.want)
			}
		})
	}
}

func TestIntegrationReconcileAutoMode(t *testing.T) {
	if testing.Short() {
		t.Skip(skipMsg)
	}

	const projectName = "kedge-test-auto"
	client := skipIfNoDocker(t, projectName)
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

	project, err := docker.LoadProject(ctx, composePath, projectName)
	if err != nil {
		t.Fatal(err)
	}

	r := New(client, project, Config{Mode: ModeAuto}, nil)
	r.SetCommit("test-commit")

	result := r.Reconcile(ctx)
	if result.Error != nil {
		t.Fatalf("reconcile failed: %v", result.Error)
	}

	if !result.Reconciled {
		t.Error("expected reconciled=true for auto mode with drift")
	}

	statuses, err := client.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(statuses) != 1 {
		t.Errorf("got %d containers, want 1", len(statuses))
	}
}

func TestIntegrationReconcileNotifyMode(t *testing.T) {
	if testing.Short() {
		t.Skip(skipMsg)
	}

	const projectName = "kedge-test-notify"
	client := skipIfNoDocker(t, projectName)
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

	project, err := docker.LoadProject(ctx, composePath, projectName)
	if err != nil {
		t.Fatal(err)
	}

	r := New(client, project, Config{Mode: ModeNotify}, nil)

	result := r.Reconcile(ctx)
	if result.Error != nil {
		t.Fatalf("reconcile failed: %v", result.Error)
	}

	if result.Reconciled {
		t.Error("expected reconciled=false for notify mode")
	}

	if len(result.Changes) == 0 {
		t.Error("expected changes to be reported")
	}

	statuses, err := client.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(statuses) != 0 {
		t.Errorf("got %d containers, want 0 (notify mode should not deploy)", len(statuses))
	}
}

func TestIntegrationSync(t *testing.T) {
	if testing.Short() {
		t.Skip(skipMsg)
	}

	const projectName = "kedge-test-sync"
	client := skipIfNoDocker(t, projectName)
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

	project, err := docker.LoadProject(ctx, composePath, projectName)
	if err != nil {
		t.Fatal(err)
	}

	r := New(client, project, Config{Mode: ModeManual}, nil)
	r.SetCommit("test-commit")

	result := r.Reconcile(ctx)
	if result.Reconciled {
		t.Error("manual mode should not reconcile automatically")
	}

	result = r.Sync(ctx)
	if result.Error != nil {
		t.Fatalf("sync failed: %v", result.Error)
	}

	if !result.Reconciled {
		t.Error("sync should always reconcile")
	}

	statuses, err := client.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(statuses) != 1 {
		t.Errorf("got %d containers after sync, want 1", len(statuses))
	}
}
