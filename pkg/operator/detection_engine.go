package operator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/executor"
	"github.com/sirupsen/logrus"
)

// DetectionEngine handles OpenShift operator detection logic
type DetectionEngine struct {
	executor *executor.CommandExecutor
}

// DetectionResult represents the result of operator detection
type DetectionResult struct {
	OperatorName string                      `json:"operator_name"`
	IsInstalled  bool                        `json:"is_installed"`
	Summary      string                      `json:"summary"`
	Details      []CheckResult               `json:"details"`
	Commands     []*executor.ExecutionResult `json:"commands"`
}

// CheckResult represents the result of a single check
type CheckResult struct {
	CheckType   string `json:"check_type"`
	Found       bool   `json:"found"`
	Description string `json:"description"`
	Details     string `json:"details"`
}

// NewDetectionEngine creates a new operator detection engine
func NewDetectionEngine() *DetectionEngine {
	return &DetectionEngine{
		executor: executor.NewCommandExecutor(),
	}
}

// NewDetectionEngineWithKubeconfig creates a new operator detection engine with kubeconfig
func NewDetectionEngineWithKubeconfig(kubeconfigPath string) *DetectionEngine {
	return &DetectionEngine{
		executor: executor.NewCommandExecutorWithKubeconfig(kubeconfigPath),
	}
}

// IsOperatorQuery detects if a query is asking about operator installation
func (de *DetectionEngine) IsOperatorQuery(query string) (bool, string) {
	lowerQuery := strings.ToLower(query)

	// Pattern 1: "is <operator> operator installed"
	operatorPattern := regexp.MustCompile(`is\s+([a-zA-Z0-9\-_]+)\s+operator\s+(installed|running|available)`)
	if matches := operatorPattern.FindStringSubmatch(lowerQuery); len(matches) > 1 {
		return true, matches[1]
	}

	// Pattern 2: "check <operator> operator"
	checkPattern := regexp.MustCompile(`check\s+([a-zA-Z0-9\-_]+)\s+operator`)
	if matches := checkPattern.FindStringSubmatch(lowerQuery); len(matches) > 1 {
		return true, matches[1]
	}

	// Pattern 3: "<operator> operator status"
	statusPattern := regexp.MustCompile(`([a-zA-Z0-9\-_]+)\s+operator\s+status`)
	if matches := statusPattern.FindStringSubmatch(lowerQuery); len(matches) > 1 {
		return true, matches[1]
	}

	// Pattern 4: General operator keywords
	operatorKeywords := []string{"operator installed", "operator running", "operator status"}
	for _, keyword := range operatorKeywords {
		if strings.Contains(lowerQuery, keyword) {
			// Extract potential operator name
			words := strings.Fields(lowerQuery)
			for i, word := range words {
				if word == "operator" && i > 0 {
					return true, words[i-1]
				}
			}
		}
	}

	return false, ""
}

// DetectOperator performs comprehensive operator detection
func (de *DetectionEngine) DetectOperator(operatorName string) *DetectionResult {
	logrus.Debugf("Starting operator detection for: %s", operatorName)

	result := &DetectionResult{
		OperatorName: operatorName,
		Details:      make([]CheckResult, 0),
		Commands:     make([]*executor.ExecutionResult, 0),
	}

	// Define the comprehensive check sequence
	checks := []struct {
		checkType   string
		command     string
		description string
	}{
		{
			checkType:   "subscription",
			command:     fmt.Sprintf("kubectl get subscription -n openshift-operators | grep %s", operatorName),
			description: "PRIMARY CHECK: Operator subscription in openshift-operators namespace",
		},
		{
			checkType:   "csv",
			command:     fmt.Sprintf("kubectl get csv -n openshift-operators | grep %s", operatorName),
			description: "ClusterServiceVersion (CSV) status in openshift-operators namespace",
		},
		{
			checkType:   "installplan",
			command:     fmt.Sprintf("kubectl get installplans -n openshift-operators | grep %s", operatorName),
			description: "Installation plans for operator in openshift-operators namespace",
		},
		{
			checkType:   "pods",
			command:     fmt.Sprintf("kubectl get pods -n openshift-operators | grep %s", operatorName),
			description: "Operator pods running in openshift-operators namespace",
		},
		{
			checkType:   "crds",
			command:     fmt.Sprintf("kubectl get crds | grep %s", operatorName),
			description: "Custom Resource Definitions (CRDs) provided by operator",
		},
	}

	// Execute all checks
	foundInAnyCheck := false
	for _, check := range checks {
		logrus.Debugf("Executing check: %s", check.checkType)

		execResult := de.executor.Execute(check.command)
		result.Commands = append(result.Commands, execResult)

		checkResult := CheckResult{
			CheckType:   check.checkType,
			Description: check.description,
		}

		if execResult.Error != "" {
			checkResult.Found = false
			checkResult.Details = fmt.Sprintf("Error: %s", execResult.Error)
		} else {
			analysis := de.analyzeOutput(check.checkType, execResult.Output, operatorName)
			checkResult.Found = analysis.found
			checkResult.Details = analysis.details
			if analysis.found {
				foundInAnyCheck = true
			}
		}

		result.Details = append(result.Details, checkResult)
	}

	// Set overall result
	result.IsInstalled = foundInAnyCheck
	result.Summary = de.generateSummary(result)

	return result
}

// outputAnalysis represents analyzed command output
type outputAnalysis struct {
	found   bool
	details string
}

// analyzeOutput analyzes the output of each check type
func (de *DetectionEngine) analyzeOutput(checkType, output, operatorName string) outputAnalysis {
	output = strings.TrimSpace(output)

	if output == "" {
		return outputAnalysis{
			found:   false,
			details: "No results found",
		}
	}

	switch checkType {
	case "subscription":
		return de.analyzeSubscriptionOutput(output, operatorName)
	case "csv":
		return de.analyzeCSVOutput(output, operatorName)
	case "installplan":
		return de.analyzeInstallPlanOutput(output, operatorName)
	case "pods":
		return de.analyzePodsOutput(output, operatorName)
	case "crds":
		return de.analyzeCRDsOutput(output, operatorName)
	default:
		return outputAnalysis{
			found:   len(output) > 0,
			details: fmt.Sprintf("Raw output: %s", output),
		}
	}
}

// analyzeSubscriptionOutput analyzes subscription command output
func (de *DetectionEngine) analyzeSubscriptionOutput(output, operatorName string) outputAnalysis {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(operatorName)) {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				return outputAnalysis{
					found:   true,
					details: fmt.Sprintf("✅ FOUND: Subscription '%s' in namespace '%s', status: %s", parts[0], parts[1], parts[3]),
				}
			}
			return outputAnalysis{
				found:   true,
				details: fmt.Sprintf("✅ FOUND: Subscription entry: %s", line),
			}
		}
	}
	return outputAnalysis{
		found:   false,
		details: "❌ No matching subscription found",
	}
}

// analyzeCSVOutput analyzes ClusterServiceVersion output
func (de *DetectionEngine) analyzeCSVOutput(output, operatorName string) outputAnalysis {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(operatorName)) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return outputAnalysis{
					found:   true,
					details: fmt.Sprintf("✅ FOUND: CSV '%s' in namespace '%s', phase: %s", parts[0], parts[1], parts[2]),
				}
			}
			return outputAnalysis{
				found:   true,
				details: fmt.Sprintf("✅ FOUND: CSV entry: %s", line),
			}
		}
	}
	return outputAnalysis{
		found:   false,
		details: "❌ No matching ClusterServiceVersion found",
	}
}

// analyzeInstallPlanOutput analyzes InstallPlan output
func (de *DetectionEngine) analyzeInstallPlanOutput(output, operatorName string) outputAnalysis {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(operatorName)) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return outputAnalysis{
					found:   true,
					details: fmt.Sprintf("✅ FOUND: InstallPlan '%s' in namespace '%s', approved: %s", parts[0], parts[1], parts[2]),
				}
			}
			return outputAnalysis{
				found:   true,
				details: fmt.Sprintf("✅ FOUND: InstallPlan entry: %s", line),
			}
		}
	}
	return outputAnalysis{
		found:   false,
		details: "❌ No matching InstallPlan found",
	}
}

// analyzePodsOutput analyzes pod output
func (de *DetectionEngine) analyzePodsOutput(output, operatorName string) outputAnalysis {
	lines := strings.Split(output, "\n")
	podCount := 0
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(operatorName)) {
			podCount++
		}
	}
	if podCount > 0 {
		return outputAnalysis{
			found:   true,
			details: fmt.Sprintf("✅ FOUND: %d operator pod(s) running", podCount),
		}
	}
	return outputAnalysis{
		found:   false,
		details: "❌ No operator pods found",
	}
}

// analyzeCRDsOutput analyzes CRD output
func (de *DetectionEngine) analyzeCRDsOutput(output, operatorName string) outputAnalysis {
	lines := strings.Split(output, "\n")
	crdCount := 0
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(operatorName)) {
			crdCount++
		}
	}
	if crdCount > 0 {
		return outputAnalysis{
			found:   true,
			details: fmt.Sprintf("✅ FOUND: %d Custom Resource Definition(s) related to operator", crdCount),
		}
	}
	return outputAnalysis{
		found:   false,
		details: "❌ No related CRDs found",
	}
}

// generateSummary creates a comprehensive summary of the detection results
func (de *DetectionEngine) generateSummary(result *DetectionResult) string {
	if result.IsInstalled {
		installedChecks := make([]string, 0)
		for _, check := range result.Details {
			if check.Found {
				installedChecks = append(installedChecks, check.CheckType)
			}
		}
		return fmt.Sprintf("🎯 OPERATOR INSTALLED: '%s' operator is installed on this OpenShift cluster. Found evidence in: %s",
			result.OperatorName, strings.Join(installedChecks, ", "))
	}

	return fmt.Sprintf("❌ OPERATOR NOT FOUND: '%s' operator is not installed on this OpenShift cluster. No evidence found in any of the standard OpenShift operator locations.",
		result.OperatorName)
}
