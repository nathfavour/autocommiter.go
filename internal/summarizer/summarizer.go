package summarizer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/nathfavour/autocommiter.go/internal/git"
)

type FileChange struct {
	File   string `json:"f"`
	Change string `json:"c"`
}

type FileChangesResponse struct {
	Files []FileChange `json:"files"`
}

func AnalyzeFileChange(cwd string, file string) (string, error) {
	// First, get the diff with context
	diff, err := git.GetStagedDiff(cwd, file)
	if err == nil && diff != "" {
		// If the diff is small enough, return it all
		if len(diff) < 2000 {
			return diff, nil
		}

		// Otherwise, get a summary of what changed
		numstat, _ := git.GetStagedDiffNumstat(cwd, file)
		return fmt.Sprintf("Large diff: %s\nFull diff omitted but here is the start:\n%s", numstat, truncateDiff(diff, 1000)), nil
	}

	return "mod", nil
}

func truncateDiff(diff string, maxLen int) string {
	if len(diff) <= maxLen {
		return diff
	}
	return diff[:maxLen] + "\n... (truncated)"
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

	// We'll try to fit as much as possible, prioritizing file names
	// Modern LLMs can handle much more than 400 chars. 
	// The caller should pass a larger maxLen (e.g., 10000)

	serialize := func(arr []FileChange, truncateLen int) string {
		var mapped []FileChange
		for _, fc := range arr {
			change := fc.Change
			if truncateLen > 0 && len(change) > truncateLen {
				change = change[:truncateLen] + "\n..."
			}
			mapped = append(mapped, FileChange{File: fc.File, Change: change})
		}
		res := FileChangesResponse{Files: mapped}
		b, _ := json.Marshal(res)
		return string(b)
	}

	// Dynamic truncation levels
	truncateLens := []int{-1, 2000, 1000, 500, 200, 100, 50}

	for _, tLen := range truncateLens {
		for keep := len(fileChanges); keep >= 1; keep-- {
			s := serialize(fileChanges[:keep], tLen)
			if len(s) <= maxLen {
				return s
			}
		}
	}

	// Ultimate fallback
	return `{"files":[{"f":"multiple files","c":"too large to summarize"}]}`
}
