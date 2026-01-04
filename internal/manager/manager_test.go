package manager

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/kedge/internal/state"
)

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
	mgr := New(store, slog.Default())

	if mgr.store != store {
		t.Error("expected store to be set")
	}
	if mgr.controllers == nil {
		t.Error("expected controllers map to be initialized")
	}
}

func TestIsReadyNoControllers(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, slog.Default())

	if mgr.IsReady() {
		t.Error("expected IsReady to return false with no controllers")
	}
}

func TestStartNoRepos(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, slog.Default())

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	err := mgr.Start(ctx, Config{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestStartMissingKedgeYaml(t *testing.T) {
	store := newTestStore(t)

	_, err := store.SaveRepo(t.Context(), "test-repo", "https://github.com/octocat/Hello-World.git", "master")
	if err != nil {
		t.Fatal(err)
	}

	mgr := New(store, slog.Default())

	err = mgr.Start(t.Context(), Config{
		StatePath: filepath.Join(t.TempDir(), "state.db"),
	})

	if err == nil {
		t.Fatal("expected error for missing kedge.yaml")
	}
	if err.Error() != "repo test-repo: kedge.yaml not found" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClose(t *testing.T) {
	store := newTestStore(t)
	mgr := New(store, slog.Default())

	err := mgr.Close()
	if err != nil {
		t.Errorf("expected no error closing empty manager, got %v", err)
	}
}
