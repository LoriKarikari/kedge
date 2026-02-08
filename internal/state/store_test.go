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
	_, err = store.SaveRepo(t.Context(), testRepoName, "https://example.com/repo.git", "main", nil)
	if err != nil {
		t.Fatal(err)
	}
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

func TestFindRepoByURL(t *testing.T) {
	store := newTestStore(t)
	ctx := t.Context()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{name: "exact match", query: "https://example.com/repo.git"},
		{name: "without .git", query: "https://example.com/repo"},
		{name: "ssh format", query: "git@example.com:repo.git"},
		{name: "trailing slash", query: "https://example.com/repo/"},
		{name: "case insensitive", query: "https://Example.COM/Repo.git"},
		{name: "no match", query: "https://other.com/repo.git", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := store.FindRepoByURL(ctx, tt.query)
			if tt.wantErr {
				if err != ErrNotFound {
					t.Errorf("error: got %v, want ErrNotFound", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if repo.Name != testRepoName {
				t.Errorf("name: got %q, want %q", repo.Name, testRepoName)
			}
		})
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/org/repo.git", "github.com/org/repo"},
		{"https://github.com/org/repo", "github.com/org/repo"},
		{"git@github.com:org/repo.git", "github.com/org/repo"},
		{"http://github.com/org/repo.git/", "github.com/org/repo"},
		{"HTTPS://GitHub.COM/Org/Repo.git", "github.com/org/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeGitURL(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
