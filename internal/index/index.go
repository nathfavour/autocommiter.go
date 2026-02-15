package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

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

	schema := `
	CREATE TABLE IF NOT EXISTS repo_cache (
		repo_path_hash TEXT PRIMARY KEY,
		account_handle TEXT,
		email TEXT,
		name TEXT,
		last_used INTEGER
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return db, nil
}
