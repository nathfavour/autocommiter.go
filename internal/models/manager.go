package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/nathfavour/autocommiter.go/internal/auth"
	"github.com/nathfavour/autocommiter.go/internal/config"
)

type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	FriendlyName *string  `json:"friendly_name"`
	Publisher    *string  `json:"publisher"`
	Summary      *string  `json:"summary"`
	Task         *string  `json:"task"`
	Tags         []string `json:"tags"`
}

type CachedModels struct {
	Models []ModelInfo `json:"models"`
}

var DEFAULT_MODELS = []struct {
	ID           string
	FriendlyName string
	Summary      string
}{
	{"gpt-4o-mini", "OpenAI GPT-4o mini", "Fast & cost-effective, great for most tasks"},
	{"gpt-4o", "OpenAI GPT-4o", "High quality, most capable model"},
	{"Phi-3-mini-128k-instruct", "Phi-3 mini 128k", "Lightweight, efficient open model"},
	{"Mistral-large", "Mistral Large", "Powerful open-source model"},
}

func GetDefaultModels() []ModelInfo {
	var models []ModelInfo
	task := "chat-completion"
	for _, dm := range DEFAULT_MODELS {
		friendlyName := dm.FriendlyName
		summary := dm.Summary
		models = append(models, ModelInfo{
			ID:           dm.ID,
			Name:         dm.ID,
			FriendlyName: &friendlyName,
			Summary:      &summary,
			Task:         &task,
		})
	}
	return models
}

func FetchAvailableModels(apiKey string) ([]ModelInfo, error) {
	url := "https://models.inference.ai.azure.com/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GetDefaultModels(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GetDefaultModels(), nil
	}

	var modelsResponse []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		return GetDefaultModels(), nil
	}

	var models []ModelInfo
	for _, m := range modelsResponse {
		task, _ := m["task"].(string)
		if task == "chat-completion" {
			name, _ := m["name"].(string)
			friendlyName, _ := m["friendly_name"].(string)
			publisher, _ := m["publisher"].(string)
			summary, _ := m["summary"].(string)

			var tags []string
			if ts, ok := m["tags"].([]interface{}); ok {
				for _, t := range ts {
					if s, ok := t.(string); ok {
						tags = append(tags, s)
					}
				}
			}

			models = append(models, ModelInfo{
				ID:           name,
				Name:         name,
				FriendlyName: &friendlyName,
				Publisher:    &publisher,
				Summary:      &summary,
				Task:         &task,
				Tags:         tags,
			})
		}
	}

	if len(models) == 0 {
		return GetDefaultModels(), nil
	}

	return models, nil
}

func GetModelsCacheFile() (string, error) {
	return config.GetModelsCacheFile()
}

func GetCachedModels() ([]ModelInfo, error) {
	cacheFile, err := GetModelsCacheFile()
	if err != nil {
		return GetDefaultModels(), err
	}

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return GetDefaultModels(), nil
	}

	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return GetDefaultModels(), err
	}

	var cached CachedModels
	if err := json.Unmarshal(content, &cached); err != nil {
		return GetDefaultModels(), nil
	}

	return cached.Models, nil
}

func UpdateCachedModels(models []ModelInfo) error {
	cacheFile, err := GetModelsCacheFile()
	if err != nil {
		return err
	}

	cached := CachedModels{Models: models}
	content, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, content, 0644)
}

func RefreshModelList(apiKey string) (bool, string, int, error) {
	token := auth.GetToken(apiKey)
	if token == "" {
		return false, "No API key provided and GitHub CLI not authenticated", 0, fmt.Errorf("authentication required")
	}

	models, err := FetchAvailableModels(token)
	if err != nil {
		return false, fmt.Sprintf("Failed to fetch models: %v", err), 0, err
	}

	if len(models) == 0 {
		return false, "No chat-completion models found", 0, nil
	}

	if err := UpdateCachedModels(models); err != nil {
		return false, fmt.Sprintf("Failed to cache models: %v", err), 0, err
	}

	return true, fmt.Sprintf("Successfully fetched and cached %d models", len(models)), len(models), nil
}

func ListAvailableModels() ([]ModelInfo, error) {
	return GetCachedModels()
}
