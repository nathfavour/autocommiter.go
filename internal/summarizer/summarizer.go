package summarizer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nathfavour/autocommiter-go/internal/git"
)

type FileChange struct {
	File   string `json:"f"`
	Change string `json:"c"`
}

type FileChangesResponse struct {
	Files []FileChange `json:"files"`
}

func AnalyzeFileChange(cwd string, file string) (string, error) {
	diff, err := git.GetStagedDiffNumstat(cwd, file)
	if err == nil && diff != "" {
		lines := strings.Split(diff, "\n")
		if len(lines) > 0 {
			parts := strings.Split(lines[0], "\t")
			if len(parts) >= 3 {
				added := parts[0]
				if added == "-" {
					added = "0"
				}
				removed := parts[1]
				if removed == "-" {
					removed = "0"
				}
				return fmt.Sprintf("%s+/%s−", added, removed), nil
			}
		}
		return "mod", nil
	}

	hunks, err := git.GetStagedDiffUnified(cwd, file)
	if err == nil {
		lines := strings.Split(hunks, "\n")
		var first string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				first = trimmed
				break
			}
		}
		if first == "" {
			first = "mod"
		}
		if len(first) > 40 {
			first = first[:40]
		}
		re := regexp.MustCompile(`\s+`)
		collapsed := re.ReplaceAllString(first, " ")
		return collapsed, nil
	}

	return "err", nil
}

func BuildFileChanges(cwd string) ([]FileChange, error) {
	files, err := git.GetStagedFiles(cwd)
	if err != nil {
		return nil, err
	}

	var changes []FileChange
	for _, file := range files {
		change, _ := AnalyzeFileChange(cwd, file)
		changes = append(changes, FileChange{File: file, Change: change})
	}

	return changes, nil
}

func CompressToJSON(fileChanges []FileChange, maxLen int) string {
	if len(fileChanges) == 0 {
		return `{"files":[]}`
	}

	serialize := func(arr []FileChange, truncateLen int) string {
		var mapped []FileChange
		for _, fc := range arr {
			change := fc.Change
			if truncateLen > 0 && len(change) > truncateLen {
				change = change[:truncateLen]
			}
			mapped = append(mapped, FileChange{File: fc.File, Change: change})
		}
		res := FileChangesResponse{Files: mapped}
		b, _ := json.Marshal(res)
		return string(b)
	}

	truncateLens := []int{-1, 12, 6, 3, 1}

	for _, tLen := range truncateLens {
		for keep := len(fileChanges); keep >= 1; keep-- {
			s := serialize(fileChanges[:keep], tLen)
			if len(s) <= maxLen {
				return s
			}
		}
	}

	// Fallback
	fileName := fileChanges[0].File
	parts := strings.Split(fileName, "/")
	fileName = parts[len(parts)-1]
	minimal := []FileChange{{File: fileName, Change: "mod"}}
	b, _ := json.Marshal(FileChangesResponse{Files: minimal})
	return string(b)
}
