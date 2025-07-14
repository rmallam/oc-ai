package llm

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"google.golang.org/api/option"
)

// Client interface for LLM operations
type Client interface {
	GenerateResponse(prompt string) (string, error)
	GenerateSpecializedResponse(req *PromptRequest) (string, error)
	GetAlternativeAnalysis(originalQuery string) (string, error)
}

// EnhancedClient extends the basic client with specialized SRE capabilities
type EnhancedClient interface {
	Client
	GenerateTroubleshootingResponse(issue, symptoms, logs string) (string, error)
	GenerateSecurityReview(yamlContent string) (string, error)
	GenerateIncidentResponse(incidentType, severity, affectedServices string) (string, error)
	GeneratePerformanceAnalysis(metrics, issues string) (string, error)
	GenerateCapacityPlanningGuidance(currentUsage, projectedGrowth string) (string, error)
}

// GeminiClient implements LLM client using Google Gemini
type GeminiClient struct {
	client        *genai.Client
	model         *genai.GenerativeModel
	config        *config.Config
	promptManager *PromptManager
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

// NewGeminiClient creates a new Gemini client with OpenShift knowledge
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

	// Configure the model for OpenShift SRE responses
	model.SetTemperature(0.2) // Lower temperature for more consistent technical responses
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(4096) // Increased for detailed SRE responses

	return &GeminiClient{
		client:        client,
		model:         model,
		config:        cfg,
		promptManager: NewPromptManager(),
	}, nil
}

// GenerateResponse generates a response for a given prompt using OpenShift knowledge
func (g *GeminiClient) GenerateResponse(prompt string) (string, error) {
	ctx := context.Background()

	// Use the prompt manager to inject OpenShift knowledge
	enhancedPrompt := g.promptManager.InjectGeneralKnowledge(prompt)

	resp, err := g.model.GenerateContent(ctx, genai.Text(enhancedPrompt))
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

// GenerateSpecializedResponse generates a specialized response for specific SRE scenarios
func (g *GeminiClient) GenerateSpecializedResponse(req *PromptRequest) (string, error) {
	ctx := context.Background()

	specializedPrompt, err := g.promptManager.GenerateSpecializedPrompt(req)
	if err != nil {
		return "", fmt.Errorf("failed to generate specialized prompt: %w", err)
	}

	resp, err := g.model.GenerateContent(ctx, genai.Text(specializedPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate specialized content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no specialized response generated")
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

// GetAlternativeAnalysis provides alternative analysis using OpenShift expertise
func (g *GeminiClient) GetAlternativeAnalysis(originalQuery string) (string, error) {
	ctx := context.Background()

	// Use specialized troubleshooting knowledge for alternative analysis
	req := &PromptRequest{
		Type:      "troubleshooting",
		UserQuery: originalQuery,
		Context: map[string]string{
			"analysis_type": "alternative",
		},
	}

	alternativePrompt, err := g.promptManager.GenerateSpecializedPrompt(req)
	if err != nil {
		return "", fmt.Errorf("failed to generate alternative analysis prompt: %w", err)
	}

	// Add specific instruction for alternative perspective
	enhancedPrompt := fmt.Sprintf(`%s

ALTERNATIVE ANALYSIS REQUEST:
The user has declined the initial analysis and wants a different approach.
Provide:
1. Alternative root cause analysis using different reasoning
2. Different solution strategies
3. Cross-correlation patterns with other cluster issues
4. Advanced troubleshooting techniques

Focus on creative, systematic approaches that might not be immediately obvious.`, alternativePrompt)

	resp, err := g.model.GenerateContent(ctx, genai.Text(enhancedPrompt))
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

	return fmt.Sprintf("🤖 **Alternative OpenShift SRE Analysis:**\n\n%s", response), nil
}

// GenerateTroubleshootingResponse generates specialized troubleshooting responses
func (g *GeminiClient) GenerateTroubleshootingResponse(issue, symptoms, logs string) (string, error) {
	req := &PromptRequest{
		Type:      "troubleshooting",
		UserQuery: issue,
		Context: map[string]string{
			"symptoms": symptoms,
			"logs":     logs,
		},
		Environment: "production",
	}

	return g.GenerateSpecializedResponse(req)
}

// GenerateSecurityReview generates comprehensive security analysis for YAML configurations
func (g *GeminiClient) GenerateSecurityReview(yamlContent string) (string, error) {
	req := &PromptRequest{
		Type:      "security",
		UserQuery: "Please perform a comprehensive security review of this OpenShift configuration",
		Context: map[string]string{
			"yaml_content":         yamlContent,
			"compliance_framework": "CIS Kubernetes Benchmark",
		},
		Environment: "production",
	}

	return g.GenerateSpecializedResponse(req)
}

// GenerateIncidentResponse generates incident-specific response guidance
func (g *GeminiClient) GenerateIncidentResponse(incidentType, severity, affectedServices string) (string, error) {
	req := &PromptRequest{
		Type:      "incident",
		UserQuery: fmt.Sprintf("Critical incident: %s", incidentType),
		Context: map[string]string{
			"affected_services": affectedServices,
			"incident_type":     incidentType,
		},
		Severity:    severity,
		Environment: "production",
	}

	return g.GenerateSpecializedResponse(req)
}

// GeneratePerformanceAnalysis generates performance-focused analysis
func (g *GeminiClient) GeneratePerformanceAnalysis(metrics, issues string) (string, error) {
	req := &PromptRequest{
		Type:      "performance",
		UserQuery: "Analyze performance issues and provide optimization recommendations",
		Context: map[string]string{
			"metrics": metrics,
			"issues":  issues,
		},
		Environment: "production",
	}

	return g.GenerateSpecializedResponse(req)
}

// GenerateCapacityPlanningGuidance generates capacity planning recommendations
func (g *GeminiClient) GenerateCapacityPlanningGuidance(currentUsage, projectedGrowth string) (string, error) {
	req := &PromptRequest{
		Type:      "capacity",
		UserQuery: "Provide capacity planning guidance for OpenShift cluster scaling",
		Context: map[string]string{
			"current_usage":    currentUsage,
			"projected_growth": projectedGrowth,
		},
		Environment: "production",
	}

	return g.GenerateSpecializedResponse(req)
}

// Close closes the client connection
func (g *GeminiClient) Close() error {
	return g.client.Close()
}
