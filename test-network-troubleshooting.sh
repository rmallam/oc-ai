#!/bin/bash

# Test script for network troubleshooting integration
# This script tests the OpenShift MCP Go network troubleshooting functionality

set -e

echo "🧪 Testing OpenShift MCP Go Network Troubleshooting Integration"
echo "============================================================="

# Start the server in the background
echo "Starting OpenShift MCP Go server..."
cd /Users/rakeshkumarmallam/openshift-mcp-go
./bin/openshift-mcp-go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test network troubleshooting prompts
echo ""
echo "Testing network troubleshooting prompts..."

# Test 1: tcpdump request
echo "Test 1: tcpdump on pod my-app-123 in namespace production"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "tcpdump on pod my-app-123 in namespace production"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 2: Packet capture with specific interface
echo "Test 2: capture packets from pod nginx-456 interface eth0"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "capture packets from pod nginx-456 interface eth0"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 3: Network connectivity testing
echo "Test 3: ping from pod frontend-789 to 8.8.8.8"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "ping from pod frontend-789 to 8.8.8.8"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 4: DNS resolution testing
echo "Test 4: test DNS resolution from pod backend-321"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "test DNS resolution from pod backend-321"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 5: HTTP testing
echo "Test 5: curl from pod web-app-666 to https://api.example.com"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "curl from pod web-app-666 to https://api.example.com"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 6: Network statistics
echo "Test 6: show network connections in pod api-server-777"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "show network connections in pod api-server-777"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 7: Advanced network debugging
echo "Test 7: debug network namespace for pod cache-999"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "debug network namespace for pod cache-999"}' | \
  jq -r '.response' | head -10

echo ""
echo "---"

# Test 8: General network troubleshooting
echo "Test 8: network troubleshooting for pod queue-111"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "network troubleshooting for pod queue-111"}' | \
  jq -r '.response' | head -10

echo ""
echo "============================================================="

# Check prompt categorization
echo "Checking prompt categorization..."
curl -s -X GET http://localhost:8080/api/v1/prompts/categories | jq '.'

echo ""
echo "============================================================="

# Clean up
echo "Cleaning up..."
kill $SERVER_PID
sleep 2

echo "✅ Network troubleshooting integration test completed!"
