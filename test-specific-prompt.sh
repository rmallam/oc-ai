#!/bin/bash

# Test the specific prompt that was highlighted by the user

echo "🧪 Testing Specific Network Troubleshooting Prompt"
echo "=================================================="

cd /Users/rakeshkumarmallam/openshift-mcp-go

# Test the specific prompt from the user's selection
echo "Testing prompt: 'packet capture for pod backend-321 host 10.0.0.1'"

# Create a quick test
cat > test_specific_prompt.go << 'EOF'
package main

import (
	"fmt"
	"strings"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/decision"
)

func main() {
	// Create a dummy engine
	engine := &decision.Engine{}
	
	// Create network troubleshooter
	nt := decision.NewNetworkTroubleshooter(engine)

	// Test the specific prompt
	query := "packet capture for pod backend-321 host 10.0.0.1"
	
	fmt.Printf("Query: %s\n", query)
	fmt.Println(strings.Repeat("=", 50))
	
	if nt.IsNetworkTroubleshootingQuery(query) {
		fmt.Println("✅ Detected as network troubleshooting query")
		fmt.Println("   - Should generate tcpdump workflow")
		fmt.Println("   - Should extract pod name: backend-321")
		fmt.Println("   - Should extract host filter: 10.0.0.1")
		fmt.Println("   - Should generate commands for packet capture")
	} else {
		fmt.Println("❌ Not detected as network troubleshooting query")
	}
}
EOF

go run test_specific_prompt.go
rm test_specific_prompt.go

echo ""
echo "✅ Specific prompt test completed!"
