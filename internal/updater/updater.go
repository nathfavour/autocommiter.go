package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/nathfavour/autocommiter.go/internal/config"
)

const repo = "nathfavour/autocommiter.go"

type release struct {
	TagName string `json:"tag_name"`
}

func CheckForUpdates(currentVersion string) {
	if currentVersion == "dev" {
		return
	}

	latest, err := getLatestTag()
	if err != nil {
		return
	}

	if latest != currentVersion && latest != "" {
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

	if latest == currentVersion && currentVersion != "dev" {
		color.Green("✓ autocommiter is already up to date (%s)", currentVersion)
		return nil
	}

	color.Cyan("🔄 Updating autocommiter to %s...", latest)

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

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a temporary file for the new binary
	tmpPath := exe + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpPath)

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to save binary: %w", err)
	}

	// Move current binary to a backup location
	oldPath := exe + ".old"
	if err := os.Rename(exe, oldPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to the original location
	if err := os.Rename(tmpPath, exe); err != nil {
		// Attempt to restore backup
		os.Rename(oldPath, exe)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Remove backup
	os.Remove(oldPath)

	color.Green("✨ Successfully updated to %s!", latest)
	return nil
}
