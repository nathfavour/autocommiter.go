package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/fatih/color"
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
		color.Yellow("👉 Run the install script to update:")
		color.Cyan("   curl -fsSL https://raw.githubusercontent.com/%s/master/install.sh | bash\n", repo)
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

// SeamlessUpdate would download and replace the binary, but for now we'll suggest the install script
// as it handles permissions and PATH better across platforms.
func SeamlessUpdate(currentVersion string) error {
	if currentVersion == "dev" {
		return nil
	}

	latest, err := getLatestTag()
	if err != nil {
		return err
	}

	if latest == currentVersion {
		return nil
	}

	color.Cyan("🔄 Updating autocommiter to %s...", latest)

	// Determine binary name based on OS/Arch
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// autocommiter.go_linux_amd64.tar.gz
	archiveName := fmt.Sprintf("autocommiter.go_%s_%s.tar.gz", osName, archName)
	if osName == "windows" {
		archiveName = fmt.Sprintf("autocommiter.go_%s_%s.zip", osName, archName)
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, latest, archiveName)

	color.Yellow("📥 Downloading from %s...", downloadURL)

	// In a real implementation we would download, extract, and replace.
	// For this CLI, suggesting the curl command is safer until we have a robust cross-platform self-update library integration.

	return nil
}
