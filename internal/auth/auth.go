package auth

import (
	"os/exec"
	"strings"

	"github.com/cli/go-gh/v2/pkg/auth"
)

// GetToken checks if an API key is provided, if not it tries to get one from gh CLI.
func GetToken(configuredKey string) string {
	if configuredKey != "" {
		return configuredKey
	}

	// Try getting token from gh CLI
	token, _ := auth.TokenForHost("github.com")
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
