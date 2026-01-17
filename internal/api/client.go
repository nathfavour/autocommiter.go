package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func CallInferenceAPI(apiKey, prompt, model string) (string, error) {
	url := "https://models.inference.ai.azure.com/chat/completions"

	request := ChatCompletionRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant that generates concise, informative git commit messages. Reply only with the commit message, nothing else.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Model: model,
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var responseData ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return "", err
	}

	if len(responseData.Choices) > 0 {
		return strings.TrimSpace(responseData.Choices[0].Message.Content), nil
	}

	return "", fmt.Errorf("unexpected API response format")
}

func GenerateCommitMessage(apiKey, fileNames, compressedJSON, model string) (string, error) {
	prompt := fmt.Sprintf(
		"reply only with a very concise but informative commit message, and nothing else:\n\nFiles:\n%s\n\nSummaryJSON:%s",
		fileNames, compressedJSON,
	)

	return CallInferenceAPI(apiKey, prompt, model)
}
