package decision

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/memory"
	"github.com/sirupsen/logrus"
)

// Engine represents the dynamic decision making engine
type Engine struct {
	config    *config.Config
	memory    *memory.Store
	llmClient llm.Client
}

// Analysis represents the result of analysis
type Analysis struct {
	Query       string                 `json:"query"`
	Response    string                 `json:"response"`
	Confidence  float64                `json:"confidence"`
	Severity    string                 `json:"severity"`
	RootCauses  []RootCause            `json:"root_causes"`
	Actions     []RecommendedAction    `json:"recommended_actions"`
	Evidence    []Evidence             `json:"evidence"`
	Timestamp   time.Time              `json:"timestamp"`
	AnalysisID  string                 `json:"analysis_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RootCause represents an identified root cause
type RootCause struct {
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Evidence    string  `json:"evidence"`
}

// RecommendedAction represents a recommended action
type RecommendedAction struct {
	Description string `json:"description"`
	Priority    string `json:"priority"` // High, Medium, Low
	Command     string `json:"command,omitempty"`
	Risk        string `json:"risk,omitempty"`
}

// Evidence represents evidence collected during analysis
type Evidence struct {
	Type        string `json:"type"`        // logs, events, status, etc.
	Source      string `json:"source"`      // pod name, node name, etc.
	Content     string `json:"content"`     // actual evidence content
	Timestamp   time.Time `json:"timestamp"`
}

// NewEngine creates a new decision engine
func NewEngine(cfg *config.Config, mem *memory.Store, llmClient llm.Client) (*Engine, error) {
	return &Engine{
		config:    cfg,
		memory:    mem,
		llmClient: llmClient,
	}, nil
}

// Analyze analyzes a user prompt and returns analysis
func (e *Engine) Analyze(prompt string) (*Analysis, error) {
	logrus.WithField("prompt", prompt).Debug("Starting analysis")

	analysis := &Analysis{
		Query:      prompt,
		Timestamp:  time.Now(),
		AnalysisID: generateAnalysisID(),
		Metadata:   make(map[string]interface{}),
	}

	// Check if this is a diagnostic query
	if e.isDiagnosticQuery(prompt) {
		return e.performDiagnosticAnalysis(analysis)
	}

	// Handle regular queries
	return e.performRegularAnalysis(analysis)
}

// isDiagnosticQuery checks if the prompt contains diagnostic keywords
func (e *Engine) isDiagnosticQuery(prompt string) bool {
	diagnosticKeywords := []string{
		"crashloop", "crash", "failing", "not working", "broken", "error",
		"troubleshoot", "debug", "diagnose", "fix", "solve", "why",
		"problem", "issue", "status", "check", "logs",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range diagnosticKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			return true
		}
	}
	return false
}

// performDiagnosticAnalysis performs diagnostic analysis
func (e *Engine) performDiagnosticAnalysis(analysis *Analysis) (*Analysis, error) {
	logrus.Debug("Performing diagnostic analysis")

	// Extract resource information from prompt
	resourceInfo := e.extractResourceInfo(analysis.Query)
	analysis.Metadata["resource_info"] = resourceInfo

	// Collect evidence
	evidence, err := e.collectEvidence(resourceInfo)
	if err != nil {
		logrus.WithError(err).Warn("Failed to collect some evidence")
	}
	analysis.Evidence = evidence

	// Analyze root causes
	rootCauses := e.analyzeRootCauses(evidence)
	analysis.RootCauses = rootCauses

	// Generate recommendations
	actions := e.generateRecommendations(rootCauses, evidence)
	analysis.Actions = actions

	// Calculate confidence and severity
	analysis.Confidence = e.calculateConfidence(rootCauses, evidence)
	analysis.Severity = e.calculateSeverity(rootCauses, evidence)

	// Generate response
	analysis.Response = e.formatDiagnosticResponse(analysis)

	return analysis, nil
}

// performRegularAnalysis performs regular non-diagnostic analysis
func (e *Engine) performRegularAnalysis(analysis *Analysis) (*Analysis, error) {
	logrus.Debug("Performing regular analysis")

	// Use LLM for regular queries
	response, err := e.llmClient.GenerateResponse(analysis.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LLM response: %w", err)
	}

	analysis.Response = response
	analysis.Confidence = 0.8 // Default confidence for regular queries
	analysis.Severity = "Low"

	return analysis, nil
}

// extractResourceInfo extracts resource information from prompt
func (e *Engine) extractResourceInfo(prompt string) map[string]string {
	info := make(map[string]string)

	// Extract pod name
	podRegex := regexp.MustCompile(`pod\s+([a-zA-Z0-9\-]+)`)
	if matches := podRegex.FindStringSubmatch(prompt); len(matches) > 1 {
		info["pod_name"] = matches[1]
	}

	// Extract namespace
	nsRegex := regexp.MustCompile(`namespace\s+([a-zA-Z0-9\-]+)`)
	if matches := nsRegex.FindStringSubmatch(prompt); len(matches) > 1 {
		info["namespace"] = matches[1]
	}

	// Extract deployment name
	deployRegex := regexp.MustCompile(`deployment\s+([a-zA-Z0-9\-]+)`)
	if matches := deployRegex.FindStringSubmatch(prompt); len(matches) > 1 {
		info["deployment"] = matches[1]
	}

	return info
}

// collectEvidence collects evidence for analysis
func (e *Engine) collectEvidence(resourceInfo map[string]string) ([]Evidence, error) {
	var evidence []Evidence

	// This would normally collect actual evidence from Kubernetes
	// For now, we'll simulate evidence collection
	if podName, exists := resourceInfo["pod_name"]; exists {
		evidence = append(evidence, Evidence{
			Type:      "pod_status",
			Source:    podName,
			Content:   "Pod is in CrashLoopBackOff state",
			Timestamp: time.Now(),
		})

		evidence = append(evidence, Evidence{
			Type:      "logs",
			Source:    podName,
			Content:   "Error: No module named 'uvicorn'",
			Timestamp: time.Now(),
		})

		evidence = append(evidence, Evidence{
			Type:      "events",
			Source:    podName,
			Content:   "Container image 'my-app:latest' is present on machine",
			Timestamp: time.Now(),
		})
	}

	return evidence, nil
}

// analyzeRootCauses analyzes evidence to identify root causes
func (e *Engine) analyzeRootCauses(evidence []Evidence) []RootCause {
	var rootCauses []RootCause

	// Analyze evidence patterns
	for _, ev := range evidence {
		if strings.Contains(ev.Content, "No module named") {
			rootCauses = append(rootCauses, RootCause{
				Description: "Missing Python module dependency",
				Confidence:  0.9,
				Evidence:    ev.Content,
			})
		}

		if strings.Contains(ev.Content, "CrashLoopBackOff") {
			rootCauses = append(rootCauses, RootCause{
				Description: "Application failing to start properly",
				Confidence:  0.8,
				Evidence:    ev.Content,
			})
		}
	}

	return rootCauses
}

// generateRecommendations generates recommended actions
func (e *Engine) generateRecommendations(rootCauses []RootCause, evidence []Evidence) []RecommendedAction {
	var actions []RecommendedAction

	for _, cause := range rootCauses {
		if strings.Contains(cause.Description, "Missing Python module") {
			actions = append(actions, RecommendedAction{
				Description: "Install missing Python dependencies",
				Priority:    "High",
				Command:     "pip install <missing_module>",
				Risk:        "Low",
			})
		}

		if strings.Contains(cause.Description, "failing to start") {
			actions = append(actions, RecommendedAction{
				Description: "Check application configuration and logs",
				Priority:    "High",
				Command:     "oc logs <pod_name>",
				Risk:        "Low",
			})
		}
	}

	return actions
}

// calculateConfidence calculates overall confidence score
func (e *Engine) calculateConfidence(rootCauses []RootCause, evidence []Evidence) float64 {
	if len(rootCauses) == 0 {
		return 0.3
	}

	var totalConfidence float64
	for _, cause := range rootCauses {
		totalConfidence += cause.Confidence
	}

	confidence := totalConfidence / float64(len(rootCauses))
	
	// Boost confidence based on evidence quality
	if len(evidence) >= 3 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateSeverity calculates severity level
func (e *Engine) calculateSeverity(rootCauses []RootCause, evidence []Evidence) string {
	if len(rootCauses) == 0 {
		return "Low"
	}

	// Check for critical issues
	for _, ev := range evidence {
		if strings.Contains(ev.Content, "CrashLoopBackOff") ||
			strings.Contains(ev.Content, "ImagePullBackOff") {
			return "High"
		}
	}

	// Check for medium issues
	if len(rootCauses) >= 2 {
		return "Medium"
	}

	return "Medium"
}

// formatDiagnosticResponse formats the diagnostic response
func (e *Engine) formatDiagnosticResponse(analysis *Analysis) string {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("🔍 **Diagnostic Analysis**\n\n"))
	response.WriteString(fmt.Sprintf("🔴 **Severity:** %s\n", analysis.Severity))
	response.WriteString(fmt.Sprintf("📊 **Confidence:** %.0f%%\n\n", analysis.Confidence*100))

	if len(analysis.RootCauses) > 0 {
		response.WriteString("## Root Causes Identified:\n")
		for i, cause := range analysis.RootCauses {
			response.WriteString(fmt.Sprintf("%d. %s (Confidence: %.0f%%)\n", 
				i+1, cause.Description, cause.Confidence*100))
		}
		response.WriteString("\n")
	}

	if len(analysis.Evidence) > 0 {
		response.WriteString("## Evidence Found:\n")
		for _, ev := range analysis.Evidence {
			response.WriteString(fmt.Sprintf("• **%s:** %s\n", ev.Type, ev.Content))
		}
		response.WriteString("\n")
	}

	if len(analysis.Actions) > 0 {
		response.WriteString("## Recommended Solutions:\n")
		for i, action := range analysis.Actions {
			response.WriteString(fmt.Sprintf("### %s Priority Action %d:\n", 
				strings.ToUpper(action.Priority), i+1))
			response.WriteString(fmt.Sprintf("%s\n", action.Description))
			if action.Command != "" {
				response.WriteString(fmt.Sprintf("```\n%s\n```\n", action.Command))
			}
		}
		response.WriteString("\n")
	}

	response.WriteString("## What would you like to do?\n")
	response.WriteString("- **Accept Analysis** ✅ → Get implementation guidance\n")
	response.WriteString("- **Get Alternative Analysis** 🤖 → AI-powered different perspective\n")
	response.WriteString("- **Get More Details** 📊 → Extended diagnostic information\n")

	return response.String()
}

// generateAnalysisID generates a unique analysis ID
func generateAnalysisID() string {
	return fmt.Sprintf("analysis_%d", time.Now().UnixNano())
}
