package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/decision"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
	mcpserver "github.com/rakeshkumarmallam/openshift-mcp-go/pkg/mcp"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/memory"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/types"
	"github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	config         *config.Config
	engine         *gin.Engine
	decisionEngine *decision.Engine
	memory         *memory.Store
	llmClient      llm.Client
	mcpServer      *mcpserver.Server
	enhancedChat   *EnhancedChatHandler
}

// ChatRequest represents a chat API request
type ChatRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

// ChatResponse represents a chat API response
type ChatResponse struct {
	Response  string                 `json:"response"`
	Analysis  *types.Analysis        `json:"analysis,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// UserChoiceRequest represents a user choice API request
type UserChoiceRequest struct {
	Choice        string `json:"choice" binding:"required,oneof=accept decline more_info"`
	OriginalQuery string `json:"original_query"`
	AnalysisID    string `json:"analysis_id,omitempty"`
}

// UserChoiceResponse represents a user choice API response
type UserChoiceResponse struct {
	Response  string                 `json:"response"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewServer creates a new API server
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize components
	memStore, err := memory.NewStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize memory store: %w", err)
	}

	llmClient, err := llm.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	decisionEngine, err := decision.NewEngine(cfg, memStore, llmClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize decision engine: %w", err)
	}

	// Set gin mode based on debug setting
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	server := &Server{
		config:         cfg,
		engine:         engine,
		decisionEngine: decisionEngine,
		memory:         memStore,
		llmClient:      llmClient,
	}

	// Initialize MCP server if enabled
	if cfg.MCP.Enabled {
		if err := server.initializeMCP(); err != nil {
			logrus.WithError(err).Warn("Failed to initialize MCP server, continuing without MCP support")
		}
	}

	// Initialize enhanced chat handler
	if server.mcpServer != nil {
		server.enhancedChat = NewEnhancedChatHandler(server.mcpServer, server.config)
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.handleHealth)

	// Direct chat endpoint for convenience
	if s.enhancedChat != nil {
		s.engine.POST("/chat", s.handleEnhancedChatDirect)
	}

	// API routes
	api := s.engine.Group("/api/v1")
	{
		// Use enhanced chat if available, otherwise fall back to basic chat
		if s.enhancedChat != nil {
			api.POST("/chat", s.handleEnhancedChatDirect)
		} else {
			api.POST("/chat", s.handleChat)
		}

		api.POST("/user-choice", s.handleUserChoice)
		api.GET("/prompts/stats", s.handlePromptStats)
		api.POST("/prompts/update", s.handleUpdatePrompts)
		api.GET("/prompts/categories", s.handlePromptCategories)
	}

	// Enhanced chat routes (with LLM intelligence)
	if s.enhancedChat != nil {
		// Register enhanced chat routes (for /api/v1/chat/enhanced endpoint)
		s.enhancedChat.RegisterRoutes(s.engine)
	}

	// Static routes for web UI (if templates exist)
	templatesPath := "web/templates"
	if _, err := os.Stat(templatesPath); err == nil {
		// Templates directory exists, check if there are any template files
		templateFiles, err := filepath.Glob("web/templates/*")
		if err == nil && len(templateFiles) > 0 {
			s.engine.Static("/static", "./web/static")
			s.engine.LoadHTMLGlob("web/templates/*")
			s.engine.GET("/", s.handleIndex)
			logrus.Info("Web UI templates loaded successfully")
		} else {
			logrus.Warn("No template files found in web/templates/, web UI disabled")
		}
	} else {
		logrus.Warn("Web templates directory not found, web UI disabled")
	}
}

// Run starts the server
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	logrus.Infof("Starting server on %s", addr)
	return s.engine.Run(addr)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}

// handleIndex serves the main web interface
func (s *Server) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "OpenShift MCP - AI SRE Assistant",
	})
}

// handleChat handles chat requests with live execution
func (s *Server) handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithField("prompt", req.Prompt).Debug("Processing chat request with live execution")

	// If MCP server is not available, return error
	if s.mcpServer == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MCP server not available"})
		return
	}

	// Execute live command directly through MCP
	result, err := s.executeLiveCommand(req.Prompt)
	if err != nil {
		logrus.WithError(err).Error("Failed to execute live command")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Live execution failed: " + err.Error()})
		return
	}

	response := ChatResponse{
		Response:  result,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"execution_type": "live",
			"mcp_enabled":    true,
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleEnhancedChatDirect handles chat requests with LLM intelligence
func (s *Server) handleEnhancedChatDirect(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithField("prompt", req.Prompt).Debug("Processing enhanced chat request with LLM intelligence")

	// If enhanced chat handler is not available, fall back to regular chat
	if s.enhancedChat == nil {
		s.handleChat(c)
		return
	}

	// Convert to enhanced chat request
	enhancedReq := EnhancedChatRequest{
		Prompt:   req.Prompt,
		MaxSteps: 10,
		Profile:  "sre",
	}

	// Execute with enhanced chat handler
	response, err := s.enhancedChat.executeIterativeQuery(c.Request.Context(), enhancedReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to execute enhanced chat request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Enhanced chat execution failed: " + err.Error()})
		return
	}

	// Convert enhanced response to regular response format for compatibility
	chatResponse := ChatResponse{
		Response:  response.Response,
		Timestamp: response.Timestamp,
		Metadata:  response.Metadata,
	}

	c.JSON(http.StatusOK, chatResponse)
}

// executeLiveCommand executes a command directly through MCP tools
func (s *Server) executeLiveCommand(prompt string) (string, error) {
	// Simple command mapping based on keywords
	prompt = strings.ToLower(prompt)

	// Determine which MCP tool to use based on the prompt
	if strings.Contains(prompt, "pods") || strings.Contains(prompt, "pod") {
		return s.executeMCPTool("list_pods", map[string]interface{}{
			"namespace": s.extractNamespace(prompt),
		})
	}

	if strings.Contains(prompt, "namespaces") || strings.Contains(prompt, "namespace") {
		return s.executeMCPTool("list_namespaces", map[string]interface{}{})
	}

	if strings.Contains(prompt, "events") || strings.Contains(prompt, "event") {
		return s.executeMCPTool("get_events", map[string]interface{}{
			"namespace": s.extractNamespace(prompt),
		})
	}

	if strings.Contains(prompt, "resources") || strings.Contains(prompt, "resource") {
		return s.executeMCPTool("get_resource", map[string]interface{}{
			"resource_type": s.extractResourceType(prompt),
			"resource_name": s.extractResourceName(prompt),
			"namespace":     s.extractNamespace(prompt),
		})
	}

	if strings.Contains(prompt, "helm") {
		return s.executeMCPTool("helm_list", map[string]interface{}{
			"namespace": s.extractNamespace(prompt),
		})
	}

	if strings.Contains(prompt, "diagnose") || strings.Contains(prompt, "debug") {
		return s.executeMCPTool("openshift_diagnose", map[string]interface{}{
			"resource_type": s.extractResourceType(prompt),
			"resource_name": s.extractResourceName(prompt),
			"namespace":     s.extractNamespace(prompt),
		})
	}

	// Default to listing pods if no specific command is detected
	return s.executeMCPTool("list_pods", map[string]interface{}{
		"namespace": "default",
	})
}

// executeMCPTool executes a specific MCP tool using the handler
func (s *Server) executeMCPTool(toolName string, arguments map[string]interface{}) (string, error) {
	ctx := context.Background()

	// Create MCP request structure
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: arguments,
		},
	}

	// Use the MCP handler to execute the tool
	handler := NewMCPHandler(s.mcpServer)
	return handler.executeTool(ctx, request)
}

// extractNamespace extracts namespace from the prompt, defaults to "default"
func (s *Server) extractNamespace(prompt string) string {
	// Look for namespace patterns
	if strings.Contains(prompt, "namespace") {
		words := strings.Fields(prompt)
		for i, word := range words {
			if word == "namespace" && i+1 < len(words) {
				return words[i+1]
			}
			// Also check for "word namespace" pattern
			if i+1 < len(words) && words[i+1] == "namespace" {
				return words[i]
			}
		}
	}

	// Look for "in the X" pattern
	if strings.Contains(prompt, "in the") {
		words := strings.Fields(prompt)
		for i, word := range words {
			if word == "the" && i+1 < len(words) {
				next := words[i+1]
				// Don't return common words
				if next != "cluster" && next != "system" && next != "pod" && next != "pods" {
					return next
				}
			}
		}
	}

	// Look for common namespace names
	if strings.Contains(prompt, "kube-system") {
		return "kube-system"
	}
	if strings.Contains(prompt, "openshift") {
		return "openshift"
	}
	if strings.Contains(prompt, "monitoring") {
		return "openshift-monitoring"
	}
	if strings.Contains(prompt, "debugger") {
		return "debugger"
	}

	return "default"
}

// extractResourceType extracts resource type from the prompt
func (s *Server) extractResourceType(prompt string) string {
	if strings.Contains(prompt, "pod") {
		return "pod"
	}
	if strings.Contains(prompt, "deployment") {
		return "deployment"
	}
	if strings.Contains(prompt, "service") {
		return "service"
	}
	if strings.Contains(prompt, "configmap") {
		return "configmap"
	}
	if strings.Contains(prompt, "secret") {
		return "secret"
	}
	return "pod" // default
}

// extractResourceName extracts resource name from the prompt
func (s *Server) extractResourceName(prompt string) string {
	// This is a simple implementation - in a real scenario, you'd use NLP
	// to extract the resource name more accurately
	words := strings.Fields(prompt)
	for i, word := range words {
		if (word == "pod" || word == "deployment" || word == "service") && i+1 < len(words) {
			return words[i+1]
		}
	}
	return "" // empty name means list all
}

// handleUserChoice handles user choice requests
func (s *Server) handleUserChoice(c *gin.Context) {
	var req UserChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithFields(logrus.Fields{
		"choice":         req.Choice,
		"original_query": req.OriginalQuery,
	}).Debug("Processing user choice")

	// Store user feedback
	if err := s.memory.StoreFeedback(req.OriginalQuery, req.Choice); err != nil {
		logrus.WithError(err).Warn("Failed to store feedback")
	}

	var response string
	switch req.Choice {
	case "accept":
		response = "Great! I'll help you implement the recommended solution. Here are the step-by-step instructions..."
	case "decline":
		// Get alternative analysis from LLM
		altResponse, err := s.llmClient.GetAlternativeAnalysis(req.OriginalQuery)
		if err != nil {
			logrus.WithError(err).Error("Failed to get alternative analysis")
			response = "I understand you'd like a different approach. Let me think of alternative solutions..."
		} else {
			response = altResponse
		}
	case "more_info":
		response = "Let me provide more detailed diagnostic information..."
	}

	c.JSON(http.StatusOK, UserChoiceResponse{
		Response:  response,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"choice": req.Choice,
		},
	})
}

// handlePromptStats handles prompt statistics requests
func (s *Server) handlePromptStats(c *gin.Context) {
	categories, err := s.memory.GetPromptCategories()
	if err != nil {
		logrus.WithError(err).Error("Failed to get prompt categories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prompt statistics"})
		return
	}

	// Calculate statistics
	stats := make(map[string]interface{})
	categoryStats := make(map[string]int)
	subcategoryStats := make(map[string]map[string]int)
	totalPrompts := 0
	successfulPrompts := 0

	for _, category := range categories {
		totalPrompts += category.Frequency
		categoryStats[category.Category] += category.Frequency

		if category.Success {
			successfulPrompts += category.Frequency
		}

		if subcategoryStats[category.Category] == nil {
			subcategoryStats[category.Category] = make(map[string]int)
		}
		subcategoryStats[category.Category][category.Subcategory] += category.Frequency
	}

	stats["total_prompts"] = totalPrompts
	stats["unique_prompts"] = len(categories)
	stats["success_rate"] = float64(successfulPrompts) / float64(totalPrompts) * 100
	stats["by_category"] = categoryStats
	stats["by_subcategory"] = subcategoryStats

	c.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"timestamp": time.Now(),
	})
}

// handleUpdatePrompts handles manual prompts.md update requests
func (s *Server) handleUpdatePrompts(c *gin.Context) {
	err := s.memory.UpdatePromptsFile()
	if err != nil {
		logrus.WithError(err).Error("Failed to update prompts file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prompts file"})
		return
	}

	categories, _ := s.memory.GetPromptCategories()
	c.JSON(http.StatusOK, gin.H{
		"message":       "Prompts file updated successfully",
		"prompts_count": len(categories),
		"timestamp":     time.Now(),
	})
}

// handlePromptCategories handles prompt categories listing
func (s *Server) handlePromptCategories(c *gin.Context) {
	categories, err := s.memory.GetPromptCategories()
	if err != nil {
		logrus.WithError(err).Error("Failed to get prompt categories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get prompt categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"count":      len(categories),
		"timestamp":  time.Now(),
	})
}

// initializeMCP sets up the MCP server integration
func (s *Server) initializeMCP() error {
	// Initialize MCP server with simple configuration
	mcpConfig := &mcpserver.Config{
		Profile: s.config.MCP.Profile,
		Debug:   s.config.Debug,
	}

	s.mcpServer = mcpserver.NewServer(mcpConfig, s.config.Kubeconfig)
	if s.mcpServer == nil {
		return fmt.Errorf("failed to create MCP server")
	}

	// Add MCP routes
	mcpHandler := NewMCPHandler(s.mcpServer)
	mcpHandler.RegisterRoutes(s.engine)

	// Initialize enhanced chat handler (routes will be registered in setupRoutes)
	s.enhancedChat = NewEnhancedChatHandler(s.mcpServer, s.config)

	logrus.Info("MCP server initialized successfully")
	return nil
}
