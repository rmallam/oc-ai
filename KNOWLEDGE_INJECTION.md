# OpenShift MCP Go - Knowledge Injection System

## Overview

This Go implementation provides the same advanced knowledge injection strategies as the Python version, transforming generic Gemini into an OpenShift expert through comprehensive context injection and specialized prompt engineering.

## Key Features

### 🧠 Comprehensive Knowledge Injection
- **Core OpenShift Concepts**: Pods, networking, storage, security, cluster operators
- **Troubleshooting Methodologies**: Systematic approaches for common issues
- **Command Reference**: Essential `oc` commands for SRE operations
- **Domain-Specific Patterns**: Security, performance, incident response expertise

### 🎯 Specialized Response Modes
- **Troubleshooting**: CrashLoopBackOff, ImagePullBackOff, networking issues
- **Security**: RBAC review, SCC analysis, compliance frameworks
- **Incident Response**: P1-P4 classification, emergency procedures
- **Performance**: Resource analysis, bottleneck identification, scaling guidance
- **Capacity Planning**: Growth projections, resource optimization

### 🔄 Advanced Prompt Engineering
- **Context-Aware Enhancement**: Dynamically adapts based on request type
- **Symptom Analysis**: Incorporates error logs and diagnostic information
- **Environment Awareness**: Production, staging, development considerations
- **Severity Classification**: Appropriate urgency and response protocols

## Architecture

```
User Query → Request Classification → Knowledge Injection → Specialized Prompt → Gemini API → Expert Response
```

### Core Components

1. **KnowledgeInjector** (`pkg/llm/knowledge_injection.go`)
   - Comprehensive OpenShift knowledge base
   - Specialized domain pattern injection
   - Context-aware prompt enhancement

2. **PromptManager** (`pkg/llm/prompt_manager.go`)
   - Request type classification
   - Specialized prompt generation
   - Template management for different scenarios

3. **Enhanced Gemini Client** (`pkg/llm/gemini.go`)
   - Knowledge-enhanced API calls
   - Specialized response methods
   - Alternative analysis capabilities

4. **SRE Assistant** (`pkg/llm/sre_assistant.go`)
   - High-level request orchestration
   - Automatic request classification
   - Context extraction and management

## Usage Examples

### Basic Troubleshooting
```go
client, _ := llm.NewGeminiClient(config)
response, _ := client.GenerateResponse("My pods are in CrashLoopBackOff")
// Returns comprehensive troubleshooting guidance with specific oc commands
```

### Specialized Security Review
```go
yamlConfig := `...` // Your OpenShift YAML configuration
response, _ := client.GenerateSecurityReview(yamlConfig)
// Returns detailed security analysis with compliance recommendations
```

### Critical Incident Response
```go
response, _ := client.GenerateIncidentResponse(
    "API server outage", 
    "P1", 
    "all services affected"
)
// Returns emergency response procedures and recovery steps
```

### Performance Analysis
```go
response, _ := client.GeneratePerformanceAnalysis(
    "CPU: 95%, Memory: 80%", 
    "high latency, timeouts"
)
// Returns performance optimization recommendations
```

## Knowledge Base Content

### Core Concepts
- Pod lifecycle and troubleshooting patterns
- Networking (Services, Routes, Ingress, NetworkPolicies)
- Storage (PV/PVC, StorageClasses, CSI drivers)
- Security (RBAC, SCCs, ServiceAccounts)
- Cluster operators and their health indicators

### Troubleshooting Patterns
- Systematic investigation methodology
- Common issue patterns with solutions
- Command sequences for different scenarios
- Root cause analysis approaches

### Command Reference
- Essential `oc` commands for SRE operations
- Diagnostic and debugging commands
- Performance monitoring commands
- Security and RBAC commands

### Specialized Knowledge
- **Security**: CIS Kubernetes Benchmark compliance, RBAC best practices
- **Performance**: Resource monitoring, bottleneck identification
- **Incident**: Response procedures, communication templates
- **Capacity**: Growth planning, resource optimization

## Comparison with XPRR

| Aspect | XPRR | OpenShift MCP Go |
|--------|------|------------------|
| Model Type | Fine-tuned CodeLlama | Generic Gemini + Knowledge Injection |
| Domain Focus | Code Review Only | Full SRE Spectrum |
| Provider Switching | Manual | Automatic (via configuration) |
| Scalability | Local/Self-hosted | Cloud-native |
| Knowledge Updates | Requires Retraining | Dynamic Context Injection |
| Specialization | Model Fine-tuning | Advanced Prompt Engineering |

## Key Advantages

### 🚀 **No Fine-tuning Required**
Instead of training a specialized model, we inject comprehensive domain knowledge directly into prompts, making any generic model an OpenShift expert.

### 📚 **Comprehensive Knowledge Base**
Covers the full spectrum of OpenShift SRE operations:
- Troubleshooting methodologies
- Security best practices  
- Incident response procedures
- Performance optimization
- Capacity planning

### 🎯 **Context-Aware Specialization**
Automatically adapts responses based on:
- Request type classification
- Severity and environment
- Available context (logs, symptoms, metrics)
- Compliance requirements

### 🔄 **Real-time Knowledge Injection**
- Always up-to-date with latest Gemini capabilities
- Easy to extend with new patterns and knowledge
- No model retraining or deployment overhead
- Cloud-native scalability

### 🛠️ **Practical SRE Focus**
Provides actionable guidance with:
- Specific `oc` commands
- Step-by-step procedures
- Decision trees for complex scenarios
- Real-world troubleshooting patterns

## Testing and Validation

Run the comprehensive test suite:
```bash
go run test_knowledge_injection.go
```

This demonstrates:
- Knowledge injection effectiveness (prompt enhancement)
- Specialized prompt generation for different scenarios
- Request classification accuracy
- Integration with Gemini API (when API key provided)

## Configuration

Set your Gemini API key:
```bash
export GEMINI_API_KEY="your-api-key-here"
```

The system will automatically use the enhanced knowledge injection for all requests.

## Future Enhancements

- Integration with OpenShift cluster APIs for real-time data
- Custom knowledge base updates from cluster observations  
- Machine learning-based request classification improvements
- Advanced context extraction from logs and metrics
- Integration with other LLM providers using the same knowledge base

## Conclusion

This Go implementation demonstrates how comprehensive knowledge injection can transform a generic LLM into a domain expert without fine-tuning. By leveraging advanced prompt engineering and systematic knowledge organization, we achieve specialized performance that rivals or exceeds fine-tuned models while maintaining the benefits of cloud-native scalability and always-current capabilities.

## Bug Fix: Network Troubleshooting Misclassification

### Issue Resolved
Previously, the query "create a namespace called test and create a service account test-sa in that namespace and that SA should have admin access to only that namespace" was incorrectly responding with network troubleshooting content (netstat workflows).

### Root Cause
The issue was in the request classification logic where:
1. **Missing resource creation category**: No specific handling for resource creation vs. configuration requests
2. **Incorrect keyword prioritization**: Security keywords (like "access") were triggering before RBAC/configuration keywords
3. **Insufficient request type coverage**: Only had troubleshooting, security, incident, and performance categories

### Solution Implemented
1. **Added new request categories**:
   - `resource-creation`: For creating OpenShift resources (namespaces, deployments, services)
   - `configuration`: For RBAC, permissions, and access control setup

2. **Enhanced keyword classification**:
   ```go
   // Resource creation keywords
   resourceCreationKeywords := []string{
       "create", "deploy", "apply", "provision", "setup", "install",
       "namespace", "service account", "deployment", "service", "route",
   }
   
   // Configuration and RBAC keywords  
   configurationKeywords := []string{
       "rbac", "role", "rolebinding", "clusterrole", "clusterrolebinding",
       "permission", "access", "policy", "admin access", "configure", "bind",
   }
   ```

3. **Corrected prioritization order**:
   - Incident (highest priority)
   - Configuration/RBAC (high priority for access control)
   - Resource creation (medium priority)  
   - Security reviews (lower priority)
   - Performance, troubleshooting (specialized scenarios)

4. **Added specialized prompt generators**:
   - `generateResourceCreationPrompt()`: Provides step-by-step resource creation with YAML manifests
   - `generateConfigurationPrompt()`: Focuses on RBAC, security, and access control

### Result
The same query now correctly:
✅ **Classifies as**: `configuration` (instead of incorrect network troubleshooting)
✅ **Provides**: Step-by-step RBAC setup with proper oc commands and YAML manifests
✅ **Includes**: Security best practices, least-privilege principles, and verification steps
✅ **Focuses on**: Namespace creation, ServiceAccount setup, and admin role binding

### Testing
```bash
# Test the fix
go run test_knowledge_injection.go

# Expected output shows proper classification and specialized RBAC guidance
```
