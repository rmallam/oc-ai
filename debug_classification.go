package main

import (
	"fmt"
	"strings"
)

func main() {
	// Test the exact problematic query
	query := "create a servicce account called test-sa in test namespace and it should have admin access only to test namespace"

	fmt.Printf("Testing Query: %s\n", query)
	fmt.Printf("Lowercased: %s\n", strings.ToLower(query))

	// Test each keyword category
	testKeywordCategories(strings.ToLower(query))
}

func testKeywordCategories(input string) {
	fmt.Println("\n🔍 Keyword Analysis:")

	// Resource creation keywords
	resourceCreationKeywords := []string{
		"create", "deploy", "apply", "provision", "setup", "install",
		"namespace", "service account", "deployment", "service", "route",
		"configmap", "secret", "pvc", "pod", "job", "cronjob",
	}

	// Configuration and RBAC keywords
	configurationKeywords := []string{
		"rbac", "role", "rolebinding", "clusterrole", "clusterrolebinding",
		"permission", "access", "policy", "scc", "security context",
		"admin access", "configure", "bind", "grant", "allow",
	}

	// Troubleshooting keywords
	troubleshootingKeywords := []string{
		"crashloopbackoff", "imagepullbackoff", "pending", "failed", "error",
		"not working", "troubleshoot", "debug", "investigate", "diagnose",
		"pod stuck", "container restart", "connection refused", "logs", "events",
	}

	// Security review keywords
	securityReviewKeywords := []string{
		"security review", "vulnerability", "compliance", "hardening",
		"scan", "audit", "cve", "security assessment", "penetration test",
	}

	// Incident keywords
	incidentKeywords := []string{
		"incident", "outage", "down", "critical", "emergency", "urgent",
		"production issue", "service unavailable", "cluster down",
	}

	// Performance keywords
	performanceKeywords := []string{
		"performance", "slow", "latency", "cpu", "memory", "optimization",
		"capacity", "scaling", "bottleneck", "resource", "monitoring",
	}

	// Test each category
	fmt.Println("\n📋 Category Matches:")

	matchedIncident := testCategoryMatches("incident", input, incidentKeywords)
	matchedConfig := testCategoryMatches("configuration", input, configurationKeywords)
	matchedResource := testCategoryMatches("resource-creation", input, resourceCreationKeywords)
	matchedSecurity := testCategoryMatches("security", input, securityReviewKeywords)
	matchedPerformance := testCategoryMatches("performance", input, performanceKeywords)
	matchedTroubleshooting := testCategoryMatches("troubleshooting", input, troubleshootingKeywords)

	fmt.Println("\n🎯 Classification Result:")
	// Apply the same logic as SRE assistant
	if matchedIncident {
		fmt.Println("✅ Final Classification: incident")
	} else if matchedConfig {
		fmt.Println("✅ Final Classification: configuration")
	} else if matchedResource {
		fmt.Println("✅ Final Classification: resource-creation")
	} else if matchedSecurity {
		fmt.Println("✅ Final Classification: security")
	} else if matchedPerformance {
		fmt.Println("✅ Final Classification: performance")
	} else if matchedTroubleshooting {
		fmt.Println("✅ Final Classification: troubleshooting")
	} else {
		fmt.Println("✅ Final Classification: general")
	}
}

func testCategoryMatches(category, input string, keywords []string) bool {
	fmt.Printf("\n%s keywords: ", category)
	matchedKeywords := []string{}

	for _, keyword := range keywords {
		if strings.Contains(input, keyword) {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}

	if len(matchedKeywords) > 0 {
		fmt.Printf("✅ MATCHED: %v", matchedKeywords)
		return true
	} else {
		fmt.Printf("❌ No matches")
		return false
	}
}
