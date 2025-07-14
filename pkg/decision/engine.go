package decision

import (
	"fmt"
	"strings"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/command"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/memory"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/network"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/operator"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/types"
	"github.com/sirupsen/logrus"
)

// Engine represents the dynamic decision making engine with specialized sub-engines
type Engine struct {
	config         *config.Config
	memory         *memory.Store
	llmClient      llm.Client
	sreAssistant   *llm.SREAssistant
	operatorEngine *operator.DetectionEngine
	commandEngine  *command.GenerationEngine
	networkEngine  *network.TroubleshootingEngine
}

// NewEngine creates a new decision engine with specialized sub-engines
func NewEngine(cfg *config.Config, mem *memory.Store, llmClient llm.Client) (*Engine, error) {
	return &Engine{
		config:         cfg,
		memory:         mem,
		llmClient:      llmClient,
		sreAssistant:   llm.NewSREAssistant(llmClient),
		operatorEngine: operator.NewDetectionEngine(),
		commandEngine:  command.NewGenerationEngine(llmClient),
		networkEngine:  network.NewTroubleshootingEngine(),
	}, nil
}

// Analyze analyzes a user prompt and returns analysis using specialized engines
func (e *Engine) Analyze(prompt string) (*types.Analysis, error) {
	logrus.WithField("prompt", prompt).Debug("Starting analysis with specialized engines")

	analysis := &types.Analysis{
		Query:      prompt,
		Timestamp:  time.Now(),
		AnalysisID: generateAnalysisID(),
		Metadata:   make(map[string]interface{}),
	}

	// Store the prompt in memory
	if e.memory != nil {
		if err := e.memory.StoreQuery(prompt); err != nil {
			logrus.WithError(err).Warn("Failed to store prompt in memory")
		}
	}

	// Check for operator queries first (highest priority)
	if ok, operatorName := e.operatorEngine.IsOperatorQuery(prompt); ok {
		return e.handleOperatorQuery(analysis, operatorName)
	}

	// Check for network troubleshooting queries (second priority)
	if e.networkEngine.IsNetworkQuery(prompt) {
		return e.handleNetworkTroubleshooting(analysis)
	}

	// Use SRE Assistant for intelligent classification (third priority)
	return e.handleSREClassification(analysis)
}

// handleOperatorQuery handles operator detection queries
func (e *Engine) handleOperatorQuery(analysis *types.Analysis, operatorName string) (*types.Analysis, error) {
	logrus.Debugf("Handling operator query for: %s", operatorName)

	// Use the dedicated operator detection engine
	detectionResult := e.operatorEngine.DetectOperator(operatorName)

	// Convert detection result to analysis format
	analysis.Response = e.formatOperatorResponse(detectionResult)
	analysis.Confidence = 0.95
	analysis.Severity = "Low"
	analysis.Metadata["operator_check"] = operatorName
	analysis.Metadata["execution_type"] = "operator_detection"
	analysis.Metadata["is_installed"] = detectionResult.IsInstalled
	analysis.Metadata["detection_details"] = detectionResult.Details
	analysis.Metadata["commands_executed"] = len(detectionResult.Commands)

	return analysis, nil
}

// handleCommandExecution handles general command execution with support for multiple commands
func (e *Engine) handleCommandExecution(analysis *types.Analysis) (*types.Analysis, error) {
	logrus.Debug("Handling general command execution")

	// Use the dedicated command generation engine
	generationResult := e.commandEngine.GenerateAndExecute(analysis.Query)

	// Convert generation result to analysis format
	analysis.Response = e.formatCommandResponse(generationResult)
	analysis.Metadata["execution_type"] = "command_execution"

	// Handle multiple commands
	if len(generationResult.GeneratedCommands) > 1 {
		analysis.Metadata["commands"] = generationResult.GeneratedCommands
		analysis.Metadata["command_count"] = len(generationResult.GeneratedCommands)

		// Calculate overall success metrics
		successCount := 0
		totalDuration := 0.0
		for _, result := range generationResult.ExecutionResults {
			if result.ExitCode == 0 {
				successCount++
			}
			totalDuration += result.Duration.Seconds()
		}

		analysis.Metadata["successful_commands"] = successCount
		analysis.Metadata["total_duration"] = totalDuration

		if successCount == len(generationResult.GeneratedCommands) {
			analysis.Confidence = 0.9
			analysis.Severity = "Low"
		} else if successCount > 0 {
			analysis.Confidence = 0.6
			analysis.Severity = "Medium"
		} else {
			analysis.Confidence = 0.3
			analysis.Severity = "High"
		}
	} else {
		// Single command (backward compatibility)
		analysis.Metadata["command"] = generationResult.GeneratedCommand
		analysis.Metadata["exit_code"] = generationResult.ExecutionResult.ExitCode
		analysis.Metadata["duration"] = generationResult.ExecutionResult.Duration.Seconds()

		if generationResult.ExecutionResult.ExitCode == 0 {
			analysis.Confidence = 0.9
			analysis.Severity = "Low"
		} else {
			analysis.Confidence = 0.3
			analysis.Severity = "Medium"
			if generationResult.Fallback != nil {
				analysis.Confidence = 0.7
				analysis.Metadata["fallback_used"] = true
				analysis.Metadata["fallback_command"] = generationResult.Fallback.Command
			}
		}
	}

	return analysis, nil
}

// handleNetworkTroubleshooting handles network troubleshooting queries
func (e *Engine) handleNetworkTroubleshooting(analysis *types.Analysis) (*types.Analysis, error) {
	logrus.Debug("Handling network troubleshooting query")

	// Use the dedicated network troubleshooting engine
	troubleshootingResult := e.networkEngine.TroubleshootNetwork(analysis.Query)

	// Convert troubleshooting result to analysis format
	analysis.Response = e.formatNetworkResponse(troubleshootingResult)
	analysis.Metadata["workflow_type"] = troubleshootingResult.WorkflowType
	analysis.Metadata["execution_type"] = "network_troubleshooting"
	analysis.Metadata["pod_info"] = troubleshootingResult.PodInfo
	analysis.Metadata["steps_executed"] = len(troubleshootingResult.Steps)
	analysis.Metadata["commands_executed"] = len(troubleshootingResult.Commands)

	if troubleshootingResult.Success {
		analysis.Confidence = 0.9
		analysis.Severity = "Low"
	} else {
		analysis.Confidence = 0.5
		analysis.Severity = "Medium"
	}

	return analysis, nil
}

// handleSREClassification uses the SRE Assistant to intelligently classify and handle requests
func (e *Engine) handleSREClassification(analysis *types.Analysis) (*types.Analysis, error) {
	logrus.Debug("Handling request through SRE classification")

	// First, classify the request type using SRE Assistant
	requestType := e.sreAssistant.ClassifyRequest(analysis.Query)
	logrus.Debugf("SRE classification result: %s", requestType)

	analysis.Metadata["sre_classification"] = requestType

	// For requests that need command execution, route to command engine
	if requestType == "resource-creation" || requestType == "configuration" {
		logrus.Debug("Routing classified request to command execution")
		return e.handleCommandExecution(analysis)
	}

	// For other types (troubleshooting, security, incident, performance), provide analysis response
	response, err := e.sreAssistant.AnalyzeIssue(analysis.Query)
	if err != nil {
		// Fallback to general command execution if SRE analysis fails
		logrus.WithError(err).Warn("SRE analysis failed, falling back to command execution")
		return e.handleCommandExecution(analysis)
	}

	// Set response and metadata
	analysis.Response = response
	analysis.Metadata["execution_type"] = "sre_analysis"
	analysis.Confidence = 0.8
	analysis.Severity = "Low"

	return analysis, nil
}

// formatOperatorResponse formats the operator detection result for display
func (e *Engine) formatOperatorResponse(result *operator.DetectionResult) string {
	var lines []string

	// Add main summary
	lines = append(lines, result.Summary)
	lines = append(lines, "")

	// Add detailed check results
	lines = append(lines, "🔍 Detailed Check Results:")
	lines = append(lines, strings.Repeat("=", 50))

	for i, check := range result.Details {
		status := "❌ FAILED"
		if check.Found {
			status = "✅ PASSED"
		}

		lines = append(lines, fmt.Sprintf("%d. %s - %s", i+1, status, check.Description))
		lines = append(lines, fmt.Sprintf("   📋 %s", check.Details))

		// Add command info if available
		if i < len(result.Commands) {
			cmd := result.Commands[i]
			if cmd.Error != "" {
				lines = append(lines, fmt.Sprintf("   🔧 Command: %s (failed: %s)", cmd.Command, cmd.Error))
			} else {
				lines = append(lines, fmt.Sprintf("   🔧 Command: %s (success)", cmd.Command))
			}
		}
		lines = append(lines, "")
	}

	// Add overall assessment
	if result.IsInstalled {
		lines = append(lines, "🎯 CONCLUSION: Operator is INSTALLED and functional")
	} else {
		lines = append(lines, "❌ CONCLUSION: Operator is NOT INSTALLED")
		lines = append(lines, "💡 TIP: To install, visit OperatorHub in OpenShift Console or use 'oc' commands")
	}

	return strings.Join(lines, "\n")
}

// formatCommandResponse formats the command execution result for display (supports multiple commands)
func (e *Engine) formatCommandResponse(result *command.GenerationResult) string {
	var lines []string

	// Handle multiple commands
	if len(result.GeneratedCommands) > 1 {
		lines = append(lines, fmt.Sprintf("🔧 Executed %d Commands:", len(result.GeneratedCommands)))
		lines = append(lines, "")

		successCount := 0
		for i, cmd := range result.GeneratedCommands {
			execResult := result.ExecutionResults[i]

			lines = append(lines, fmt.Sprintf("Command %d: %s", i+1, cmd))

			if execResult.ExitCode == 0 {
				lines = append(lines, "  ✅ Status: SUCCESS")
				successCount++
				if execResult.Output != "" {
					lines = append(lines, fmt.Sprintf("  📋 Output: %s", execResult.Output))
				}
			} else {
				lines = append(lines, fmt.Sprintf("  ❌ Status: FAILED (exit code: %d)", execResult.ExitCode))
				if execResult.Error != "" {
					lines = append(lines, fmt.Sprintf("  🚨 Error: %s", execResult.Error))
				}
				if execResult.Output != "" {
					lines = append(lines, fmt.Sprintf("  📋 Output: %s", execResult.Output))
				}
			}
			lines = append(lines, "")
		}

		// Summary
		if successCount == len(result.GeneratedCommands) {
			lines = append(lines, "✅ All commands executed successfully!")
		} else {
			lines = append(lines, fmt.Sprintf("⚠️  %d of %d commands succeeded", successCount, len(result.GeneratedCommands)))
		}
	} else {
		// Single command (backward compatibility)
		lines = append(lines, fmt.Sprintf("🔧 Executed Command: %s", result.GeneratedCommand))

		if result.ExecutionResult.ExitCode == 0 {
			lines = append(lines, "✅ Status: SUCCESS")
			lines = append(lines, "")
			lines = append(lines, "📋 Output:")
			lines = append(lines, strings.Repeat("-", 40))
			if result.ExecutionResult.Output != "" {
				lines = append(lines, result.ExecutionResult.Output)
			} else {
				lines = append(lines, "(No output)")
			}
		} else {
			lines = append(lines, fmt.Sprintf("❌ Status: FAILED (exit code: %d)", result.ExecutionResult.ExitCode))
			if result.ExecutionResult.Error != "" {
				lines = append(lines, fmt.Sprintf("🚨 Error: %s", result.ExecutionResult.Error))
			}

			// Show fallback if used
			if result.Fallback != nil {
				lines = append(lines, "")
				lines = append(lines, fmt.Sprintf("🔄 Fallback Command: %s", result.Fallback.Command))
				if result.Fallback.ExitCode == 0 {
					lines = append(lines, "✅ Fallback Status: SUCCESS")
					lines = append(lines, "")
					lines = append(lines, "📋 Fallback Output:")
					lines = append(lines, strings.Repeat("-", 40))
					if result.Fallback.Output != "" {
						lines = append(lines, result.Fallback.Output)
					} else {
						lines = append(lines, "(No output)")
					}
				} else {
					lines = append(lines, fmt.Sprintf("❌ Fallback Status: FAILED (exit code: %d)", result.Fallback.ExitCode))
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// formatNetworkResponse formats the network troubleshooting result for display
func (e *Engine) formatNetworkResponse(result *network.TroubleshootingResult) string {
	return result.Summary
}

// generateAnalysisID generates a unique analysis ID
func generateAnalysisID() string {
	return fmt.Sprintf("analysis-%d", time.Now().UnixNano())
}
