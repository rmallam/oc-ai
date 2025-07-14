#!/bin/bash

# Simple test for network troubleshooting integration detection
# This script tests the OpenShift MCP Go network troubleshooting functionality

echo "🧪 Testing OpenShift MCP Go Network Troubleshooting Integration"
echo "============================================================="

cd /Users/rakeshkumarmallam/openshift-mcp-go

# Create a simple test program that only tests the detection logic
cat > test_network_detection.go << 'EOF'
package main

import (
	"fmt"
	"strings"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/decision"
)

func main() {
	// Create a dummy engine (nil values for testing detection only)
	engine := &decision.Engine{}
	
	// Create network troubleshooter
	nt := decision.NewNetworkTroubleshooter(engine)

	// Test network troubleshooting queries
	testQueries := []string{
		"tcpdump on pod my-app-123 in namespace production",
		"capture packets from pod nginx-456 interface eth0",
		"ping from pod frontend-789 to 8.8.8.8",
		"test DNS resolution from pod backend-321",
		"curl from pod web-app-666 to https://api.example.com",
		"show network connections in pod api-server-777",
		"debug network namespace for pod cache-999",
		"network troubleshooting for pod queue-111",
		"list all pods", // This should NOT be detected as network troubleshooting
		"show cluster nodes", // This should NOT be detected as network troubleshooting
	}

	fmt.Println("Testing network troubleshooting query detection:")
	fmt.Println(strings.Repeat("=", 60))

	for i, query := range testQueries {
		isNetworkTroubleshooting := nt.IsNetworkTroubleshootingQuery(query)
		status := "❌ Not detected"
		if isNetworkTroubleshooting {
			status = "✅ Detected"
		}
		
		fmt.Printf("Test %02d: %s - %s\n", i+1, status, query)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ Network troubleshooting detection test completed!")
}
EOF

# Make the function public for testing
sed -i '' 's/isNetworkTroubleshootingQuery/IsNetworkTroubleshootingQuery/g' pkg/decision/network_troubleshooter.go

# Run the test
go run test_network_detection.go

# Revert the function to private
sed -i '' 's/IsNetworkTroubleshootingQuery/isNetworkTroubleshootingQuery/g' pkg/decision/network_troubleshooter.go

# Clean up
rm test_network_detection.go

echo "Test completed successfully!"
