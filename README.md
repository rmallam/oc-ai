# OpenShift MCP Go

[![Go Report Card](https://goreportcard.com/badge/github.com/rakeshkumarmallam/openshift-mcp-go)](https://goreportcard.com/report/github.com/rakeshkumarmallam/openshift-mcp-go)
![GitHub License](https://img.shields.io/github/license/rakeshkumarmallam/openshift-mcp-go)

OpenShift MCP Go is an AI-powered OpenShift SRE assistant that provides intelligent cluster management, diagnostics, and automation through a conversational interface.

## Features

- 🤖 **AI-Powered Diagnostics**: Advanced decision engine with confidence scoring and severity assessment
- 🔄 **User Feedback Loop**: Accept/decline/more info workflow for continuous improvement
- 🧩 **Plugin Architecture**: Extensible system for custom diagnostic handlers
- 📊 **Evidence Collection**: Automated gathering of logs, events, and cluster state
- 🎯 **Root Cause Analysis**: Intelligent pattern recognition for common issues
- 💾 **Learning System**: Stores queries and feedback for future improvement
- 🔒 **RBAC Aware**: Respects OpenShift permissions and security policies

## Quick Start

### Installation

#### Build from Source

```bash
git clone https://github.com/rakeshkumarmallam/openshift-mcp-go.git
cd openshift-mcp-go
go mod download
go build -o openshift-mcp ./cmd/openshift-mcp
```

#### Container Installation (Podman)

```bash
# Build with Podman using Red Hat UBI
podman build -t openshift-mcp:latest .

# Or use the provided build script
./build-container.sh --tag latest

# Run with Podman
podman run --rm -p 8080:8080 \
  -e GEMINI_API_KEY=your_key \
  openshift-mcp:latest
```

#### Install Binary

```bash
sudo mv openshift-mcp /usr/local/bin/
```

### Usage

First, ensure you have access to an OpenShift cluster and set your Gemini API key:

```bash
export GEMINI_API_KEY=your_api_key_here
```

#### Start the Server

```bash
openshift-mcp --port 8080
```

#### Configuration

Create a configuration file at `~/.config/openshift-mcp/config.yaml`:

```yaml
# LLM configuration
gemini-api-key: "your_api_key_here"
model: "gemini-2.5-pro-preview-06-05"
llm-provider: "gemini"

# Server configuration
host: "0.0.0.0"
port: "8080"
debug: false

# Kubernetes configuration
kubeconfig: "~/.kube/config"

# Decision engine configuration
confidence-threshold: 0.7
evidence-limit: 10

# Database configuration
database-path: "~/.config/openshift-mcp/memory.db"
```

#### API Usage

##### Chat Endpoint

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Why is pod openshift-mcp-ai-xyz in crashloop?"}'
```

##### User Choice Endpoint

```bash
curl -X POST http://localhost:8080/api/v1/user-choice \
  -H "Content-Type: application/json" \
  -d '{
    "choice": "decline",
    "original_query": "Why is pod openshift-mcp-ai-xyz in crashloop?"
  }'
```

## Example Interactions

### Crashloop Diagnosis

**Request:**
```json
{
  "prompt": "check why pod nginx-deployment-abc123 is in crashloop in default namespace"
}
```

**Response:**
```json
{
  "response": "🔍 **Diagnostic Analysis**\n\n🔴 **Severity:** High\n📊 **Confidence:** 85%\n\n## Root Causes Identified:\n1. Missing Python module dependency (Confidence: 90%)\n2. Application failing to start properly (Confidence: 80%)\n\n## Evidence Found:\n• **pod_status:** Pod is in CrashLoopBackOff state\n• **logs:** Error: No module named 'uvicorn'\n\n## Recommended Solutions:\n### HIGH Priority Action 1:\nInstall missing Python dependencies\n```\npip install <missing_module>\n```\n\n## What would you like to do?\n- **Accept Analysis** ✅ → Get implementation guidance\n- **Get Alternative Analysis** 🤖 → AI-powered different perspective\n- **Get More Details** 📊 → Extended diagnostic information",
  "analysis": {
    "query": "check why pod nginx-deployment-abc123 is in crashloop in default namespace",
    "confidence": 0.85,
    "severity": "High",
    "root_causes": [...],
    "recommended_actions": [...],
    "evidence": [...]
  },
  "timestamp": "2025-07-04T10:30:00Z"
}
```

### User Feedback

**Request:**
```json
{
  "choice": "decline",
  "original_query": "check why pod nginx-deployment-abc123 is in crashloop"
}
```

**Response:**
```json
{
  "response": "🤖 **Alternative AI Analysis:**\n\nInstead of focusing on missing modules, consider this systematic approach:\n\n## 🎯 Alternative Root Cause Analysis:\n1. Dependency management issue - the application environment may need restructuring\n2. Container security context mismatch - OpenShift security policies may be restrictive\n\n## Alternative Solution Strategy:\n1. 📊 Baseline Assessment - Compare with working deployments\n2. 🔍 Pattern Correlation - Check cluster-wide issues\n3. 📈 Resource Impact Analysis - Analyze consumption patterns\n4. 🛡️ Security Context Validation - Review OpenShift SCCs",
  "timestamp": "2025-07-04T10:31:00Z"
}
```

## Container Deployment

### Using Podman (Recommended for OpenShift)

```bash
# Build the container image
./build-container.sh --tag v1.0.0

# Run with Podman
podman run --rm -p 8080:8080 \
  -e GEMINI_API_KEY=your_api_key \
  -e KUBECONFIG=/tmp/kubeconfig \
  -v ~/.kube/config:/tmp/kubeconfig:ro \
  quay.io/openshift-community/openshift-mcp:v1.0.0

# Deploy to OpenShift
oc new-app quay.io/openshift-community/openshift-mcp:v1.0.0
oc expose service openshift-mcp --port=8080
```

### Container Features

- **Red Hat UBI Base**: Built on Red Hat Universal Base Images for enterprise compatibility
- **Security**: Runs as non-root user (UID 1001) with minimal privileges
- **OpenShift Ready**: Compatible with OpenShift security context constraints
- **Health Checks**: Built-in health monitoring endpoint
- **Multi-arch**: Supports AMD64 and ARM64 platforms

| Flag | Environment Variable | Config File | Default | Description |
|------|---------------------|-------------|---------|-------------|
| `--port` | `OPENSHIFT_MCP_PORT` | `port` | `8080` | Server port |
| `--host` | `OPENSHIFT_MCP_HOST` | `host` | `0.0.0.0` | Server host |
| `--debug` | `OPENSHIFT_MCP_DEBUG` | `debug` | `false` | Enable debug logging |
| `--gemini-api-key` | `GEMINI_API_KEY` | `gemini-api-key` | | Gemini API key |
| `--kubeconfig` | `OPENSHIFT_MCP_KUBECONFIG` | `kubeconfig` | `~/.kube/config` | Kubeconfig path |

## Plugin Development

Create custom diagnostic plugins by implementing the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Description() string
    CanHandle(prompt string) bool
    Handle(prompt string, context map[string]interface{}) (*decision.Analysis, error)
}
```

Example plugin:

```go
type CustomPlugin struct{}

func (p *CustomPlugin) Name() string {
    return "custom-handler"
}

func (p *CustomPlugin) CanHandle(prompt string) bool {
    return strings.Contains(strings.ToLower(prompt), "custom-issue")
}

func (p *CustomPlugin) Handle(prompt string, context map[string]interface{}) (*decision.Analysis, error) {
    // Custom diagnostic logic
    return &decision.Analysis{...}, nil
}
```

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │ Decision Engine │    │   LLM Client    │
│                 │────│                 │────│                 │
│ - Chat          │    │ - Evidence      │    │ - Gemini        │
│ - User Choice   │    │ - Root Cause    │    │ - Alternative   │
└─────────────────┘    │ - Confidence    │    │   Analysis      │
                       └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐    ┌─────────────────┐
                       │ Plugin System   │    │ Memory Store    │
                       │                 │    │                 │
                       │ - CrashLoop     │    │ - Queries       │
                       │ - Network       │    │ - Responses     │
                       │ - Custom        │    │ - Feedback      │
                       └─────────────────┘    └─────────────────┘
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Testing

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

*Built with ❤️ for the OpenShift community*
