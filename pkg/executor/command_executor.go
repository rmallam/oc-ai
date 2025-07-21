package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/kubeconfig"
	"github.com/sirupsen/logrus"
)

// CommandExecutor handles safe execution of shell commands
type CommandExecutor struct {
	allowedCommands []string
	timeout         time.Duration
	kubeconfigPath  string
}

// ExecutionResult represents the result of command execution
type ExecutionResult struct {
	Command   string        `json:"command"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	ExitCode  int           `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{
		allowedCommands: []string{
			"kubectl", "oc", "helm", "docker", "podman",
			"curl", "ping", "nslookup", "dig", "telnet",
			"cat", "grep", "awk", "sed", "head", "tail",
		},
		timeout: 10 * time.Second, // Reduced timeout for faster feedback
	}
}

// NewCommandExecutorWithKubeconfig creates a new command executor with kubeconfig
func NewCommandExecutorWithKubeconfig(kubeconfigPath string) *CommandExecutor {
	return &CommandExecutor{
		allowedCommands: []string{
			"kubectl", "oc", "helm", "docker", "podman",
			"curl", "ping", "nslookup", "dig", "telnet",
			"cat", "grep", "awk", "sed", "head", "tail",
		},
		timeout:        10 * time.Second,
		kubeconfigPath: kubeconfigPath,
	}
}

// SetTimeout sets the command execution timeout
func (ce *CommandExecutor) SetTimeout(timeout time.Duration) {
	ce.timeout = timeout
}

// IsCommandSafe validates if a command is safe to execute
func (ce *CommandExecutor) IsCommandSafe(command string) bool {
	// Trim and split command
	parts := strings.Fields(strings.TrimSpace(command))
	if len(parts) == 0 {
		return false
	}

	baseCommand := parts[0]

	// Check if base command is in allowed list
	for _, allowed := range ce.allowedCommands {
		if baseCommand == allowed {
			return ce.validateCommandArgs(command)
		}
	}

	return false
}

// validateCommandArgs performs additional validation on command arguments
func (ce *CommandExecutor) validateCommandArgs(command string) bool {
	// Block dangerous patterns
	dangerousPatterns := []string{
		"rm -rf", "dd if=", "mkfs", "fdisk", "parted",
		"shutdown", "reboot", "halt", "poweroff",
		"passwd", "sudo", "su -", "chmod 777",
		">/dev/", "curl.*|.*sh", "wget.*|.*sh",
	}

	lowerCmd := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, pattern) {
			logrus.Warnf("Blocked dangerous command pattern: %s", pattern)
			return false
		}
	}

	return true
}

// Execute runs a command and returns the result
func (ce *CommandExecutor) Execute(command string) *ExecutionResult {
	startTime := time.Now()
	result := &ExecutionResult{
		Command:   command,
		Timestamp: startTime,
	}

	// Validate command safety
	if !ce.IsCommandSafe(command) {
		result.Error = "Command rejected for security reasons"
		result.ExitCode = 1
		result.Duration = time.Since(startTime)
		return result
	}

	// Prepare kubectl/oc commands with proper authentication
	command = ce.prepareKubernetesCommand(command)

	logrus.Debugf("Executing command: %s", command)

	// Check if this is a shell command (contains pipes, redirects, etc.)
	if ce.isShellCommand(command) {
		return ce.executeShellCommand(command, startTime)
	}

	// Execute as regular command
	return ce.executeRegularCommand(command, startTime)
}

// prepareKubernetesCommand sets up kubectl/oc commands with proper authentication
func (ce *CommandExecutor) prepareKubernetesCommand(command string) string {
	parts := strings.Fields(strings.TrimSpace(command))
	if len(parts) == 0 {
		return command
	}

	// Check if this is a kubectl or oc command
	firstCommand := parts[0]
	if firstCommand != "kubectl" && firstCommand != "oc" {
		return command
	}

	// Check if we're running in cluster (service account available)
	if kubeconfig.IsInCluster() {
		// When running in cluster, kubectl/oc will automatically use service account token
		// No additional flags needed
		return command
	}

	// When running outside cluster, ensure kubeconfig is specified if we have one
	if ce.kubeconfigPath != "" {
		// Check if --kubeconfig is already specified
		kubeconfigPresent := false
		for _, part := range parts {
			if strings.HasPrefix(part, "--kubeconfig") {
				kubeconfigPresent = true
				break
			}
		}

		// Add kubeconfig flag if not present
		if !kubeconfigPresent {
			// Insert after the command name
			newParts := []string{parts[0], "--kubeconfig", ce.kubeconfigPath}
			if len(parts) > 1 {
				newParts = append(newParts, parts[1:]...)
			}
			return strings.Join(newParts, " ")
		}
	}

	return command
}

// isShellCommand detects if a command needs shell execution
func (ce *CommandExecutor) isShellCommand(command string) bool {
	shellOperators := []string{"|", "&&", "||", ">", ">>", "<", ";"}
	shellCommands := []string{"grep", "awk", "sed", "head", "tail", "sort", "uniq", "wc"}

	for _, op := range shellOperators {
		if strings.Contains(command, op) {
			return true
		}
	}

	parts := strings.Fields(command)
	if len(parts) > 0 {
		for _, shellCmd := range shellCommands {
			if parts[0] == shellCmd {
				return true
			}
		}
	}

	return false
}

// executeShellCommand executes command through shell with timeout
func (ce *CommandExecutor) executeShellCommand(command string, startTime time.Time) *ExecutionResult {
	result := &ExecutionResult{
		Command:   command,
		Timestamp: startTime,
	}

	// Use bash for shell commands with timeout
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Env = ce.prepareEnvironment()

	// Set up timeout
	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case err := <-done:
		result.Duration = time.Since(startTime)
		result.Output = strings.TrimSpace(string(output))

		if err != nil {
			result.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitError.ExitCode()
			} else {
				result.ExitCode = 1
			}
		}
	case <-time.After(ce.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		result.Duration = ce.timeout
		result.Error = fmt.Sprintf("command timed out after %v", ce.timeout)
		result.ExitCode = 124 // Standard timeout exit code
	}

	return result
}

// executeRegularCommand executes command directly with timeout
func (ce *CommandExecutor) executeRegularCommand(command string, startTime time.Time) *ExecutionResult {
	result := &ExecutionResult{
		Command:   command,
		Timestamp: startTime,
	}

	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = ce.prepareEnvironment()

	// Set up timeout
	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case err := <-done:
		result.Duration = time.Since(startTime)
		result.Output = strings.TrimSpace(string(output))

		if err != nil {
			result.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitError.ExitCode()
			} else {
				result.ExitCode = 1
			}
		}
	case <-time.After(ce.timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		result.Duration = ce.timeout
		result.Error = fmt.Sprintf("command timed out after %v", ce.timeout)
		result.ExitCode = 124 // Standard timeout exit code
	}

	return result
}

// prepareEnvironment sets up environment variables for command execution
func (ce *CommandExecutor) prepareEnvironment() []string {
	env := os.Environ()

	// Ensure PATH includes common binary locations
	pathFound := false
	for i, v := range env {
		if strings.HasPrefix(v, "PATH=") {
			env[i] = v + ":/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin"
			pathFound = true
			break
		}
	}

	if !pathFound {
		env = append(env, "PATH=/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin")
	}

	return env
}

// ExecuteMultiple executes multiple commands and returns results
func (ce *CommandExecutor) ExecuteMultiple(commands []string) []*ExecutionResult {
	results := make([]*ExecutionResult, len(commands))

	for i, cmd := range commands {
		results[i] = ce.Execute(cmd)
		logrus.Debugf("Command %d completed: %s (exit: %d)", i+1, cmd, results[i].ExitCode)
	}

	return results
}

// GetSupportedCommands returns list of supported command prefixes
func (ce *CommandExecutor) GetSupportedCommands() []string {
	return append([]string{}, ce.allowedCommands...)
}
