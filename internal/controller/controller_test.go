package controller

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/kedge/internal/git"
	"github.com/LoriKarikari/kedge/internal/reconcile"
)

func TestNew(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" && os.Getenv("CI") != "" {
		t.Skip("skipping: docker not available in CI")
	}

	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.db")

	watcher := git.NewWatcher(
		"https://github.com/octocat/Hello-World.git",
		"master",
		filepath.Join(tmpDir, "repo"),
		time.Minute,
		nil,
	)

	cfg := Config{
		ProjectName: "test-project",
		ComposePath: "docker-compose.yaml",
		StatePath:   statePath,
		ReconcileCfg: reconcile.Config{
			Mode: reconcile.ModeManual,
		},
	}

	ctrl, err := New(t.Context(), watcher, cfg, nil, nil)
	if err != nil {
		t.Skipf("docker not available: %v", err)
	}
	defer ctrl.Close()

	if ctrl.watcher != watcher {
		t.Error("watcher not set")
	}
	if ctrl.client == nil {
		t.Error("client not initialized")
	}
	if ctrl.reconciler == nil {
		t.Error("reconciler not initialized")
	}
	if ctrl.store == nil {
		t.Error("store not initialized")
	}
}
