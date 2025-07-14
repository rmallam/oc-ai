package main

import (
	"fmt"
	"strings"
)

func main() {
	query := "create a test namespace and a service account called test-sa, this service account should have admin access to test namespace"
	lowerQuery := strings.ToLower(query)

	fmt.Printf("Testing query: %s\n\n", query)

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
		"troubleshoot", "debug", "diagnose", "fix", "analyze",
		"pod failing", "pod not working", "pod crash", "pod error",
		"pod stuck", "pod pending", "pod evicted", "pod terminating",
		"crashloopbackoff", "imagepullbackoff", "oomkilled",
		"pod status", "pod health", "pod issues", "pod problems",
		"why is pod", "whats wrong with", "check pod", "examine pod",
		"container creating", "containercreating", "stuck in container",
		"creating container", "pod initializing", "init container",
		"pulling image", "waiting for", "pod not starting",
	}

	fmt.Println("Checking network keywords:")
	networkMatched := false
	for _, keyword := range networkKeywords {
		if strings.Contains(lowerQuery, keyword) {
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
		if strings.Contains(lowerQuery, keyword) {
			fmt.Printf("  MATCH: %s\n", keyword)
			podMatched = true
		}
	}
	if !podMatched {
		fmt.Println("  No pod troubleshooting keywords matched")
	}

	fmt.Println("\nChecking pod + context condition:")
	if strings.Contains(lowerQuery, "pod") &&
		(strings.Contains(lowerQuery, "troubleshoot") ||
			strings.Contains(lowerQuery, "debug") ||
			strings.Contains(lowerQuery, "diagnose") ||
			strings.Contains(lowerQuery, "analyze") ||
			strings.Contains(lowerQuery, "check") ||
			strings.Contains(lowerQuery, "examine")) {
		fmt.Println("  MATCH: pod + troubleshooting context")
	} else {
		fmt.Println("  No pod + troubleshooting context match")
	}

	// Overall result
	isNetworkQuery := networkMatched || podMatched || (strings.Contains(lowerQuery, "pod") &&
		(strings.Contains(lowerQuery, "troubleshoot") ||
			strings.Contains(lowerQuery, "debug") ||
			strings.Contains(lowerQuery, "diagnose") ||
			strings.Contains(lowerQuery, "analyze") ||
			strings.Contains(lowerQuery, "check") ||
			strings.Contains(lowerQuery, "examine")))

	fmt.Printf("\nFinal result: IsNetworkQuery = %t\n", isNetworkQuery)
}
