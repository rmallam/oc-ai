package main

import (
	"fmt"
	"strings"
)

func main() {
	testClassification()
}

func testClassification() {
	fmt.Println("🧪 Testing Request Classification")
	fmt.Println("================================")

	// Test queries to check classification
	testQueries := []string{
		"create a namespace called test and create a service account test-sa in that namespace and that SA should have admin access to only that namespace",
		"my pods are in CrashLoopBackOff state",
		"review this YAML for security issues",
		"API server is down, cluster unavailable",
		"high CPU usage, pods running slow",
		"create a deployment for nginx",
		"setup RBAC for developer access",
		"configure role binding for service account",
	}

	for i, query := range testQueries {
		// Use reflection to access the private method
		// For testing, we'll just show what classification would happen
		fmt.Printf("\n%d. Query: %s\n", i+1, query)

		// Manually check keywords to demonstrate classification
		classification := classifyTestQuery(query)
		fmt.Printf("   Classification: %s\n", classification)
	}
}

// Simple classification test function
func classifyTestQuery(input string) string {
	input = strings.ToLower(input)

	// Resource creation keywords
	resourceCreationKeywords := []string{
		"create", "deploy", "apply", "provision", "setup", "install",
		"namespace", "service account", "deployment", "service", "route",
	}

	// Configuration and RBAC keywords
	configurationKeywords := []string{
		"rbac", "role", "rolebinding", "clusterrole", "clusterrolebinding",
		"permission", "access", "policy", "admin access", "configure", "bind",
	}

	// Troubleshooting keywords
	troubleshootingKeywords := []string{
		"crashloopbackoff", "imagepullbackoff", "pending", "failed", "error",
		"not working", "troubleshoot", "debug", "investigate", "diagnose",
	}

	// Security review keywords
	securityReviewKeywords := []string{
		"security review", "review", "vulnerability", "security issues",
	}

	// Incident keywords
	incidentKeywords := []string{
		"incident", "outage", "down", "critical", "emergency", "urgent",
		"unavailable", "cluster down", "api server down",
	}

	// Performance keywords
	performanceKeywords := []string{
		"performance", "slow", "latency", "cpu", "memory", "high cpu",
	}

	// Check for matches - order matters
	for _, keyword := range incidentKeywords {
		if strings.Contains(input, keyword) {
			return "incident"
		}
	}

	// Check for configuration/RBAC requests first (higher priority)
	for _, keyword := range configurationKeywords {
		if strings.Contains(input, keyword) {
			return "configuration"
		}
	}

	for _, keyword := range resourceCreationKeywords {
		if strings.Contains(input, keyword) {
			return "resource-creation"
		}
	}

	for _, keyword := range securityReviewKeywords {
		if strings.Contains(input, keyword) {
			return "security"
		}
	}

	for _, keyword := range performanceKeywords {
		if strings.Contains(input, keyword) {
			return "performance"
		}
	}

	for _, keyword := range troubleshootingKeywords {
		if strings.Contains(input, keyword) {
			return "troubleshooting"
		}
	}

	return "general"
}
