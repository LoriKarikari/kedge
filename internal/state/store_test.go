package state

import (
	"path/filepath"
	"testing"
)

const (
	testCommitFmt     = "commit: got %q, want %q"
	testDeploymentMsg = "deployment failed"
	testRepoName      = "test-repo"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := New(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestNew(t *testing.T) {
	store := newTestStore(t)
	if store.db == nil {
		t.Error("expected db to be initialized")
	}
}

func TestSaveDeployment(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	d, err := store.SaveDeployment(ctx, testRepoName, "abc123", "services:\n  web:\n    image: nginx", StatusSuccess, "deployed successfully")
	if err != nil {
		t.Fatal(err)
	}

	if d.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if d.CommitHash != "abc123" {
		t.Errorf(testCommitFmt, d.CommitHash, "abc123")
	}
	if d.Status != StatusSuccess {
		t.Errorf("status: got %q, want %q", d.Status, StatusSuccess)
	}
}

func TestGetLastDeployment(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	_, err := store.SaveDeployment(ctx, testRepoName, "commit1", "content1", StatusSuccess, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.SaveDeployment(ctx, testRepoName, "commit2", "content2", StatusSuccess, "")
	if err != nil {
		t.Fatal(err)
	}

	last, err := store.GetLastDeployment(ctx, testRepoName)
	if err != nil {
		t.Fatal(err)
	}

	if last.CommitHash != "commit2" {
		t.Errorf(testCommitFmt, last.CommitHash, "commit2")
	}
}

func TestGetLastDeploymentEmpty(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetLastDeployment(t.Context(), testRepoName)
	if err != ErrNotFound {
		t.Errorf("error: got %v, want ErrNotFound", err)
	}
}

func TestGetDeploymentByCommit(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	_, err := store.SaveDeployment(ctx, testRepoName, "abc123", "content", StatusSuccess, "")
	if err != nil {
		t.Fatal(err)
	}

	d, err := store.GetDeploymentByCommit(ctx, testRepoName, "abc123")
	if err != nil {
		t.Fatal(err)
	}

	if d.CommitHash != "abc123" {
		t.Errorf(testCommitFmt, d.CommitHash, "abc123")
	}
}

func TestGetDeploymentByCommitNotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetDeploymentByCommit(t.Context(), testRepoName, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("error: got %v, want ErrNotFound", err)
	}
}

func TestListDeployments(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	for i := range 5 {
		_, err := store.SaveDeployment(ctx, testRepoName, "commit"+string(rune('0'+i)), "content", StatusSuccess, "")
		if err != nil {
			t.Fatal(err)
		}
	}

	deployments, err := store.ListDeployments(ctx, testRepoName, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(deployments) != 3 {
		t.Errorf("count: got %d, want 3", len(deployments))
	}

	if deployments[0].CommitHash != "commit4" {
		t.Errorf("first commit: got %q, want %q", deployments[0].CommitHash, "commit4")
	}
}

func TestUpdateDeploymentStatus(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	d, err := store.SaveDeployment(ctx, testRepoName, "abc123", "content", StatusPending, "")
	if err != nil {
		t.Fatal(err)
	}

	err = store.UpdateDeploymentStatus(ctx, d.ID, StatusFailed, testDeploymentMsg)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := store.GetDeployment(ctx, d.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Status != StatusFailed {
		t.Errorf("status: got %q, want %q", updated.Status, StatusFailed)
	}
	if updated.Message != testDeploymentMsg {
		t.Errorf("message: got %q, want %q", updated.Message, testDeploymentMsg)
	}
}

func TestUpdateDeploymentStatusNotFound(t *testing.T) {
	store := newTestStore(t)

	err := store.UpdateDeploymentStatus(t.Context(), 999, StatusFailed, "")
	if err != ErrNotFound {
		t.Errorf("error: got %v, want ErrNotFound", err)
	}
}

func TestStatusRolledBack(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	d, err := store.SaveDeployment(ctx, testRepoName, "abc123", "content", StatusSuccess, "")
	if err != nil {
		t.Fatal(err)
	}

	err = store.UpdateDeploymentStatus(ctx, d.ID, StatusRolledBack, "rolled back to previous")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := store.GetDeployment(ctx, d.ID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Status != StatusRolledBack {
		t.Errorf("status: got %q, want %q", updated.Status, StatusRolledBack)
	}
}

func TestSaveDeploymentInvalidStatus(t *testing.T) {
	store := newTestStore(t)

	_, err := store.SaveDeployment(t.Context(), testRepoName, "abc123", "content", "invalid", "")
	if err != ErrInvalidStatus {
		t.Errorf("error: got %v, want ErrInvalidStatus", err)
	}
}

func TestUpdateDeploymentStatusInvalid(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	d, err := store.SaveDeployment(ctx, testRepoName, "abc123", "content", StatusPending, "")
	if err != nil {
		t.Fatal(err)
	}

	err = store.UpdateDeploymentStatus(ctx, d.ID, "invalid", "")
	if err != ErrInvalidStatus {
		t.Errorf("error: got %v, want ErrInvalidStatus", err)
	}
}
