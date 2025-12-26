package git

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	testFileName       = "test.txt"
	testWorkDir        = "watcher-work"
	testCloneFailedFmt = "Clone failed: %v"
	testSecondCommit   = "second commit"
)

type testRepo struct {
	bareRepoPath string
	worktree     *git.Worktree
	clone        *git.Repository
	clonePath    string
	tmpDir       string
}

func setupTestRepo(t *testing.T) *testRepo {
	t.Helper()
	tmpDir := t.TempDir()

	clonePath := filepath.Join(tmpDir, "origin")
	repo, err := git.PlainInit(clonePath, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("failed to get worktree: %v", err)
	}

	testFile := filepath.Join(clonePath, testFileName)
	if err := os.WriteFile(testFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if _, err := wt.Add(testFileName); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	bareRepoPath := filepath.Join(tmpDir, "bare.git")
	_, err = git.PlainClone(bareRepoPath, true, &git.CloneOptions{
		URL: clonePath,
	})
	if err != nil {
		t.Fatalf("failed to create bare clone: %v", err)
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{bareRepoPath},
	})
	if err != nil {
		t.Fatalf("failed to add remote: %v", err)
	}

	return &testRepo{
		bareRepoPath: bareRepoPath,
		worktree:     wt,
		clone:        repo,
		clonePath:    clonePath,
		tmpDir:       tmpDir,
	}
}

func (r *testRepo) addCommit(t *testing.T, message string) string {
	t.Helper()

	testFile := filepath.Join(r.clonePath, testFileName)
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	if err := os.WriteFile(testFile, append(data, []byte("\n"+message)...), 0o644); err != nil {
		t.Fatalf("failed to update test file: %v", err)
	}

	if _, err := r.worktree.Add(testFileName); err != nil {
		t.Fatalf("failed to add file: %v", err)
	}

	hash, err := r.worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	if err := r.clone.Push(&git.PushOptions{}); err != nil {
		t.Fatalf("failed to push: %v", err)
	}

	return hash.String()
}

func TestWatcherCloneAndPull(t *testing.T) {
	tr := setupTestRepo(t)

	workDir := filepath.Join(tr.tmpDir, testWorkDir)
	w := NewWatcher(tr.bareRepoPath, "master", workDir, time.Second)

	ctx := t.Context()
	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	initialCommit := w.LastCommit()
	if initialCommit == "" {
		t.Error("LastCommit should not be empty after clone")
	}

	changed, hash, err := w.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}
	if changed {
		t.Error("Pull returned changed=true, expected false (no new commits)")
	}
	if hash != initialCommit {
		t.Errorf("Pull hash = %s, want %s", hash, initialCommit)
	}

	newCommitHash := tr.addCommit(t, testSecondCommit)

	changed, hash, err = w.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull after new commit failed: %v", err)
	}
	if !changed {
		t.Error("Pull returned changed=false, expected true after new commit")
	}
	if hash != newCommitHash {
		t.Errorf("Pull hash = %s, want %s", hash, newCommitHash)
	}
}

func TestWatcherCloneExistingWorkDir(t *testing.T) {
	tr := setupTestRepo(t)

	workDir := filepath.Join(tr.tmpDir, testWorkDir)
	w := NewWatcher(tr.bareRepoPath, "master", workDir, time.Second)

	ctx := t.Context()
	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	firstCommit := w.LastCommit()

	w2 := NewWatcher(tr.bareRepoPath, "master", workDir, time.Second)
	if err := w2.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	if w2.LastCommit() != firstCommit {
		t.Errorf("LastCommit after reopen = %s, want %s", w2.LastCommit(), firstCommit)
	}
}

func TestWatcherCloneNonGitDir(t *testing.T) {
	tmpDir := t.TempDir()

	workDir := filepath.Join(tmpDir, "not-a-repo")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	testFile := filepath.Join(workDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	w := NewWatcher("https://github.com/example/repo.git", "main", workDir, time.Second)

	err := w.Clone(t.Context())
	if err == nil {
		t.Error("Clone should fail for non-git directory")
	}

	if _, statErr := os.Stat(testFile); os.IsNotExist(statErr) {
		t.Error("Clone deleted existing file in non-git directory")
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file after Clone: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("file content changed, got %q, want %q", string(data), "data")
	}
}

func TestWatcherWatch(t *testing.T) {
	tr := setupTestRepo(t)

	pollInterval := 50 * time.Millisecond
	workDir := filepath.Join(tr.tmpDir, testWorkDir)
	w := NewWatcher(tr.bareRepoPath, "master", workDir, pollInterval)

	ctx := t.Context()

	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	received := make(chan ChangeEvent, 1)
	go w.Watch(ctx, func(event ChangeEvent) {
		select {
		case received <- event:
		default:
		}
	})

	newCommitHash := tr.addCommit(t, testSecondCommit)

	select {
	case event := <-received:
		if event.Commit != newCommitHash {
			t.Errorf("Watch event commit = %s, want %s", event.Commit, newCommitHash)
		}
		if event.Message != testSecondCommit {
			t.Errorf("Watch event message = %q, want %q", event.Message, testSecondCommit)
		}
	case <-time.After(pollInterval*3 + 500*time.Millisecond):
		t.Error("Watch did not receive event within timeout")
	}
}

func TestWatcherWatchBackpressure(t *testing.T) {
	tr := setupTestRepo(t)

	pollInterval := 50 * time.Millisecond
	workDir := filepath.Join(tr.tmpDir, testWorkDir)
	w := NewWatcher(tr.bareRepoPath, "master", workDir, pollInterval)

	ctx := t.Context()

	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	blockHandler := make(chan struct{})
	handlerStarted := make(chan struct{})
	received := make(chan ChangeEvent, 10)

	go w.Watch(ctx, func(event ChangeEvent) {
		select {
		case handlerStarted <- struct{}{}:
		default:
		}
		<-blockHandler
		received <- event
	})

	tr.addCommit(t, "first commit")

	select {
	case <-handlerStarted:
	case <-time.After(pollInterval * 3):
		t.Fatal("handler not started")
	}

	for i := range 5 {
		tr.addCommit(t, fmt.Sprintf("queued commit %d", i))
		time.Sleep(pollInterval / 2)
	}

	time.Sleep(pollInterval * 2)
	close(blockHandler)

	deadline := time.After(pollInterval * 10)

	var events []ChangeEvent
collect:
	for {
		select {
		case event := <-received:
			events = append(events, event)
		case <-deadline:
			break collect
		}
	}

	if len(events) < 1 {
		t.Error("expected at least one event to be processed")
	}
}

func TestWatcherWatchPanicRecovery(t *testing.T) {
	tr := setupTestRepo(t)

	workDir := filepath.Join(tr.tmpDir, testWorkDir)
	w := NewWatcher(tr.bareRepoPath, "master", workDir, 50*time.Millisecond)

	ctx := t.Context()

	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	var callCount atomic.Int32
	received := make(chan ChangeEvent, 2)

	go w.Watch(ctx, func(event ChangeEvent) {
		count := callCount.Add(1)
		if count == 1 {
			panic("test panic")
		}
		received <- event
	})

	tr.addCommit(t, "first commit after clone")

	time.Sleep(100 * time.Millisecond)

	tr.addCommit(t, "second commit after panic")

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Error("Watch stopped processing after handler panic")
	}
}

func TestWatcherIntegrationCloneFromGitHub(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	workDir := filepath.Join(t.TempDir(), "repo")
	w := NewWatcher("https://github.com/octocat/Hello-World.git", "master", workDir, time.Second)

	ctx := t.Context()

	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	if w.LastCommit() == "" {
		t.Error("LastCommit should not be empty after clone")
	}

	if len(w.LastCommit()) != 40 {
		t.Errorf("LastCommit should be 40 char hash, got %d chars", len(w.LastCommit()))
	}

	changed, hash, err := w.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	if changed {
		t.Error("Pull should not detect changes on fresh clone")
	}

	if hash != w.LastCommit() {
		t.Errorf("Pull hash %s != LastCommit %s", hash, w.LastCommit())
	}

	t.Logf("Successfully cloned and pulled from GitHub, commit: %s", hash[:8])
}

func TestWatcherHardResetOnDirtyWorktree(t *testing.T) {
	tr := setupTestRepo(t)

	workDir := filepath.Join(tr.tmpDir, testWorkDir)
	w := NewWatcher(tr.bareRepoPath, "master", workDir, time.Second)

	ctx := t.Context()
	if err := w.Clone(ctx); err != nil {
		t.Fatalf(testCloneFailedFmt, err)
	}

	dirtyFile := filepath.Join(workDir, testFileName)
	if err := os.WriteFile(dirtyFile, []byte("local changes"), 0o644); err != nil {
		t.Fatalf("failed to write dirty file: %v", err)
	}

	tr.addCommit(t, "remote commit")

	changed, _, err := w.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull with dirty worktree failed: %v", err)
	}

	if !changed {
		t.Error("Pull should detect change after remote commit")
	}

	data, err := os.ReadFile(dirtyFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) == "local changes" {
		t.Error("Local changes should be reset after hard reset")
	}
}
