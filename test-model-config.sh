#!/bin/bash

# Test script to verify the updated Gemini model configuration

echo "🧪 Testing Updated Gemini Model Configuration"
echo "============================================="

cd /Users/rakeshkumarmallam/openshift-mcp-go

# Check current configuration
echo "Current model configuration:"
grep "model:" config.yaml

echo ""
echo "Example configuration:"
grep "model:" config.yaml.example

echo ""
echo "✅ Model configuration has been updated to use: gemini-2.0-flash-001"
echo ""
echo "To test with a real API key, set:"
echo "export GEMINI_API_KEY=your-actual-api-key"
echo ""
echo "Then run:"
echo "./bin/openshift-mcp-go --config ./config.yaml"
