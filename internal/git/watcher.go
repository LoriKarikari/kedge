package git

import (
	"context"
	"errors"
	"fmt"
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
			return fmt.Errorf("failed to open existing repo at %s: %w", w.workDir, err)
		}
		w.repo = repo
		_, _, err = w.Pull(ctx)
		return err
	}

	return w.clone(ctx)
}

func (w *Watcher) clone(ctx context.Context) error {
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

func (w *Watcher) Pull(ctx context.Context) (changed bool, hash string, err error) {
	worktree, err := w.repo.Worktree()
	if err != nil {
		return false, "", err
	}

	prevCommit := w.LastCommit()

	err = worktree.PullContext(ctx, &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(w.branch),
		SingleBranch:  true,
	})

	if err == nil || errors.Is(err, git.NoErrAlreadyUpToDate) {
		if err := w.updateLastCommit(); err != nil {
			return false, "", err
		}
		newCommit := w.LastCommit()
		return newCommit != prevCommit, newCommit, nil
	}

	if isRecoverableError(err) {
		if resetErr := w.hardReset(ctx); resetErr != nil {
			slog.Warn("hard reset failed", "error", resetErr)
			return false, "", err
		}
		if err := w.updateLastCommit(); err != nil {
			return false, "", err
		}
		newCommit := w.LastCommit()
		return newCommit != prevCommit, newCommit, nil
	}

	return false, "", err
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
			return fmt.Errorf("branch not found: %s", w.branch)
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

func (w *Watcher) Watch(ctx context.Context, onChange func(ChangeEvent)) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	const eventQueueSize = 16
	events := make(chan ChangeEvent, eventQueueSize)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							slog.Error("panic in onChange handler", "error", r)
						}
					}()
					onChange(event)
				}()
			}
		}
	}()

	defer func() {
		close(events)
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			changed, hash, err := w.Pull(ctx)
			if err != nil {
				slog.Error("failed to pull", "error", err)
				continue
			}

			if changed {
				event := ChangeEvent{
					Commit:    hash,
					Timestamp: time.Now(),
					Message:   w.getCommitMessage(hash),
				}
				enqueued := false
				for !enqueued {
					select {
					case events <- event:
						enqueued = true
					case <-ctx.Done():
						return
					case <-time.After(w.pollInterval):
						slog.Warn("event queue full; waiting for handler", "commit", hash)
					}
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
	w.mu.RLock()
	h := plumbing.NewHash(hash)
	commit, err := w.repo.CommitObject(h)
	w.mu.RUnlock()
	if err != nil {
		return ""
	}
	return commit.Message
}
