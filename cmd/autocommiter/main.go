package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/nathfavour/autocommiter-go/internal/auth"
	"github.com/nathfavour/autocommiter-go/internal/config"
	"github.com/nathfavour/autocommiter-go/internal/models"
	"github.com/nathfavour/autocommiter-go/internal/processor"
	"github.com/nathfavour/autocommiter-go/internal/updater"
	"github.com/spf13/cobra"
)

var (
	repoPath string
	noPush   bool
	force    bool

	// Version metadata (injected by GoReleaser)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg, _ := config.LoadConfig()
	if cfg.AutoUpdate != nil && *cfg.AutoUpdate {
		updater.CheckForUpdates(version)
	}

	var rootCmd = &cobra.Command{
		Use:     "autocommiter",
		Short:   "Auto-generate git commit messages using AI",
		Version: fmt.Sprintf("%s (%s, %s)", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			return processor.GenerateCommit(repoPath, noPush, force)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&repoPath, "repo", "r", "", "Path to git repository (defaults to current directory)")
	rootCmd.PersistentFlags().BoolVarP(&noPush, "no-push", "n", false, "Skip pushing after commit")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Don't ask for confirmation before committing")

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate commit message and commit changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return processor.GenerateCommit(repoPath, noPush, force)
		},
	}
	rootCmd.AddCommand(generateCmd)

	var generateMessageCmd = &cobra.Command{
		Use:   "generate-message",
		Short: "Generate and output only the commit message (no commit/push)",
		RunE: func(cmd *cobra.Command, args []string) error {
			msg, err := processor.GenerateMessage(repoPath)
			if err != nil {
				return err
			}
			fmt.Print(msg)
			return nil
		},
	}
	rootCmd.AddCommand(generateMessageCmd)

	var setApiKeyCmd = &cobra.Command{
		Use:   "set-api-key [KEY]",
		Short: "Set GitHub API key (optional if 'gh auth login' is used)",
		Long:  "Set your GitHub Models API key. Note: This is optional if you are already authenticated with the GitHub CLI ('gh auth login').",
		RunE: func(cmd *cobra.Command, args []string) error {
			var key string
			if len(args) > 0 {
				key = args[0]
			} else {
				fmt.Print(color.CyanString("Enter GitHub API key (will be stored securely): "))
				reader := bufio.NewReader(os.Stdin)
				key, _ = reader.ReadString('\n')
				key = strings.TrimSpace(key)
			}

			if key == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			if err := config.SetAPIKey(key); err != nil {
				return err
			}
			color.Green("✓ API key saved!")
			return nil
		},
	}
	rootCmd.AddCommand(setApiKeyCmd)

	var rawKey bool
	var getApiKeyCmd = &cobra.Command{
		Use:   "get-api-key",
		Short: "Get stored API key",
		Run: func(cmd *cobra.Command, args []string) {
			key, _ := config.GetAPIKey()
			token := auth.GetToken(key)

			if rawKey {
				if token != "" {
					fmt.Print(token)
				}
				return
			}

			if key != "" {
				masked := "****"
				if len(key) > 8 {
					masked = key[:4] + "..." + key[len(key)-4:]
				}
				color.Cyan("🔑 API Key: %s", color.YellowString(masked))
			} else {
				if token != "" {
					user := auth.GetGithubUser()
					color.Green("✓ Authenticated via GitHub CLI (%s)", user)
				} else {
					color.Yellow("ℹ️ No API key set and GitHub CLI not authenticated.")
					color.Cyan("👉 Use 'set-api-key' or run 'gh auth login'")
				}
			}
		},
	}
	getApiKeyCmd.Flags().BoolVar(&rawKey, "raw", false, "Output raw token (useful for scripting)")
	rootCmd.AddCommand(getApiKeyCmd)

	var refreshModelsCmd = &cobra.Command{
		Use:   "refresh-models",
		Short: "Refresh available AI models from GitHub Models API",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, _ := config.GetAPIKey()
			token := auth.GetToken(apiKey)
			if token == "" {
				return fmt.Errorf("API key not set and GitHub CLI not authenticated. Use 'set-api-key' or 'gh auth login'")
			}

			color.Cyan("🔄 Fetching models from GitHub Models API...")
			success, msg, count, err := models.RefreshModelList(apiKey)
			if err != nil {
				return err
			}

			if success {
				color.Green("✓ %d models cached", count)
			} else {
				color.Red("✗ %s", msg)
			}
			return nil
		},
	}
	rootCmd.AddCommand(refreshModelsCmd)

	var listModelsCmd = &cobra.Command{
		Use:   "list-models",
		Short: "List available AI models",
		RunE: func(cmd *cobra.Command, args []string) error {
			available, err := models.ListAvailableModels()
			if err != nil {
				return err
			}
			current, _ := config.GetSelectedModel()

			color.New(color.FgCyan, color.Bold).Println("📋 Available Models:")
			for _, m := range available {
				marker := " "
				if m.ID == current {
					marker = "→"
				}
				color.Green("%s %s", marker, color.CyanString(m.Name))
				if m.FriendlyName != nil {
					color.New(color.Faint).Printf("   %s\n", *m.FriendlyName)
				}
				if m.Summary != nil {
					color.New(color.Faint).Printf("   %s\n", *m.Summary)
				}
				fmt.Println()
			}
			return nil
		},
	}
	rootCmd.AddCommand(listModelsCmd)

	var selectModelCmd = &cobra.Command{
		Use:   "select-model",
		Short: "Select default AI model",
		RunE: func(cmd *cobra.Command, args []string) error {
			available, err := models.ListAvailableModels()
			if err != nil || len(available) == 0 {
				return fmt.Errorf("no models available")
			}

			color.New(color.FgCyan, color.Bold).Println("🤖 Select a Model:")
			for i, m := range available {
				friendly := m.Name
				if m.FriendlyName != nil {
					friendly = *m.FriendlyName
				}
				fmt.Printf("%d. %s (%s)\n", i+1, color.CyanString(m.Name), color.New(color.Faint).Sprint(friendly))
			}

			fmt.Print(color.CyanString("\nEnter choice (1-%d): ", len(available)))
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			choice, err := strconv.Atoi(strings.TrimSpace(input))
			if err != nil || choice < 1 || choice > len(available) {
				return fmt.Errorf("invalid choice")
			}

			selected := available[choice-1]
			if err := config.SetSelectedModel(selected.ID); err != nil {
				return err
			}
			color.Green("✓ Selected: %s", color.CyanString(selected.Name))
			return nil
		},
	}
	rootCmd.AddCommand(selectModelCmd)

	var rawModel bool
	var getModelCmd = &cobra.Command{
		Use:   "get-model",
		Short: "Get current default model",
		Run: func(cmd *cobra.Command, args []string) {
			model, _ := config.GetSelectedModel()
			if rawModel {
				fmt.Print(model)
				return
			}
			color.Cyan("🤖 Current Model: %s", color.YellowString(model))
		},
	}
	getModelCmd.Flags().BoolVar(&rawModel, "raw", false, "Output raw model ID")
	rootCmd.AddCommand(getModelCmd)

	var toggleGitmojiCmd = &cobra.Command{
		Use:   "toggle-gitmoji",
		Short: "Enable/disable gitmoji prefixes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.LoadConfig()
			current := false
			if cfg.EnableGitmoji != nil {
				current = *cfg.EnableGitmoji
			}
			newVal := !current
			cfg.EnableGitmoji = &newVal
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}

			if newVal {
				color.Green("✓ Gitmoji enabled")
			} else {
				color.Green("✓ Gitmoji %s", color.YellowString("disabled"))
			}
			return nil
		},
	}
	rootCmd.AddCommand(toggleGitmojiCmd)

	var toggleSkipConfirmationCmd = &cobra.Command{
		Use:   "toggle-skip-confirmation",
		Short: "Enable/disable skipping commit confirmation",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.LoadConfig()
			current := false
			if cfg.SkipConfirmation != nil {
				current = *cfg.SkipConfirmation
			}
			newVal := !current
			cfg.SkipConfirmation = &newVal
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}

			if newVal {
				color.Green("✓ Skip Confirmation enabled")
			} else {
				color.Green("✓ Skip Confirmation %s", color.YellowString("disabled"))
			}
			return nil
		},
	}
	rootCmd.AddCommand(toggleSkipConfirmationCmd)

	var toggleAutoUpdateCmd = &cobra.Command{
		Use:   "toggle-auto-update",
		Short: "Enable/disable automatic update checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.LoadConfig()
			current := true
			if cfg.AutoUpdate != nil {
				current = *cfg.AutoUpdate
			}
			newVal := !current
			cfg.AutoUpdate = &newVal
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}

			if newVal {
				color.Green("✓ Auto-update checks enabled")
			} else {
				color.Green("✓ Auto-update checks %s", color.YellowString("disabled"))
			}
			return nil
		},
	}
	rootCmd.AddCommand(toggleAutoUpdateCmd)

	var getConfigCmd = &cobra.Command{
		Use:   "get-config",
		Short: "Display current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.LoadConfig()
			color.New(color.FgCyan, color.Bold).Println("⚙️  Configuration:")

			color.Cyan("Authentication:")
			if cfg.APIKey != nil && *cfg.APIKey != "" {
				key := *cfg.APIKey
				masked := "****"
				if len(key) > 8 {
					masked = key[:4] + "..." + key[len(key)-4:]
				}
				fmt.Printf("  Manual: %s\n", color.YellowString(masked))
			} else {
				token := auth.GetToken("")
				if token != "" {
					user := auth.GetGithubUser()
					color.Green("  GitHub CLI: Authenticated (%s)", user)
				} else {
					color.Red("  Not Authenticated")
					color.New(color.Faint).Println("  (Use 'set-api-key' or 'gh auth login')")
				}
			}

			color.Cyan("\nSelected Model:")
			model := "gpt-4o-mini"
			if cfg.SelectedModel != nil {
				model = *cfg.SelectedModel
			}
			fmt.Printf("  %s\n", color.YellowString(model))

			color.Cyan("\nGitmoji Enabled:")
			gitmoji := false
			if cfg.EnableGitmoji != nil {
				gitmoji = *cfg.EnableGitmoji
			}
			if gitmoji {
				color.Green("  Yes")
			} else {
				color.Red("  No")
			}

			color.Cyan("\nUpdate Gitignore:")
			update := false
			if cfg.UpdateGitignore != nil {
				update = *cfg.UpdateGitignore
			}
			if update {
				color.Green("  Yes")
			} else {
				color.Red("  No")
			}

			color.Cyan("\nSkip Confirmation:")
			skip := false
			if cfg.SkipConfirmation != nil {
				skip = *cfg.SkipConfirmation
			}
			if skip {
				color.Green("  Yes")
			} else {
				color.Red("  No")
			}

			color.Cyan("\nUpdate Channel:")
			buildFromSource := false
			if cfg.BuildFromSource != nil {
				buildFromSource = *cfg.BuildFromSource
			}
			if buildFromSource {
				color.Green("  Build From Source (Beta)")
			} else {
				color.Yellow("  Stable (Binary)")
			}

			color.Cyan("\nAuto-Update Enabled:")
			auto := true
			if cfg.AutoUpdate != nil {
				auto = *cfg.AutoUpdate
			}
			if auto {
				color.Green("  Yes")
			} else {
				color.Red("  No")
			}
		},
	}
	rootCmd.AddCommand(getConfigCmd)

	var resetConfigCmd = &cobra.Command{
		Use:   "reset-config",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			defaultCfg := config.DefaultConfig()
			if err := config.SaveConfig(defaultCfg); err != nil {
				return err
			}
			color.Green("✓ Configuration reset to defaults!")
			return nil
		},
	}
	rootCmd.AddCommand(resetConfigCmd)

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of autocommiter",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("autocommiter version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("build date: %s\n", date)
		},
	}
	rootCmd.AddCommand(versionCmd)

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Self-update autocommiter to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updater.SeamlessUpdate(version)
		},
	}
	rootCmd.AddCommand(updateCmd)

	var cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Remove all application data and reset configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return performClean()
		},
	}
	rootCmd.AddCommand(cleanCmd)

	var shouldClean bool
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Remove autocommiter binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return performUninstall(shouldClean)
		},
	}
	uninstallCmd.Flags().BoolVar(&shouldClean, "clean", false, "Remove autocommiter binary and all configuration data")
	rootCmd.AddCommand(uninstallCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func performClean() error {
	// Clean up application data directory
	dataDir, _ := config.GetDataDir()
	if dataDir != "" {
		if _, err := os.Stat(dataDir); err == nil {
			color.Cyan("🗑️ Removing application data directory: %s", dataDir)
			err := os.RemoveAll(dataDir)
			if err != nil {
				return fmt.Errorf("failed to remove data directory: %w", err)
			}
		}
	}

	// Also clean up legacy files if they exist
	home, _ := os.UserHomeDir()
	if home != "" {
		legacyConfig := filepath.Join(home, ".autocommiter.json")
		legacyModels := filepath.Join(home, ".autocommiter.models.json")
		_ = os.Remove(legacyConfig)
		_ = os.Remove(legacyModels)
	}

	color.Green("✨ All application data and configuration have been cleared.")
	return nil
}

func performUninstall(clean bool) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}

	if clean {
		if err := performClean(); err != nil {
			return err
		}
	}

	color.Red("🧼 Uninstalling autocommiter...")
	color.New(color.Faint).Printf("Removing binary at: %s\n", exe)

	err = os.Remove(exe)
	if err != nil {
		return fmt.Errorf("failed to remove binary: %w (try running with sudo)", err)
	}

	color.Green("✨ autocommiter has been uninstalled.")
	return nil
}
