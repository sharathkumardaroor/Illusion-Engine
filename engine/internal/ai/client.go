package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	Enabled bool
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func New(enabled bool, baseURL, apiKey, model string) *Client {
	return &Client{
		Enabled: enabled,
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	}
}

func (c *Client) GenerateCommitMessage(diff string, context string) (string, error) {
	if !c.Enabled {
		return "chore: incremental update", nil
	}

	prompt := fmt.Sprintf("Generate a concise, professional git commit message for the following changes. Context: %s\n\nDiff:\n%s", context, diff)

	return c.chat(prompt)
}

func (c *Client) chat(prompt string) (string, error) {
	reqBody, _ := json.Marshal(ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: "You are a senior software engineer generating git history."},
			{Role: "user", Content: prompt},
		},
	})

	url := c.BaseURL
	if url == "" {
		url = "https://text.pollinations.ai/v1/chat/completions"
	} else {
		// URL Normalization
		if !strings.HasSuffix(url, "/chat/completions") {
			if !strings.HasSuffix(url, "/") {
				url += "/"
			}
			url += "chat/completions"
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API returned status %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from AI")
}
