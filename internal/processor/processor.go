package processor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/nathfavour/autocommiter.go/internal/api"
	"github.com/nathfavour/autocommiter.go/internal/auth"
	"github.com/nathfavour/autocommiter.go/internal/config"
	"github.com/nathfavour/autocommiter.go/internal/git"
	"github.com/nathfavour/autocommiter.go/internal/gitmoji"
	"github.com/nathfavour/autocommiter.go/internal/summarizer"
)

func GenerateCommit(repoPath string, noPush bool, force bool) error {
	startDir := repoPath
	if startDir == "" {
		startDir = "."
	}

	color.Cyan("🪄 Autocommiter: Discovering repositories...")
	repos := git.DiscoverRepositories(startDir)

	if len(repos) == 0 {
		return fmt.Errorf("no git repositories found in %s", startDir)
	}

	if len(repos) > 1 {
		color.Green("✓ Found %d repositories\n", len(repos))
	}

	for _, repo := range repos {
		if err := ProcessSingleRepo(repo, noPush, force); err != nil {
			color.Red("✗ Error processing %s: %v\n", repo, err)
		}
	}

	color.New(color.FgGreen, color.Bold).Println("✨ All done!")
	return nil
}

func SetupUser(repoPath string, user string) error {
	if user == "" {
		return nil
	}

	startDir := repoPath
	if startDir == "" {
		startDir = "."
	}

	repos := git.DiscoverRepositories(startDir)
	if len(repos) == 0 {
		return fmt.Errorf("no git repositories found in %s", startDir)
	}

	for _, repoRoot := range repos {
		color.Cyan("👤 Setting default user for %s: %s", repoRoot, user)

		// 1. Switch Account
		if err := auth.SwitchAccount(user); err != nil {
			return fmt.Errorf("failed to switch to account %s: %v", user, err)
		}

		// 2. Save to local config
		repoCfg, _ := config.LoadMergedConfig(repoRoot)
		repoCfg.DefaultUser = &user
		if err := config.SaveRepoConfig(repoRoot, repoCfg); err != nil {
			return fmt.Errorf("failed to save repo config: %v", err)
		}

		// 3. Sync Git Config
		preferNoReply := true
		if repoCfg.PreferNoReplyEmail != nil {
			preferNoReply = *repoCfg.PreferNoReplyEmail
		}

		name, email, err := auth.GetAccountIdentity(preferNoReply)
		if err != nil {
			color.Yellow("⚠️ Could not fetch identity for %s: %v", user, err)
		} else {
			if err := git.SyncLocalConfig(repoRoot, name, email); err != nil {
				color.Yellow("⚠️ Could not sync git config: %v", err)
			}
		}
		color.Green("✓ User %s is now the default for this repository", user)
	}

	return nil
}

func ProcessSingleRepo(repoRoot string, noPush bool, force bool) error {
	color.Cyan("📂 Repository: %s", color.New(color.Bold).Sprint(repoRoot))

	// Start account discovery in background
	accMgr := NewAccountManager(repoRoot)
	accMgr.StartDiscovery()

	// Ensure gitignore safety
	color.Cyan("🛡️ Ensure .gitignore safety...")
	if err := EnsureGitignoreSafety(repoRoot); err != nil {
		return err
	}

	// Check for staged files first
	stagedFiles, err := git.GetStagedFiles(repoRoot)
	if err != nil {
		return err
	}

	if len(stagedFiles) == 0 {
		// Stage changes only if nothing is staged
		color.Cyan("📦 No changes staged. Staging all changes...")
		if err := git.StageAllChanges(repoRoot); err != nil {
			return err
		}
		
		stagedFiles, err = git.GetStagedFiles(repoRoot)
		if err != nil {
			return err
		}
	} else {
		color.Green("✓ Using %d already staged files", len(stagedFiles))
	}

	if len(stagedFiles) == 0 {
		color.Yellow("ℹ️ No changes to commit — Autocommit skipped.\n")
		return nil
	}
	for _, file := range stagedFiles {
		color.New(color.Faint).Printf("  - %s\n", file)
	}

	// Generate message
	message, err := GenerateMessage(repoRoot, accMgr)
	if err != nil {
		return err
	}
	color.Cyan("💬 Message: %s", color.New(color.Italic).Sprint(message))

	// Ask for confirmation
	cfg, _ := config.LoadMergedConfig(repoRoot)
	skipConf := false
	if cfg.SkipConfirmation != nil {
		skipConf = *cfg.SkipConfirmation
	}

	if !force && !skipConf {
		fmt.Print(color.CyanString("\n🤔 Proceed with commit? (y/n): "))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(input), "y") {
			color.Red("❌ Cancelled.\n")
			return nil
		}
	}

	// Commit
	color.Cyan("✍️ Committing changes...")
	if err := git.CommitWithMessage(repoRoot, message); err != nil {
		return err
	}
	color.Green("✓ Commit successful!")

	// Push
	if !noPush {
		color.Cyan("🚀 Pushing to remote...")
		if err := PushWithRetry(repoRoot, accMgr); err != nil {
			return err
		}
		color.Green("✓ Push successful!")
	}

	color.Green("✓ Done with this repository!\n")
	return nil
}

func PushWithRetry(repoRoot string, accMgr *AccountManager) error {
	err := git.PushChanges(repoRoot)
	if err == nil {
		return nil
	}

	// If it's not an auth error, just return it
	errMsg := err.Error()
	if !strings.Contains(errMsg, "403") && !strings.Contains(errMsg, "401") && !strings.Contains(errMsg, "Permission denied") && !strings.Contains(errMsg, "Authentication failed") {
		return err
	}

	color.Yellow("⚠️ Push failed with authentication error. Attempting reactive account switch...")

	// Reactive strategy: try other accounts
	accounts, listErr := auth.ListAccounts()
	if listErr != nil || len(accounts) <= 1 {
		return err // Give up if we can't list or only have 1
	}

	activeUser := auth.GetGithubUser()
	for _, acc := range accounts {
		if acc == activeUser {
			continue
		}

		color.Cyan("🔄 Trying account: %s...", acc)
		if switchErr := auth.SwitchAccount(acc); switchErr == nil {
			// Try fetching identity and sync config for this account too
			cfg, _ := config.LoadMergedConfig(repoRoot)
			preferNoReply := true
			if cfg.PreferNoReplyEmail != nil {
				preferNoReply = *cfg.PreferNoReplyEmail
			}

			if name, email, identErr := auth.GetAccountIdentity(preferNoReply); identErr == nil {
				_ = git.SyncLocalConfig(repoRoot, name, email)
			}

			if retryErr := git.PushChanges(repoRoot); retryErr == nil {
				// Success! Cache this for next time
				cfg, _ := config.LoadMergedConfig(repoRoot)
				preferNoReply := true
				if cfg.PreferNoReplyEmail != nil {
					preferNoReply = *cfg.PreferNoReplyEmail
				}
				if name, email, identErr := auth.GetAccountIdentity(preferNoReply); identErr == nil {
					accMgr.CacheAccount(acc, email, name)
				}
				return nil
			}
		}
	}

	// If we tried everything and still failed, switch back to original or just return last error
	_ = auth.SwitchAccount(activeUser)
	return err
}

func GenerateMessage(repoRoot string, accMgr *AccountManager) (string, error) {
	cfg, _ := config.LoadMergedConfig(repoRoot)

	apiKey := ""
	if cfg.APIKey != nil {
		apiKey = *cfg.APIKey
	}

	// Wait for account discovery to finish
	if accMgr != nil {
		if err := accMgr.Wait(); err == nil {
			if err := accMgr.Sync(); err != nil {
				color.Yellow("⚠️ Account sync warning: %v", err)
			}
		}
	}

	token := auth.GetToken(apiKey)
	if token == "" {
		return "", fmt.Errorf("authentication failed: please run 'gh auth login' or use 'autocommiter set-api-key'")
	}

	return TryAPIGeneration(repoRoot, token, cfg)
}

func TryAPIGeneration(repoRoot string, apiKey string, cfg config.Config) (string, error) {
	model := "gpt-4o-mini"
	if cfg.SelectedModel != nil {
		model = *cfg.SelectedModel
	}

	color.New(color.FgCyan).Fprint(os.Stderr, "🤖 Generating with model: ")
	color.New(color.FgCyan, color.Faint).Fprintln(os.Stderr, model, "...")

	branch, _ := git.GetCurrentBranch(repoRoot)
	fileChanges, err := summarizer.BuildFileChanges(repoRoot)
	if err != nil {
		return "", err
	}

	var fileNamesList []string
	for i, fc := range fileChanges {
		if i >= 100 { // Increased from 50
			break
		}
		fileNamesList = append(fileNamesList, fc.File)
	}
	fileNames := strings.Join(fileNamesList, "\n")
	
	// Increased limit from 400 to 12000 to give the LLM much more context
	compressedJSON := summarizer.CompressToJSON(fileChanges, 12000)

	message, err := api.GenerateCommitMessage(apiKey, branch, fileNames, compressedJSON, model)
	if err != nil {
		return "", err
	}

	enableGitmoji := false
	if cfg.EnableGitmoji != nil {
		enableGitmoji = *cfg.EnableGitmoji
	}

	if enableGitmoji {
		message = gitmoji.GetGitmojifiedMessage(message)
	}

	return message, nil
}

func GetSummarizedChanges(repoRoot string) (string, error) {
	fileChanges, err := summarizer.BuildFileChanges(repoRoot)
	if err != nil {
		return "", err
	}
	// Default to a large maxLen because the extension can handle it
	return summarizer.CompressToJSON(fileChanges, 12000), nil
}

func FixLastCommit(repoRoot string, targetUser string) error {
	if targetUser == "" {
		return fmt.Errorf("please specify a user with --user")
	}

	color.Cyan("🔧 Repairing last commit for: %s", targetUser)

	// 1. Switch Account
	if err := auth.SwitchAccount(targetUser); err != nil {
		return fmt.Errorf("failed to switch to account %s: %v", targetUser, err)
	}

	// 2. Get Identity
	cfg, _ := config.LoadMergedConfig(repoRoot)
	preferNoReply := true
	if cfg.PreferNoReplyEmail != nil {
		preferNoReply = *cfg.PreferNoReplyEmail
	}

	name, email, err := auth.GetAccountIdentity(preferNoReply)
	if err != nil {
		return fmt.Errorf("failed to get identity for %s: %v", targetUser, err)
	}

	// 3. Sync Local Config
	if err := git.SyncLocalConfig(repoRoot, name, email); err != nil {
		return fmt.Errorf("failed to sync git config: %v", err)
	}

	// 4. Amend Commit
	color.Cyan("✍️ Amending commit author...")
	authorStr := fmt.Sprintf("%s <%s>", name, email)
	_, err = git.RunGitCommand(repoRoot, "commit", "--amend", "--no-edit", "--author", authorStr)
	if err != nil {
		return fmt.Errorf("failed to amend commit: %v", err)
	}

	// 5. Force Push (Safety first)
	color.Yellow("⚠️ Force-pushing changes to remote...")
	_, err = git.RunGitCommand(repoRoot, "push", "--force-with-lease")
	if err != nil {
		return fmt.Errorf("failed to push changes: %v", err)
	}

	color.Green("✨ Commit repaired successfully!")
	return nil
}

func EnsureGitignoreSafety(repoRoot string) error {
	cfg, _ := config.LoadMergedConfig(repoRoot)
	shouldUpdate := false
	if cfg.UpdateGitignore != nil {
		shouldUpdate = *cfg.UpdateGitignore
	}

	if !shouldUpdate {
		return nil
	}

	gitignorePath := filepath.Join(repoRoot, ".gitignore")
	existing, _ := os.ReadFile(gitignorePath)
	content := string(existing)

	lines := strings.Split(content, "\n")
	lineMap := make(map[string]bool)
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			lineMap[trimmed] = true
		}
	}

	var toAppend []string
	for _, pattern := range cfg.GitignorePatterns {
		if !lineMap[pattern] {
			toAppend = append(toAppend, fmt.Sprintf("# Added by Autocommiter: ensure %s", pattern))
			toAppend = append(toAppend, pattern)
		}
	}

	if len(toAppend) > 0 {
		newContent := content
		if !strings.HasSuffix(content, "\n") && content != "" {
			newContent += "\n"
		}
		newContent += strings.Join(toAppend, "\n") + "\n"
		return os.WriteFile(gitignorePath, []byte(newContent), 0644)
	}

	return nil
}
