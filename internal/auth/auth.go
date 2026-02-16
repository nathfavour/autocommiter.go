package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ghauth "github.com/cli/go-gh/v2/pkg/auth"
)

type GithubAccount struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetToken checks if an API key is provided, if not it tries to get one from gh CLI.
func GetToken(configuredKey string) string {
	if configuredKey != "" {
		return configuredKey
	}

	// Try getting token from gh CLI
	token, _ := ghauth.TokenForHost("github.com")
	return token
}

// GetGithubUser returns the current authenticated GitHub user login.
func GetGithubUser() string {
	cmd := exec.Command("gh", "api", "user", "--template", "{{.login}}")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ListAccounts returns a list of all logged-in GitHub accounts.
func ListAccounts() ([]string, error) {
	// We can check ~/.config/gh/hosts.yml or use gh auth status
	// Using 'gh auth status' is more reliable but slower.
	// Let's try to parse the config file for speed if it exists.
	home, _ := os.UserHomeDir()
	hostsPath := filepath.Join(home, ".config", "gh", "hosts.yml")
	
	// If we can't find/parse the file, fallback to command
	if _, err := os.Stat(hostsPath); err == nil {
		// Simple parsing of hosts.yml (this is a bit crude but fast)
		// Better way: use gh auth status --json (available in newer gh versions)
		cmd := exec.Command("gh", "auth", "status", "--json")
		out, err := cmd.CombinedOutput()
		if err == nil {
			var data struct {
				ActiveUser string `json:"active_user"`
				Accounts   []struct {
					User string `json:"user"`
				} `json:"accounts"`
			}
			if err := json.Unmarshal(out, &data); err == nil {
				var users []string
				for _, acc := range data.Accounts {
					users = append(users, acc.User)
				}
				return users, nil
			}
		}
	}

	// Fallback to old school status parsing if JSON fails
	cmd := exec.Command("gh", "auth", "status")
	out, _ := cmd.CombinedOutput()
	lines := strings.Split(string(out), "\n")
	var users []string
	for _, line := range lines {
		if strings.Contains(line, "Logged in to github.com as") {
			parts := strings.Fields(line)
			if len(parts) >= 7 {
				users = append(users, parts[6])
			}
		}
	}
	return users, nil
}

// GetAccountIdentity fetches the primary verified email, name and login for the active account.
func GetAccountIdentity(preferNoReply bool) (name string, email string, login string, err error) {
	// Get User Data (Name, Login, ID)
	cmd := exec.Command("gh", "api", "user", "--jq", "{name: (.name // .login), login: .login, id: .id}")
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", err
	}

	var data struct {
		Name  string `json:"name"`
		Login string `json:"login"`
		ID    int64  `json:"id"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return "", "", "", err
	}

	if preferNoReply {
		email = fmt.Sprintf("%d+%s@users.noreply.github.com", data.ID, data.Login)
		return data.Name, email, data.Login, nil
	}

	// Fallback to real email if requested
	cmdEmail := exec.Command("gh", "api", "user/emails", "--jq", ".[] | select(.primary == true and .verified == true) | .email")
	outEmail, err := cmdEmail.Output()
	if err != nil {
		return data.Name, "", data.Login, nil
	}
	email = strings.TrimSpace(string(outEmail))

	return data.Name, email, data.Login, nil
}

// SwitchAccount switches the active gh account.
func SwitchAccount(handle string) error {
	cmd := exec.Command("gh", "auth", "switch", "-u", handle)
	return cmd.Run()
}
