package processor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/nathfavour/autocommiter.go/internal/git"
)

const (
	MaxCodeFileSize = 2 * 1024 * 1024 // 2MB
	BinaryThreshold = 5 * 1024 * 1024 // 5MB
)

var sensitivePatterns = []string{
	".env",
	".pem",
	".key",
	".p12",
	".pfx",
	".crt",
	".ca-bundle",
	"id_rsa",
	"id_dsa",
	"id_ecdsa",
	"id_ed25519",
	"secrets.json",
	"credentials.json",
}

func RunSecurityCheck(repoRoot string) ([]string, error) {
	stagedFiles, err := git.GetStagedFiles(repoRoot)
	if err != nil {
		return nil, err
	}

	var insecureFiles []string

	for _, file := range stagedFiles {
		fullPath := filepath.Join(repoRoot, file)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if isInsecure(file, fullPath, info) {
			insecureFiles = append(insecureFiles, file)
		}
	}

	if len(insecureFiles) > 0 {
		color.Yellow("⚠️  SECURE_MODE: Detected potentially insecure or bulky files staged for commit:")
		for _, f := range insecureFiles {
			color.Red("   - %s", f)
		}

		fmt.Print(color.CyanString("🛡️  Adding these to .gitignore and unstaging them... "))
		if err := handleInsecureFiles(repoRoot, insecureFiles); err != nil {
			fmt.Println(color.RedString("Failed: %v", err))
			return nil, err
		}
		fmt.Println(color.GreenString("Done!"))
	}

	return insecureFiles, nil
}

func isInsecure(relPath string, fullPath string, info os.FileInfo) bool {
	lowerPath := strings.ToLower(relPath)

	// 1. Check sensitive patterns
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	// 2. Check for bulky files that might be binaries
	if info.Size() > BinaryThreshold {
		return isBinary(fullPath)
	}

	return false
}

func isBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Check for null bytes in the first 512 bytes
	return bytes.Contains(buffer[:n], []byte{0})
}

func handleInsecureFiles(repoRoot string, files []string) error {
	// 1. Unstage files
	for _, f := range files {
		_, err := git.RunGitCommand(repoRoot, "reset", "HEAD", "--", f)
		if err != nil {
			return err
		}
	}

	// 2. Add to .gitignore
	gitignorePath := filepath.Join(repoRoot, ".gitignore")
	existing, _ := os.ReadFile(gitignorePath)
	content := string(existing)

	var toAppend []string
	lineMap := make(map[string]bool)
	for _, l := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			lineMap[trimmed] = true
		}
	}

	for _, f := range files {
		if !lineMap[f] {
			toAppend = append(toAppend, f)
		}
	}

	if len(toAppend) > 0 {
		newContent := content
		if !strings.HasSuffix(content, "\n") && content != "" {
			newContent += "\n"
		}
		newContent += "\n# Added by Autocommiter SECURE_MODE:\n"
		newContent += strings.Join(toAppend, "\n") + "\n"
		return os.WriteFile(gitignorePath, []byte(newContent), 0644)
	}

	return nil
}
