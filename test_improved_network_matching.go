package main

import (
	"fmt"
	"strings"
)

func main() {
	query := "create a test namespace and a service account called test-sa, this service account should have admin access to test namespace"
	lowerQuery := strings.ToLower(query)

	fmt.Printf("Testing query: %s\n\n", query)

	// Replicate the improved network detection logic
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

	podTroubleshootingKeywords := []string{
		"pod failing", "pod not working", "pod crash", "pod error",
		"pod stuck", "pod pending", "pod evicted", "pod terminating",
		"crashloopbackoff", "imagepullbackoff", "oomkilled",
		"pod status", "pod health", "pod issues", "pod problems",
		"why is pod", "what's wrong with", "check pod", "examine pod",
		"container creating", "containercreating", "stuck in container",
		"creating container", "pod initializing", "init container",
		"pulling image", "waiting for", "pod not starting",
		"troubleshoot pod", "debug pod", "diagnose pod",
	}

	// Helper function for improved keyword matching
	containsKeyword := func(text, keyword string) bool {
		if strings.Contains(keyword, " ") {
			return strings.Contains(text, keyword)
		}

		words := strings.Fields(text)
		for _, word := range words {
			cleanWord := strings.Trim(word, ".,!?;:-")
			if cleanWord == keyword {
				return true
			}
		}
		return false
	}

	fmt.Println("Checking network keywords:")
	networkMatched := false
	for _, keyword := range networkKeywords {
		if containsKeyword(lowerQuery, keyword) {
			fmt.Printf("  MATCH: %s\n", keyword)
			networkMatched = true
		}
	}
	if !networkMatched {
		fmt.Println("  No network keywords matched")
	}

	fmt.Println("\nChecking pod troubleshooting keywords:")
	podMatched := false
	for _, keyword := range podTroubleshootingKeywords {
		if containsKeyword(lowerQuery, keyword) {
			fmt.Printf("  MATCH: %s\n", keyword)
			podMatched = true
		}
	}
	if !podMatched {
		fmt.Println("  No pod troubleshooting keywords matched")
	}

	fmt.Println("\nChecking pod + specific troubleshooting context:")
	podContextMatch := strings.Contains(lowerQuery, "pod") &&
		(strings.Contains(lowerQuery, "failing") ||
			strings.Contains(lowerQuery, "not working") ||
			strings.Contains(lowerQuery, "crash") ||
			strings.Contains(lowerQuery, "error") ||
			strings.Contains(lowerQuery, "stuck") ||
			strings.Contains(lowerQuery, "pending") ||
			strings.Contains(lowerQuery, "what's wrong") ||
			strings.Contains(lowerQuery, "why is") ||
			strings.Contains(lowerQuery, "troubleshoot pod") ||
			strings.Contains(lowerQuery, "debug pod"))

	if podContextMatch {
		fmt.Println("  MATCH: pod + specific troubleshooting context")
	} else {
		fmt.Println("  No pod + troubleshooting context match")
	}

	// Overall result
	isNetworkQuery := networkMatched || podMatched || podContextMatch

	fmt.Printf("\nFinal result: IsNetworkQuery = %t\n", isNetworkQuery)

	if isNetworkQuery {
		fmt.Println("❌ ERROR: Query incorrectly classified as network troubleshooting!")
	} else {
		fmt.Println("✅ SUCCESS: Query correctly NOT classified as network troubleshooting!")
	}
}
