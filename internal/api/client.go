package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nathfavour/autocommiter.go/internal/netutil"
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

const SystemPrompt = `You are an expert software engineer specializing in high-quality git commit messages.
Your task is to generate a concise, professional, and descriptive commit message based on the provided diffs and file changes.

Follow these rules:
1. Format: Use the Conventional Commits specification (e.g., feat: ..., fix: ..., docs: ..., refactor: ..., chore: ..., style: ..., test: ...).
2. Subject Line:
   - Must be under 50 characters if possible, never exceeding 72.
   - Use the imperative mood (e.g., "add", not "added" or "adds").
   - Do not end with a period.
3. Body (Optional):
   - Only include a body if the changes are complex and require explanation.
   - Separate the subject from the body with a blank line.
   - Explain 'what' and 'why', not 'how'.
4. Specificity: Be specific. Instead of "update files", say "refactor auth logic in client.go".
5. Output: Return ONLY the commit message text. No markdown, no "Commit message:", no quotes.

Context:
- Current branch: %s
`

func CallInferenceAPI(apiKey, branch, prompt, model string) (string, error) {
	url := "https://models.inference.ai.azure.com/chat/completions"

	request := ChatCompletionRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: fmt.Sprintf(SystemPrompt, branch),
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

	client := netutil.GetHttpClient()
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

func GenerateCommitMessage(apiKey, branch, fileNames, compressedJSON, model string) (string, error) {
	prompt := fmt.Sprintf(
		"Generate a commit message for the following changes:\n\nFiles changed:\n%s\n\nDetailed changes (JSON):\n%s",
		fileNames, compressedJSON,
	)

	return CallInferenceAPI(apiKey, branch, prompt, model)
}
