package diagnostics

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// AnalysisEngine performs analysis on collected diagnostic data
type AnalysisEngine struct {
	logger *logrus.Logger
}

// AnalysisResult represents the result of diagnostic analysis
type AnalysisResult struct {
	Type            string                 `json:"type"`
	FilePath        string                 `json:"file_path"`
	Issues          []Issue                `json:"issues"`
	Metrics         map[string]interface{} `json:"metrics"`
	Summary         string                 `json:"summary"`
	Recommendations []string               `json:"recommendations"`
	Timestamp       time.Time              `json:"timestamp"`
}

// Issue represents a discovered issue
type Issue struct {
	Severity    string            `json:"severity"` // critical, warning, info
	Category    string            `json:"category"` // performance, error, security, configuration
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Location    string            `json:"location"` // file:line or component
	Evidence    []string          `json:"evidence"` // relevant log lines or data
	Resolution  string            `json:"resolution"`
	Metadata    map[string]string `json:"metadata"`
}

// LogPattern represents a pattern to match in logs
type LogPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Severity    string
	Category    string
	Description string
	Resolution  string
}

// NewAnalysisEngine creates a new analysis engine
func NewAnalysisEngine(logger *logrus.Logger) *AnalysisEngine {
	return &AnalysisEngine{
		logger: logger,
	}
}

// AnalyzeMustGather analyzes must-gather data
func (ae *AnalysisEngine) AnalyzeMustGather(ctx context.Context, mustGatherPath string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Type:      "must-gather-analysis",
		FilePath:  mustGatherPath,
		Issues:    []Issue{},
		Metrics:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	ae.logger.Infof("Starting must-gather analysis: %s", mustGatherPath)

	// Analyze cluster version and health
	if err := ae.analyzeClusterHealth(mustGatherPath, result); err != nil {
		ae.logger.Warnf("Failed to analyze cluster health: %v", err)
	}

	// Analyze node health
	if err := ae.analyzeNodeHealth(mustGatherPath, result); err != nil {
		ae.logger.Warnf("Failed to analyze node health: %v", err)
	}

	// Analyze pod issues
	if err := ae.analyzePodIssues(mustGatherPath, result); err != nil {
		ae.logger.Warnf("Failed to analyze pod issues: %v", err)
	}

	// Analyze events
	if err := ae.analyzeEvents(mustGatherPath, result); err != nil {
		ae.logger.Warnf("Failed to analyze events: %v", err)
	}

	// Analyze operator logs
	if err := ae.analyzeOperatorLogs(mustGatherPath, result); err != nil {
		ae.logger.Warnf("Failed to analyze operator logs: %v", err)
	}

	// Generate summary and recommendations
	ae.generateSummaryAndRecommendations(result)

	ae.logger.Infof("Must-gather analysis completed: found %d issues", len(result.Issues))
	return result, nil
}

// AnalyzeLogs analyzes collected log files
func (ae *AnalysisEngine) AnalyzeLogs(ctx context.Context, logPath string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Type:      "log-analysis",
		FilePath:  logPath,
		Issues:    []Issue{},
		Metrics:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	ae.logger.Infof("Starting log analysis: %s", logPath)

	// Get log patterns for analysis
	patterns := ae.getLogPatterns()

	// Analyze log files
	err := filepath.Walk(logPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".log") || strings.HasSuffix(path, ".txt")) {
			if err := ae.analyzeLogFile(path, patterns, result); err != nil {
				ae.logger.Warnf("Failed to analyze log file %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk log directory: %v", err)
	}

	// Analyze log metrics
	ae.calculateLogMetrics(result)

	// Generate summary
	ae.generateSummaryAndRecommendations(result)

	ae.logger.Infof("Log analysis completed: found %d issues", len(result.Issues))
	return result, nil
}

// AnalyzeTcpdump analyzes packet capture data
func (ae *AnalysisEngine) AnalyzeTcpdump(ctx context.Context, pcapPath string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Type:      "tcpdump-analysis",
		FilePath:  pcapPath,
		Issues:    []Issue{},
		Metrics:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	ae.logger.Infof("Starting tcpdump analysis: %s", pcapPath)

	// Use tshark for packet analysis if available
	if err := ae.analyzePcapWithTshark(pcapPath, result); err != nil {
		// Fallback to basic file analysis
		ae.logger.Warnf("Tshark analysis failed, using basic analysis: %v", err)
		if err := ae.analyzePcapBasic(pcapPath, result); err != nil {
			return nil, fmt.Errorf("pcap analysis failed: %v", err)
		}
	}

	ae.generateSummaryAndRecommendations(result)

	ae.logger.Infof("Tcpdump analysis completed: found %d issues", len(result.Issues))
	return result, nil
}

// analyzeClusterHealth analyzes cluster health from must-gather
func (ae *AnalysisEngine) analyzeClusterHealth(mustGatherPath string, result *AnalysisResult) error {
	// Look for cluster version info
	versionPath := filepath.Join(mustGatherPath, "cluster-scoped-resources", "config.openshift.io", "clusterversions.yaml")
	if data, err := os.ReadFile(versionPath); err == nil {
		if strings.Contains(string(data), "Degraded: \"True\"") {
			result.Issues = append(result.Issues, Issue{
				Severity:    "critical",
				Category:    "cluster",
				Title:       "Cluster Version Degraded",
				Description: "Cluster version operator reports degraded state",
				Location:    versionPath,
				Evidence:    []string{"Degraded: \"True\" found in cluster version"},
				Resolution:  "Check cluster version operator logs and resolve blocking conditions",
			})
		}
	}

	// Check cluster operators
	operatorsPath := filepath.Join(mustGatherPath, "cluster-scoped-resources", "config.openshift.io", "clusteroperators.yaml")
	if data, err := os.ReadFile(operatorsPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(line, "status: \"False\"") && i > 0 {
				// Look for the operator name in previous lines
				for j := i - 1; j >= 0 && j >= i-10; j-- {
					if strings.Contains(lines[j], "name:") {
						operatorName := strings.TrimSpace(strings.Split(lines[j], ":")[1])
						result.Issues = append(result.Issues, Issue{
							Severity:    "warning",
							Category:    "operator",
							Title:       fmt.Sprintf("Operator %s Not Available", operatorName),
							Description: fmt.Sprintf("Cluster operator %s reports status: False", operatorName),
							Location:    fmt.Sprintf("%s:line %d", operatorsPath, i+1),
							Evidence:    []string{line},
							Resolution:  fmt.Sprintf("Check %s operator logs and resolve issues", operatorName),
						})
						break
					}
				}
			}
		}
	}

	return nil
}

// analyzeNodeHealth analyzes node health from must-gather
func (ae *AnalysisEngine) analyzeNodeHealth(mustGatherPath string, result *AnalysisResult) error {
	nodesPath := filepath.Join(mustGatherPath, "cluster-scoped-resources", "core", "nodes.yaml")
	if data, err := os.ReadFile(nodesPath); err == nil {
		// Check for node conditions
		if strings.Contains(string(data), "Ready: \"False\"") {
			result.Issues = append(result.Issues, Issue{
				Severity:    "critical",
				Category:    "node",
				Title:       "Node Not Ready",
				Description: "One or more nodes are not in Ready state",
				Location:    nodesPath,
				Evidence:    []string{"Ready: \"False\" found in node status"},
				Resolution:  "Check node conditions and resolve underlying issues",
			})
		}

		if strings.Contains(string(data), "DiskPressure: \"True\"") {
			result.Issues = append(result.Issues, Issue{
				Severity:    "warning",
				Category:    "node",
				Title:       "Node Disk Pressure",
				Description: "Node experiencing disk pressure",
				Location:    nodesPath,
				Evidence:    []string{"DiskPressure: \"True\" found in node status"},
				Resolution:  "Free up disk space on the affected node",
			})
		}

		if strings.Contains(string(data), "MemoryPressure: \"True\"") {
			result.Issues = append(result.Issues, Issue{
				Severity:    "warning",
				Category:    "node",
				Title:       "Node Memory Pressure",
				Description: "Node experiencing memory pressure",
				Location:    nodesPath,
				Evidence:    []string{"MemoryPressure: \"True\" found in node status"},
				Resolution:  "Check memory usage and consider scaling or optimizing workloads",
			})
		}
	}

	return nil
}

// analyzePodIssues analyzes pod issues from must-gather
func (ae *AnalysisEngine) analyzePodIssues(mustGatherPath string, result *AnalysisResult) error {
	// Walk through namespaces and look for pod issues
	namespacesPath := filepath.Join(mustGatherPath, "namespaces")

	return filepath.Walk(namespacesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		if strings.HasSuffix(path, "pods.yaml") {
			if data, err := os.ReadFile(path); err == nil {
				content := string(data)

				// Check for crash loop backoff
				if strings.Contains(content, "CrashLoopBackOff") {
					result.Issues = append(result.Issues, Issue{
						Severity:    "critical",
						Category:    "pod",
						Title:       "Pod in CrashLoopBackOff",
						Description: "Pod is repeatedly crashing",
						Location:    path,
						Evidence:    []string{"CrashLoopBackOff status found"},
						Resolution:  "Check pod logs for error messages and fix the underlying issue",
					})
				}

				// Check for image pull errors
				if strings.Contains(content, "ImagePullBackOff") || strings.Contains(content, "ErrImagePull") {
					result.Issues = append(result.Issues, Issue{
						Severity:    "warning",
						Category:    "pod",
						Title:       "Image Pull Error",
						Description: "Pod cannot pull container image",
						Location:    path,
						Evidence:    []string{"ImagePullBackOff or ErrImagePull status found"},
						Resolution:  "Check image name, registry access, and authentication",
					})
				}

				// Check for pending pods
				if strings.Contains(content, "phase: Pending") {
					result.Issues = append(result.Issues, Issue{
						Severity:    "warning",
						Category:    "pod",
						Title:       "Pod Pending",
						Description: "Pod is stuck in Pending state",
						Location:    path,
						Evidence:    []string{"phase: Pending found"},
						Resolution:  "Check resource availability, node selectors, and scheduling constraints",
					})
				}
			}
		}
		return nil
	})
}

// analyzeEvents analyzes cluster events
func (ae *AnalysisEngine) analyzeEvents(mustGatherPath string, result *AnalysisResult) error {
	eventsPath := filepath.Join(mustGatherPath, "cluster-scoped-resources", "core", "events.yaml")
	if data, err := os.ReadFile(eventsPath); err == nil {
		content := string(data)

		// Look for error events
		errorPatterns := []string{
			"Failed",
			"Error",
			"Warning",
			"FailedScheduling",
			"FailedMount",
			"Unhealthy",
		}

		for _, pattern := range errorPatterns {
			if strings.Contains(content, pattern) {
				result.Issues = append(result.Issues, Issue{
					Severity:    "info",
					Category:    "events",
					Title:       fmt.Sprintf("Event: %s", pattern),
					Description: fmt.Sprintf("Found events containing %s", pattern),
					Location:    eventsPath,
					Evidence:    []string{fmt.Sprintf("Events containing '%s' found", pattern)},
					Resolution:  "Review events for details and address underlying issues",
				})
			}
		}
	}

	return nil
}

// analyzeOperatorLogs analyzes operator logs
func (ae *AnalysisEngine) analyzeOperatorLogs(mustGatherPath string, result *AnalysisResult) error {
	return filepath.Walk(mustGatherPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if strings.Contains(path, "openshift-") && strings.HasSuffix(path, ".log") {
			if err := ae.analyzeLogFile(path, ae.getOperatorLogPatterns(), result); err != nil {
				ae.logger.Warnf("Failed to analyze operator log %s: %v", path, err)
			}
		}
		return nil
	})
}

// analyzeLogFile analyzes a single log file
func (ae *AnalysisEngine) analyzeLogFile(filePath string, patterns []LogPattern, result *AnalysisResult) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, pattern := range patterns {
			if pattern.Pattern.MatchString(line) {
				result.Issues = append(result.Issues, Issue{
					Severity:    pattern.Severity,
					Category:    pattern.Category,
					Title:       pattern.Name,
					Description: pattern.Description,
					Location:    fmt.Sprintf("%s:line %d", filePath, lineNum),
					Evidence:    []string{line},
					Resolution:  pattern.Resolution,
				})
			}
		}
	}

	return scanner.Err()
}

// getLogPatterns returns common log patterns to match
func (ae *AnalysisEngine) getLogPatterns() []LogPattern {
	return []LogPattern{
		{
			Name:        "OutOfMemory Error",
			Pattern:     regexp.MustCompile(`(?i)(out of memory|oom killed|memory limit exceeded)`),
			Severity:    "critical",
			Category:    "memory",
			Description: "Application killed due to memory limit",
			Resolution:  "Increase memory limits or optimize memory usage",
		},
		{
			Name:        "Connection Refused",
			Pattern:     regexp.MustCompile(`(?i)(connection refused|connection reset)`),
			Severity:    "warning",
			Category:    "network",
			Description: "Network connection issues detected",
			Resolution:  "Check network connectivity and service availability",
		},
		{
			Name:        "DNS Resolution Failure",
			Pattern:     regexp.MustCompile(`(?i)(dns resolution failed|no such host|name resolution)`),
			Severity:    "warning",
			Category:    "network",
			Description: "DNS resolution failures detected",
			Resolution:  "Check DNS configuration and network policies",
		},
		{
			Name:        "Disk Space Error",
			Pattern:     regexp.MustCompile(`(?i)(no space left|disk full|storage full)`),
			Severity:    "critical",
			Category:    "storage",
			Description: "Disk space exhaustion detected",
			Resolution:  "Free up disk space or increase storage capacity",
		},
		{
			Name:        "Permission Denied",
			Pattern:     regexp.MustCompile(`(?i)(permission denied|access denied|unauthorized)`),
			Severity:    "warning",
			Category:    "security",
			Description: "Permission or access issues detected",
			Resolution:  "Check RBAC permissions and security contexts",
		},
	}
}

// getOperatorLogPatterns returns patterns specific to operator logs
func (ae *AnalysisEngine) getOperatorLogPatterns() []LogPattern {
	patterns := ae.getLogPatterns()

	// Add operator-specific patterns
	operatorPatterns := []LogPattern{
		{
			Name:        "Reconcile Error",
			Pattern:     regexp.MustCompile(`(?i)(reconcile.*error|failed to reconcile)`),
			Severity:    "warning",
			Category:    "operator",
			Description: "Operator reconciliation errors",
			Resolution:  "Check operator logs and resource configurations",
		},
		{
			Name:        "Controller Error",
			Pattern:     regexp.MustCompile(`(?i)(controller.*error|controller failed)`),
			Severity:    "warning",
			Category:    "operator",
			Description: "Controller errors detected",
			Resolution:  "Review controller logs and resource states",
		},
	}

	return append(patterns, operatorPatterns...)
}

// Helper functions for different analysis types
func (ae *AnalysisEngine) analyzePcapWithTshark(pcapPath string, result *AnalysisResult) error {
	// This would use tshark to analyze packet captures
	// For now, return error to trigger fallback
	return fmt.Errorf("tshark analysis not implemented yet")
}

func (ae *AnalysisEngine) analyzePcapBasic(pcapPath string, result *AnalysisResult) error {
	// Basic file analysis
	info, err := os.Stat(pcapPath)
	if err != nil {
		return err
	}

	result.Metrics["file_size"] = info.Size()
	result.Metrics["file_modified"] = info.ModTime()

	if info.Size() == 0 {
		result.Issues = append(result.Issues, Issue{
			Severity:    "warning",
			Category:    "capture",
			Title:       "Empty Capture File",
			Description: "Packet capture file is empty",
			Location:    pcapPath,
			Resolution:  "Check tcpdump command and network interfaces",
		})
	}

	return nil
}

func (ae *AnalysisEngine) calculateLogMetrics(result *AnalysisResult) {
	// Calculate metrics based on issues found
	severityCounts := make(map[string]int)
	categoryCounts := make(map[string]int)

	for _, issue := range result.Issues {
		severityCounts[issue.Severity]++
		categoryCounts[issue.Category]++
	}

	result.Metrics["severity_counts"] = severityCounts
	result.Metrics["category_counts"] = categoryCounts
	result.Metrics["total_issues"] = len(result.Issues)
}

func (ae *AnalysisEngine) generateSummaryAndRecommendations(result *AnalysisResult) {
	criticalCount := 0
	warningCount := 0

	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			criticalCount++
		case "warning":
			warningCount++
		}
	}

	// Generate summary
	if criticalCount > 0 {
		result.Summary = fmt.Sprintf("CRITICAL: Found %d critical issues and %d warnings that require immediate attention",
			criticalCount, warningCount)
	} else if warningCount > 0 {
		result.Summary = fmt.Sprintf("WARNING: Found %d warnings that should be addressed", warningCount)
	} else {
		result.Summary = "No critical issues found in the analysis"
	}

	// Generate recommendations
	if criticalCount > 0 {
		result.Recommendations = append(result.Recommendations,
			"Address critical issues immediately to prevent system instability")
	}

	if warningCount > 5 {
		result.Recommendations = append(result.Recommendations,
			"Multiple warnings detected - consider a comprehensive system review")
	}

	// Add specific recommendations based on issue categories
	categories := make(map[string]bool)
	for _, issue := range result.Issues {
		categories[issue.Category] = true
	}

	for category := range categories {
		switch category {
		case "memory":
			result.Recommendations = append(result.Recommendations,
				"Review memory usage patterns and consider resource optimization")
		case "network":
			result.Recommendations = append(result.Recommendations,
				"Investigate network connectivity and DNS configuration")
		case "storage":
			result.Recommendations = append(result.Recommendations,
				"Monitor disk usage and implement storage management policies")
		}
	}
}
