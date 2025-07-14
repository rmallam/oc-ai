package network

import (
	"strings"
	"testing"
)

func TestPodDiagnosticsWorkflow(t *testing.T) {
	engine := NewTroubleshootingEngine()

	// Test detection of pod troubleshooting queries
	testCases := []struct {
		query    string
		expected bool
	}{
		{"troubleshoot the httpd pod in app1 namespace", true},
		{"debug pod nginx in default namespace", true},
		{"analyze the failing pod my-app", true},
		{"check pod status for web-server", true},
		{"why is pod frontend crashing", true},
		{"list all pods", false},
		{"create deployment", false},
	}

	for _, tc := range testCases {
		result := engine.IsNetworkQuery(tc.query)
		if result != tc.expected {
			t.Errorf("IsNetworkQuery(%q) = %v, expected %v", tc.query, result, tc.expected)
		}
	}
}

func TestDetermineWorkflowType(t *testing.T) {
	engine := NewTroubleshootingEngine()

	testCases := []struct {
		query    string
		expected string
	}{
		{"troubleshoot the httpd pod in app1 namespace", "pod_diagnostics"},
		{"debug nginx pod issues", "pod_diagnostics"},
		{"analyze failing pod", "pod_diagnostics"},
		{"tcpdump on pod", "tcpdump"},
		{"ping from pod", "ping"},
		{"general network issue", "general"},
		{"netstat connections", "netstat"},
	}

	for _, tc := range testCases {
		result := engine.determineWorkflowType(tc.query)
		if result != tc.expected {
			t.Errorf("determineWorkflowType(%q) = %v, expected %v", tc.query, result, tc.expected)
		}
	}
}

func TestPodInfoExtraction(t *testing.T) {
	engine := NewTroubleshootingEngine()

	testCases := []struct {
		query         string
		expectedPod   string
		expectedNS    string
		expectedFound bool
	}{
		{"troubleshoot the httpd pod in app1 namespace", "httpd", "app1", true},
		{"debug pod nginx in default namespace", "nginx", "default", true},
		{"check app1/httpd", "httpd", "app1", true},
		{"analyze pod web-server", "web-server", "", false},
		{"troubleshoot in production namespace", "", "production", false},
	}

	for _, tc := range testCases {
		result := engine.extractPodInfo(tc.query)
		if result.PodName != tc.expectedPod {
			t.Errorf("extractPodInfo(%q).PodName = %v, expected %v", tc.query, result.PodName, tc.expectedPod)
		}
		if result.Namespace != tc.expectedNS {
			t.Errorf("extractPodInfo(%q).Namespace = %v, expected %v", tc.query, result.Namespace, tc.expectedNS)
		}
		if result.Found != tc.expectedFound {
			t.Errorf("extractPodInfo(%q).Found = %v, expected %v", tc.query, result.Found, tc.expectedFound)
		}
	}
}

func TestPodStatusParsing(t *testing.T) {
	engine := NewTroubleshootingEngine()

	// Mock kubectl get pod output for CrashLoopBackOff
	crashLoopOutput := `NAME    READY   STATUS             RESTARTS   AGE
httpd   0/1     CrashLoopBackOff   5          10m`

	diagnostic := DiagnosticResult{
		Issues:    make([]Issue, 0),
		NextSteps: make([]string, 0),
	}

	engine.parsePodStatus(crashLoopOutput, &diagnostic)

	if diagnostic.PodStatus != "CrashLoopBackOff" {
		t.Errorf("Expected PodStatus to be CrashLoopBackOff, got %s", diagnostic.PodStatus)
	}

	if len(diagnostic.Issues) == 0 {
		t.Errorf("Expected issues to be found, got none")
	}

	// Check for CrashLoopBackOff issue
	foundCrashLoop := false
	foundRestarts := false
	for _, issue := range diagnostic.Issues {
		if strings.Contains(issue.Message, "CrashLoopBackOff") {
			foundCrashLoop = true
			if issue.Severity != "critical" {
				t.Errorf("Expected CrashLoopBackOff to have critical severity, got %s", issue.Severity)
			}
			if issue.Category != "stability" {
				t.Errorf("Expected CrashLoopBackOff to have stability category, got %s", issue.Category)
			}
		}
		if strings.Contains(issue.Message, "restarted 5 times") {
			foundRestarts = true
		}
	}

	if !foundCrashLoop {
		t.Errorf("Expected to find CrashLoopBackOff issue")
	}
	if !foundRestarts {
		t.Errorf("Expected to find restart count issue")
	}
}

func TestImagePullBackOffParsing(t *testing.T) {
	engine := NewTroubleshootingEngine()

	imagePullOutput := `NAME    READY   STATUS             RESTARTS   AGE
nginx   0/1     ImagePullBackOff   0          5m`

	diagnostic := DiagnosticResult{
		Issues:    make([]Issue, 0),
		NextSteps: make([]string, 0),
	}

	engine.parsePodStatus(imagePullOutput, &diagnostic)

	if diagnostic.PodStatus != "ImagePullBackOff" {
		t.Errorf("Expected PodStatus to be ImagePullBackOff, got %s", diagnostic.PodStatus)
	}

	foundImagePull := false
	for _, issue := range diagnostic.Issues {
		if strings.Contains(issue.Message, "Cannot pull container image") {
			foundImagePull = true
			if issue.Severity != "critical" {
				t.Errorf("Expected ImagePullBackOff to have critical severity, got %s", issue.Severity)
			}
			if issue.Category != "image" {
				t.Errorf("Expected ImagePullBackOff to have image category, got %s", issue.Category)
			}
		}
	}

	if !foundImagePull {
		t.Errorf("Expected to find ImagePullBackOff issue")
	}
}

func TestLogsParsing(t *testing.T) {
	engine := NewTroubleshootingEngine()

	logsWithErrors := `2024-01-01T10:00:00Z Starting application
2024-01-01T10:00:01Z ERROR: Failed to connect to database
2024-01-01T10:00:02Z FATAL: Application crashed due to unhandled exception
2024-01-01T10:00:03Z Application exiting with code 1`

	diagnostic := DiagnosticResult{
		Issues:    make([]Issue, 0),
		NextSteps: make([]string, 0),
	}

	engine.parseLogsOutput(logsWithErrors, &diagnostic, "current")

	if len(diagnostic.Issues) == 0 {
		t.Errorf("Expected issues to be found in logs, got none")
	}

	foundError := false
	foundFatal := false
	for _, issue := range diagnostic.Issues {
		if strings.Contains(issue.Message, "ERROR: Failed to connect") {
			foundError = true
			if issue.Category != "application" {
				t.Errorf("Expected error to have application category, got %s", issue.Category)
			}
		}
		if strings.Contains(issue.Message, "FATAL: Application crashed") {
			foundFatal = true
			if issue.Severity != "high" {
				t.Errorf("Expected fatal error to have high severity, got %s", issue.Severity)
			}
		}
	}

	if !foundError {
		t.Errorf("Expected to find ERROR in logs")
	}
	if !foundFatal {
		t.Errorf("Expected to find FATAL in logs")
	}
}

func TestRootCauseAnalysis(t *testing.T) {
	engine := NewTroubleshootingEngine()

	diagnostic := DiagnosticResult{
		Issues: []Issue{
			{
				Type:       "error",
				Source:     "status",
				Message:    "Pod is in CrashLoopBackOff state",
				Severity:   "critical",
				Category:   "stability",
				Actionable: true,
				Suggestion: "Check application logs for crash reasons",
			},
			{
				Type:       "warning",
				Source:     "status",
				Message:    "Pod has restarted 5 times",
				Severity:   "medium",
				Category:   "stability",
				Actionable: true,
				Suggestion: "Check pod logs and events for crash reasons",
			},
		},
		NextSteps: make([]string, 0),
	}

	engine.analyzeRootCause(&diagnostic)

	if diagnostic.RootCause != "Pod is in CrashLoopBackOff state" {
		t.Errorf("Expected root cause to be the critical issue, got %s", diagnostic.RootCause)
	}

	if diagnostic.Recommendation != "Check application logs for crash reasons" {
		t.Errorf("Expected recommendation from critical issue, got %s", diagnostic.Recommendation)
	}

	if len(diagnostic.NextSteps) == 0 {
		t.Errorf("Expected next steps to be generated")
	}

	if !diagnostic.LogsNeeded {
		t.Errorf("Expected LogsNeeded to be true for stability issues")
	}
}

func TestGeneratePodDiagnosticsSteps(t *testing.T) {
	engine := NewTroubleshootingEngine()

	podInfo := PodInfo{
		PodName:   "httpd",
		Namespace: "app1",
		Found:     true,
	}

	steps := engine.generatePodDiagnosticsSteps(podInfo, "troubleshoot httpd pod")

	if len(steps) != 5 {
		t.Errorf("Expected 5 diagnostic steps, got %d", len(steps))
	}

	expectedCommands := []string{
		"kubectl get pod httpd -n app1 -o wide",
		"kubectl describe pod httpd -n app1",
		"kubectl get events --field-selector involvedObject.name=httpd -n app1",
		"kubectl logs httpd -n app1 --tail=50",
		"kubectl logs httpd -n app1 --previous --tail=50",
	}

	for i, step := range steps {
		if !strings.Contains(step.Command, expectedCommands[i]) {
			t.Errorf("Step %d command mismatch. Expected to contain %q, got %q",
				i+1, expectedCommands[i], step.Command)
		}
	}
}

func TestFullPodDiagnosticsWorkflow(t *testing.T) {
	engine := NewTroubleshootingEngine()

	result := engine.TroubleshootNetwork("troubleshoot the httpd pod in app1 namespace")

	if result.WorkflowType != "pod_diagnostics" {
		t.Errorf("Expected workflow type pod_diagnostics, got %s", result.WorkflowType)
	}

	if !result.PodInfo.Found {
		t.Errorf("Expected pod info to be found")
	}

	if result.PodInfo.PodName != "httpd" {
		t.Errorf("Expected pod name httpd, got %s", result.PodInfo.PodName)
	}

	if result.PodInfo.Namespace != "app1" {
		t.Errorf("Expected namespace app1, got %s", result.PodInfo.Namespace)
	}

	if len(result.Steps) == 0 {
		t.Errorf("Expected diagnostic steps to be generated")
	}

	// The summary should be intelligent for pod diagnostics
	if !strings.Contains(result.Summary, "POD DIAGNOSTIC") ||
		strings.Contains(result.Summary, "Network Troubleshooting") {
		t.Errorf("Expected intelligent pod diagnostic summary, got: %s", result.Summary)
	}
}
