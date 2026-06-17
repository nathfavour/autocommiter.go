package processor

import (
        "bytes"
        "fmt"
        "io"
        "os"
        "path/filepath"
        "strings"

        "github.com/fatih/color"
        "github.com/nathfavour/autocommiter.go/internal/config"
        "github.com/nathfavour/autocommiter.go/internal/git"
        "regexp"
)

const (
        MaxCodeFileSize = 2 * 1024 * 1024 // 2MB
        BinaryThreshold = 5 * 1024 * 1024 // 5MB
)

type LeakMatch struct {
        File    string
        Line    int
        Content string
        Type    string
}

var (
        emailRegex    = regexp.MustCompile(`(?i)[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}`)
        apiKeyRegex   = regexp.MustCompile(`(?i)(?:key|token|secret|password|auth|api)[a-z0-9_]*["']?\s*[:=]\s*["']?([a-zA-Z0-9-._~+/]{32,})["']?`)
        awsKeyRegex   = regexp.MustCompile(`(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`)
        stripeKeyRegex = regexp.MustCompile(`sk_live_[0-9a-zA-Z]{24}`)
)

var sensitiveExtensions = []string{
        ".pem",
        ".key",
        ".p12",
        ".pfx",
        ".crt",
        ".ca-bundle",
}

var sensitiveFileNames = []string{
        ".env",
        "id_rsa",
        "id_dsa",
        "id_ecdsa",
        "id_ed25519",
        "secrets.json",
        "credentials.json",
}

func RunSecurityCheck(repoRoot string) ([]string, error) {
        cfg, _ := config.LoadMergedConfig(repoRoot)
        stagedFiles, err := git.GetStagedFiles(repoRoot)
        if err != nil {
                return nil, err
        }

        var insecureFiles []string
        var leaks []LeakMatch

        for _, file := range stagedFiles {
                fullPath := filepath.Join(repoRoot, file)
                info, err := os.Stat(fullPath)
                if err != nil {
                        continue
                }

                // 1. Check for sensitive/bulky files
                detectBulky := true
                if cfg.SecureDetectBulky != nil {
                        detectBulky = *cfg.SecureDetectBulky
                }

                if detectBulky && isInsecure(file, fullPath, info) {
                        insecureFiles = append(insecureFiles, file)
                        continue
                }

                // 2. Check for PII/Leaks in code diffs
                detectPII := true
                if cfg.SecureDetectPII != nil {
                        detectPII = *cfg.SecureDetectPII
                }

                if detectPII && info.Size() < MaxCodeFileSize {
                        fileLeaks, _ := scanFileForLeaks(repoRoot, file)
                        if len(fileLeaks) > 0 {
                                leaks = append(leaks, fileLeaks...)
                        }
                }
        }

        if len(leaks) > 0 {
                color.Red("\n🚨 SECURE_MODE: Potential PII or Secrets detected in code diffs:")
                for _, leak := range leaks {
                        color.Yellow("   - %s:%d [%s]: %s", leak.File, leak.Line, leak.Type, color.New(color.Faint).Sprint(leak.Content))
                }
                color.Cyan("\n🛡️  Action Required: Please review these lines for sensitive data.")
                color.Cyan("👉 To skip this check for this run, use --no-secure")
                color.Cyan("👉 To disable this permanently, use 'autocommiter toggle-secure-pii'")
                return nil, fmt.Errorf("security check failed: PII/Leaks detected")
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

func scanFileForLeaks(repoRoot string, file string) ([]LeakMatch, error) {
        diff, err := git.GetStagedDiffUnified(repoRoot, file)
        if err != nil {
                return nil, err
        }

        var leaks []LeakMatch
        lines := strings.Split(diff, "\n")
        currentLine := 0

        for _, line := range lines {
                if strings.HasPrefix(line, "@@") {
                        // Parse chunk header to get line number if needed, but for now simple
                        continue
                }
                if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
                        continue
                }

                content := strings.TrimPrefix(line, "+")
                currentLine++ // This is naive, but works for identifying relative position

                if awsKeyRegex.MatchString(content) {
                        leaks = append(leaks, LeakMatch{File: file, Line: currentLine, Content: content, Type: "AWS Key"})
                } else if stripeKeyRegex.MatchString(content) {
                        leaks = append(leaks, LeakMatch{File: file, Line: currentLine, Content: content, Type: "Stripe Key"})
                } else if apiKeyRegex.MatchString(content) {
                        leaks = append(leaks, LeakMatch{File: file, Line: currentLine, Content: content, Type: "Potential API Key/Secret"})
                } else if emailRegex.MatchString(content) {
                        // Optional: only alert if it looks like a leak (e.g. not a known contributor email)
                        leaks = append(leaks, LeakMatch{File: file, Line: currentLine, Content: content, Type: "PII (Email)"})
                }
        }

        return leaks, nil
}

func isInsecure(relPath string, fullPath string, info os.FileInfo) bool {
        lowerPath := strings.ToLower(relPath)
        ext := filepath.Ext(lowerPath)
        base := filepath.Base(lowerPath)

        // 1. Check sensitive extensions
        for _, targetExt := range sensitiveExtensions {
                if ext == targetExt {
                        return true
                }
        }

        // 2. Check sensitive exact filenames
        for _, targetName := range sensitiveFileNames {
                if base == targetName || strings.HasPrefix(base, targetName+".") {
                        return true
                }
        }

        // 3. Check for bulky files that might be binaries
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
