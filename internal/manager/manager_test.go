package manager

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LoriKarikari/kedge/internal/state"
)

const testRepoName = "test-repo"

func newTestStore(t *testing.T) *state.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := state.New(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestNew(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, nil, slog.Default())

	if mgr.store != store {
		t.Error("expected store to be set")
	}
	if mgr.controllers == nil {
		t.Error("expected controllers map to be initialized")
	}
	if mgr.repoStatus == nil {
		t.Error("expected repoStatus map to be initialized")
	}
}

func TestIsReadyNoControllers(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, nil, slog.Default())

	if mgr.IsReady() {
		t.Error("expected IsReady to return false with no controllers")
	}
}

func TestStartNoRepos(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, nil, slog.Default())

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	err := mgr.Start(ctx, Config{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestStartAllReposFail(t *testing.T) {
	store := newTestStore(t)

	_, err := store.SaveRepo(t.Context(), testRepoName, "https://github.com/octocat/Hello-World.git", "master")
	if err != nil {
		t.Fatal(err)
	}

	mgr := New(store, nil, slog.Default())

	err = mgr.Start(t.Context(), Config{
		StatePath: filepath.Join(t.TempDir(), "state.db"),
	})

	if err == nil {
		t.Fatal("expected error when all repos fail")
	}
	if !strings.Contains(err.Error(), "all repos failed to start") {
		t.Errorf("expected 'all repos failed' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "kedge.yaml not found") {
		t.Errorf("expected error to mention kedge.yaml, got: %v", err)
	}

	status := mgr.Status()
	if len(status) != 1 {
		t.Errorf("expected 1 status entry, got %d", len(status))
	}
	if status[testRepoName] == nil {
		t.Error("expected status for " + testRepoName)
	} else if status[testRepoName].Running {
		t.Error("expected " + testRepoName + " to not be running")
	}
}

func TestStartMultipleReposFail(t *testing.T) {
	store := newTestStore(t)

	_, err := store.SaveRepo(t.Context(), "repo1", "https://github.com/octocat/Hello-World.git", "master")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.SaveRepo(t.Context(), "repo2", "https://github.com/octocat/Spoon-Knife.git", "main")
	if err != nil {
		t.Fatal(err)
	}

	mgr := New(store, nil, slog.Default())

	err = mgr.Start(t.Context(), Config{
		StatePath: filepath.Join(t.TempDir(), "state.db"),
	})

	if err == nil {
		t.Fatal("expected error when all repos fail")
	}
	if !strings.Contains(err.Error(), "all repos failed to start") {
		t.Errorf("expected 'all repos failed' error, got: %v", err)
	}

	status := mgr.Status()
	if len(status) != 2 {
		t.Errorf("expected 2 status entries, got %d", len(status))
	}
	for name, s := range status {
		if s.Running {
			t.Errorf("expected %s to not be running", name)
		}
		if s.Error == nil {
			t.Errorf("expected %s to have an error", name)
		}
	}
}

func TestStatusEmpty(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, nil, slog.Default())

	status := mgr.Status()
	if len(status) != 0 {
		t.Errorf("expected empty status, got %d entries", len(status))
	}
}

func TestClose(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, nil, slog.Default())

	err := mgr.Close()
	if err != nil {
		t.Errorf("expected no error closing empty manager, got %v", err)
	}
}
