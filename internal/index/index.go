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

	// Use sqlite driver from modernc.org
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

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

	// Migration: Add default_user column if it doesn't exist
	_, _ = db.Exec("ALTER TABLE repo_cache ADD COLUMN default_user TEXT")

	return db, nil
}

func GetRepoHash(repoRoot string) string {
	abs, err := filepath.Abs(repoRoot)
	if err == nil {
		repoRoot = abs
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(repoRoot)))
	// fmt.Printf("DEBUG: Path=%s HASH=%s\n", repoRoot, hash)
	return hash
}

func SetDefaultUser(repoRoot string, user string) error {
	db, err := InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	repoHash := GetRepoHash(repoRoot)
	fmt.Printf("DEBUG: Saving DefaultUser=%s for Hash=%s\n", user, repoHash)
	// We use the handle as both the account_handle and default_user for consistency in manual setup
	_, err = db.Exec("INSERT INTO repo_cache (repo_path_hash, account_handle, default_user, last_used) VALUES (?, ?, ?, ?) ON CONFLICT(repo_path_hash) DO UPDATE SET default_user = excluded.default_user, account_handle = excluded.account_handle, last_used = excluded.last_used",
		repoHash, user, user, time.Now().Unix())
	return err
}

func GetDefaultUser(repoRoot string) (string, error) {
	db, err := InitDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	repoHash := GetRepoHash(repoRoot)
	var user sql.NullString
	err = db.QueryRow("SELECT default_user FROM repo_cache WHERE repo_path_hash = ?", repoHash).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("DEBUG: No DefaultUser found for Hash=%s\n", repoHash)
			return "", nil
		}
		return "", err
	}
	fmt.Printf("DEBUG: Found DefaultUser=%s for Hash=%s\n", user.String, repoHash)
	return user.String, nil
}

func ListAllCache() {
	db, _ := InitDB()
	defer db.Close()
	rows, _ := db.Query("SELECT repo_path_hash, account_handle, default_user FROM repo_cache")
	for rows.Next() {
		var hash, handle, def sql.NullString
		rows.Scan(&hash, &handle, &def)
		fmt.Printf("HASH: %s | HANDLE: %s | DEFAULT: %s\n", hash.String, handle.String, def.String)
	}
}
