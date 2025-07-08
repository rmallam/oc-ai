# OpenShift MCP Go - Project Context

## Project Overview

**OpenShift MCP Go** is a complete rewrite of the Python-based `openshift-mcp` project, following modern Go conventions and architectural patterns inspired by `kubectl-ai`. It serves as an AI-powered OpenShift SRE assistant providing intelligent cluster management, diagnostics, and automation through a conversational REST API interface.

## Project Structure

```
openshift-mcp-go/
├── cmd/
│   └── openshift-mcp/           # Main application entrypoint
│       └── main.go              # CLI and server startup logic
├── pkg/                         # Public packages (reusable)
│   ├── api/                     # REST API server and handlers
│   │   └── server.go            # Gin-based HTTP server with chat/user-choice endpoints
│   ├── decision/                # Dynamic decision making engine
│   │   └── engine.go            # Core diagnostic analysis, confidence scoring
│   ├── llm/                     # LLM integration layer
│   │   └── gemini.go            # Google Gemini API client implementation
│   ├── memory/                  # Persistent storage layer
│   │   └── store.go             # BoltDB-based query/response/feedback storage
│   ├── plugins/                 # Plugin system for extensibility
│   │   └── manager.go           # Plugin management and default plugins
│   └── utils/                   # Utility functions
│       └── helpers.go           # Resource parsing, validation, formatting
├── internal/                    # Private packages (internal only)
│   └── config/                  # Configuration management
│       └── config.go            # Viper-based config loading from files/env/flags
├── test/                        # Test files
│   ├── main_test.go             # Main function and config tests
│   └── decision_test.go         # Decision engine unit tests
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── README.md                    # Project documentation
```

## Core Architecture

### 1. REST API Layer (`pkg/api/`)

**Purpose**: HTTP server providing RESTful endpoints for chat interactions and user feedback.

**Key Components**:
- **Server**: Gin-based HTTP server with middleware for logging and recovery
- **Endpoints**:
  - `POST /api/v1/chat`: Main conversation endpoint
  - `POST /api/v1/user-choice`: User feedback handling (accept/decline/more_info)
  - `GET /health`: Health check endpoint
  - `GET /`: Web UI (future enhancement)

**Request/Response Models**:
```go
type ChatRequest struct {
    Prompt string `json:"prompt" binding:"required"`
}

type ChatResponse struct {
    Response  string                 `json:"response"`
    Analysis  *decision.Analysis     `json:"analysis,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

### 2. Decision Engine (`pkg/decision/`)

**Purpose**: Core intelligence system that analyzes user queries and provides diagnostic insights.

**Key Features**:
- **Diagnostic Detection**: Identifies queries requiring diagnostic analysis vs. general queries
- **Evidence Collection**: Gathers logs, events, pod status, and cluster state
- **Root Cause Analysis**: Pattern matching for common OpenShift issues
- **Confidence Scoring**: Calculates confidence levels (0.0-1.0) for recommendations
- **Severity Assessment**: Categorizes issues as Low/Medium/High/Critical
- **Recommendation Generation**: Provides actionable solutions with priority levels

**Analysis Workflow**:
```
User Query → Diagnostic Detection → Evidence Collection → Root Cause Analysis → 
Confidence Scoring → Severity Assessment → Response Generation
```

**Core Types**:
```go
type Analysis struct {
    Query       string                 `json:"query"`
    Response    string                 `json:"response"`
    Confidence  float64                `json:"confidence"`
    Severity    string                 `json:"severity"`
    RootCauses  []RootCause            `json:"root_causes"`
    Actions     []RecommendedAction    `json:"recommended_actions"`
    Evidence    []Evidence             `json:"evidence"`
    // ... additional fields
}
```

### 3. LLM Integration (`pkg/llm/`)

**Purpose**: Provides AI-powered analysis using various LLM providers.

**Current Implementation**:
- **Gemini Client**: Google Generative AI integration
- **Interface-based Design**: Extensible for multiple LLM providers
- **System Prompting**: OpenShift-specific context and expertise
- **Alternative Analysis**: Provides different perspectives when users decline initial recommendations

**Usage Patterns**:
- Regular queries: Direct LLM interaction
- Diagnostic queries: LLM + rule-based analysis
- Alternative analysis: Creative problem-solving when users want different approaches

### 4. Memory System (`pkg/memory/`)

**Purpose**: Persistent storage for learning and conversation history.

**Storage Backend**: BoltDB (embedded key-value database)
**Data Types**:
- **Queries**: User prompts with metadata
- **Responses**: Analysis results and AI responses
- **Feedback**: User choices (accept/decline/more_info) for learning

**Learning Capabilities**:
- Stores successful analysis patterns
- Tracks user satisfaction through feedback
- Enables future query optimization
- Provides analytics on common issues

### 5. Plugin System (`pkg/plugins/`)

**Purpose**: Extensible architecture for specialized diagnostic handlers.

**Plugin Interface**:
```go
type Plugin interface {
    Name() string
    Description() string
    CanHandle(prompt string) bool
    Handle(prompt string, context map[string]interface{}) (*decision.Analysis, error)
}
```

**Default Plugins**:
- **CrashLoopPlugin**: Specialized crashloop diagnostics
- **NetworkPlugin**: Network connectivity and service issues
- **Custom Plugins**: Extensible for organization-specific needs

### 6. Configuration System (`internal/config/`)

**Purpose**: Flexible configuration management supporting multiple sources.

**Configuration Sources** (in order of precedence):
1. Command-line flags
2. Environment variables
3. Configuration files (YAML)
4. Default values

**Key Configuration Areas**:
- Server settings (host, port, debug)
- LLM configuration (API keys, models, providers)
- Kubernetes settings (kubeconfig path)
- Decision engine tuning (confidence thresholds, evidence limits)
- Storage paths (database location)

## Key Features

### 1. Dynamic Decision Making Engine

**Intelligence Features**:
- **Keyword Detection**: Identifies diagnostic vs. operational queries
- **Resource Extraction**: Parses pod names, namespaces, deployments from natural language
- **Evidence Analysis**: Examines logs, events, status for patterns
- **Confidence Mathematics**: Statistical confidence based on evidence quality and pattern matching
- **Severity Calculation**: Risk assessment based on impact and urgency

**Diagnostic Patterns**:
- Exit code analysis (125, 126, 127)
- Log pattern recognition (import errors, permission issues, port conflicts)
- Event correlation (ImagePullBackOff, CrashLoopBackOff, FailedMount)
- Resource constraint detection
- Security context issues

### 2. User Feedback Loop

**Workflow**:
```
Analysis Presentation → User Choice → Action
├── Accept → Implementation guidance
├── Decline → Alternative AI analysis
└── More Info → Extended diagnostics
```

**Learning Integration**:
- Feedback stored in memory system
- Successful patterns reinforced
- Alternative approaches learned
- User preferences adapted

### 3. RBAC and Security Awareness

**Design Principles**:
- Respects user's Kubernetes permissions
- Uses client-go for cluster interactions
- Security context validation
- Safe command recommendations
- Audit trail for all operations

## Technology Stack

### Core Dependencies

```go
// Web Framework
github.com/gin-gonic/gin v1.9.1

// CLI Framework
github.com/spf13/cobra v1.8.0

// Configuration Management
github.com/spf13/viper v1.18.2

// Database
go.etcd.io/bbolt v1.3.8

// Kubernetes Client
k8s.io/client-go v0.29.0

// Logging
github.com/sirupsen/logrus v1.9.3

// LLM Integration
github.com/google/generativeai-go v0.8.0
```

### Development Tools
- Go 1.21+
- Standard testing framework
- Go modules for dependency management
- Conventional project layout

## Usage Patterns

### 1. Diagnostic Queries

**Examples**:
```
"Why is pod nginx-123 in crashloop?"
"Check why deployment webapp is failing"
"Troubleshoot service connectivity issues"
"Diagnose ImagePullBackOff in production namespace"
```

**System Response**:
- Evidence collection from cluster
- Root cause identification
- Confidence scoring
- Actionable recommendations
- User choice options

### 2. Operational Queries

**Examples**:
```
"List all pods in default namespace"
"Scale deployment to 5 replicas"
"Create a service for my app"
"Show cluster nodes status"
```

**System Response**:
- Direct LLM processing
- Command generation
- Explanation and context
- Safe execution guidance

### 3. Learning and Feedback

**User Interactions**:
- Accept recommendations → Implementation steps provided
- Decline recommendations → Alternative analysis via AI
- Request more info → Extended diagnostic details
- Implicit learning from choices

## Extension Points

### 1. Custom Plugins

**Development Pattern**:
```go
type CustomDiagnosticPlugin struct{}

func (p *CustomDiagnosticPlugin) CanHandle(prompt string) bool {
    return strings.Contains(prompt, "custom-issue")
}

func (p *CustomDiagnosticPlugin) Handle(prompt string, context map[string]interface{}) (*decision.Analysis, error) {
    // Custom logic here
    return analysis, nil
}
```

### 2. Additional LLM Providers

**Interface Implementation**:
```go
type OpenAIClient struct{}

func (o *OpenAIClient) GenerateResponse(prompt string) (string, error) {
    // OpenAI implementation
}

func (o *OpenAIClient) GetAlternativeAnalysis(query string) (string, error) {
    // Alternative analysis implementation
}
```

### 3. Custom Evidence Collectors

**Future Enhancement**:
- Prometheus metrics collection
- Custom log aggregation
- External monitoring integration
- Performance data gathering

## Deployment Considerations

### 1. Environment Setup

**Required**:
- OpenShift/Kubernetes cluster access
- Gemini API key
- Appropriate RBAC permissions

**Optional**:
- Custom configuration file
- Persistent storage for database
- TLS certificates for production

### 2. Security Considerations

**Implementation**:
- API key protection via environment variables
- Cluster permission validation
- Input sanitization for logging
- Secure defaults for all configurations

### 3. Scaling and Performance

**Design Features**:
- Stateless server design (except for memory store)
- Concurrent request handling via Gin
- Efficient BoltDB operations
- Configurable confidence thresholds
- LLM request optimization

## Testing Strategy

### 1. Unit Tests

**Coverage Areas**:
- Configuration loading
- Decision engine logic
- Resource extraction patterns
- Confidence calculations
- Plugin system functionality

### 2. Integration Tests

**Future Implementation**:
- API endpoint testing
- LLM integration validation
- Database operations
- Plugin loading and execution

### 3. End-to-End Tests

**Scenarios**:
- Complete diagnostic workflows
- User feedback loops
- Multi-turn conversations
- Error handling paths

## Future Enhancements

### 1. Advanced Features

**Planned**:
- Multi-resource analysis (deployments + services + ingress)
- Predictive issue detection
- Performance optimization recommendations
- Cost optimization analysis
- Security posture assessment

### 2. Integration Expansions

**Possibilities**:
- Slack/Teams bot integration
- VS Code extension
- kubectl plugin mode
- Grafana dashboard integration
- Webhook notifications

### 3. AI Improvements

**Evolution**:
- Fine-tuned models for OpenShift
- Advanced pattern recognition
- Temporal analysis (issue trends)
- Collaborative filtering based on feedback
- Multi-modal inputs (logs + metrics + events)

## Migration from Python Version

### Key Improvements

**Architecture**:
- Cleaner separation of concerns
- Better testability
- Standard Go project layout
- Interface-driven design

**Performance**:
- Faster startup time
- Lower memory usage
- Better concurrent handling
- Efficient database operations

**Maintainability**:
- Strong typing
- Clear module boundaries
- Comprehensive documentation
- Standard Go tooling

This context file provides a comprehensive overview of the `openshift-mcp-go` project, serving as a reference for development, deployment, and future enhancements.
