#!/bin/bash

# Test script for OpenShift MCP Go command execution

echo "🚀 Testing OpenShift MCP Go Command Execution"
echo "=============================================="

# Check if binary exists
if [ ! -f "./bin/openshift-mcp" ]; then
    echo "❌ Binary not found. Please run: go build -o bin/openshift-mcp ./cmd/openshift-mcp"
    exit 1
fi

# Check if GEMINI_API_KEY is set
if [ -z "$GEMINI_API_KEY" ]; then
    echo "❌ GEMINI_API_KEY not set. Please set it first:"
    echo "   export GEMINI_API_KEY=your-api-key"
    exit 1
fi

echo "✅ Binary found"
echo "✅ GEMINI_API_KEY set"

# Start the server in background
echo "🔄 Starting OpenShift MCP server..."
./bin/openshift-mcp &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test command execution
echo "🧪 Testing command execution..."
echo ""

# Test 1: Get namespaces
echo "📋 Test 1: Get namespaces"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "get all namespaces"}' | jq -r '.response' | head -10
echo ""

# Test 2: Get pods
echo "📋 Test 2: Get pods in default namespace"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "get pods in default namespace"}' | jq -r '.response' | head -5
echo ""

# Test 3: Get nodes
echo "📋 Test 3: Get nodes"
curl -s -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "show me the nodes"}' | jq -r '.response' | head -5
echo ""

# Cleanup
echo "🧹 Cleaning up..."
kill $SERVER_PID 2>/dev/null

echo "✅ Testing completed!"
echo ""
echo "💡 The application now executes actual cluster commands instead of just providing explanations!"
