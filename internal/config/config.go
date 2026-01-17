package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey            *string  `json:"api_key"`
	SelectedModel     *string  `json:"selected_model"`
	EnableGitmoji     *bool    `json:"enable_gitmoji"`
	UpdateGitignore   *bool    `json:"update_gitignore"`
	SkipConfirmation  *bool    `json:"skip_confirmation"`
	GitignorePatterns []string `json:"gitignore_patterns"`
}

func DefaultConfig() Config {
	apiKey := ""
	selectedModel := "gpt-4o-mini"
	enableGitmoji := false
	updateGitignore := false
	skipConfirmation := false

	return Config{
		APIKey:            &apiKey,
		SelectedModel:     &selectedModel,
		EnableGitmoji:     &enableGitmoji,
		UpdateGitignore:   &updateGitignore,
		SkipConfirmation:  &skipConfirmation,
		GitignorePatterns: []string{"*.env*", ".env*", "docx/", ".docx/"},
	}
}

func GetConfigFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".autocommiter.json"), nil
}

func LoadConfig() (Config, error) {
	configFile, err := GetConfigFile()
	if err != nil {
		return DefaultConfig(), err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return DefaultConfig(), err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return DefaultConfig(), nil
	}

	return config, nil
}

func SaveConfig(config Config) error {
	configFile, err := GetConfigFile()
	if err != nil {
		return err
	}

	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, content, 0644)
}

func GetAPIKey() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	if config.APIKey == nil {
		return "", nil
	}
	return *config.APIKey, nil
}

func SetAPIKey(key string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.APIKey = &key
	return SaveConfig(config)
}

func GetSelectedModel() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "gpt-4o-mini", err
	}
	if config.SelectedModel == nil {
		return "gpt-4o-mini", nil
	}
	return *config.SelectedModel, nil
}

func SetSelectedModel(model string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.SelectedModel = &model
	return SaveConfig(config)
}
