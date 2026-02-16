package index

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nathfavour/autocommiter.go/internal/config"
	_ "modernc.org/sqlite"
)

func GetDBPath() (string, error) {
	dir, err := config.GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "index.db"), nil
}

func GetSentinelPath() (string, error) {
	dir, err := config.GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".single_account"), nil
}

func HasSingleAccountSentinel() bool {
	path, err := GetSentinelPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func SetSingleAccountSentinel(isSingle bool) error {
	path, err := GetSentinelPath()
	if err != nil {
		return err
	}
	if isSingle {
		return os.WriteFile(path, []byte("true"), 0644)
	}
	return os.Remove(path)
}

func InitDB() (*sql.DB, error) {
	path, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	// Performance and reliability optimizations
	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA synchronous=NORMAL;")

	schema := `
	CREATE TABLE IF NOT EXISTS repo_cache (
		repo_path_hash TEXT PRIMARY KEY,
		account_handle TEXT,
		email TEXT,
		name TEXT,
		last_used INTEGER,
		default_user TEXT
	);
	CREATE TABLE IF NOT EXISTS gravity (
		dir_path TEXT PRIMARY KEY,
		account_handle TEXT,
		weight INTEGER
	);
	CREATE TABLE IF NOT EXISTS global_stats (
		key TEXT PRIMARY KEY,
		value TEXT
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	// Migration: Ensure default_user column exists
	_, _ = db.Exec("ALTER TABLE repo_cache ADD COLUMN default_user TEXT")

	return db, nil
}

func GetRepoHash(repoRoot string) string {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(repoRoot)))
}

func SetDefaultUser(repoRoot string, user string) error {
	db, err := InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	repoHash := GetRepoHash(repoRoot)
	// Use explicit INSERT OR REPLACE with all columns we care about
	_, err = db.Exec("INSERT OR REPLACE INTO repo_cache (repo_path_hash, account_handle, default_user, last_used) VALUES (?, ?, ?, ?)",
		repoHash, user, user, time.Now().Unix())
	return err
}

func GetDefaultUser(repoRoot string) (string, error) {
	db, err := InitDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	abs, _ := filepath.Abs(repoRoot)
	curr := abs
	for {
		repoHash := GetRepoHash(curr)
		var user sql.NullString
		err = db.QueryRow("SELECT default_user FROM repo_cache WHERE repo_path_hash = ?", repoHash).Scan(&user)
		if err == nil && user.String != "" {
			return user.String, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr || parent == "." || parent == "/" || parent == "" {
			break
		}
		curr = parent
	}

	return "", nil
}

func ListAllCache() {
	db, err := InitDB()
	if err != nil {
		fmt.Printf("Error opening DB: %v\n", err)
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT repo_path_hash, account_handle, default_user FROM repo_cache")
	if err != nil {
		fmt.Printf("Error querying DB: %v\n", err)
		return
	}
	defer rows.Close()
	fmt.Printf("Listing all entries in repo_cache:\n")
	for rows.Next() {
		var hash, handle, def sql.NullString
		rows.Scan(&hash, &handle, &def)
		fmt.Printf("HASH: %s | HANDLE: %s | DEFAULT: %s\n", hash.String, handle.String, def.String)
	}
}
