# OpenShift MCP Go

[![Go Report Card](https://goreportcard.com/badge/github.com/rakeshkumarmallam/openshift-mcp-go)](https://goreportcard.com/report/github.com/rakeshkumarmallam/openshift-mcp-go)
![GitHub License](https://img.shields.io/github/license/rakeshkumarmallam/openshift-mcp-go)

OpenShift MCP Go is an AI-powered OpenShift SRE assistant that provides intelligent cluster management, diagnostics, and automation through a conversational interface.

## Quick Start: Running OpenShift MCP Go

### Prerequisites
- Go 1.21 or newer installed
- Access to an OpenShift/Kubernetes cluster (with kubeconfig)
- Gemini API key (for LLM integration)

### 1. Clone the Repository
```sh
git clone <repo-url>
cd openshift-mcp-go
```

### 2. Build the Application
```sh
go build -o bin/openshift-mcp ./cmd/openshift-mcp
```

### 3. Configure Environment
- Copy or create a configuration file if needed (see `internal/config/config.go` for options)
- Set required environment variables:
  - `GEMINI_API_KEY` (for LLM)
  - `KUBECONFIG` (if not using default location)

Example:
```sh
export GEMINI_API_KEY=your-gemini-api-key
export KUBECONFIG=~/.kube/config
```

**Note**: Make sure your `GEMINI_API_KEY` is valid and has access to the Gemini models. You can get an API key from [Google AI Studio](https://makersuite.google.com/app/apikey).

**Model Selection**: The default model is `gemini-2.0-flash-001` which is the latest and most capable. You can override this with:
```sh
export GEMINI_MODEL=gemini-2.0-flash-001  # Latest model (default)
# or
export GEMINI_MODEL=gemini-1.5-pro  # For more stable responses
```

### 4. Run the Server
```sh
./bin/openshift-mcp
```

By default, the server listens on `localhost:8080`.

### 5. Test the API
The application supports three types of queries:

1. **Command Execution** - For operational tasks (returns actual cluster data):
```sh
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "get all namespaces"}'

curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "show me all pods"}'
```

2. **Diagnostic Analysis** - For troubleshooting (returns AI analysis):
```sh
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "Why is my pod crashlooping?"}'
```

3. **Network Troubleshooting** - For advanced network debugging:
```sh
# Tcpdump packet capture
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "tcpdump on pod my-app-123 in namespace production"}'

# Network connectivity testing
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "ping from pod nginx-456 to 8.8.8.8"}'

# DNS resolution testing
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "test DNS resolution from pod backend-321"}'

# HTTP/HTTPS testing
curl -X POST http://localhost:8080/api/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"prompt": "curl from pod web-app-666 to https://api.example.com"}'
```

- Health check:
```sh
curl http://localhost:8080/health
```

### 6. (Optional) Run Tests
```sh
go test ./...
```

## Troubleshooting

### Common Issues

**1. Gemini API Model Not Found Error**
```
Error 404: models/gemini-xxx is not found for API version v1
```
**Solution**: Update the model name in your configuration. Current supported models include:
- `gemini-2.0-flash-001` (recommended, latest and most capable)
- `gemini-1.5-pro` (more capable, slower)
- `gemini-1.5-flash` (fast and efficient)

You can also specify the model via environment variable:
```sh
export GEMINI_MODEL=gemini-2.0-flash-001
```

**2. Missing API Key Error**
```
Gemini API key is required
```
**Solution**: Set your API key:
```sh
export GEMINI_API_KEY=your-actual-api-key
```

**3. Kubeconfig Not Found**
```sh
export KUBECONFIG=/path/to/your/kubeconfig
```

**4. Test Available Models**
If you're unsure which model to use, you can test different models:
```sh
# Test with different models
export GEMINI_MODEL=gemini-1.5-flash
./bin/openshift-mcp

# Or test with pro model
export GEMINI_MODEL=gemini-1.5-pro
./bin/openshift-mcp
```

**5. Debug Mode**
Enable debug logging to see more details:
```sh
export OPENSHIFT_MCP_DEBUG=true
./bin/openshift-mcp --debug
```

---

For more details, see the project context in `PROJECT_CONTEXT.md` or the code comments in each package.

## Features

- 🤖 **AI-Powered Diagnostics**: Advanced decision engine with confidence scoring and severity assessment
- 🔄 **User Feedback Loop**: Accept/decline/more info workflow for continuous improvement
- 🧩 **Plugin Architecture**: Extensible system for custom diagnostic handlers
- 📊 **Evidence Collection**: Automated gathering of logs, events, and cluster state
- 🎯 **Root Cause Analysis**: Intelligent pattern recognition for common issues
- 💾 **Learning System**: Stores queries and feedback for future improvement
- 🔒 **RBAC Aware**: Respects OpenShift permissions and security policies
- ⚡ **Command Execution**: Direct execution of kubectl/oc commands with real cluster data
- 🛡️ **Security Validation**: Safe command execution with built-in security checks
- 🔍 **Network Troubleshooting**: Advanced tcpdump/nsenter workflows for pod network debugging
  - Tcpdump packet capture from pod network namespaces
  - Network connectivity testing (ping, curl, DNS resolution)
  - Network statistics and interface analysis
  - Automated OpenShift 4.8+ and 4.9+ version detection
  - No SSH access required - uses `oc debug node` sessions

## Installation

### Build from Source

```bash
git clone https://github.com/rakeshkumarmallam/openshift-mcp-go.git
cd openshift-mcp-go
go mod download
go build -o openshift-mcp ./cmd/openshift-mcp
```

### Container Installation (Podman)

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

### Install Binary

```bash
sudo mv openshift-mcp /usr/local/bin/
```

## Usage

First, ensure you have access to an OpenShift cluster and set your Gemini API key:

```bash
export GEMINI_API_KEY=your_api_key_here
```

### Start the Server

```bash
openshift-mcp --port 8080
```

### Configuration

Create a configuration file at `~/.config/openshift-mcp/config.yaml`:

```yaml
# LLM configuration
gemini-api-key: "your_api_key_here"
model: "gemini-2.0-flash-001"
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

### API Usage

#### Chat Endpoint

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Why is pod openshift-mcp-ai-xyz in crashloop?"}'
```

#### User Choice Endpoint

```bash
curl -X POST http://localhost:8080/api/v1/user-choice \
  -H "Content-Type: application/json" \
  -d '{
    "choice": "decline",
    "original_query": "Why is pod openshift-mcp-ai-xyz in crashloop?"
  }'
```

## Example Interactions

### Command Execution

**Request:**
```json
{
  "prompt": "get all namespaces"
}
```

**Response:**
```json
{
  "response": "NAME                                               STATUS   AGE\ndefault                                            Active   123d\nkube-node-lease                                    Active   123d\nkube-public                                        Active   123d\nkube-system                                        Active   123d\nopenshift-apiserver                                Active   123d\nopenshift-apiserver-operator                       Active   123d\nopenshift-authentication                           Active   123d\nopenshift-authentication-operator                 Active   123d\nopenshift-cloud-controller-manager                Active   123d\nopenshift-cloud-controller-manager-operator       Active   123d\nopenshift-cluster-machine-approver                Active   123d\nopenshift-cluster-node-tuning-operator            Active   123d\nopenshift-cluster-samples-operator                Active   123d\nopenshift-cluster-storage-operator                Active   123d\nopenshift-cluster-version                          Active   123d\nopenshift-config                                   Active   123d\nopenshift-config-managed                           Active   123d\nopenshift-config-operator                          Active   123d\nopenshift-console                                  Active   123d\nopenshift-console-operator                         Active   123d\nopenshift-console-user-settings                   Active   123d\nopenshift-controller-manager                       Active   123d\nopenshift-controller-manager-operator             Active   123d\nopenshift-dns                                      Active   123d\nopenshift-dns-operator                            Active   123d\nopenshift-etcd                                     Active   123d\nopenshift-etcd-operator                            Active   123d\nopenshift-image-registry                           Active   123d\nopenshift-infra                                    Active   123d\nopenshift-ingress                                  Active   123d\nopenshift-ingress-canary                           Active   123d\nopenshift-ingress-operator                         Active   123d\nopenshift-insights                                 Active   123d\nopenshift-kube-apiserver                           Active   123d\nopenshift-kube-apiserver-operator                 Active   123d\nopenshift-kube-controller-manager                 Active   123d\nopenshift-kube-controller-manager-operator        Active   123d\nopenshift-kube-scheduler                           Active   123d\nopenshift-kube-scheduler-operator                 Active   123d\nopenshift-kube-storage-version-migrator           Active   123d\nopenshift-kube-storage-version-migrator-operator  Active   123d\nopenshift-machine-api                              Active   123d\nopenshift-machine-config-operator                 Active   123d\nopenshift-marketplace                              Active   123d\nopenshift-monitoring                               Active   123d\nopenshift-multus                                   Active   123d\nopenshift-network-diagnostics                     Active   123d\nopenshift-network-operator                         Active   123d\nopenshift-node                                     Active   123d\nopenshift-oauth-apiserver                          Active   123d\nopenshift-operator-lifecycle-manager              Active   123d\nopenshift-operators                                Active   123d\nopenshift-ovn-kubernetes                           Active   123d\nopenshift-service-ca                               Active   123d\nopenshift-service-ca-operator                      Active   123d\nopenshift-user-workload-monitoring                Active   123d",
  "analysis": {
    "query": "get all namespaces",
    "confidence": 0.9,
    "severity": "Low",
    "timestamp": "2025-07-09T10:30:00Z"
  },
  "metadata": {
    "command": "oc get namespaces",
    "execution_type": "command",
    "model": "gemini-2.0-flash-001",
    "provider": "gemini"
  }
}
```

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
