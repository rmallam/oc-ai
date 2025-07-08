# Development Guide - OpenShift MCP Go

## Prerequisites

### Required Software
- **Go 1.21+**: [Install Go](https://golang.org/doc/install)
- **Git**: Version control
- **kubectl/oc**: OpenShift/Kubernetes CLI tools
- **Make**: Build automation (optional but recommended)

### API Keys
- **Gemini API Key**: Get from [Google AI Studio](https://aistudio.google.com)

### Cluster Access
- Access to an OpenShift or Kubernetes cluster
- Appropriate RBAC permissions for resource inspection

## Getting Started

### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/rakeshkumarmallam/openshift-mcp-go.git
cd openshift-mcp-go

# Download dependencies
go mod download

# Verify setup
go mod tidy
```

### 2. Environment Configuration

```bash
# Set required environment variables
export GEMINI_API_KEY="your_gemini_api_key_here"
export KUBECONFIG="path/to/your/kubeconfig"

# Optional: Set custom configuration
export OPENSHIFT_MCP_DEBUG=true
export OPENSHIFT_MCP_PORT=8080
```

### 3. Build and Run

```bash
# Build the application
go build -o bin/openshift-mcp ./cmd/openshift-mcp

# Run directly with go
go run ./cmd/openshift-mcp --debug

# Or run the built binary
./bin/openshift-mcp --port 8080 --debug
```

## Development Workflow

### 1. Code Organization

**Follow these patterns**:
- Put reusable code in `pkg/` packages
- Keep internal-only code in `internal/` packages
- Write tests alongside your code
- Use interfaces for better testability
- Follow Go naming conventions

### 2. Adding New Features

#### Adding a New Plugin

1. **Create the plugin**:
```go
// pkg/plugins/my_plugin.go
type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) CanHandle(prompt string) bool {
    return strings.Contains(strings.ToLower(prompt), "my-keyword")
}

func (p *MyPlugin) Handle(prompt string, context map[string]interface{}) (*decision.Analysis, error) {
    // Your diagnostic logic here
    return &decision.Analysis{...}, nil
}
```

2. **Register the plugin**:
```go
// In pkg/plugins/manager.go InitializeDefaultPlugins function
manager.Register(&MyPlugin{})
```

3. **Test the plugin**:
```go
// test/my_plugin_test.go
func TestMyPlugin_CanHandle(t *testing.T) {
    plugin := &MyPlugin{}
    // Test cases...
}
```

#### Adding a New LLM Provider

1. **Implement the interface**:
```go
// pkg/llm/openai.go
type OpenAIClient struct {
    apiKey string
    model  string
}

func (o *OpenAIClient) GenerateResponse(prompt string) (string, error) {
    // OpenAI API implementation
}

func (o *OpenAIClient) GetAlternativeAnalysis(originalQuery string) (string, error) {
    // Alternative analysis implementation
}
```

2. **Update the factory function**:
```go
// In pkg/llm/gemini.go NewClient function
func NewClient(cfg *config.Config) (Client, error) {
    switch cfg.LLMProvider {
    case "gemini":
        return NewGeminiClient(cfg)
    case "openai":
        return NewOpenAIClient(cfg)
    default:
        return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLMProvider)
    }
}
```

#### Adding New API Endpoints

1. **Define request/response types**:
```go
// In pkg/api/server.go
type NewEndpointRequest struct {
    Field string `json:"field" binding:"required"`
}

type NewEndpointResponse struct {
    Result string `json:"result"`
}
```

2. **Add the handler**:
```go
func (s *Server) handleNewEndpoint(c *gin.Context) {
    var req NewEndpointRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Process request
    response := NewEndpointResponse{Result: "success"}
    c.JSON(http.StatusOK, response)
}
```

3. **Register the route**:
```go
// In setupRoutes method
api.POST("/new-endpoint", s.handleNewEndpoint)
```

### 3. Testing

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/decision

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...
```

#### Writing Tests

**Unit Test Example**:
```go
func TestDecisionEngine_AnalyzePrompt(t *testing.T) {
    cfg := &config.Config{
        ConfidenceThreshold: 0.7,
    }
    
    engine := &Engine{config: cfg}
    
    tests := []struct {
        name     string
        prompt   string
        expected bool
    }{
        {"diagnostic query", "pod is crashlooping", true},
        {"regular query", "list pods", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := engine.isDiagnosticQuery(tt.prompt)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**Integration Test Example**:
```go
func TestAPI_ChatEndpoint(t *testing.T) {
    // Setup test server
    cfg := &config.Config{
        GeminiAPIKey: "test-key",
        Debug:        true,
    }
    
    server, err := api.NewServer(cfg)
    if err != nil {
        t.Fatalf("Failed to create server: %v", err)
    }
    
    // Test request
    reqBody := `{"prompt": "test prompt"}`
    req, _ := http.NewRequest("POST", "/api/v1/chat", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute and validate
    w := httptest.NewRecorder()
    server.engine.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

### 4. Configuration Management

#### Adding New Configuration Options

1. **Add to Config struct**:
```go
// internal/config/config.go
type Config struct {
    // ...existing fields...
    NewField string `mapstructure:"new-field"`
}
```

2. **Add default value**:
```go
func setDefaults(v *viper.Viper) {
    // ...existing defaults...
    v.SetDefault("new-field", "default-value")
}
```

3. **Add command-line flag**:
```go
// cmd/openshift-mcp/main.go
rootCmd.PersistentFlags().String("new-field", "", "description of new field")
```

## Debugging

### 1. Enable Debug Logging

```bash
# Via command line
./openshift-mcp --debug

# Via environment variable
export OPENSHIFT_MCP_DEBUG=true

# Via config file
debug: true
```

### 2. Common Debug Techniques

**Add logging**:
```go
import "github.com/sirupsen/logrus"

logrus.WithFields(logrus.Fields{
    "prompt": prompt,
    "user_id": userID,
}).Debug("Processing chat request")

logrus.WithError(err).Error("Failed to analyze prompt")
```

**Use delve debugger**:
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug with delve
dlv debug ./cmd/openshift-mcp
```

### 3. Testing API Endpoints

**Using curl**:
```bash
# Test chat endpoint
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "why is pod nginx-123 failing?"}'

# Test user choice endpoint
curl -X POST http://localhost:8080/api/v1/user-choice \
  -H "Content-Type: application/json" \
  -d '{
    "choice": "decline",
    "original_query": "pod issue"
  }'

# Test health endpoint
curl http://localhost:8080/health
```

**Using httpie** (if installed):
```bash
# Install httpie
brew install httpie  # macOS
# or pip install httpie

# Test endpoints
http POST localhost:8080/api/v1/chat prompt="pod is crashlooping"
http GET localhost:8080/health
```

## Performance Optimization

### 1. Profiling

```bash
# Build with profiling
go build -o bin/openshift-mcp ./cmd/openshift-mcp

# Run with CPU profiling
./bin/openshift-mcp --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### 2. Memory Management

```go
// Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Close resources properly
defer client.Close()
defer db.Close()

// Use buffered channels for async processing
results := make(chan Result, 100)
```

### 3. Database Optimization

```go
// Batch operations when possible
err := db.Update(func(tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("queries"))
    for _, record := range records {
        data, _ := json.Marshal(record)
        bucket.Put([]byte(record.ID), data)
    }
    return nil
})

// Use read-only transactions for queries
err := db.View(func(tx *bolt.Tx) error {
    // Read operations
    return nil
})
```

## Code Quality

### 1. Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

### 2. Code Formatting

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Vet code for common issues
go vet ./...
```

### 3. Security Scanning

```bash
# Install gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

## Deployment

### 1. Building for Production

```bash
# Build with version info
VERSION=$(git describe --tags --always)
COMMIT=$(git rev-parse HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o bin/openshift-mcp ./cmd/openshift-mcp
```

### 2. Docker Container

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o openshift-mcp ./cmd/openshift-mcp

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/openshift-mcp .
CMD ["./openshift-mcp"]
```

### 3. Kubernetes Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openshift-mcp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openshift-mcp
  template:
    metadata:
      labels:
        app: openshift-mcp
    spec:
      containers:
      - name: openshift-mcp
        image: openshift-mcp:latest
        ports:
        - containerPort: 8080
        env:
        - name: GEMINI_API_KEY
          valueFrom:
            secretKeyRef:
              name: openshift-mcp-secrets
              key: gemini-api-key
```

## Troubleshooting

### Common Issues

1. **"Go not found" error**:
   ```bash
   # Install Go from https://golang.org/doc/install
   # Or using package manager:
   brew install go  # macOS
   sudo apt install golang-go  # Ubuntu
   ```

2. **Module dependency issues**:
   ```bash
   go clean -modcache
   go mod download
   go mod tidy
   ```

3. **Permission denied accessing cluster**:
   ```bash
   # Check kubeconfig
   kubectl cluster-info
   
   # Verify permissions
   kubectl auth can-i get pods
   ```

4. **Database lock errors**:
   ```bash
   # Ensure only one instance is running
   pkill openshift-mcp
   
   # Remove lock file if exists
   rm ~/.config/openshift-mcp/memory.db-lock
   ```

This development guide provides comprehensive information for working with the `openshift-mcp-go` project, from initial setup through advanced development and deployment scenarios.
