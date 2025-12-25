package git

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type ChangeEvent struct {
	Commit    string
	Message   string
	Timestamp time.Time
}

type Watcher struct {
	repoURL      string
	branch       string
	workDir      string
	pollInterval time.Duration
	repo         *git.Repository

	mu         sync.RWMutex
	lastCommit string
}

func NewWatcher(repoURL, branch, workDir string, pollInterval time.Duration) *Watcher {
	return &Watcher{
		repoURL:      repoURL,
		branch:       branch,
		workDir:      workDir,
		pollInterval: pollInterval,
	}
}

func (w *Watcher) Clone(ctx context.Context) error {
	if _, err := os.Stat(w.workDir); err == nil {
		repo, err := git.PlainOpen(w.workDir)
		if err != nil {
			return err
		}
		w.repo = repo
		return w.Pull(ctx)
	}

	repo, err := git.PlainCloneContext(ctx, w.workDir, false, &git.CloneOptions{
		URL:           w.repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(w.branch),
		SingleBranch:  true,
	})
	if err != nil {
		return err
	}

	w.repo = repo
	return w.updateLastCommit()
}

func (w *Watcher) Pull(ctx context.Context) error {
	worktree, err := w.repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.PullContext(ctx, &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(w.branch),
		SingleBranch:  true,
	})

	if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
		return w.updateLastCommit()
	}

	if isRecoverableError(err) {
		if resetErr := w.hardReset(ctx); resetErr != nil {
			slog.Warn("hard reset failed", "error", resetErr)
			return err
		}
		return w.updateLastCommit()
	}

	return err
}

func isRecoverableError(err error) bool {
	return errors.Is(err, git.ErrNonFastForwardUpdate) ||
		errors.Is(err, git.ErrUnstagedChanges) ||
		errors.Is(err, git.ErrWorktreeNotClean)
}

func (w *Watcher) hardReset(ctx context.Context) error {
	if err := w.repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	ref, err := w.repo.Reference(plumbing.NewRemoteReferenceName("origin", w.branch), true)
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return errors.New("branch not found: " + w.branch)
		}
		return err
	}

	worktree, err := w.repo.Worktree()
	if err != nil {
		return err
	}

	return worktree.Reset(&git.ResetOptions{
		Commit: ref.Hash(),
		Mode:   git.HardReset,
	})
}

// Watch polls the repository and sends change events.
// The caller must provide a buffered channel and keep consuming from it.
// The caller must NOT close onChange; cancel ctx instead.
func (w *Watcher) Watch(ctx context.Context, onChange chan<- ChangeEvent) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			prevCommit := w.LastCommit()
			if err := w.Pull(ctx); err != nil {
				slog.Error("failed to pull", "error", err)
				continue
			}

			currentCommit := w.LastCommit()
			if currentCommit != prevCommit {
				event := ChangeEvent{
					Commit:    currentCommit,
					Timestamp: time.Now(),
					Message:   w.getCommitMessage(currentCommit),
				}

				select {
				case onChange <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (w *Watcher) LastCommit() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastCommit
}

func (w *Watcher) updateLastCommit() error {
	ref, err := w.repo.Head()
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.lastCommit = ref.Hash().String()
	w.mu.Unlock()
	return nil
}

func (w *Watcher) getCommitMessage(hash string) string {
	h := plumbing.NewHash(hash)
	commit, err := w.repo.CommitObject(h)
	if err != nil {
		return ""
	}
	return commit.Message
}
