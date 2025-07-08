package llm

import (
	"context"
	"fmt"

	"github.com/google/generativeai-go/genai"
	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"google.golang.org/api/option"
)

// Client interface for LLM operations
type Client interface {
	GenerateResponse(prompt string) (string, error)
	GetAlternativeAnalysis(originalQuery string) (string, error)
}

// GeminiClient implements LLM client using Google Gemini
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
	config *config.Config
}

// NewClient creates a new LLM client based on configuration
func NewClient(cfg *config.Config) (Client, error) {
	switch cfg.LLMProvider {
	case "gemini":
		return NewGeminiClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLMProvider)
	}
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(cfg *config.Config) (*GeminiClient, error) {
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(cfg.Model)
	
	// Configure the model
	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(2048)

	return &GeminiClient{
		client: client,
		model:  model,
		config: cfg,
	}, nil
}

// GenerateResponse generates a response for a given prompt
func (g *GeminiClient) GenerateResponse(prompt string) (string, error) {
	ctx := context.Background()

	// Create a system prompt for OpenShift SRE context
	systemPrompt := `You are an expert OpenShift SRE assistant. You help users manage OpenShift clusters using natural language commands. 
Provide clear, actionable responses for OpenShift operations including:
- Listing and managing resources (pods, deployments, services, etc.)
- Troubleshooting and diagnostics
- Deployment and scaling operations
- Security and RBAC management

Always provide specific OpenShift/kubectl commands when applicable.`

	fullPrompt := fmt.Sprintf("%s\n\nUser Query: %s", systemPrompt, prompt)

	resp, err := g.model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	// Extract text from the first candidate
	var response string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			response += string(textPart)
		}
	}

	return response, nil
}

// GetAlternativeAnalysis provides alternative analysis using AI
func (g *GeminiClient) GetAlternativeAnalysis(originalQuery string) (string, error) {
	ctx := context.Background()

	alternativePrompt := fmt.Sprintf(`You are providing an alternative analysis perspective for this OpenShift issue.
The user has declined the initial analysis and wants a different approach.

Original Query: %s

Please provide:
1. Alternative root cause analysis using different reasoning
2. Different solution strategies 
3. Cross-correlation patterns with other cluster issues
4. Advanced troubleshooting techniques

Focus on creative, systematic approaches that might not be immediately obvious.`, originalQuery)

	resp, err := g.model.GenerateContent(ctx, genai.Text(alternativePrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate alternative analysis: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no alternative analysis generated")
	}

	// Extract text from the first candidate
	var response string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			response += string(textPart)
		}
	}

	return fmt.Sprintf("🤖 **Alternative AI Analysis:**\n\n%s", response), nil
}

// Close closes the client connection
func (g *GeminiClient) Close() error {
	return g.client.Close()
}
