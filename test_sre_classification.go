package main

import (
	"fmt"
	"strings"
)

func main() {
	query := "create a namespace testing and create a service account test-sa in testing namespace"
	input := strings.ToLower(query)

	fmt.Printf("Testing query: %s\n", query)
	fmt.Printf("Lowercase: %s\n\n", input)

	// Configuration and RBAC keywords
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

	fmt.Println("Testing configuration keywords:")
	configMatched := false
	for _, keyword := range configurationKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("  MATCH: %s\n", keyword)
			configMatched = true
		}
	}
	if !configMatched {
		fmt.Println("  No configuration keywords matched")
	}

	fmt.Println("\nTesting resource creation keywords:")
	resourceMatched := false
	for _, keyword := range resourceCreationKeywords {
		if strings.Contains(input, keyword) {
			fmt.Printf("  MATCH: %s\n", keyword)
			resourceMatched = true
		}
	}
	if !resourceMatched {
		fmt.Println("  No resource creation keywords matched")
	}

	// Determine classification based on priority
	classification := "general"
	if configMatched {
		classification = "configuration"
	} else if resourceMatched {
		classification = "resource-creation"
	}

	fmt.Printf("\nFinal classification: %s\n", classification)
}
