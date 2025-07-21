package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
)

// OpenAIClient handles OpenAI API interactions
type OpenAIClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
	Stream      bool            `json:"stream"`
}

// OpenAIMessage represents a message in the OpenAI API
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int           `json:"index"`
		Message      OpenAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient() *OpenAIClient {
	return &OpenAIClient{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		BaseURL: "https://api.openai.com/v1",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Enhanced chat handler with real OpenAI integration
func (h *EnhancedChatHandler) callOpenAIGPT4Real(prompt string) (string, error) {
	client := NewOpenAIClient()

	if client.APIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	request := OpenAIRequest{
		Model: "gpt-4",
		Messages: []OpenAIMessage{
			{
				Role:    "system",
				Content: "You are an expert OpenShift/Kubernetes administrator. Respond only with valid JSON.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1,
		MaxTokens:   1500,
		Stream:      false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST",
		client.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.APIKey)

	resp, err := client.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// Ollama integration for local LLMs
func (h *EnhancedChatHandler) callOllamaReal(prompt string) (string, error) {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.1"
	}

	request := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1,
			"top_p":       0.9,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Post(endpoint+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Response, nil
}

// Gemini integration for real LLM calls
func (h *EnhancedChatHandler) callGeminiReal(prompt string) (string, error) {
	if h.config == nil || h.config.LLM.Gemini.APIKey == "" {
		return "", fmt.Errorf("Gemini API key not configured")
	}

	// Create a temporary config for the Gemini client
	geminiConfig := &config.Config{
		GeminiAPIKey: h.config.LLM.Gemini.APIKey,
		Model:        h.config.LLM.Gemini.Model,
		LLMProvider:  "gemini",
	}

	// Create Gemini client
	client, err := llm.NewGeminiClient(geminiConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Generate response
	response, err := client.GenerateResponse(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate Gemini response: %w", err)
	}

	return response, nil
}

// Updated callLLMForPlanning with real integrations
func (h *EnhancedChatHandler) callLLMForPlanningReal(prompt string) (string, error) {
	var provider string
	if h.config != nil {
		provider = h.config.LLM.Provider
	} else {
		provider = os.Getenv("LLM_PROVIDER")
	}

	if provider == "" {
		provider = "mock" // Default to mock for backward compatibility
	}

	switch provider {
	case "openai":
		return h.callOpenAIGPT4Real(prompt)
	case "claude":
		// TODO: Implement real Claude integration
		return h.generateIntelligentMockResponse(prompt)
	case "gemini":
		return h.callGeminiReal(prompt)
	case "ollama":
		return h.callOllamaReal(prompt)
	case "mock":
		return h.generateIntelligentMockResponse(prompt)
	default:
		return "", fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

// Enhanced planning with real LLM integration
func (h *EnhancedChatHandler) planExecutionWithLLM(query string) (*ExecutionPlan, error) {
	// Build the planning prompt
	prompt := h.buildPlanningPrompt(query)

	// Try LLM first
	llmResponse, err := h.callLLMForPlanningReal(prompt)
	if err != nil {
		fmt.Printf("LLM planning failed, falling back to static patterns: %v\n", err)
		return h.planWithStaticPatterns(query)
	}

	// Parse LLM response
	plan, err := h.parseLLMPlanResponse(llmResponse)
	if err != nil {
		fmt.Printf("LLM response parsing failed, falling back to static patterns: %v\n", err)
		return h.planWithStaticPatterns(query)
	}

	return plan, nil
}
