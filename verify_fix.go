package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test the exact problematic query
	query := "create a servicce account called test-sa in test namespace and it should have admin access only to test namespace"

	fmt.Printf("🧪 Testing Classification Fix\n")
	fmt.Printf("=============================\n")
	fmt.Printf("Query: %s\n\n", query)

	// Apply the same logic as the updated SRE assistant
	classification := classifyRequestLikeAssistant(query)
	fmt.Printf("✅ Classification Result: %s\n", classification)

	if classification == "configuration" {
		fmt.Printf("🎉 SUCCESS: Query correctly classified as configuration!\n")
		fmt.Printf("📋 Expected Response: RBAC setup guidance with namespace, ServiceAccount, and Role configuration\n")
	} else {
		fmt.Printf("❌ ISSUE: Query still misclassified as: %s\n", classification)
	}
}

func classifyRequestLikeAssistant(input string) string {
	input = strings.ToLower(input)

	// Configuration and RBAC keywords (updated with typo handling)
	configurationKeywords := []string{
		"rbac", "role", "rolebinding", "clusterrole", "clusterrolebinding",
		"permission", "access", "policy", "scc", "security context",
		"admin access", "configure", "bind", "grant", "allow", "authorize",
		"service account", "servicce account", // handle typos
	}

	// Resource creation keywords
	resourceCreationKeywords := []string{
		"create", "deploy", "apply", "provision", "setup", "install",
		"namespace", "service account", "servicce account", "deployment", "service", "route",
		"configmap", "secret", "pvc", "pod", "job", "cronjob",
	}

	// Incident keywords
	incidentKeywords := []string{
		"incident", "outage", "down", "critical", "emergency", "urgent",
		"production issue", "service unavailable", "cluster down",
	}

	// Security review keywords
	securityReviewKeywords := []string{
		"security review", "vulnerability", "compliance", "hardening",
		"scan", "audit", "cve", "security assessment", "penetration test",
	}

	// Performance keywords
	performanceKeywords := []string{
		"performance", "slow", "latency", "cpu", "memory", "optimization",
		"capacity", "scaling", "bottleneck", "resource", "monitoring",
	}

	// Troubleshooting keywords
	troubleshootingKeywords := []string{
		"crashloopbackoff", "imagepullbackoff", "pending", "failed", "error",
		"not working", "troubleshoot", "debug", "investigate", "diagnose",
		"pod stuck", "container restart", "connection refused", "logs", "events",
	}

	fmt.Printf("🔍 Checking keyword matches:\n")

	// Check incident first (highest priority)
	for _, keyword := range incidentKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("   ✅ Incident keyword matched: %s\n", keyword)
			return "incident"
		}
	}

	// Check configuration/RBAC (high priority)
	for _, keyword := range configurationKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("   ✅ Configuration keyword matched: %s\n", keyword)
			return "configuration"
		}
	}

	// Check resource creation
	for _, keyword := range resourceCreationKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("   ✅ Resource creation keyword matched: %s\n", keyword)
			return "resource-creation"
		}
	}

	// Check security reviews
	for _, keyword := range securityReviewKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("   ✅ Security keyword matched: %s\n", keyword)
			return "security"
		}
	}

	// Check performance
	for _, keyword := range performanceKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("   ✅ Performance keyword matched: %s\n", keyword)
			return "performance"
		}
	}

	// Check troubleshooting
	for _, keyword := range troubleshootingKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("   ✅ Troubleshooting keyword matched: %s\n", keyword)
			return "troubleshooting"
		}
	}

	fmt.Printf("   ❌ No specific keywords matched\n")
	return "general"
}
