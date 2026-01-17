package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func RunGitCommand(cmdStr string, cwd string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %s, error: %w", string(output), err)
	}

	return strings.TrimSpace(string(output)), nil
}

func StageAllChanges(cwd string) error {
	_, err := RunGitCommand("git add .", cwd)
	return err
}

func GetStagedFiles(cwd string) ([]string, error) {
	output, err := RunGitCommand("git diff --staged --name-only", cwd)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	lines := strings.Split(output, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}

func GetStagedDiffNumstat(cwd string, file string) (string, error) {
	cmd := fmt.Sprintf("git diff --staged --numstat -- %q", file)
	return RunGitCommand(cmd, cwd)
}

func GetStagedDiffUnified(cwd string, file string) (string, error) {
	cmd := fmt.Sprintf("git diff --staged --unified=0 -- %q", file)
	return RunGitCommand(cmd, cwd)
}

func CommitWithMessage(cwd string, message string) error {
	tmpFile, err := os.CreateTemp("", "commit-msg-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(message); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	cmd := fmt.Sprintf("git commit -F %q", tmpFile.Name())
	_, err = RunGitCommand(cmd, cwd)
	return err
}

func PushChanges(cwd string) error {
	_, err := RunGitCommand("git push", cwd)
	return err
}

func GetRepoRoot(cwd string) (string, error) {
	return RunGitCommand("git rev-parse --show-toplevel", cwd)
}

func DiscoverRepositories(root string) []string {
	var repos []string

	// 1. Check if we are inside a git repo
	if toplevel, err := GetRepoRoot(root); err == nil {
		if absToplevel, err := filepath.Abs(toplevel); err == nil {
			repos = append(repos, absToplevel)
		}
	}

	// 2. Search for sub-repositories
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		name := info.Name()
		if name == ".git" || name == "node_modules" || name == "target" {
			return filepath.SkipDir
		}

		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			if absPath, err := filepath.Abs(path); err == nil {
				repos = append(repos, absPath)
			}
		}
		return nil
	})

	sort.Strings(repos)
	if len(repos) > 0 {
		uniqueRepos := make([]string, 0, len(repos))
		seen := make(map[string]bool)
		for _, repo := range repos {
			if !seen[repo] {
				seen[repo] = true
				uniqueRepos = append(uniqueRepos, repo)
			}
		}
		repos = uniqueRepos
	}

	return repos
}
