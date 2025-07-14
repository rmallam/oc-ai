package main

import (
	"fmt"
	"regexp"
	"strings"
)

type PodInfo struct {
	PodName   string
	Namespace string
	Found     bool
}

func extractPodInfo(query string) PodInfo {
	podInfo := PodInfo{}

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

func isNetworkQuery(query string) bool {
	podTroubleshootingKeywords := []string{
		"troubleshoot", "debug", "diagnose", "fix", "analyze",
		"pod failing", "pod not working", "pod crash", "pod error",
		"pod stuck", "pod pending", "pod evicted", "pod terminating",
		"crashloopbackoff", "imagepullbackoff", "oomkilled",
		"pod status", "pod health", "pod issues", "pod problems",
		"why is pod", "what's wrong with", "check pod", "examine pod",
		"container creating", "containercreating", "stuck in container",
		"creating container", "pod initializing", "init container",
		"pulling image", "waiting for", "pod not starting",
	}

	lowerQuery := strings.ToLower(query)

	// Check for pod troubleshooting keywords
	for _, keyword := range podTroubleshootingKeywords {
		if strings.Contains(lowerQuery, keyword) {
			return true
		}
	}

	// Check for pod mention with troubleshooting context
	if strings.Contains(lowerQuery, "pod") &&
		(strings.Contains(lowerQuery, "troubleshoot") ||
			strings.Contains(lowerQuery, "debug") ||
			strings.Contains(lowerQuery, "diagnose") ||
			strings.Contains(lowerQuery, "analyze") ||
			strings.Contains(lowerQuery, "check") ||
			strings.Contains(lowerQuery, "examine") ||
			strings.Contains(lowerQuery, "why") ||
			strings.Contains(lowerQuery, "stuck")) {
		return true
	}

	return false
}

func main() {
	fmt.Println("🧪 Testing Pod Info Extraction for Container Creating Issue")
	fmt.Println(strings.Repeat("=", 60))

	testQueries := []string{
		"why is httpd pod stuck in container creating",
		"why is httpd pod stuck in container creating in app1 namespace",
		"troubleshoot the httpd pod in app1 namespace",
		"debug nginx pod stuck",
		"check my-app pod in production",
	}

	for _, query := range testQueries {
		fmt.Printf("\nQuery: %s\n", query)

		isNetwork := isNetworkQuery(query)
		fmt.Printf("Is Network Query: %t\n", isNetwork)

		if isNetwork {
			podInfo := extractPodInfo(query)
			fmt.Printf("Extracted Pod Info: Name=%s, Namespace=%s, Found=%t\n",
				podInfo.PodName, podInfo.Namespace, podInfo.Found)
		}
	}

	fmt.Println("\n✅ Testing complete!")
}
