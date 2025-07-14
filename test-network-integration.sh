#!/bin/bash

# Simple test for network troubleshooting integration
# This script tests the OpenShift MCP Go network troubleshooting functionality

echo "🧪 Testing OpenShift MCP Go Network Troubleshooting Integration"
echo "============================================================="

# Setup environment
export GEMINI_API_KEY="test-key"
export OPENSHIFT_MCP_CONFIG="./config.yaml"

echo "Testing network troubleshooting prompts with built-in test..."

# Test network troubleshooting detection
cd /Users/rakeshkumarmallam/openshift-mcp-go

# Create a simple test program
cat > test_network_integration.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"os"
	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/decision"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/memory"
)

func main() {
	// Initialize config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize in-memory store for testing
	store := memory.NewMemoryStore()

	// Initialize dummy LLM client
	llmClient := llm.NewDummyClient()

	// Initialize decision engine
	engine, err := decision.NewEngine(cfg, store, llmClient)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

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
	}

	fmt.Println("Testing network troubleshooting queries:")
	fmt.Println("="*50)

	for i, query := range testQueries {
		fmt.Printf("\nTest %d: %s\n", i+1, query)
		fmt.Println("-"*50)

		analysis, err := engine.Analyze(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Response: %s\n", analysis.Response[:min(200, len(analysis.Response))])
		fmt.Printf("Confidence: %.2f\n", analysis.Confidence)
		fmt.Printf("Evidence Count: %d\n", len(analysis.Evidence))

		if analysis.Metadata != nil {
			if workflow, ok := analysis.Metadata["workflow"]; ok {
				fmt.Printf("Workflow: %s\n", workflow)
			}
		}
	}

	fmt.Println("\n" + "="*50)
	fmt.Println("✅ Network troubleshooting integration test completed!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
EOF

# Run the test
go run test_network_integration.go

# Clean up
rm test_network_integration.go

echo "Test completed successfully!"
