package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	mcpserver "github.com/rakeshkumarmallam/openshift-mcp-go/pkg/mcp"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/types"
)

// EnhancedChatRequest represents an enhanced chat request with iteration support
type EnhancedChatRequest struct {
	Prompt      string `json:"prompt" binding:"required"`
	MaxSteps    int    `json:"max_steps,omitempty"`   // Maximum number of iterative steps
	Interactive bool   `json:"interactive,omitempty"` // Whether to support interactive mode
	Profile     string `json:"profile,omitempty"`     // Profile to use (sre, developer, admin)
}

// EnhancedChatResponse represents an enhanced chat response with step-by-step execution
type EnhancedChatResponse struct {
	Response       string                 `json:"response"`
	Steps          []ExecutionStep        `json:"steps"`
	Analysis       *types.Analysis        `json:"analysis,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
	Completed      bool                   `json:"completed"`
	NextSuggestion string                 `json:"next_suggestion,omitempty"`
}

// ExecutionStep represents a single step in the execution process
type ExecutionStep struct {
	StepNumber int                    `json:"step_number"`
	Action     string                 `json:"action"`
	ToolUsed   string                 `json:"tool_used"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     string                 `json:"result"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
}

// EnhancedChatHandler handles enhanced chat requests with MCP tool integration
type EnhancedChatHandler struct {
	server         *mcpserver.Server
	maxSteps       int
	defaultProfile string
	config         *config.Config
}

// NewEnhancedChatHandler creates a new enhanced chat handler
func NewEnhancedChatHandler(server *mcpserver.Server, cfg *config.Config) *EnhancedChatHandler {
	return &EnhancedChatHandler{
		server:         server,
		maxSteps:       10, // Default maximum steps
		defaultProfile: "sre",
		config:         cfg,
	}
}

// extractNamespaceFromQuery extracts namespace from the query string
func (h *EnhancedChatHandler) extractNamespaceFromQuery(query string) string {
	// Common namespace patterns
	namespacePatterns := []string{
		"debugger",
		"kube-system",
		"openshift-",
		"default",
		"monitoring",
		"logging",
	}

	// Check for explicit namespace mentions
	for _, pattern := range namespacePatterns {
		if strings.Contains(query, pattern) {
			if pattern == "openshift-" {
				// Find the full openshift namespace
				words := strings.Fields(query)
				for _, word := range words {
					if strings.HasPrefix(word, "openshift-") {
						return word
					}
				}
			} else {
				return pattern
			}
		}
	}

	// Default to debugger namespace if no specific namespace found
	return "debugger"
}

// buildPlanningPrompt creates a prompt for LLM-based planning
func (h *EnhancedChatHandler) buildPlanningPrompt(query string) string {
	availableTools := []string{
		"list_pods - List pods in a namespace (parameters: namespace)",
		"list_namespaces - List all namespaces (no parameters needed)",
		"get_events - Get events from a namespace (parameters: namespace)",
		"get_resource - Get details about a specific resource (parameters: resource_type, name, namespace)",
		"create_configmap - Create a ConfigMap (parameters: name, namespace, data)",
		"create_namespace - Create a new namespace (parameters: namespace_name)",
		"create_resource - Create any Kubernetes resource (parameters: yaml, namespace)",
		"delete_resource - Delete a Kubernetes resource (parameters: resource_type, name, namespace)",
		"scale_deployment - Scale a deployment (parameters: name, namespace, replicas)",
		"apply_yaml - Apply YAML configuration (parameters: yaml, namespace)",
		"generate_yaml - Generate YAML for common resources (parameters: resource_type, name, namespace, image, replicas, data)",
		"openshift_diagnose - Diagnose OpenShift cluster issues",
	}

	prompt := fmt.Sprintf(`You are an expert OpenShift SRE. Given a user query, create an execution plan using ONLY the available MCP tools listed below.

User Query: "%s"

Available Tools (USE ONLY THESE):
%s

IMPORTANT: 
- Only use tools from the list above
- Do not create custom_script, manual_input, or any other tools
- Keep plans simple and direct
- For creating namespaces, use create_namespace tool
- For creating service accounts, use create_resource tool with proper YAML
- For simple queries like "list pods", use a single step with list_pods tool
- For common resources (deployment, service, configmap), use generate_yaml tool first, then apply_yaml
- For complex applications (skupper, operators, etc.), create specific YAML content and use apply_yaml
- When using apply_yaml, provide actual YAML content in the yaml parameter, not placeholder names

YAML Content Guidelines:
- Always provide complete, valid YAML content in the yaml parameter
- For Skupper v2: Use "quay.io/skupper/skupper-router:2.0" image with proper deployment YAML
- Include all required metadata: name, namespace, labels
- Use proper resource specifications (cpu, memory, ports)

Generate a JSON execution plan with this structure:
{
  "description": "Brief description of what will be done",
  "category": "troubleshooting|exploration|maintenance",
  "complexity": "low|medium|high",
  "steps": [
    {
      "action": "descriptive_action_name",
      "tool": "exact_tool_name_from_list",
      "parameters": {"param": "value"},
      "description": "What this step does",
      "required": true
    }
  ]
}

Return only the JSON, no explanations.`, query, strings.Join(availableTools, "\n"))

	return prompt
}

// callLLMForPlanning calls the LLM service for planning
func (h *EnhancedChatHandler) callLLMForPlanning(prompt string) (string, error) {
	var provider string
	if h.config != nil {
		provider = h.config.LLM.Provider
	} else {
		provider = os.Getenv("LLM_PROVIDER")
	}
	logrus.Debugf("LLM Provider: %s", provider)

	hasReal := h.hasRealLLMIntegration()
	logrus.Debugf("Has real LLM integration: %v", hasReal)

	// Check if we have real LLM integration available
	if hasReal {
		logrus.Debugf("Using real LLM integration")
		return h.callLLMForPlanningReal(prompt)
	}

	// Fall back to intelligent mock response
	logrus.Debugf("Falling back to mock response")
	return h.generateIntelligentMockResponse(prompt)
}

// hasRealLLMIntegration checks if real LLM integration is available
func (h *EnhancedChatHandler) hasRealLLMIntegration() bool {
	if h.config == nil {
		return false
	}

	provider := h.config.LLM.Provider
	switch provider {
	case "openai":
		return h.config.LLM.OpenAI.APIKey != ""
	case "claude":
		return h.config.LLM.Claude.APIKey != ""
	case "gemini":
		return h.config.LLM.Gemini.APIKey != ""
	case "ollama":
		return true // Ollama doesn't need API key
	default:
		return false
	}
}

// callOpenAIGPT4 integrates with OpenAI GPT-4
func (h *EnhancedChatHandler) callOpenAIGPT4(prompt string) (string, error) {
	// Example implementation - you'll need to add OpenAI client
	/*
		client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

		req := openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.1, // Low temperature for consistent planning
			MaxTokens:   1000,
		}

		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			return "", err
		}

		return resp.Choices[0].Message.Content, nil
	*/

	// For demonstration, return a dynamic response based on the query
	return h.generateIntelligentMockResponse(prompt)
}

// callGemini integrates with Google Gemini (already available in your server)
func (h *EnhancedChatHandler) callGemini(prompt string) (string, error) {
	// You can use the existing Gemini client from your server
	// This would integrate with the LLM client already configured
	return h.generateIntelligentMockResponse(prompt)
}

// callClaude integrates with Anthropic Claude
func (h *EnhancedChatHandler) callClaude(prompt string) (string, error) {
	// You can use the existing Claude client from your server
	// This would integrate with the Anthropic client already configured
	return h.generateIntelligentMockResponse(prompt)
}

// callOllama integrates with local Ollama LLM
func (h *EnhancedChatHandler) callOllama(prompt string) (string, error) {
	// Example Ollama integration
	/*
		payload := map[string]interface{}{
			"model":  "llama3.1",
			"prompt": prompt,
			"stream": false,
		}

		jsonPayload, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:11434/api/generate",
			"application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		var result struct {
			Response string `json:"response"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		return result.Response, nil
	*/

	return h.generateIntelligentMockResponse(prompt)
}

// generateIntelligentMockResponse creates context-aware mock responses
func (h *EnhancedChatHandler) generateIntelligentMockResponse(prompt string) (string, error) {
	// Extract the user query from the prompt
	query := h.extractQueryFromPrompt(prompt)
	queryLower := strings.ToLower(query)

	// Generate intelligent responses based on query analysis
	if strings.Contains(queryLower, "fix") && strings.Contains(queryLower, "pod") {
		return `{
  "description": "Diagnose and fix failing pods with comprehensive analysis",
  "category": "troubleshooting",
  "complexity": "medium",
  "steps": [
    {
      "action": "list_pods",
      "tool": "list_pods",
      "parameters": {"namespace": "debugger"},
      "description": "List all pods to identify failing ones",
      "required": true
    },
    {
      "action": "get_events",
      "tool": "get_events",
      "parameters": {"namespace": "debugger"},
      "description": "Get events to understand failure reasons",
      "required": true
    },
    {
      "action": "openshift_diagnose",
      "tool": "openshift_diagnose",
      "parameters": {"resource_type": "pod", "resource_name": "failing-pod", "namespace": "debugger"},
      "description": "Perform detailed diagnosis with specific recommendations",
      "required": true
    }
  ]
}`, nil
	}

	if strings.Contains(queryLower, "scale") && strings.Contains(queryLower, "deployment") {
		return `{
  "description": "Scale deployment with validation",
  "category": "maintenance",
  "complexity": "low",
  "steps": [
    {
      "action": "get_deployment_status",
      "tool": "get_resource",
      "parameters": {"resource_type": "deployment", "resource_name": "target-deployment", "namespace": "default"},
      "description": "Check current deployment status",
      "required": true
    },
    {
      "action": "scale_deployment",
      "tool": "scale_deployment",
      "parameters": {"deployment_name": "target-deployment", "namespace": "default", "replicas": "3"},
      "description": "Scale deployment to desired replicas",
      "required": true
    }
  ]
}`, nil
	}

	// Check for namespace-specific queries
	if strings.Contains(queryLower, "namespace") || strings.Contains(queryLower, "namepsace") {
		return `{
  "description": "List all namespaces in the cluster",
  "category": "exploration",
  "complexity": "low",
  "steps": [
    {
      "action": "list_namespaces",
      "tool": "list_namespaces",
      "parameters": {},
      "description": "List all namespaces in the cluster",
      "required": true
    }
  ]
}`, nil
	}

	// Check for pod-specific queries
	if strings.Contains(queryLower, "pod") {
		// Check if they want all pods or specific namespace
		if strings.Contains(queryLower, "all") {
			return `{
  "description": "List pods across all namespaces",
  "category": "exploration",
  "complexity": "low",
  "steps": [
    {
      "action": "list_namespaces",
      "tool": "list_namespaces",
      "parameters": {},
      "description": "List all namespaces first",
      "required": true
    },
    {
      "action": "list_pods",
      "tool": "list_pods",
      "parameters": {"namespace": "debugger"},
      "description": "List pods in debugger namespace",
      "required": true
    }
  ]
}`, nil
		} else {
			return `{
  "description": "List pods in specific namespace",
  "category": "exploration",
  "complexity": "low",
  "steps": [
    {
      "action": "list_pods",
      "tool": "list_pods",
      "parameters": {"namespace": "debugger"},
      "description": "List pods in debugger namespace",
      "required": true
    }
  ]
}`, nil
		}
	}

	// Default exploratory response
	return `{
  "description": "Explore cluster resources and status",
  "category": "exploration",
  "complexity": "low",
  "steps": [
    {
      "action": "list_namespaces",
      "tool": "list_namespaces",
      "parameters": {},
      "description": "List all namespaces in the cluster",
      "required": true
    },
    {
      "action": "list_pods",
      "tool": "list_pods",
      "parameters": {"namespace": "debugger"},
      "description": "List pods in the debugger namespace",
      "required": true
    }
  ]
}`, nil
}

// extractQueryFromPrompt extracts the user query from the planning prompt
func (h *EnhancedChatHandler) extractQueryFromPrompt(prompt string) string {
	// Simple extraction - in production, make this more robust
	lines := strings.Split(prompt, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "User Query:") {
			return strings.Trim(strings.TrimPrefix(line, "User Query:"), " \"")
		}
	}
	return ""
}

// parseLLMPlanResponse parses LLM response into ExecutionPlan
func (h *EnhancedChatHandler) parseLLMPlanResponse(response string) (*ExecutionPlan, error) {
	var plan ExecutionPlan

	// Clean up markdown formatting if present
	jsonStr := strings.TrimSpace(response)

	// Remove markdown code block formatting
	if strings.HasPrefix(jsonStr, "```json") {
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
	}
	if strings.HasPrefix(jsonStr, "```") {
		jsonStr = strings.TrimPrefix(jsonStr, "```")
	}
	if strings.HasSuffix(jsonStr, "```") {
		jsonStr = strings.TrimSuffix(jsonStr, "```")
	}

	jsonStr = strings.TrimSpace(jsonStr)

	logrus.Debugf("Cleaned JSON string: %s", jsonStr)

	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %v", err)
	}

	// Validate the parsed plan
	if plan.Description == "" {
		return nil, fmt.Errorf("plan description is required")
	}

	if len(plan.Steps) == 0 {
		return nil, fmt.Errorf("plan must have at least one step")
	}

	return &plan, nil
}

// parseLLMPlanResponseWithQuery parses LLM response and adds query field
func (h *EnhancedChatHandler) parseLLMPlanResponseWithQuery(query, response string) (*ExecutionPlan, error) {
	logrus.Debugf("Parsing LLM response for query: %s", query)
	logrus.Debugf("LLM response: %s", response)

	plan, err := h.parseLLMPlanResponse(response)
	if err != nil {
		logrus.Debugf("Failed to parse LLM response: %v", err)
		return nil, err
	}

	// Add the query field which is missing from the LLM response
	plan.Query = query

	return plan, nil
}

// HandleEnhancedChat handles enhanced chat requests
func (h *EnhancedChatHandler) HandleEnhancedChat(c *gin.Context) {
	var req EnhancedChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.MaxSteps == 0 {
		req.MaxSteps = h.maxSteps
	}
	if req.Profile == "" {
		req.Profile = h.defaultProfile
	}

	logrus.WithFields(logrus.Fields{
		"prompt":      req.Prompt,
		"max_steps":   req.MaxSteps,
		"interactive": req.Interactive,
		"profile":     req.Profile,
	}).Debug("Processing enhanced chat request")

	// Execute the request with iterative capability
	response, err := h.executeIterativeQuery(c.Request.Context(), req)
	if err != nil {
		logrus.WithError(err).Error("Failed to execute iterative query")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// executeIterativeQuery executes a query with iterative capability like Claude Desktop
func (h *EnhancedChatHandler) executeIterativeQuery(ctx context.Context, req EnhancedChatRequest) (*EnhancedChatResponse, error) {
	response := &EnhancedChatResponse{
		Steps:     make([]ExecutionStep, 0),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"profile":     req.Profile,
			"max_steps":   req.MaxSteps,
			"interactive": req.Interactive,
		},
	}

	// Parse the initial query to determine the execution plan
	executionPlan, err := h.planExecution(req.Prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to plan execution: %w", err)
	}

	// Execute the plan step by step
	for i, step := range executionPlan.Steps {
		if i >= req.MaxSteps {
			response.Response += fmt.Sprintf("\n⚠️  Maximum steps (%d) reached. Execution truncated.", req.MaxSteps)
			break
		}

		executionStep := h.executeStep(ctx, i+1, step)
		response.Steps = append(response.Steps, executionStep)

		// Check if we should continue based on the step result
		if !executionStep.Success {
			response.Response += fmt.Sprintf("\n❌ Step %d failed: %s", i+1, executionStep.Error)
			response.Completed = false
			return response, nil
		}

		// Add step result to overall response
		response.Response += fmt.Sprintf("\n📋 Step %d: %s", i+1, executionStep.Result)
	}

	// Generate final summary
	response.Response = h.generateSummary(executionPlan, response.Steps) + response.Response
	response.Completed = true

	// Generate next suggestion if applicable
	if req.Interactive {
		response.NextSuggestion = h.generateNextSuggestion(executionPlan, response.Steps)
	}

	return response, nil
}

// ExecutionPlan represents a plan for executing a complex query
type ExecutionPlan struct {
	Query       string        `json:"query"`
	Description string        `json:"description"`
	Steps       []PlannedStep `json:"steps"`
	Category    string        `json:"category"`
	Complexity  string        `json:"complexity"`
}

// PlannedStep represents a planned step in the execution
type PlannedStep struct {
	Action      string                 `json:"action"`
	Tool        string                 `json:"tool"`
	Parameters  map[string]interface{} `json:"parameters"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
}

// planExecution creates an execution plan for a given query
func (h *EnhancedChatHandler) planExecution(query string) (*ExecutionPlan, error) {
	// Try LLM-powered planning first, fallback to static patterns
	plan, err := h.planWithLLM(query)
	if err == nil {
		logrus.Debugf("LLM planning succeeded for query: %s", query)
		return plan, nil
	}

	logrus.Debugf("LLM planning failed: %v, falling back to static patterns", err)
	// Fallback to static pattern matching
	return h.planWithStaticPatterns(query)
}

// planWithLLM uses LLM to generate intelligent execution plans
func (h *EnhancedChatHandler) planWithLLM(query string) (*ExecutionPlan, error) {
	// Create a prompt for the LLM to generate execution plan
	prompt := h.buildPlanningPrompt(query)

	// Call your LLM service (you'll need to implement this)
	llmResponse, err := h.callLLMForPlanning(prompt)
	if err != nil {
		return nil, err
	}

	// Parse the LLM response into an ExecutionPlan
	plan, err := h.parseLLMPlanResponseWithQuery(query, llmResponse)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

// planWithStaticPatterns - the existing static pattern matching logic
func (h *EnhancedChatHandler) planWithStaticPatterns(query string) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Query: query,
		Steps: make([]PlannedStep, 0),
	}
	// Analyze the query to determine appropriate steps
	queryLower := strings.ToLower(query)

	// Extract namespace from query
	namespace := h.extractNamespaceFromQuery(queryLower)

	// Pod troubleshooting queries (fix, debug, diagnose, troubleshoot)
	if strings.Contains(queryLower, "fix") && strings.Contains(queryLower, "pod") {
		plan.Description = "Diagnose and fix failing pods"
		plan.Category = "troubleshooting"
		plan.Complexity = "medium"
		plan.Steps = []PlannedStep{
			{
				Action:      "list_pods",
				Tool:        "list_pods",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("List all pods in %s namespace to identify failing ones", namespace),
				Required:    true,
			},
			{
				Action:      "get_events",
				Tool:        "get_events",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("Get events from %s namespace for diagnostic clues", namespace),
				Required:    true,
			},
			{
				Action:      "openshift_diagnose",
				Tool:        "openshift_diagnose",
				Parameters:  map[string]interface{}{"resource_type": "pod", "resource_name": "failing-pod", "namespace": namespace},
				Description: fmt.Sprintf("Diagnose pod issues in %s namespace", namespace),
				Required:    true,
			},
		}
	} else if (strings.Contains(queryLower, "debug") || strings.Contains(queryLower, "diagnose") || strings.Contains(queryLower, "troubleshoot")) && strings.Contains(queryLower, "pod") {
		plan.Description = "Debug and diagnose pod issues"
		plan.Category = "troubleshooting"
		plan.Complexity = "medium"
		plan.Steps = []PlannedStep{
			{
				Action:      "list_pods",
				Tool:        "list_pods",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("List all pods in %s namespace", namespace),
				Required:    true,
			},
			{
				Action:      "get_events",
				Tool:        "get_events",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("Get events from %s namespace", namespace),
				Required:    true,
			},
			{
				Action:      "openshift_diagnose",
				Tool:        "openshift_diagnose",
				Parameters:  map[string]interface{}{"resource_type": "pod", "resource_name": "pod", "namespace": namespace},
				Description: fmt.Sprintf("Diagnose pod issues in %s namespace", namespace),
				Required:    true,
			},
		}
	} else if strings.Contains(queryLower, "pod") {
		plan.Description = "List and analyze pods"
		plan.Category = "exploration"
		plan.Complexity = "low"
		plan.Steps = []PlannedStep{
			{
				Action:      "list_pods",
				Tool:        "list_pods",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("List all pods in %s namespace", namespace),
				Required:    true,
			},
		}
	} else if strings.Contains(queryLower, "event") {
		plan.Description = "Check cluster events"
		plan.Category = "monitoring"
		plan.Complexity = "low"
		plan.Steps = []PlannedStep{
			{
				Action:      "get_events",
				Tool:        "get_events",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("Get events from %s namespace", namespace),
				Required:    true,
			},
		}
	} else if strings.Contains(queryLower, "namespace") {
		plan.Description = "List and analyze namespaces"
		plan.Category = "exploration"
		plan.Complexity = "low"
		plan.Steps = []PlannedStep{
			{
				Action:      "list_namespaces",
				Tool:        "list_namespaces",
				Parameters:  map[string]interface{}{},
				Description: "List all namespaces",
				Required:    true,
			},
		}
	} else {
		// Generic query - use basic pod listing
		plan.Description = "General cluster exploration"
		plan.Category = "exploration"
		plan.Complexity = "low"
		plan.Steps = []PlannedStep{
			{
				Action:      "list_pods",
				Tool:        "list_pods",
				Parameters:  map[string]interface{}{"namespace": namespace},
				Description: fmt.Sprintf("List pods in %s namespace", namespace),
				Required:    true,
			},
		}
	}

	return plan, nil
}

// executeStep executes a single step using the appropriate MCP tool
func (h *EnhancedChatHandler) executeStep(ctx context.Context, stepNumber int, step PlannedStep) ExecutionStep {
	start := time.Now()

	logrus.Debugf("Executing step %d: %s using tool %s", stepNumber, step.Action, step.Tool)

	executionStep := ExecutionStep{
		StepNumber: stepNumber,
		Action:     step.Action,
		ToolUsed:   step.Tool,
		Parameters: step.Parameters,
		Timestamp:  start,
	}

	// Create MCP tool call request
	callRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      step.Tool,
			Arguments: step.Parameters,
		},
	}

	logrus.Debugf("About to call MCP tool: %s with params: %v", step.Tool, step.Parameters)

	// Execute the tool (this would need to be implemented to call the actual MCP tools)
	result, err := h.callMCPTool(ctx, callRequest)

	logrus.Debugf("MCP tool call completed for step %d: success=%v, error=%v", stepNumber, err == nil, err)

	executionStep.Duration = time.Since(start)

	if err != nil {
		executionStep.Success = false
		executionStep.Error = err.Error()
		executionStep.Result = fmt.Sprintf("Failed to execute %s: %s", step.Tool, err.Error())
	} else {
		executionStep.Success = true
		executionStep.Result = result
	}

	return executionStep
}

// callMCPTool calls an MCP tool using the dynamic tool execution pattern (like Claude)
func (h *EnhancedChatHandler) callMCPTool(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	// Use the same dynamic tool calling approach as Claude Desktop
	// This leverages the existing MCP infrastructure without hardcoded switch statements!

	logrus.Debugf("Dynamically calling MCP tool: %s with params: %v", request.Params.Name, request.Params.Arguments)

	// Use the MCP handler to execute the tool - this is the same mechanism Claude uses
	handler := NewMCPHandler(h.server)
	result, err := handler.executeTool(ctx, request)

	if err != nil {
		return "", fmt.Errorf("MCP tool execution failed: %w", err)
	}

	return result, nil
}

// extractTextFromMCPResult extracts text content from MCP result
func (h *EnhancedChatHandler) extractTextFromMCPResult(result *mcp.CallToolResult) string {
	if result != nil && len(result.Content) > 0 {
		// Try different ways to extract text content
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			return content.Text
		case mcp.TextContent:
			return content.Text
		default:
			// If it's not a TextContent, try to convert to string
			return fmt.Sprintf("%v", content)
		}
	}
	return "No result returned"
}

// generateSummary generates a summary of the execution
func (h *EnhancedChatHandler) generateSummary(plan *ExecutionPlan, steps []ExecutionStep) string {
	successCount := 0
	for _, step := range steps {
		if step.Success {
			successCount++
		}
	}

	summary := fmt.Sprintf("🎯 **%s**\n", plan.Description)
	summary += fmt.Sprintf("📊 Executed %d/%d steps successfully\n", successCount, len(steps))
	summary += fmt.Sprintf("⏱️  Total execution time: %s\n", h.calculateTotalDuration(steps))

	if successCount == len(steps) {
		summary += "✅ All steps completed successfully\n"
	} else {
		summary += "⚠️  Some steps failed - check details below\n"
	}

	return summary
}

// generateNextSuggestion generates a suggestion for the next action
func (h *EnhancedChatHandler) generateNextSuggestion(plan *ExecutionPlan, steps []ExecutionStep) string {
	if plan.Category == "troubleshooting" {
		return "Would you like me to help fix the identified issues or provide more detailed diagnostics?"
	} else if plan.Category == "operator_check" {
		return "Would you like me to check the operator logs or installation status?"
	}
	return "Would you like me to explore other aspects of your cluster?"
}

// calculateTotalDuration calculates the total duration of all steps
func (h *EnhancedChatHandler) calculateTotalDuration(steps []ExecutionStep) time.Duration {
	var total time.Duration
	for _, step := range steps {
		total += step.Duration
	}
	return total
}

// RegisterRoutes registers the enhanced chat routes
func (h *EnhancedChatHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/chat/enhanced", h.HandleEnhancedChat)
	}
}
