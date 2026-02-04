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
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nathfavour/autocommiter.go/internal/config"
	"github.com/nathfavour/autocommiter.go/internal/netutil"
)

const repo = "nathfavour/autocommiter.go"
const updateThrottle = 4 * time.Hour // Check GitHub at most every 4 hours

type release struct {
	TagName string `json:"tag_name"`
}

func CheckForUpdates(currentVersion string) {
	cfg, _ := config.LoadConfig()
	
	// If we've checked recently, use the cached result to notify
	if cfg.LastUpdateCheck > 0 && time.Since(time.Unix(cfg.LastUpdateCheck, 0)) < updateThrottle {
		if cfg.LatestVersionFound != "" && isNewer(cfg.LatestVersionFound, currentVersion) {
			color.Yellow("\n🔔 A new version is available: %s (current: %s)", cfg.LatestVersionFound, currentVersion)
			color.Yellow("👉 It will be automatically installed after this task completes.\n")
		}
		return
	}

	if currentVersion == "dev" || currentVersion == "" {
		if cfg.BuildFromSource == nil || !*cfg.BuildFromSource {
			return
		}
	}

	latest, err := getLatestTag()
	if err != nil {
		return
	}

	// Update the cache
	cfg.LastUpdateCheck = time.Now().Unix()
	cfg.LatestVersionFound = latest
	_ = config.SaveConfig(cfg)

	if isNewer(latest, currentVersion) {
		color.Yellow("\n🔔 A new version is available: %s (current: %s)", latest, currentVersion)
		color.Yellow("👉 It will be automatically installed after this task completes.\n")
	}
}

func AutoUpdate(currentVersion string) {
	cfg, _ := config.LoadConfig()
	
	// We only update if we have a version in the cache that is newer
	if cfg.LatestVersionFound == "" || !isNewer(cfg.LatestVersionFound, currentVersion) {
		return
	}

	latest := cfg.LatestVersionFound
	color.Cyan("\n🚀 New version %s detected. Performing automatic update...", latest)
	if err := SeamlessUpdate(currentVersion); err != nil {
		color.Red("❌ Automatic update failed: %v", err)
	} else {
		// After a successful update, we clear the "latest version found" so we don't 
		// keep trying to update to the same thing in a loop if the version string didn't change.
		cfg.LatestVersionFound = ""
		_ = config.SaveConfig(cfg)
	}
}

func getLatestTag() (string, error) {
	// Instead of /releases/latest which only returns full releases,
	// we fetch all releases and pick the first one (usually the newest)
	resp, err := netutil.GetHttpClient().Get(fmt.Sprintf("https://api.github.com/repos/%s/releases", repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	// The first release in the list is usually the latest by creation date
	latest := releases[0].TagName
	
	// If "latest" tag exists but there is also a semantic version, prioritize the semantic one if it's newer
	// for now we just return the most recent one.
	return latest, nil
}

func isNewer(latest, current string) bool {
	if latest == "" || latest == current {
		return false
	}
	
	// If latest is a rolling tag, we only update if we aren't already on "dev" 
	// or if we have specialized logic. For now, assume "latest" is always newer 
	// than a semantic version, but avoid loops.
	if latest == "latest" {
		// If current is dev, it's hard to tell if latest is newer without commit hashes.
		// We'll rely on the throttle to prevent blind loops.
		return current != "latest"
	}

	if current == "dev" || current == "" {
		return true
	}

	// Basic string comparison (v0.1.1 != v0.1.0)
	// For production, a semver library would be better.
	return latest != current
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
		color.Yellow("📥 Updating via 'go build' to ensure installation in ~/.local/bin...")

		// Ensure primary directory exists
		if err := os.MkdirAll(primaryDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", primaryDir, err)
		}

		// Download source or use temporary directory to clone and build
		// For simplicity and to avoid 'go install' putting it in GOBIN,
		// we clone to a temp dir and build directly to the target path.
		tmpDir, err := os.MkdirTemp("", "autocommiter-build-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		color.Cyan("📂 Cloning latest source...")
		cloneCmd := exec.Command("git", "clone", "--depth", "1", "https://github.com/"+repo+".git", tmpDir)
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("failed to clone source: %w", err)
		}

		color.Cyan("🔨 Building binary...")
		buildCmd := exec.Command("go", "build", "-o", primaryPath, "./cmd/autocommiter")
		buildCmd.Dir = tmpDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build: %w", err)
		}

		// Conflict Cleanup Logic (even in source build)
		cleanupConflicts(primaryPath)

		color.Green("✨ Successfully built and installed latest version to %s!", primaryPath)
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
	cleanupConflicts(primaryPath)

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

	resp, err := netutil.GetHttpClient().Get(downloadURL)
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

func cleanupConflicts(primaryPath string) {
	color.Cyan("🔍 Checking for conflicting binaries in $PATH...")
	
	// Get all instances of autocommiter in PATH
	allPaths := getAllInstances("autocommiter")

	for _, p := range allPaths {
		absP, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		
		// Resolve symlinks to find the real binary
		realPath, err := filepath.EvalSymlinks(absP)
		if err == nil {
			absP = realPath
		}

		if absP == primaryPath {
			continue
		}

		// If it's not our primary path, it's a conflict
		if info, err := os.Stat(absP); err == nil && !info.IsDir() {
			color.Yellow("🗑️  Removing conflicting binary at: %s", absP)
			err := os.Remove(absP)
			if err != nil {
				color.Red("⚠️  Could not remove %s: %v", absP, err)
			}
		}
	}

	// Specifically check GOBIN/go bin as well
	home, _ := os.UserHomeDir()
	goBin := filepath.Join(home, "go", "bin", "autocommiter")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	if info, err := os.Stat(goBin); err == nil && !info.IsDir() {
		if absGoBin, err := filepath.Abs(goBin); err == nil {
			realGoBin, err := filepath.EvalSymlinks(absGoBin)
			if err == nil {
				absGoBin = realGoBin
			}
			if absGoBin != primaryPath {
				color.Yellow("🗑️  Removing GOBIN binary at: %s", absGoBin)
				_ = os.Remove(absGoBin)
			}
		}
	}
}

func getAllInstances(exe string) []string {
	var paths []string
	
	// Try 'which -a' first
	cmdWhich := exec.Command("which", "-a", exe)
	output, err := cmdWhich.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, l := range lines {
			if l != "" {
				paths = append(paths, l)
			}
		}
	}

	// Also manually scan PATH to be sure
	envPath := os.Getenv("PATH")
	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	
	for _, dir := range strings.Split(envPath, sep) {
		full := filepath.Join(dir, exe)
		if runtime.GOOS == "windows" && !strings.HasSuffix(full, ".exe") {
			full += ".exe"
		}
		if _, err := os.Stat(full); err == nil {
			paths = append(paths, full)
		}
	}

	// Unique paths
	unique := make(map[string]bool)
	var result []string
	for _, p := range paths {
		if !unique[p] {
			unique[p] = true
			result = append(result, p)
		}
	}
	return result
}
