package llm

import (
	"context"
	"fmt"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/sashabaranov/go-openai"
)

// Client represents an LLM client
type Client interface {
	GenerateResponse(prompt string) (string, error)
	GetAlternativeAnalysis(prompt string) (string, error)
}

// OpenAIClient represents an OpenAI client
type OpenAIClient struct {
	client *openai.Client
	config *config.Config
}

// NewClient creates a new LLM client
func NewClient(cfg *config.Config) (Client, error) {
	if cfg.LLMProvider == "openai" {
		return NewOpenAIClient(cfg)
	}
	return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLMProvider)
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(cfg *config.Config) (*OpenAIClient, error) {
	client := openai.NewClient(cfg.GeminiAPIKey)
	return &OpenAIClient{
		client: client,
		config: cfg,
	}, nil
}

// GenerateResponse generates a response from the LLM
func (c *OpenAIClient) GenerateResponse(prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: c.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// GetAlternativeAnalysis gets an alternative analysis from the LLM
func (c *OpenAIClient) GetAlternativeAnalysis(prompt string) (string, error) {
	return c.GenerateResponse("Provide an alternative analysis for the following problem: " + prompt)
}
