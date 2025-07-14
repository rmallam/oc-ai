#!/bin/bash

# Test script for the OpenShift MCP Go application

echo "Testing OpenShift MCP Go Application"
echo "====================================="

# Make sure the API key is set
if [ -z "$GEMINI_API_KEY" ]; then
    echo "ERROR: Please set GEMINI_API_KEY environment variable"
    exit 1
fi

echo "✓ GEMINI_API_KEY is set"

# Test 1: Simple namespace listing
echo ""
echo "Test 1: List all namespaces"
echo "curl -X POST http://localhost:8080/api/v1/chat -H 'Content-Type: application/json' -d '{\"prompt\": \"get all namespaces\"}'"
echo ""

# Test 2: List all pods
echo "Test 2: List all pods"
echo "curl -X POST http://localhost:8080/api/v1/chat -H 'Content-Type: application/json' -d '{\"prompt\": \"show me all pods\"}'"
echo ""

# Test 3: Check nodes
echo "Test 3: Check nodes"
echo "curl -X POST http://localhost:8080/api/v1/chat -H 'Content-Type: application/json' -d '{\"prompt\": \"list nodes\"}'"
echo ""

# Test 4: Show crashing pods (IMPORTANT TEST)
echo "Test 4: Show crashing pods"
echo "curl -X POST http://localhost:8080/api/v1/chat -H 'Content-Type: application/json' -d '{\"prompt\": \"show me crashing pods in the cluster\"}'"
echo ""

# Test 5: Show failed pods
echo "Test 5: Show failed pods"
echo "curl -X POST http://localhost:8080/api/v1/chat -H 'Content-Type: application/json' -d '{\"prompt\": \"show failed pods\"}'"
echo ""

# Test 6: Show pending pods
echo "Test 6: Show pending pods"
echo "curl -X POST http://localhost:8080/api/v1/chat -H 'Content-Type: application/json' -d '{\"prompt\": \"show pending pods\"}'"
echo ""

echo "Instructions:"
echo "1. Start the server: ./bin/openshift-mcp"
echo "2. In another terminal, run the above curl commands"
echo "3. For crashing pods test, the response should show ONLY pods with issues (CrashLoopBackOff, ImagePullBackOff, etc.)"
echo "4. If no crashing pods exist, you should see an empty result or 'no lines selected' message"
echo ""
echo "For health check: curl http://localhost:8080/health"
