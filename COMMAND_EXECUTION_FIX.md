# OpenShift MCP Go - Command Execution Fix

## What Was Fixed

The original issue was that the OpenShift MCP Go application was returning explanations about kubectl commands instead of actually executing them and returning cluster data.

## Changes Made

### 1. Enhanced Command Generation (pkg/decision/engine.go)
- Improved the LLM system prompt to generate simpler, more reliable commands
- Added explicit instructions to avoid complex go-templates and advanced formatting
- Enhanced error handling for command generation

### 2. Added Fallback Command System
- Created `getFallbackCommand()` method with predefined commands for common queries
- Provides reliable fallback when LLM generates complex/invalid commands
- Covers common operations like listing namespaces, pods, nodes, etc.

### 3. Better Command Execution Flow
- Added proper error handling and retry logic
- Improved command validation and security checks
- Enhanced logging for debugging issues

### 4. Updated Configuration
- Changed default model to `gemini-2.0-flash-001` (working model)
- Added environment variable support for `GEMINI_MODEL`
- Updated all documentation to reflect correct model names

## How It Works Now

1. **User Query**: "get all namespaces"
2. **LLM Generation**: System prompts LLM to generate: `kubectl get namespaces`
3. **Fallback Check**: If LLM command fails, tries fallback: `kubectl get namespaces`
4. **Execution**: Runs the command against the cluster
5. **Response**: Returns actual cluster data (namespace list)

## Test Commands

```bash
# Test basic operations
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "get all namespaces"}'

# Test pod filtering (IMPORTANT: Should only show problematic pods)
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "show me crashing pods in the cluster"}'

# Test failed pods
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "show failed pods"}'

# Test node listing
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "list nodes"}'
```

## Expected Response

Instead of explanations, you should now get actual cluster data:

```json
{
  "response": "NAME               STATUS   AGE\ndefault            Active   123d\nkube-system        Active   123d\n...",
  "analysis": {
    "query": "get all namespaces",
    "confidence": 0.9,
    "severity": "Low",
    "metadata": {
      "command": "kubectl get namespaces",
      "execution_type": "command"
    }
  }
}
```

## Next Steps

1. Set your GEMINI_API_KEY environment variable
2. Start the server: `./bin/openshift-mcp`
3. Test with the provided curl commands
4. Verify you get actual cluster data instead of explanations

The application now behaves like your Python version - it generates commands via LLM and then executes them to return real cluster data.
