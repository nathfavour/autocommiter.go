package processor

import (
	"strings"
	"time"

	"github.com/nathfavour/autocommiter.go/internal/auth"
	"github.com/nathfavour/autocommiter.go/internal/config"
	"github.com/nathfavour/autocommiter.go/internal/git"
	"github.com/nathfavour/autocommiter.go/internal/index"
)

type AccountManager struct {
	repoRoot      string
	result        chan error
	TargetAccount string
	TargetEmail   string
	TargetName    string
	IsSingle      bool
}

func NewAccountManager(repoRoot string) *AccountManager {
	return &AccountManager{
		repoRoot: repoRoot,
		result:   make(chan error, 1),
	}
}

func (m *AccountManager) StartDiscovery() {
	go func() {
		m.result <- m.discover()
	}()
}

func (m *AccountManager) Wait() error {
	return <-m.result
}

func (m *AccountManager) discover() error {
	// 1. Check for Default User in index DB
	if defUser, err := index.GetDefaultUser(m.repoRoot); err == nil && defUser != "" {
		m.TargetAccount = defUser
		return nil
	}

	// 2. Fast-Exit Sentinel
	if index.HasSingleAccountSentinel() {
		m.IsSingle = true
		return nil
	}

	// 2. Directory Gravity (High Confidence)
	// Check if any parent directory name matches a logged-in account
	accounts, err := auth.ListAccounts()
	if err != nil {
		return err
	}
	if len(accounts) <= 1 {
		_ = index.SetSingleAccountSentinel(true)
		m.IsSingle = true
		return nil
	}

	for _, acc := range accounts {
		if strings.Contains(m.repoRoot, "/"+acc+"/") || strings.HasSuffix(m.repoRoot, "/"+acc) {
			m.TargetAccount = acc
			// If we find a gravity match, we are confident enough to brute
			return nil
		}
	}

	// 3. Affinity Mapping (Normal Path)
	db, err := index.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()
    
	repoHash := index.GetRepoHash(m.repoRoot)

	// 3.1 SQLite Cache
	var cachedAccount, cachedEmail, cachedName string
	err = db.QueryRow("SELECT account_handle, email, name FROM repo_cache WHERE repo_path_hash = ?", repoHash).Scan(&cachedAccount, &cachedEmail, &cachedName)
	if err == nil {
		m.TargetAccount = cachedAccount
		m.TargetEmail = cachedEmail
		m.TargetName = cachedName
		return nil
	}

	// 3.2 Local Git Config/History
	_, localEmail := git.GetLocalIdentity(m.repoRoot)
	_, histEmail := git.GetHistoryIdentity(m.repoRoot)
	
	owner := git.GetRemoteOwner(m.repoRoot)

	// Try to match emails/owner to an account
	activeUser := auth.GetGithubUser()
	
	// Heuristic: If we can't find a strong match, we'll stay with activeUser
	m.TargetAccount = activeUser

	// If owner matches one of the accounts, that's a strong signal
	for _, acc := range accounts {
		if stringsEqual(acc, owner) {
			m.TargetAccount = acc
			break
		}
	}

	// But history is stronger
	for _, acc := range accounts {
		// This is a bit of a leap, but if account handle is in email, it's a match
		if stringsEqual(acc, localEmail) || stringsEqual(acc, histEmail) || 
		   (localEmail != "" && contains(localEmail, acc)) {
			m.TargetAccount = acc
			break
		}
	}

	return nil
}

func (m *AccountManager) Sync() error {
	if m.IsSingle || m.TargetAccount == "" {
		return nil
	}

	activeUser := auth.GetGithubUser()
	if activeUser != m.TargetAccount {
		if err := auth.SwitchAccount(m.TargetAccount); err != nil {
			return err
		}
	}

	// If we don't have email/name (not in cache), fetch them
	if m.TargetEmail == "" {
		cfg, _ := config.LoadMergedConfig(m.repoRoot)
		preferNoReply := true
		if cfg.PreferNoReplyEmail != nil {
			preferNoReply = *cfg.PreferNoReplyEmail
		}

		name, email, login, err := auth.GetAccountIdentity(preferNoReply)
		if err == nil {
			m.TargetEmail = email
			m.TargetName = name
			m.TargetAccount = login // Ensure we use the actual handle
			
			// Cache it
			db, err := index.InitDB()
			if err == nil {
				defer db.Close()
				repoHash := index.GetRepoHash(m.repoRoot)
				_, _ = db.Exec("INSERT OR REPLACE INTO repo_cache (repo_path_hash, account_handle, email, name, last_used) VALUES (?, ?, ?, ?, ?)",
					repoHash, m.TargetAccount, m.TargetEmail, m.TargetName, time.Now().Unix())
			}
		}
	}

	// Sync local git config if it differs
	_, currentEmail := git.GetLocalIdentity(m.repoRoot)
	if m.TargetEmail != "" && currentEmail != m.TargetEmail {
		return git.SyncLocalConfig(m.repoRoot, m.TargetName, m.TargetEmail)
	}

	return nil
}

func (m *AccountManager) CacheAccount(account, email, name string) {
	db, err := index.InitDB()
	if err == nil {
		defer db.Close()
		repoHash := index.GetRepoHash(m.repoRoot)
		_, _ = db.Exec("INSERT OR REPLACE INTO repo_cache (repo_path_hash, account_handle, email, name, last_used) VALUES (?, ?, ?, ?, ?)",
			repoHash, account, email, name, time.Now().Unix())
	}
}

func stringsEqual(a, b string) bool {
	return (a != "" && b != "") && (a == b)
}

func contains(s, substr string) bool {
	return (s != "" && substr != "") && (len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
