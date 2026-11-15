package control

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func sanitizePathToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "task"
	}
	return out
}

func isGitRepo(root string) bool {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// prepareTaskWorktree creates a detached git worktree under <root>/.axis/worktrees/<taskID>.
// Returns the worktree directory path, or an error if git is unavailable / root is not a git repo.
func prepareTaskWorktree(root, taskID string) (string, error) {
	if !isGitRepo(root) {
		return "", fmt.Errorf("project root is not a git repository")
	}
	wtDir := filepath.Join(root, ".axis", "worktrees", sanitizePathToken(taskID))
	if err := os.MkdirAll(filepath.Dir(wtDir), 0o755); err != nil {
		return "", err
	}
	if _, err := os.Stat(wtDir); err == nil {
		return wtDir, nil
	}
	cmd := exec.Command("git", "-C", root, "worktree", "add", "--detach", wtDir, "HEAD")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git worktree add failed: %v (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return wtDir, nil
}

// captureWorktreeDiff returns a compact git diff + untracked status summary inside wtDir.
func captureWorktreeDiff(wtDir string) string {
	if wtDir == "" {
		return ""
	}
	diffCmd := exec.Command("git", "-C", wtDir, "diff", "--no-ext-diff", "HEAD")
	diffOut, _ := diffCmd.Output()

	statusCmd := exec.Command("git", "-C", wtDir, "status", "--short")
	statusOut, _ := statusCmd.Output()

	var combined strings.Builder
	if s := strings.TrimSpace(string(statusOut)); s != "" {
		combined.WriteString("# status\n")
		combined.WriteString(s)
		combined.WriteString("\n")
	}
	if d := strings.TrimSpace(string(diffOut)); d != "" {
		if combined.Len() > 0 {
			combined.WriteString("\n")
		}
		combined.WriteString(d)
	}
	res := combined.String()
	const maxDiffBytes = 32 * 1024
	if len(res) > maxDiffBytes {
		res = res[:maxDiffBytes] + "\n...(diff truncated)"
	}
	return res
}

// cleanupTaskWorktree removes a previously created git worktree for taskID.
func cleanupTaskWorktree(root, taskID string) error {
	wtDir := filepath.Join(root, ".axis", "worktrees", sanitizePathToken(taskID))
	if _, err := os.Stat(wtDir); os.IsNotExist(err) {
		return nil
	}
	if isGitRepo(root) {
		cmd := exec.Command("git", "-C", root, "worktree", "remove", "--force", wtDir)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return os.RemoveAll(wtDir)
}
