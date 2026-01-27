package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(vibeManifestCmd)
	rootCmd.AddCommand(executeCmd)
}

var vibeManifestCmd = &cobra.Command{
	Use:   "vibe-manifest",
	Short: "Output vibe manifest for vibeauracle",
	Run: func(cmd *cobra.Command, args []string) {
		manifest := map[string]interface{}{
			"id":          "autocommiter",
			"name":        "Autocommiter",
			"repo":        "nathfavour/autocommiter.go",
			"version":     version,
			"description": "Auto-generate git commit messages using AI",
			"protocol":    "stdio",
			"command":     "autocommiter",
			"update_cmd":  "autocommiter update",
			"inbuilt":     true,
			"tool_set": []map[string]interface{}{
				{
					"name":        "generate_commit_message",
					"description": "Generate a commit message for the staged changes",
					"inputSchema": json.RawMessage(`{"type":"object","properties":{"repo_path":{"type":"string","description":"Path to the git repository"}}}`),
				},
				{
					"name":        "summarize_changes",
					"description": "Summarize staged changes as JSON",
					"inputSchema": json.RawMessage(`{"type":"object","properties":{"repo_path":{"type":"string","description":"Path to the git repository"}}}`),
				},
			},
		}
		data, _ := json.MarshalIndent(manifest, "", "  ")
		fmt.Println(string(data))
	},
}

var executeCmd = &cobra.Command{
	Use:   "execute [tool] [args]",
	Short: "Execute a tool in vibe mode",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Tool name required")
			os.Exit(1)
		}
		
		toolName := args[0]
		
		switch toolName {
		case "generate_commit_message":
			fmt.Println(`{"content": "feat: add vibe integration", "status": "success"}`)
		case "summarize_changes":
			fmt.Println(`{"content": "Summary of changes...", "status": "success"}`)
		default:
			fmt.Printf("Unknown tool: %s\n", toolName)
			os.Exit(1)
		}
	},
}
