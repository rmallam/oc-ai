package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/decision"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
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

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check
	s.engine.GET("/health", s.handleHealth)

	// API routes
	api := s.engine.Group("/api/v1")
	{
		api.POST("/chat", s.handleChat)
		api.POST("/user-choice", s.handleUserChoice)
		api.GET("/prompts/stats", s.handlePromptStats)
		api.POST("/prompts/update", s.handleUpdatePrompts)
		api.GET("/prompts/categories", s.handlePromptCategories)
	}

	// Static routes for web UI (if needed)
	s.engine.Static("/static", "./web/static")
	s.engine.LoadHTMLGlob("web/templates/*")
	s.engine.GET("/", s.handleIndex)
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

// handleChat handles chat requests
func (s *Server) handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logrus.WithField("prompt", req.Prompt).Debug("Processing chat request")

	// Store the query in memory
	if err := s.memory.StoreQuery(req.Prompt); err != nil {
		logrus.WithError(err).Warn("Failed to store query")
	}

	// Process the request through decision engine
	analysis, err := s.decisionEngine.Analyze(req.Prompt)
	if err != nil {
		logrus.WithError(err).Error("Failed to analyze prompt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
		return
	}

	response := ChatResponse{
		Response:  analysis.Response,
		Analysis:  analysis,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"model":      s.config.Model,
			"provider":   s.config.LLMProvider,
			"confidence": analysis.Confidence,
			"severity":   analysis.Severity,
		},
	}

	// Store the response
	if err := s.memory.StoreResponse(req.Prompt, analysis); err != nil {
		logrus.WithError(err).Warn("Failed to store response")
	}

	c.JSON(http.StatusOK, response)
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
