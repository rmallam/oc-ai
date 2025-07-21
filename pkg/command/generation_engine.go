package command

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/executor"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
	"github.com/sirupsen/logrus"
)

// GenerationEngine handles command generation and execution
type GenerationEngine struct {
	executor  *executor.CommandExecutor
	llmClient llm.Client
}

// GenerationResult represents the result of command generation and execution
type GenerationResult struct {
	Query             string                      `json:"query"`
	GeneratedCommand  string                      `json:"generated_command"`
	GeneratedCommands []string                    `json:"generated_commands,omitempty"`
	ExecutionResult   *executor.ExecutionResult   `json:"execution_result"`
	ExecutionResults  []*executor.ExecutionResult `json:"execution_results,omitempty"`
	Fallback          *executor.ExecutionResult   `json:"fallback,omitempty"`
	Summary           string                      `json:"summary"`
}

// NewGenerationEngine creates a new command generation engine
func NewGenerationEngine(llmClient llm.Client) *GenerationEngine {
	return &GenerationEngine{
		executor:  executor.NewCommandExecutor(),
		llmClient: llmClient,
	}
}

// NewGenerationEngineWithKubeconfig creates a new command generation engine with kubeconfig
func NewGenerationEngineWithKubeconfig(llmClient llm.Client, kubeconfigPath string) *GenerationEngine {
	return &GenerationEngine{
		executor:  executor.NewCommandExecutorWithKubeconfig(kubeconfigPath),
		llmClient: llmClient,
	}
}

// GenerateAndExecute generates commands from query and executes them
func (ge *GenerationEngine) GenerateAndExecute(query string) *GenerationResult {
	logrus.Debugf("Generating commands for query: %s", query)

	result := &GenerationResult{
		Query: query,
	}

	// Generate commands using LLM
	commandResponse, err := ge.llmClient.GenerateResponse(ge.buildPrompt(query))
	if err != nil {
		result.Summary = fmt.Sprintf("Failed to generate commands: %v", err)
		return result
	}

	// Extract all commands from LLM response
	commands := ge.extractCommands(commandResponse)
	if len(commands) == 0 {
		result.Summary = "No valid commands found in LLM response"
		return result
	}

	result.GeneratedCommands = commands
	result.GeneratedCommand = commands[0] // For backward compatibility

	// Execute all commands sequentially
	var allResults []*executor.ExecutionResult
	var successCount int
	var failedCommand string

	for i, command := range commands {
		logrus.Debugf("Executing command %d/%d: %s", i+1, len(commands), command)
		execResult := ge.executor.Execute(command)
		allResults = append(allResults, execResult)

		if execResult.ExitCode == 0 {
			successCount++
		} else {
			logrus.Warnf("Command %d failed: %s (exit code: %d)", i+1, execResult.Error, execResult.ExitCode)
			if failedCommand == "" {
				failedCommand = command
			}
			// Continue executing remaining commands even if one fails
		}
	}

	result.ExecutionResults = allResults
	result.ExecutionResult = allResults[0] // For backward compatibility

	// Generate summary
	if successCount == len(commands) {
		result.Summary = fmt.Sprintf("All %d commands executed successfully", len(commands))
	} else if successCount > 0 {
		result.Summary = fmt.Sprintf("%d of %d commands succeeded. Failed command: %s", successCount, len(commands), failedCommand)
	} else {
		result.Summary = fmt.Sprintf("All %d commands failed. First failure: %s", len(commands), allResults[0].Error)

		// Try fallback for the first command only
		fallbackCommand := ge.getFallbackCommand(query)
		if fallbackCommand != "" && fallbackCommand != commands[0] {
			logrus.Debugf("Trying fallback command: %s", fallbackCommand)
			fallbackResult := ge.executor.Execute(fallbackCommand)
			if fallbackResult.ExitCode == 0 {
				result.Fallback = fallbackResult
				result.Summary = fmt.Sprintf("All commands failed, but fallback succeeded: %s", fallbackResult.Output)
			}
		}
	}

	return result
}

// buildPrompt builds the system prompt for command generation
func (ge *GenerationEngine) buildPrompt(query string) string {
	systemPrompt := `You are an expert OpenShift/Kubernetes SRE assistant. 
Given a user request, provide the necessary commands to accomplish the task.
Use basic, reliable commands without complex go-templates or advanced formatting.

For multiple steps, you can provide multiple commands separated by newlines.
Format your response as: COMMAND: <command1>
COMMAND: <command2>
COMMAND: <command3>

IMPORTANT: For pod status queries, use proper filtering:
- For crashing/failing pods: kubectl get pods --all-namespaces | grep -E "(CrashLoopBackOff|ImagePullBackOff|Error|Evicted|OOMKilled)"
- For failed pods: kubectl get pods --all-namespaces --field-selector=status.phase=Failed
- For pending pods: kubectl get pods --all-namespaces --field-selector=status.phase=Pending
- For all pods: kubectl get pods --all-namespaces

For resource creation:
- You can provide multiple commands to create multiple resources
- For helm install: helm install <release-name> <repo>/<chart> -n <namespace> --create-namespace [flags]
- For kubectl create: kubectl create <resource> <name> -n <namespace> [flags]
- For kubectl apply: kubectl apply -f <file> -n <namespace>

Examples:
- For 'create namespace and service account': 
  COMMAND: kubectl create namespace testing
  COMMAND: kubectl create serviceaccount test-sa -n testing
- For 'list pods in all namespaces': COMMAND: kubectl get pods --all-namespaces
- For 'show nodes': COMMAND: kubectl get nodes
- For 'create namespace with RBAC': 
  COMMAND: kubectl create namespace myapp
  COMMAND: kubectl create serviceaccount myapp-sa -n myapp
  COMMAND: kubectl create rolebinding myapp-admin --clusterrole=admin --serviceaccount=myapp:myapp-sa -n myapp

Provide all necessary commands to complete the user's request.`

	return fmt.Sprintf("%s\n\nUser request: %s", systemPrompt, query)
}

// extractCommand extracts the first command from LLM response (for backward compatibility)
func (ge *GenerationEngine) extractCommand(response string) string {
	commands := ge.extractCommands(response)
	if len(commands) > 0 {
		return commands[0]
	}
	return ""
}

// extractCommands extracts all commands from LLM response
func (ge *GenerationEngine) extractCommands(response string) []string {
	lines := strings.Split(response, "\n")
	var commands []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "COMMAND:") {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "COMMAND:"))
			if cmd != "" {
				commands = append(commands, cmd)
			}
		} else if strings.HasPrefix(line, "```") {
			// Skip markdown code block markers
			continue
		} else if strings.HasPrefix(line, "kubectl ") || strings.HasPrefix(line, "oc ") || strings.HasPrefix(line, "helm ") {
			// Direct command detection
			commands = append(commands, line)
		}
	}

	return commands
}

// getFallbackCommand provides fallback commands for common queries
func (ge *GenerationEngine) getFallbackCommand(query string) string {
	lowerQuery := strings.ToLower(query)

	// Common fallback patterns
	fallbacks := map[string]string{
		"pods":        "kubectl get pods --all-namespaces",
		"namespaces":  "kubectl get namespaces",
		"nodes":       "kubectl get nodes",
		"services":    "kubectl get services --all-namespaces",
		"deployments": "kubectl get deployments --all-namespaces",
		"crashing":    "kubectl get pods --all-namespaces | grep -E '(CrashLoopBackOff|ImagePullBackOff|Error|Evicted)'",
		"failing":     "kubectl get pods --all-namespaces | grep -E '(CrashLoopBackOff|ImagePullBackOff|Error|Evicted)'",
		"failed":      "kubectl get pods --all-namespaces --field-selector=status.phase=Failed",
		"pending":     "kubectl get pods --all-namespaces --field-selector=status.phase=Pending",
		"helm list":   "helm list --all-namespaces",
	}

	for keyword, command := range fallbacks {
		if strings.Contains(lowerQuery, keyword) {
			return command
		}
	}

	// Check for namespace-specific queries
	namespacePattern := regexp.MustCompile(`in\s+([a-zA-Z0-9\-]+)\s+namespace`)
	if matches := namespacePattern.FindStringSubmatch(lowerQuery); len(matches) > 1 {
		namespace := matches[1]
		if strings.Contains(lowerQuery, "pods") {
			return fmt.Sprintf("kubectl get pods -n %s", namespace)
		}
	}

	return ""
}
