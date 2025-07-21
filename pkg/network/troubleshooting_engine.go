package network

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/executor"
	"github.com/sirupsen/logrus"
)

// TroubleshootingEngine handles network troubleshooting and debugging workflows
type TroubleshootingEngine struct {
	executor *executor.CommandExecutor
}

// TroubleshootingResult represents the result of network troubleshooting
type TroubleshootingResult struct {
	Query        string                      `json:"query"`
	WorkflowType string                      `json:"workflow_type"`
	PodInfo      PodInfo                     `json:"pod_info"`
	Steps        []WorkflowStep              `json:"steps"`
	Commands     []*executor.ExecutionResult `json:"commands"`
	Summary      string                      `json:"summary"`
	Success      bool                        `json:"success"`
}

// PodInfo contains extracted pod information
type PodInfo struct {
	PodName   string `json:"pod_name"`
	Namespace string `json:"namespace"`
	NodeName  string `json:"node_name,omitempty"`
	Found     bool   `json:"found"`
}

// WorkflowStep represents a step in the troubleshooting workflow
type WorkflowStep struct {
	StepNumber  int    `json:"step_number"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Purpose     string `json:"purpose"`
}

// DiagnosticResult represents the result of diagnostic analysis
type DiagnosticResult struct {
	PodStatus      string   `json:"pod_status"`
	Phase          string   `json:"phase"`
	Issues         []Issue  `json:"issues"`
	RootCause      string   `json:"root_cause"`
	Recommendation string   `json:"recommendation"`
	LogsNeeded     bool     `json:"logs_needed"`
	NextSteps      []string `json:"next_steps"`
}

// Issue represents a specific problem found during analysis
type Issue struct {
	Type       string `json:"type"`   // "error", "warning", "info"
	Source     string `json:"source"` // "describe", "logs", "events"
	Message    string `json:"message"`
	Severity   string `json:"severity"` // "critical", "high", "medium", "low"
	Category   string `json:"category"` // "image", "network", "storage", "compute", "config"
	Actionable bool   `json:"actionable"`
	Suggestion string `json:"suggestion"`
}

// NewTroubleshootingEngine creates a new network troubleshooting engine
func NewTroubleshootingEngine() *TroubleshootingEngine {
	return &TroubleshootingEngine{
		executor: executor.NewCommandExecutor(),
	}
}

// NewTroubleshootingEngineWithKubeconfig creates a new network troubleshooting engine with kubeconfig
func NewTroubleshootingEngineWithKubeconfig(kubeconfigPath string) *TroubleshootingEngine {
	return &TroubleshootingEngine{
		executor: executor.NewCommandExecutorWithKubeconfig(kubeconfigPath),
	}
}

// IsNetworkQuery detects if a query is related to network troubleshooting or pod diagnostics
func (nt *TroubleshootingEngine) IsNetworkQuery(query string) bool {
	// More specific network keywords to avoid false positives
	networkKeywords := []string{
		"tcpdump", "packet capture", "network capture", "wireshark",
		"ping from pod", "connectivity test", "network test",
		"traceroute", "nslookup", "dig", "curl from pod",
		"network debug", "network troubleshoot", "capture packets",
		"network analysis", "packet analysis", "traffic capture",
		"dns resolution", "dns test", "http test", "https test",
		"netstat", "ss", "lsof", "netcat", "nc", "telnet",
		"network connections", "socket connections", "network routes",
	}

	// Specific pod troubleshooting keywords - removed overly broad terms
	podTroubleshootingKeywords := []string{
		"pod failing", "pod not working", "pod crash", "pod error",
		"pod stuck", "pod pending", "pod evicted", "pod terminating",
		"crashloopbackoff", "imagepullbackoff", "oomkilled",
		"pod status", "pod health", "pod issues", "pod problems",
		"why is pod", "what's wrong with", "check pod", "examine pod",
		"container creating", "containercreating", "stuck in container",
		"creating container", "pod initializing", "init container",
		"pulling image", "waiting for", "pod not starting",
		"troubleshoot pod", "debug pod", "diagnose pod", // More specific versions
	}

	lowerQuery := strings.ToLower(query)

	// Helper function to check for more precise keyword matching
	containsKeyword := func(text, keyword string) bool {
		// For multi-word keywords, use exact phrase matching
		if strings.Contains(keyword, " ") {
			return strings.Contains(text, keyword)
		}

		// For single words, use word boundary checking to avoid false positives
		// This prevents "test" from matching "test namespace" when looking for "network test"
		words := strings.Fields(text)
		for _, word := range words {
			// Remove common punctuation
			cleanWord := strings.Trim(word, ".,!?;:-")
			if cleanWord == keyword {
				return true
			}
		}
		return false
	}

	// Check for network-specific keywords with improved matching
	for _, keyword := range networkKeywords {
		if containsKeyword(lowerQuery, keyword) {
			return true
		}
	}

	// Check for pod troubleshooting keywords with improved matching
	for _, keyword := range podTroubleshootingKeywords {
		if containsKeyword(lowerQuery, keyword) {
			return true
		}
	}

	// Check for pod mention with specific troubleshooting context only
	// Remove overly broad checks that could match resource creation requests
	if strings.Contains(lowerQuery, "pod") &&
		(strings.Contains(lowerQuery, "failing") ||
			strings.Contains(lowerQuery, "not working") ||
			strings.Contains(lowerQuery, "crash") ||
			strings.Contains(lowerQuery, "error") ||
			strings.Contains(lowerQuery, "stuck") ||
			strings.Contains(lowerQuery, "pending") ||
			strings.Contains(lowerQuery, "what's wrong") ||
			strings.Contains(lowerQuery, "why is") ||
			strings.Contains(lowerQuery, "troubleshoot pod") ||
			strings.Contains(lowerQuery, "debug pod")) {
		return true
	}

	return false
}

// TroubleshootNetwork performs comprehensive network troubleshooting
func (nt *TroubleshootingEngine) TroubleshootNetwork(query string) *TroubleshootingResult {
	logrus.Debugf("Starting network troubleshooting for: %s", query)

	result := &TroubleshootingResult{
		Query:    query,
		Commands: make([]*executor.ExecutionResult, 0),
		Steps:    make([]WorkflowStep, 0),
	}

	// Extract pod information
	result.PodInfo = nt.extractPodInfo(query)

	// Determine workflow type and generate steps
	result.WorkflowType = nt.determineWorkflowType(query)
	result.Steps = nt.generateWorkflowSteps(result.WorkflowType, result.PodInfo, query)

	// Execute the workflow
	result.Success = nt.executeWorkflow(result)
	result.Summary = nt.generateSummary(result)

	return result
}

// extractPodInfo extracts pod and namespace information from the query
func (nt *TroubleshootingEngine) extractPodInfo(query string) PodInfo {
	podInfo := PodInfo{}

	// Pattern 1: "the <pod> pod in <namespace> namespace" or similar variations
	pattern1 := regexp.MustCompile(`(?:the\s+)?([a-zA-Z0-9\-]+)\s+pod\s+in\s+([a-zA-Z0-9\-]+)\s+namespace`)
	if matches := pattern1.FindStringSubmatch(query); len(matches) > 2 {
		podInfo.PodName = matches[1]
		podInfo.Namespace = matches[2]
		podInfo.Found = true
		return podInfo
	}

	// Pattern 2: "pod <pod> in <namespace>" (with optional "the")
	pattern2 := regexp.MustCompile(`pod\s+([a-zA-Z0-9\-]+)\s+in\s+(?:the\s+)?([a-zA-Z0-9\-]+)`)
	if matches := pattern2.FindStringSubmatch(query); len(matches) > 2 {
		podInfo.PodName = matches[1]
		podInfo.Namespace = matches[2]
		podInfo.Found = true
		return podInfo
	}

	// Pattern 3: "<namespace>/<pod>"
	pattern3 := regexp.MustCompile(`([a-zA-Z0-9\-]+)/([a-zA-Z0-9\-]+)`)
	if matches := pattern3.FindStringSubmatch(query); len(matches) > 2 {
		podInfo.Namespace = matches[1]
		podInfo.PodName = matches[2]
		podInfo.Found = true
		return podInfo
	}

	// Pattern 4: Just namespace mentioned
	namespacePattern := regexp.MustCompile(`(?:in|namespace)\s+([a-zA-Z0-9\-]+)`)
	if matches := namespacePattern.FindStringSubmatch(query); len(matches) > 1 {
		podInfo.Namespace = matches[1]
	}

	// Pattern 5: Just pod name mentioned
	podPattern := regexp.MustCompile(`pod\s+([a-zA-Z0-9\-]+)`)
	if matches := podPattern.FindStringSubmatch(query); len(matches) > 1 {
		podInfo.PodName = matches[1]
	}

	// Pattern 0: "why is <pod> pod" or "<pod> pod stuck/failing/etc"
	pattern0 := regexp.MustCompile(`(?:why\s+is\s+|check\s+|debug\s+|troubleshoot\s+)?([a-zA-Z0-9\-]+)\s+pod(?:\s+stuck|\s+failing|\s+not|\s+in)?`)
	if matches := pattern0.FindStringSubmatch(query); len(matches) > 1 {
		podInfo.PodName = matches[1]
		// Try to find namespace in the rest of the query
		namespacePattern := regexp.MustCompile(`(?:in|namespace)\s+([a-zA-Z0-9\-]+)`)
		if nsMatches := namespacePattern.FindStringSubmatch(query); len(nsMatches) > 1 {
			podInfo.Namespace = nsMatches[1]
		}
		podInfo.Found = true
		return podInfo
	}

	return podInfo
}

// determineWorkflowType determines the type of network troubleshooting workflow
func (nt *TroubleshootingEngine) determineWorkflowType(query string) string {
	lowerQuery := strings.ToLower(query)

	// Check for pod diagnostics first (most specific)
	if strings.Contains(lowerQuery, "troubleshoot") || strings.Contains(lowerQuery, "debug") ||
		strings.Contains(lowerQuery, "diagnose") || strings.Contains(lowerQuery, "analyze") ||
		strings.Contains(lowerQuery, "check") || strings.Contains(lowerQuery, "examine") ||
		strings.Contains(lowerQuery, "failing") || strings.Contains(lowerQuery, "error") ||
		strings.Contains(lowerQuery, "crash") || strings.Contains(lowerQuery, "issues") {
		// If it mentions "pod" in context of troubleshooting, it's pod diagnostics
		if strings.Contains(lowerQuery, "pod") {
			return "pod_diagnostics"
		}
	}

	// Then check for specific network tools
	switch {
	case strings.Contains(lowerQuery, "tcpdump") || strings.Contains(lowerQuery, "packet capture") || strings.Contains(lowerQuery, "capture packets"):
		return "tcpdump"
	case strings.Contains(lowerQuery, "ping") && !strings.Contains(lowerQuery, "troubleshoot"):
		return "ping"
	case strings.Contains(lowerQuery, "dns") || strings.Contains(lowerQuery, "nslookup") || strings.Contains(lowerQuery, "dig"):
		return "dns"
	case strings.Contains(lowerQuery, "curl") || (strings.Contains(lowerQuery, "http") && !strings.Contains(lowerQuery, "troubleshoot")):
		return "http"
	case (strings.Contains(lowerQuery, "netstat") || strings.Contains(lowerQuery, "ss") || strings.Contains(lowerQuery, "lsof")) && !strings.Contains(lowerQuery, "general"):
		return "netstat"
	default:
		return "general"
	}
}

// generateWorkflowSteps generates the appropriate workflow steps
func (nt *TroubleshootingEngine) generateWorkflowSteps(workflowType string, podInfo PodInfo, query string) []WorkflowStep {
	switch workflowType {
	case "tcpdump":
		return nt.generateTcpdumpSteps(podInfo, query)
	case "ping":
		return nt.generatePingSteps(podInfo, query)
	case "dns":
		return nt.generateDNSSteps(podInfo, query)
	case "http":
		return nt.generateHTTPSteps(podInfo, query)
	case "netstat":
		return nt.generateNetstatSteps(podInfo, query)
	case "pod_diagnostics":
		return nt.generatePodDiagnosticsSteps(podInfo, query)
	default:
		return nt.generateGeneralSteps(podInfo, query)
	}
}

// generateTcpdumpSteps generates tcpdump-specific workflow steps
func (nt *TroubleshootingEngine) generateTcpdumpSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}

	if !podInfo.Found || podInfo.PodName == "" {
		// If no specific pod, list pods to help user
		steps = append(steps, WorkflowStep{
			StepNumber:  1,
			Description: "List pods to identify target",
			Command:     fmt.Sprintf("kubectl get pods -n %s", getNamespaceOrDefault(podInfo.Namespace)),
			Purpose:     "Find the correct pod name for packet capture",
		})
		return steps
	}

	// Verify pod exists and get details
	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "Verify pod exists and get details",
		Command:     fmt.Sprintf("kubectl get pod %s -n %s -o wide", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Confirm pod is running and get node information",
	})

	// Create debug pod for packet capture
	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "Create privileged debug pod for packet capture",
		Command:     fmt.Sprintf("kubectl debug %s -n %s -it --image=registry.redhat.io/rhel8/support-tools -- tcpdump -i any -n -c 100", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Launch tcpdump in the pod's network namespace",
	})

	return steps
}

// generatePingSteps generates ping connectivity test steps
func (nt *TroubleshootingEngine) generatePingSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}

	if !podInfo.Found || podInfo.PodName == "" {
		steps = append(steps, WorkflowStep{
			StepNumber:  1,
			Description: "List pods for connectivity testing",
			Command:     fmt.Sprintf("kubectl get pods -n %s", getNamespaceOrDefault(podInfo.Namespace)),
			Purpose:     "Find pods to test connectivity",
		})
		return steps
	}

	// Test basic connectivity
	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "Test basic pod connectivity",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- ping -c 3 8.8.8.8", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Test external connectivity",
	})

	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "Test DNS resolution",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- nslookup kubernetes.default.svc.cluster.local", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Test internal DNS resolution",
	})

	return steps
}

// generateDNSSteps generates DNS troubleshooting steps
func (nt *TroubleshootingEngine) generateDNSSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}

	if !podInfo.Found || podInfo.PodName == "" {
		steps = append(steps, WorkflowStep{
			StepNumber:  1,
			Description: "List pods for DNS testing",
			Command:     fmt.Sprintf("kubectl get pods -n %s", getNamespaceOrDefault(podInfo.Namespace)),
			Purpose:     "Find pods to test DNS resolution",
		})
		return steps
	}

	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "Test DNS configuration",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- cat /etc/resolv.conf", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Check DNS server configuration",
	})

	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "Test internal DNS resolution",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- nslookup kubernetes.default.svc.cluster.local", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Test cluster internal DNS",
	})

	steps = append(steps, WorkflowStep{
		StepNumber:  3,
		Description: "Test external DNS resolution",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- nslookup google.com", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Test external DNS resolution",
	})

	return steps
}

// generateHTTPSteps generates HTTP testing steps
func (nt *TroubleshootingEngine) generateHTTPSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}

	if !podInfo.Found || podInfo.PodName == "" {
		steps = append(steps, WorkflowStep{
			StepNumber:  1,
			Description: "List pods for HTTP testing",
			Command:     fmt.Sprintf("kubectl get pods -n %s", getNamespaceOrDefault(podInfo.Namespace)),
			Purpose:     "Find pods to test HTTP connectivity",
		})
		return steps
	}

	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "Test HTTP connectivity to external service",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- curl -I http://httpbin.org/get", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Test external HTTP connectivity",
	})

	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "Test HTTPS connectivity",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- curl -I https://httpbin.org/get", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Test external HTTPS connectivity",
	})

	return steps
}

// generateNetstatSteps generates network statistics steps
func (nt *TroubleshootingEngine) generateNetstatSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}

	if !podInfo.Found || podInfo.PodName == "" {
		steps = append(steps, WorkflowStep{
			StepNumber:  1,
			Description: "List pods for network analysis",
			Command:     fmt.Sprintf("kubectl get pods -n %s", getNamespaceOrDefault(podInfo.Namespace)),
			Purpose:     "Find pods to analyze network connections",
		})
		return steps
	}

	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "Show network connections",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- netstat -tulpn", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Display active network connections and listening ports",
	})

	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "Show network interfaces",
		Command:     fmt.Sprintf("kubectl exec %s -n %s -- ip addr show", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Display network interface configuration",
	})

	return steps
}

// generateGeneralSteps generates general network troubleshooting steps
func (nt *TroubleshootingEngine) generateGeneralSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}

	namespace := getNamespaceOrDefault(podInfo.Namespace)

	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "List pods and their status",
		Command:     fmt.Sprintf("kubectl get pods -n %s -o wide", namespace),
		Purpose:     "Get overview of pods and their network configuration",
	})

	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "List services",
		Command:     fmt.Sprintf("kubectl get svc -n %s", namespace),
		Purpose:     "Check available services and their endpoints",
	})

	if podInfo.Found && podInfo.PodName != "" {
		steps = append(steps, WorkflowStep{
			StepNumber:  3,
			Description: "Describe specific pod networking",
			Command:     fmt.Sprintf("kubectl describe pod %s -n %s", podInfo.PodName, podInfo.Namespace),
			Purpose:     "Get detailed pod network information",
		})
	}

	return steps
}

// generatePodDiagnosticsSteps generates pod diagnostic troubleshooting steps
func (nt *TroubleshootingEngine) generatePodDiagnosticsSteps(podInfo PodInfo, query string) []WorkflowStep {
	steps := []WorkflowStep{}
	namespace := getNamespaceOrDefault(podInfo.Namespace)

	if !podInfo.Found || podInfo.PodName == "" {
		// If no specific pod, list pods to help user identify the problematic one
		steps = append(steps, WorkflowStep{
			StepNumber:  1,
			Description: "List all pods to identify problematic pods",
			Command:     fmt.Sprintf("kubectl get pods -n %s", namespace),
			Purpose:     "Identify pods with issues (CrashLoopBackOff, ImagePullBackOff, etc.)",
		})

		steps = append(steps, WorkflowStep{
			StepNumber:  2,
			Description: "Get detailed pod status for all pods",
			Command:     fmt.Sprintf("kubectl get pods -n %s -o wide", namespace),
			Purpose:     "Get detailed status, restarts, and node information",
		})
		return steps
	}

	// Step 1: Get basic pod information
	steps = append(steps, WorkflowStep{
		StepNumber:  1,
		Description: "Get pod status and basic information",
		Command:     fmt.Sprintf("kubectl get pod %s -n %s -o wide", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Check current status, restarts, and node assignment",
	})

	// Step 2: Describe pod for detailed diagnostics
	steps = append(steps, WorkflowStep{
		StepNumber:  2,
		Description: "Get detailed pod diagnostics",
		Command:     fmt.Sprintf("kubectl describe pod %s -n %s", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Analyze events, conditions, and container states for root cause",
	})

	// Step 3: Get pod events
	steps = append(steps, WorkflowStep{
		StepNumber:  3,
		Description: "Get recent events for the pod",
		Command:     fmt.Sprintf("kubectl get events --field-selector involvedObject.name=%s -n %s --sort-by='.lastTimestamp'", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Check for recent events that might explain the issue",
	})

	// Step 4: Get pod logs (current and previous if available)
	steps = append(steps, WorkflowStep{
		StepNumber:  4,
		Description: "Get current pod logs",
		Command:     fmt.Sprintf("kubectl logs %s -n %s --tail=50", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Check application logs for errors or crash information",
	})

	// Step 5: Get previous logs if pod is restarting
	steps = append(steps, WorkflowStep{
		StepNumber:  5,
		Description: "Get previous pod logs (if restarted)",
		Command:     fmt.Sprintf("kubectl logs %s -n %s --previous --tail=50 2>/dev/null || echo 'No previous logs available'", podInfo.PodName, podInfo.Namespace),
		Purpose:     "Check logs from previous container instance to understand crashes",
	})

	return steps
}

// executeWorkflow executes the workflow steps and parses results for pod diagnostics
func (nt *TroubleshootingEngine) executeWorkflow(result *TroubleshootingResult) bool {
	successCount := 0

	for _, step := range result.Steps {
		logrus.Debugf("Executing step %d: %s", step.StepNumber, step.Description)

		execResult := nt.executor.Execute(step.Command)
		result.Commands = append(result.Commands, execResult)

		if execResult.ExitCode == 0 {
			successCount++
		}
	}

	// Perform intelligent parsing for pod diagnostics
	if result.WorkflowType == "pod_diagnostics" && len(result.Commands) >= 2 {
		nt.parsePodDiagnostics(result)
	}

	return successCount > 0
}

// generateSummary generates a summary of the troubleshooting results
func (nt *TroubleshootingEngine) generateSummary(result *TroubleshootingResult) string {
	var lines []string

	// For pod diagnostics, check if we have diagnostic analysis and prioritize it
	if result.WorkflowType == "pod_diagnostics" {
		// Look for diagnostic analysis result (should be the last command)
		for i := len(result.Commands) - 1; i >= 0; i-- {
			if result.Commands[i].Command == "diagnostic_analysis" {
				// Return the diagnostic analysis directly - it contains the parsed results
				return result.Commands[i].Output
			}
		}
	}

	// Fall back to standard network troubleshooting summary
	lines = append(lines, fmt.Sprintf("🔍 Network Troubleshooting: %s", strings.ToUpper(result.WorkflowType)))
	lines = append(lines, strings.Repeat("=", 60))

	if result.PodInfo.Found {
		lines = append(lines, fmt.Sprintf("📍 Target: Pod '%s' in namespace '%s'", result.PodInfo.PodName, result.PodInfo.Namespace))
	} else {
		lines = append(lines, "📍 Target: General network troubleshooting")
	}
	lines = append(lines, "")

	successCount := 0
	for i, step := range result.Steps {
		if i < len(result.Commands) {
			cmd := result.Commands[i]
			// Skip diagnostic analysis in the step-by-step output
			if cmd.Command == "diagnostic_analysis" {
				continue
			}

			status := "❌ FAILED"
			if cmd.ExitCode == 0 {
				status = "✅ SUCCESS"
				successCount++
			}

			lines = append(lines, fmt.Sprintf("%d. %s - %s", step.StepNumber, status, step.Description))
			lines = append(lines, fmt.Sprintf("   🔧 Command: %s", step.Command))
			lines = append(lines, fmt.Sprintf("   📋 Purpose: %s", step.Purpose))

			if cmd.ExitCode == 0 && cmd.Output != "" {
				lines = append(lines, fmt.Sprintf("   📤 Output: %s", truncateOutput(cmd.Output, 200)))
			} else if cmd.Error != "" {
				lines = append(lines, fmt.Sprintf("   🚨 Error: %s", cmd.Error))
			}
			lines = append(lines, "")
		}
	}

	// Summary
	if result.Success {
		lines = append(lines, fmt.Sprintf("🎯 Summary: %d/%d steps completed successfully", successCount, len(result.Steps)))
	} else {
		lines = append(lines, "❌ Summary: All steps failed - check pod name, namespace, and permissions")
	}

	return strings.Join(lines, "\n")
}

// parsePodDiagnostics analyzes pod diagnostic output and provides intelligent insights
func (nt *TroubleshootingEngine) parsePodDiagnostics(result *TroubleshootingResult) {
	if len(result.Commands) < 2 {
		return
	}

	diagnostic := DiagnosticResult{
		Issues:    make([]Issue, 0),
		NextSteps: make([]string, 0),
	}

	// Parse kubectl get pod output (first command)
	if len(result.Commands) >= 1 && result.Commands[0].ExitCode == 0 {
		nt.parsePodStatus(result.Commands[0].Output, &diagnostic)
	}

	// Parse kubectl describe pod output (second command)
	if len(result.Commands) >= 2 && result.Commands[1].ExitCode == 0 {
		nt.parseDescribePodOutput(result.Commands[1].Output, &diagnostic)
	}

	// Parse events output (third command)
	if len(result.Commands) >= 3 && result.Commands[2].ExitCode == 0 {
		nt.parseEventsOutput(result.Commands[2].Output, &diagnostic)
	}

	// Parse logs output (fourth and fifth commands)
	if len(result.Commands) >= 4 && result.Commands[3].ExitCode == 0 {
		nt.parseLogsOutput(result.Commands[3].Output, &diagnostic, "current")
	}

	if len(result.Commands) >= 5 && result.Commands[4].ExitCode == 0 {
		nt.parseLogsOutput(result.Commands[4].Output, &diagnostic, "previous")
	}

	// Determine root cause and recommendations
	nt.analyzeRootCause(&diagnostic)

	// Store diagnostic result in the result
	result.Commands = append(result.Commands, &executor.ExecutionResult{
		Command:   "diagnostic_analysis",
		Output:    nt.formatDiagnosticResult(&diagnostic),
		ExitCode:  0,
		Duration:  0,
		Timestamp: time.Now(),
	})
}

// parsePodStatus extracts status information from kubectl get pod output
func (nt *TroubleshootingEngine) parsePodStatus(output string, diagnostic *DiagnosticResult) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "NAME") {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			status := fields[2]
			diagnostic.PodStatus = status

			// Extract restart count if available
			if len(fields) >= 4 {
				restarts := fields[3]
				if restarts != "0" {
					diagnostic.Issues = append(diagnostic.Issues, Issue{
						Type:       "warning",
						Source:     "status",
						Message:    fmt.Sprintf("Pod has restarted %s times", restarts),
						Severity:   "medium",
						Category:   "stability",
						Actionable: true,
						Suggestion: "Check pod logs and events for crash reasons",
					})
				}
			}

			// Analyze status
			switch {
			case strings.Contains(status, "CrashLoopBackOff"):
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "error",
					Source:     "status",
					Message:    "Pod is in CrashLoopBackOff state",
					Severity:   "critical",
					Category:   "stability",
					Actionable: true,
					Suggestion: "Check application logs for crash reasons",
				})
			case strings.Contains(status, "ImagePullBackOff") || strings.Contains(status, "ErrImagePull"):
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "error",
					Source:     "status",
					Message:    "Cannot pull container image",
					Severity:   "critical",
					Category:   "image",
					Actionable: true,
					Suggestion: "Verify image name, tag, and registry access",
				})
			case strings.Contains(status, "Pending"):
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "warning",
					Source:     "status",
					Message:    "Pod is stuck in Pending state",
					Severity:   "high",
					Category:   "scheduling",
					Actionable: true,
					Suggestion: "Check node resources and scheduling constraints",
				})
			case strings.Contains(status, "Terminating"):
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "info",
					Source:     "status",
					Message:    "Pod is terminating",
					Severity:   "medium",
					Category:   "lifecycle",
					Actionable: false,
					Suggestion: "Wait for graceful shutdown or check for stuck processes",
				})
			case strings.Contains(status, "ContainerCreating"):
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "warning",
					Source:     "status",
					Message:    "Pod is stuck in ContainerCreating state",
					Severity:   "high",
					Category:   "scheduling",
					Actionable: true,
					Suggestion: "Check image pull progress, node resources, and storage availability",
				})
			}
		}
	}
}

// parseDescribePodOutput analyzes kubectl describe pod output for detailed issues
func (nt *TroubleshootingEngine) parseDescribePodOutput(output string, diagnostic *DiagnosticResult) {
	lines := strings.Split(output, "\n")
	inEvents := false
	inContainerStatus := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for Events section
		if strings.HasPrefix(trimmedLine, "Events:") {
			inEvents = true
			continue
		}

		// Check for Container Status
		if strings.Contains(trimmedLine, "Container ID:") || strings.Contains(trimmedLine, "State:") {
			inContainerStatus = true
		}

		// Parse container state information
		if inContainerStatus {
			if strings.Contains(trimmedLine, "State:") && strings.Contains(trimmedLine, "Waiting") {
				if strings.Contains(trimmedLine, "ImagePullBackOff") {
					diagnostic.Issues = append(diagnostic.Issues, Issue{
						Type:       "error",
						Source:     "describe",
						Message:    "Container waiting due to image pull failure",
						Severity:   "critical",
						Category:   "image",
						Actionable: true,
						Suggestion: "Check image registry credentials and image availability",
					})
				}
				if strings.Contains(trimmedLine, "CrashLoopBackOff") {
					diagnostic.Issues = append(diagnostic.Issues, Issue{
						Type:       "error",
						Source:     "describe",
						Message:    "Container in crash loop",
						Severity:   "critical",
						Category:   "stability",
						Actionable: true,
						Suggestion: "Examine application logs for startup failures",
					})
				}
			}

			if strings.Contains(trimmedLine, "Last State:") && strings.Contains(trimmedLine, "Terminated") {
				if strings.Contains(trimmedLine, "OOMKilled") {
					diagnostic.Issues = append(diagnostic.Issues, Issue{
						Type:       "error",
						Source:     "describe",
						Message:    "Container was killed due to out of memory",
						Severity:   "high",
						Category:   "compute",
						Actionable: true,
						Suggestion: "Increase memory limits or optimize application memory usage",
					})
				}

				// Extract exit code
				exitCodeRegex := regexp.MustCompile(`Exit Code: (\d+)`)
				if matches := exitCodeRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
					exitCode := matches[1]
					if exitCode != "0" {
						diagnostic.Issues = append(diagnostic.Issues, Issue{
							Type:       "error",
							Source:     "describe",
							Message:    fmt.Sprintf("Container exited with non-zero code: %s", exitCode),
							Severity:   "high",
							Category:   "stability",
							Actionable: true,
							Suggestion: "Check application logs for error details",
						})
					}
				}
			}
		}

		// Parse events for additional context
		if inEvents && strings.Contains(trimmedLine, "Warning") {
			if strings.Contains(trimmedLine, "Failed") {
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "warning",
					Source:     "events",
					Message:    extractEventMessage(trimmedLine),
					Severity:   "medium",
					Category:   "events",
					Actionable: true,
					Suggestion: "Review the specific failure details",
				})
			}
		}
	}
}

// parseEventsOutput analyzes kubectl get events output
func (nt *TroubleshootingEngine) parseEventsOutput(output string, diagnostic *DiagnosticResult) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "LAST SEEN") {
			continue // Skip header
		}

		if strings.Contains(line, "Warning") || strings.Contains(line, "Error") {
			diagnostic.Issues = append(diagnostic.Issues, Issue{
				Type:       "warning",
				Source:     "events",
				Message:    extractEventMessage(line),
				Severity:   "medium",
				Category:   "events",
				Actionable: true,
				Suggestion: "Review event details for context",
			})
		}
	}
}

// parseLogsOutput analyzes pod logs for errors and issues
func (nt *TroubleshootingEngine) parseLogsOutput(output string, diagnostic *DiagnosticResult, logType string) {
	if strings.Contains(output, "No previous logs available") {
		return
	}

	lines := strings.Split(output, "\n")
	errorPatterns := []struct {
		pattern    *regexp.Regexp
		category   string
		severity   string
		suggestion string
	}{
		{regexp.MustCompile(`(?i)(error|exception|fatal|panic|crash)`), "application", "high", "Fix application error"},
		{regexp.MustCompile(`(?i)(out of memory|oom|memory exceeded)`), "compute", "high", "Increase memory limits"},
		{regexp.MustCompile(`(?i)(connection refused|connection timeout|network unreachable)`), "network", "medium", "Check network connectivity"},
		{regexp.MustCompile(`(?i)(permission denied|access denied|unauthorized)`), "config", "medium", "Check RBAC and permissions"},
		{regexp.MustCompile(`(?i)(disk|storage|volume|mount.*fail)`), "storage", "medium", "Check storage configuration"},
	}

	for _, line := range lines {
		for _, pattern := range errorPatterns {
			if pattern.pattern.MatchString(line) {
				diagnostic.Issues = append(diagnostic.Issues, Issue{
					Type:       "error",
					Source:     fmt.Sprintf("logs_%s", logType),
					Message:    fmt.Sprintf("Found in %s logs: %s", logType, strings.TrimSpace(line)),
					Severity:   pattern.severity,
					Category:   pattern.category,
					Actionable: true,
					Suggestion: pattern.suggestion,
				})
				break // Only match first pattern per line
			}
		}
	}
}

// analyzeRootCause determines the primary root cause and recommendations
func (nt *TroubleshootingEngine) analyzeRootCause(diagnostic *DiagnosticResult) {
	if len(diagnostic.Issues) == 0 {
		diagnostic.RootCause = "No critical issues detected"
		diagnostic.Recommendation = "Pod appears to be healthy, monitor for intermittent issues"
		return
	}

	// Priority analysis based on severity and category
	criticalIssues := []Issue{}
	highIssues := []Issue{}

	for _, issue := range diagnostic.Issues {
		if issue.Severity == "critical" {
			criticalIssues = append(criticalIssues, issue)
		} else if issue.Severity == "high" {
			highIssues = append(highIssues, issue)
		}
	}

	// Determine root cause based on most critical issues
	if len(criticalIssues) > 0 {
		primary := criticalIssues[0]
		diagnostic.RootCause = primary.Message
		diagnostic.Recommendation = primary.Suggestion

		// Add specific next steps based on category
		switch primary.Category {
		case "image":
			diagnostic.NextSteps = []string{
				"Verify the container image name and tag are correct",
				"Check if the image registry is accessible",
				"Verify image pull secrets if using private registry",
				"Test image pull manually: kubectl debug node/<node> -it --image=<image>",
			}
		case "stability":
			diagnostic.NextSteps = []string{
				"Examine pod logs for startup errors: kubectl logs <pod> -n <namespace>",
				"Check resource limits and requests",
				"Verify application health check endpoints",
				"Review application configuration and dependencies",
			}
		case "compute":
			diagnostic.NextSteps = []string{
				"Increase memory limits in pod specification",
				"Analyze memory usage patterns in the application",
				"Consider using horizontal pod autoscaling",
				"Review application memory optimization opportunities",
			}
		case "scheduling":
			diagnostic.NextSteps = []string{
				"Check if container image is accessible: kubectl describe pod <pod> -n <namespace>",
				"Verify node has sufficient resources (CPU, memory, disk space)",
				"Check if persistent volumes are available and accessible",
				"Review pod events for specific error messages",
				"Verify image pull secrets are correctly configured",
				"Check if any admission controllers are blocking pod creation",
			}
		}
	} else if len(highIssues) > 0 {
		primary := highIssues[0]
		diagnostic.RootCause = primary.Message
		diagnostic.Recommendation = primary.Suggestion
	} else {
		diagnostic.RootCause = "Multiple minor issues detected"
		diagnostic.Recommendation = "Review all issues and address systematically"
	}

	// Set logs needed flag
	for _, issue := range diagnostic.Issues {
		if issue.Category == "stability" || issue.Category == "application" {
			diagnostic.LogsNeeded = true
			break
		}
	}
}

// formatDiagnosticResult formats the diagnostic analysis for display
func (nt *TroubleshootingEngine) formatDiagnosticResult(diagnostic *DiagnosticResult) string {
	var lines []string

	lines = append(lines, "🔍 POD DIAGNOSTIC ANALYSIS")
	lines = append(lines, strings.Repeat("=", 50))

	if diagnostic.PodStatus != "" {
		lines = append(lines, fmt.Sprintf("📊 Current Status: %s", diagnostic.PodStatus))
	}

	lines = append(lines, fmt.Sprintf("🎯 Root Cause: %s", diagnostic.RootCause))
	lines = append(lines, fmt.Sprintf("💡 Recommendation: %s", diagnostic.Recommendation))
	lines = append(lines, "")

	if len(diagnostic.Issues) > 0 {
		lines = append(lines, "🚨 ISSUES FOUND:")
		for i, issue := range diagnostic.Issues {
			severity := getSeverityEmoji(issue.Severity)
			lines = append(lines, fmt.Sprintf("%d. %s [%s] %s", i+1, severity, strings.ToUpper(issue.Category), issue.Message))
			if issue.Suggestion != "" {
				lines = append(lines, fmt.Sprintf("   💡 %s", issue.Suggestion))
			}
		}
		lines = append(lines, "")
	}

	if len(diagnostic.NextSteps) > 0 {
		lines = append(lines, "📋 NEXT STEPS:")
		for i, step := range diagnostic.NextSteps {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, step))
		}
	}

	return strings.Join(lines, "\n")
}

// Helper functions
func extractEventMessage(eventLine string) string {
	// Extract the message part from the event line
	parts := strings.Fields(eventLine)
	if len(parts) > 6 {
		return strings.Join(parts[6:], " ")
	}
	return eventLine
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🔵"
	default:
		return "ℹ️"
	}
}

// truncateOutput truncates the output string to the specified max length
func truncateOutput(output string, maxLength int) string {
	if len(output) > maxLength {
		return output[:maxLength] + "..."
	}
	return output
}

// getNamespaceOrDefault returns the namespace or a default value
func getNamespaceOrDefault(namespace string) string {
	if namespace == "" {
		return "default"
	}
	return namespace
}
