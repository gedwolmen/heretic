package team

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
)

// WorktreeManager creates a per-member git worktree.
type WorktreeManager struct {
	teamDir string
	mu      sync.Mutex
}

// NewWorktreeManager returns a manager rooted at the team's directory.
func NewWorktreeManager(teamDir string) *WorktreeManager {
	return &WorktreeManager{teamDir: teamDir}
}

// WorktreeDir returns the path for a member's worktree.
func (w *WorktreeManager) WorktreeDir(member string) string {
	return filepath.Join(w.teamDir, "worktrees", member)
}

// shortHash returns a deterministic short hash from a string (used
// for branch names).
func shortHash(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])[:8]
}

// BranchName returns the git branch name for a team + member.
func (w *WorktreeManager) BranchName(team, member string) string {
	return fmt.Sprintf("team/%s/%s-%s", team, member, shortHash(member))
}

// Create creates a git worktree for a member.
// repoDir is the existing repo to fork from; baseBranch is the
// starting ref (defaults to "main" if empty).
func (w *WorktreeManager) Create(repoDir, team, member, baseBranch string) error {
	if baseBranch == "" {
		baseBranch = "main"
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	wt := w.WorktreeDir(member)
	branch := w.BranchName(team, member)
	cmd := exec.Command("git", "worktree", "add", "-b", branch, wt, baseBranch)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add failed: %w: %s", err, string(out))
	}
	return nil
}

// Remove deletes a member's worktree.
func (w *WorktreeManager) Remove(repoDir, member string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	wt := w.WorktreeDir(member)
	cmd := exec.Command("git", "worktree", "remove", "--force", wt)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// Fall back to plain RemoveAll — the worktree may not be
		// registered with git (e.g. if git failed mid-create).
		if rmErr := exec.Command("rm", "-rf", wt).Run(); rmErr != nil {
			return fmt.Errorf("worktree remove failed: %w (%s) and rm -rf failed: %v", err, string(out), rmErr)
		}
	}
	return nil
}

// ErrGitNotAvailable is returned when the git binary is not on PATH.
var ErrGitNotAvailable = errors.New("team: git binary not available")

// EnsureGit verifies the git binary is available.
func EnsureGit() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return ErrGitNotAvailable
	}
	return nil
}
