package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/nathfavour/autocommiter.go/internal/config"
)

const repo = "nathfavour/autocommiter.go"

type release struct {
	TagName string `json:"tag_name"`
}

func CheckForUpdates(currentVersion string) {
	if currentVersion == "v0.0.0-dev" || currentVersion == "dev" {
		return
	}

	latest, err := getLatestTag()
	if err != nil {
		return
	}

	if latest != currentVersion && latest != "" {
		if latest == "latest" && (currentVersion == "v0.0.0-dev" || currentVersion == "dev") {
			// Don't bother if we are already in a dev state and latest is just 'latest'
			// but wait, if it's a dev build, we might want to update anyway.
		}
		color.Yellow("\n🔔 A new version is available: %s (current: %s)", latest, currentVersion)
		color.Yellow("👉 Run 'autocommiter update' to upgrade instantly.\n")
	}
}

func getLatestTag() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}

	return rel.TagName, nil
}

func SeamlessUpdate(currentVersion string) error {
	home, _ := os.UserHomeDir()
	primaryDir := filepath.Join(home, ".local", "bin")
	primaryPath := filepath.Join(primaryDir, "autocommiter")

	cfg, _ := config.LoadConfig()
	buildFromSource := false
	if cfg.BuildFromSource != nil {
		buildFromSource = *cfg.BuildFromSource
	}

	if buildFromSource {
		color.Cyan("🛠️  Build From Source mode enabled.")
		color.Yellow("📥 Updating via 'go install'...")
		cmd := exec.Command("go", "install", "github.com/nathfavour/autocommiter.go/cmd/autocommiter@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build from source: %w", err)
		}
		color.Green("✨ Successfully built and installed latest version from source!")
		return nil
	}

	latest, err := getLatestTag()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	isLatestRolling := latest == "latest"
	if latest == currentVersion && !isLatestRolling && currentVersion != "v0.0.0-dev" && currentVersion != "dev" {
		color.Green("✓ autocommiter is already up to date (%s)", currentVersion)
		return nil
	}

	if isLatestRolling {
		color.Cyan("🔄 Updating autocommiter to latest rolling build...")
	} else {
		color.Cyan("🔄 Updating autocommiter to %s...", latest)
	}

	// Conflict Cleanup Logic using 'which -a'
	color.Cyan("🔍 Checking for conflicting binaries...")
	cmdWhich := exec.Command("which", "-a", "autocommiter")
	output, _ := cmdWhich.Output()
	foundPaths := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, p := range foundPaths {
		if p == "" {
			continue
		}
		absP, _ := filepath.Abs(p)
		if absP == primaryPath {
			continue
		}
		
		// If it's not our primary path, it's a conflict
		if info, err := os.Stat(absP); err == nil && !info.IsDir() {
			color.Yellow("🗑️  Removing conflicting binary at: %s", absP)
			_ = os.Remove(absP)
		}
	}

	// Determine binary name based on OS/Arch
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Check for Android (Termux)
	if osName == "linux" {
		if _, err := os.Stat("/data/data/com.termux"); err == nil {
			osName = "android"
		}
	}

	binaryName := fmt.Sprintf("autocommiter-%s-%s", osName, archName)
	if osName == "windows" {
		binaryName += ".exe"
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, latest, binaryName)

	color.Yellow("📥 Downloading binary...")

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download binary: %s", resp.Status)
	}

	if err := os.MkdirAll(primaryDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", primaryDir, err)
	}

	// Create a temporary file for the new binary
	tmpPath := primaryPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to save binary: %w", err)
	}

	// Move new binary to the primary location
	// If the file exists, we try to move it to .old first to avoid "text file busy" errors if running
	if _, err := os.Stat(primaryPath); err == nil {
		oldPath := primaryPath + ".old"
		_ = os.Remove(oldPath)
		_ = os.Rename(primaryPath, oldPath)
		defer os.Remove(oldPath)
	}

	if err := os.Rename(tmpPath, primaryPath); err != nil {
		return fmt.Errorf("failed to install to %s: %w", primaryPath, err)
	}

	color.Green("✨ Successfully installed to %s", primaryPath)
	color.Green("✓ Updated to %s!", latest)
	return nil
}
