package processor

import (
	"crypto/sha256"
	"fmt"
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
	targetAccount string
	targetEmail   string
	targetName    string
	isSingle      bool
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
	// 1. Fast-Exit Sentinel
	if index.HasSingleAccountSentinel() {
		m.isSingle = true
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
		m.isSingle = true
		return nil
	}

	for _, acc := range accounts {
		if strings.Contains(m.repoRoot, "/"+acc+"/") || strings.HasSuffix(m.repoRoot, "/"+acc) {
			m.targetAccount = acc
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
    
    // ... rest of existing discovery logic (Cache -> Config -> History)

	repoHash := fmt.Sprintf("%x", sha256.Sum256([]byte(m.repoRoot)))

	// 3.1 SQLite Cache
	var cachedAccount, cachedEmail, cachedName string
	err = db.QueryRow("SELECT account_handle, email, name FROM repo_cache WHERE repo_path_hash = ?", repoHash).Scan(&cachedAccount, &cachedEmail, &cachedName)
	if err == nil {
		m.targetAccount = cachedAccount
		m.targetEmail = cachedEmail
		m.targetName = cachedName
		return nil
	}

	// 3.2 Local Git Config/History
	_, localEmail := git.GetLocalIdentity(m.repoRoot)
	_, histEmail := git.GetHistoryIdentity(m.repoRoot)
	
	owner := git.GetRemoteOwner(m.repoRoot)

	// Try to match emails/owner to an account
	activeUser := auth.GetGithubUser()
	
	// Heuristic: If we can't find a strong match, we'll stay with activeUser
	m.targetAccount = activeUser

	// If owner matches one of the accounts, that's a strong signal
	for _, acc := range accounts {
		if stringsEqual(acc, owner) {
			m.targetAccount = acc
			break
		}
	}

	// But history is stronger
	for _, acc := range accounts {
		// This is a bit of a leap, but if account handle is in email, it's a match
		if stringsEqual(acc, localEmail) || stringsEqual(acc, histEmail) || 
		   (localEmail != "" && contains(localEmail, acc)) {
			m.targetAccount = acc
			break
		}
	}

	return nil
}

func (m *AccountManager) Sync() error {
	if m.isSingle || m.targetAccount == "" {
		return nil
	}

	activeUser := auth.GetGithubUser()
	if activeUser != m.targetAccount {
		if err := auth.SwitchAccount(m.targetAccount); err != nil {
			return err
		}
	}

	// If we don't have email/name (not in cache), fetch them
	if m.targetEmail == "" {
		cfg, _ := config.LoadMergedConfig(m.repoRoot)
		preferNoReply := true
		if cfg.PreferNoReplyEmail != nil {
			preferNoReply = *cfg.PreferNoReplyEmail
		}

		name, email, err := auth.GetAccountIdentity(preferNoReply)
		if err == nil {
			m.targetEmail = email
			m.targetName = name
			
			// Cache it
			db, err := index.InitDB()
			if err == nil {
				defer db.Close()
				repoHash := fmt.Sprintf("%x", sha256.Sum256([]byte(m.repoRoot)))
				_, _ = db.Exec("INSERT OR REPLACE INTO repo_cache (repo_path_hash, account_handle, email, name, last_used) VALUES (?, ?, ?, ?, ?)",
					repoHash, m.targetAccount, m.targetEmail, m.targetName, time.Now().Unix())
			}
		}
	}

	// Sync local git config if it differs
	_, currentEmail := git.GetLocalIdentity(m.repoRoot)
	if m.targetEmail != "" && currentEmail != m.targetEmail {
		return git.SyncLocalConfig(m.repoRoot, m.targetName, m.targetEmail)
	}

	return nil
}

func (m *AccountManager) CacheAccount(account, email, name string) {
	db, err := index.InitDB()
	if err == nil {
		defer db.Close()
		repoHash := fmt.Sprintf("%x", sha256.Sum256([]byte(m.repoRoot)))
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
