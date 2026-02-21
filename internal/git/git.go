package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func RunGitCommand(cwd string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %s, error: %w", strings.Join(args, " "), string(output), err)
	}

	return strings.TrimSpace(string(output)), nil
}

func StageAllChanges(cwd string) error {
	_, err := RunGitCommand(cwd, "add", ".")
	return err
}

func GetStagedFiles(cwd string) ([]string, error) {
	output, err := RunGitCommand(cwd, "diff", "--staged", "--name-only")
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

func GetStagedDiff(cwd string, file string) (string, error) {
	// Get the actual diff content for a file
	return RunGitCommand(cwd, "diff", "--staged", "--unified=3", "--", file)
}

func GetStagedDiffNumstat(cwd string, file string) (string, error) {
	return RunGitCommand(cwd, "diff", "--staged", "--numstat", "--", file)
}

func GetStagedDiffUnified(cwd string, file string) (string, error) {
	return RunGitCommand(cwd, "diff", "--staged", "--unified=0", "--", file)
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

	_, err = RunGitCommand(cwd, "commit", "-F", tmpFile.Name())
	return err
}

func PushChanges(cwd string) error {
	_, err := RunGitCommand(cwd, "push")
	return err
}

func GetRepoRoot(cwd string) (string, error) {
	return RunGitCommand(cwd, "rev-parse", "--show-toplevel")
}

func GetCurrentBranch(cwd string) (string, error) {
	return RunGitCommand(cwd, "rev-parse", "--abbrev-ref", "HEAD")
}

func GetLocalIdentity(cwd string) (string, string) {
	email, _ := RunGitCommand(cwd, "config", "user.email")
	name, _ := RunGitCommand(cwd, "config", "user.name")
	return name, email
}

func GetHistoryIdentity(cwd string) (string, string) {
	// Get the last committer email
	email, _ := RunGitCommand(cwd, "log", "-1", "--pretty=format:%ae")
	name, _ := RunGitCommand(cwd, "log", "-1", "--pretty=format:%an")
	return name, email
}

func GetRemoteOwner(cwd string) string {
	url, err := RunGitCommand(cwd, "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	// Parse git@github.com:owner/repo.git or https://github.com/owner/repo.git
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	if len(parts) > 1 {
		ownerPart := parts[len(parts)-2]
		if strings.Contains(ownerPart, ":") {
			subParts := strings.Split(ownerPart, ":")
			return subParts[len(subParts)-1]
		}
		return ownerPart
	}
	return ""
}

func GetRepoName(cwd string) string {
	url, err := RunGitCommand(cwd, "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	// Parse git@github.com:owner/repo.git or https://github.com/owner/repo.git
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func SyncFork(cwd string, target string) error {
	cmd := exec.Command("gh", "repo", "sync", target)
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh repo sync failed: %s, error: %w", string(output), err)
	}
	return nil
}

func SyncLocalConfig(cwd string, name string, email string) error {
	if name != "" {
		_, err := RunGitCommand(cwd, "config", "--local", "user.name", name)
		if err != nil {
			return err
		}
	}
	if email != "" {
		_, err := RunGitCommand(cwd, "config", "--local", "user.email", email)
		if err != nil {
			return err
		}
	}
	return nil
}

func DiscoverRepositories(roots string) []string {
	var repos []string
	rootList := strings.Split(roots, ",")

	for _, root := range rootList {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}

		// 1. Check if we are inside a git repo
		if toplevel, err := GetRepoRoot(root); err == nil {
			if absToplevel, err := filepath.Abs(toplevel); err == nil {
				repos = append(repos, absToplevel)
			}
		}

		// 2. Search for sub-repositories (only if not already in one)
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			name := d.Name()
			if name == ".git" {
				if absPath, err := filepath.Abs(filepath.Dir(path)); err == nil {
					repos = append(repos, absPath)
				}
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "target" || name == ".venv" || name == "vendor" || name == "dist" || name == "out" || name == "bin" || name == "obj" {
				return filepath.SkipDir
			}

			return nil
		})
	}

	sort.Strings(repos)
	if len(repos) > 1 {
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
