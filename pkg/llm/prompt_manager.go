package llm

import (
	"fmt"
	"strings"
)

// PromptManager handles specialized prompt generation for OpenShift SRE tasks
type PromptManager struct {
	knowledgeBase     *OpenShiftKnowledgeBase
	knowledgeInjector *KnowledgeInjector
}

// NewPromptManager creates a new prompt manager with OpenShift knowledge
func NewPromptManager() *PromptManager {
	return &PromptManager{
		knowledgeBase:     NewOpenShiftKnowledgeBase(),
		knowledgeInjector: NewKnowledgeInjector(),
	}
}

// PromptRequest represents a specialized prompt request
type PromptRequest struct {
	Type        string            // troubleshooting, security, incident, performance, capacity
	UserQuery   string            // Original user request
	Context     map[string]string // Additional context (symptoms, logs, environment)
	Severity    string            // For incidents: P1, P2, P3, P4
	Environment string            // production, staging, development
}

// GenerateSpecializedPrompt creates a specialized prompt based on request type
func (pm *PromptManager) GenerateSpecializedPrompt(req *PromptRequest) (string, error) {
	if req.UserQuery == "" {
		return "", fmt.Errorf("user query is required")
	}

	switch strings.ToLower(req.Type) {
	case "troubleshooting":
		return pm.generateTroubleshootingPrompt(req), nil
	case "security":
		return pm.generateSecurityPrompt(req), nil
	case "incident":
		return pm.generateIncidentPrompt(req), nil
	case "performance":
		return pm.generatePerformancePrompt(req), nil
	case "capacity":
		return pm.generateCapacityPrompt(req), nil
	case "config":
		return pm.generateConfigReviewPrompt(req), nil
	case "resource-creation":
		return pm.generateResourceCreationPrompt(req), nil
	case "configuration":
		return pm.generateConfigurationPrompt(req), nil
	default:
		// Default to general OpenShift expertise injection
		return pm.knowledgeInjector.InjectOpenShiftKnowledge(req.UserQuery), nil
	}
}

// generateTroubleshootingPrompt creates a troubleshooting-specific prompt
func (pm *PromptManager) generateTroubleshootingPrompt(req *PromptRequest) string {
	symptoms := req.Context["symptoms"]
	logs := req.Context["logs"]
	environment := req.Environment
	if environment == "" {
		environment = "OpenShift 4.x"
	}

	basePrompt := `You are a senior OpenShift SRE with 10+ years of production experience.

CORE EXPERTISE:
- OpenShift 4.x architecture and operations
- Kubernetes container orchestration
- Enterprise security and compliance
- Production incident response
- Capacity planning and performance tuning

TROUBLESHOOTING METHODOLOGY:
1. Quick health assessment: 'oc get nodes', 'oc get co'
2. Focused investigation: logs, events, resource usage
3. Root cause analysis: systematic hypothesis testing
4. Solution implementation: step-by-step remediation
5. Prevention: monitoring, alerting, process improvement`

	contextInfo := fmt.Sprintf(`
TROUBLESHOOTING REQUEST:
Issue: %s
Symptoms: %s
Environment: %s
Error Logs: %s`,
		req.UserQuery,
		symptoms,
		environment,
		logs)

	responseFormat := `
RESPONSE FORMAT:
## 🔍 IMMEDIATE DIAGNOSTIC STEPS
[Specific oc/kubectl commands for initial assessment]

## 🛠️ ROOT CAUSE ANALYSIS  
[Systematic investigation approach with decision trees]

## 💡 SOLUTION STEPS
[Step-by-step remediation with verification commands]

## 🚀 PREVENTION MEASURES
[Monitoring alerts and process improvements]

## 📚 DOCUMENTATION REFERENCES
[Relevant OpenShift docs and best practices]

Provide expert-level OpenShift SRE guidance with specific commands and procedures.`

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
		basePrompt,
		pm.knowledgeBase.CoreConcepts,
		pm.knowledgeBase.TroubleshootingPatterns,
		contextInfo,
		responseFormat)
}

// generateSecurityPrompt creates a security review prompt
func (pm *PromptManager) generateSecurityPrompt(req *PromptRequest) string {
	yamlContent := req.Context["yaml"]
	complianceFramework := req.Context["compliance"]
	environment := req.Environment
	if environment == "" {
		environment = "production"
	}

	basePrompt := `You are a senior OpenShift security expert with deep expertise in:
- Pod Security Standards and Security Context Constraints
- RBAC and access control best practices
- Network security and policies
- Container and image security
- Compliance frameworks (SOC2, PCI-DSS, HIPAA)
- Enterprise security hardening`

	contextInfo := fmt.Sprintf(`
SECURITY REVIEW REQUEST:
Configuration: %s
Compliance Framework: %s
Environment: %s`,
		req.UserQuery,
		complianceFramework,
		environment)

	if yamlContent != "" {
		contextInfo += fmt.Sprintf(`

YAML CONFIGURATION:
%s`, yamlContent)
	}

	responseFormat := `
RESPONSE FORMAT:
## 🔒 SECURITY ASSESSMENT

### 1. Pod Security Analysis
- SecurityContext configuration review
- Privilege escalation checks
- Capability analysis

### 2. RBAC Evaluation
- Permission analysis
- Least privilege compliance
- Service account security

### 3. Network Security
- NetworkPolicy evaluation
- Service exposure analysis
- Traffic flow security

### 4. Resource Security
- Resource limits and quotas
- Storage security
- Secret and ConfigMap handling

## ⚠️ SECURITY FINDINGS
[List issues with severity, impact, and specific fixes]

## ✅ HARDENED CONFIGURATION
[Provide improved, security-hardened YAML]

## 📋 COMPLIANCE CHECKLIST
[Compliance requirements and validation steps]

Focus on production-ready, enterprise security standards.`

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		basePrompt,
		pm.knowledgeBase.SecurityBestPractices,
		contextInfo,
		responseFormat)
}

// generateIncidentPrompt creates an incident response prompt
func (pm *PromptManager) generateIncidentPrompt(req *PromptRequest) string {
	incidentType := req.Context["incident_type"]
	severity := req.Severity
	if severity == "" {
		severity = "P2"
	}
	affectedServices := req.Context["affected_services"]
	currentStatus := req.Context["current_status"]

	basePrompt := `You are an experienced OpenShift SRE handling production incidents.
You follow ITIL incident management practices and focus on:
- Rapid service restoration
- Clear communication
- Systematic investigation
- Evidence preservation
- Process improvement`

	contextInfo := fmt.Sprintf(`
🚨 INCIDENT DETAILS:
Type: %s
Severity: %s
Query: %s
Affected Services: %s
Current Status: %s
Environment: Production OpenShift Cluster`,
		incidentType,
		severity,
		req.UserQuery,
		affectedServices,
		currentStatus)

	responseFormat := `
RESPONSE FORMAT:
## 🚨 IMMEDIATE ACTIONS (0-5 minutes)
1. **Impact Assessment**
   [Commands to assess scope and impact]
2. **Initial Containment**
   [Steps to prevent escalation]
3. **Communication**
   [Stakeholder notification requirements]

## 🔍 INVESTIGATION PHASE (5-30 minutes)
1. **Data Collection**
   [Essential logs, metrics, and evidence]
2. **Root Cause Analysis**
   [Systematic investigation approach]
3. **Hypothesis Testing**
   [Diagnostic commands and validation]

## 🛠️ RESOLUTION PHASE (30+ minutes)
1. **Solution Implementation**
   [Step-by-step remediation]
2. **Service Validation**
   [Verification and testing procedures]
3. **Monitoring**
   [Post-resolution monitoring]

## 📊 COMMUNICATION UPDATES
[Status update templates and escalation criteria]

## 📝 POST-INCIDENT ACTIONS
[Documentation and follow-up requirements]

Prioritize service restoration while preserving incident evidence.`

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		basePrompt,
		pm.knowledgeBase.IncidentResponse,
		contextInfo,
		responseFormat)
}

// generatePerformancePrompt creates a performance analysis prompt
func (pm *PromptManager) generatePerformancePrompt(req *PromptRequest) string {
	currentMetrics := req.Context["current_metrics"]
	performanceIssues := req.Context["performance_issues"]
	workloadType := req.Context["workload_type"]

	basePrompt := `You are an OpenShift performance specialist with expertise in:
- Resource optimization and right-sizing
- Application performance tuning
- Cluster capacity planning
- Monitoring and observability
- Scalability patterns`

	contextInfo := fmt.Sprintf(`
PERFORMANCE ANALYSIS REQUEST:
Query: %s
Current Metrics: %s
Performance Issues: %s
Workload Type: %s
Environment: %s`,
		req.UserQuery,
		currentMetrics,
		performanceIssues,
		workloadType,
		req.Environment)

	responseFormat := `
RESPONSE FORMAT:
## 📊 PERFORMANCE ANALYSIS

### 1. Current State Assessment
- Resource utilization analysis
- Performance bottleneck identification
- Capacity constraint evaluation

### 2. Root Cause Analysis
- Performance issue investigation
- Resource contention analysis
- Application behavior assessment

## 🔧 OPTIMIZATION RECOMMENDATIONS

### 1. Immediate Optimizations
- Quick wins and immediate improvements
- Resource allocation adjustments
- Configuration tuning

### 2. Long-term Improvements
- Architectural optimizations
- Scaling strategy enhancements
- Monitoring improvements

## 📈 IMPLEMENTATION PLAN
- Step-by-step optimization procedures
- Performance validation methods
- Risk mitigation strategies

## 📊 MONITORING SETUP
- Key performance indicators
- Alerting thresholds
- Performance dashboards

Provide specific metrics, thresholds, and actionable recommendations.`

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		basePrompt,
		pm.knowledgeBase.PerformanceTuning,
		contextInfo,
		responseFormat)
}

// generateCapacityPrompt creates a capacity planning prompt
func (pm *PromptManager) generateCapacityPrompt(req *PromptRequest) string {
	currentUsage := req.Context["current_usage"]
	growthProjections := req.Context["growth_projections"]
	budgetConstraints := req.Context["budget_constraints"]

	basePrompt := `You are an OpenShift capacity planning expert specializing in:
- Resource forecasting and planning
- Cost optimization strategies
- Scaling architecture design
- Performance capacity modeling
- Infrastructure right-sizing`

	contextInfo := fmt.Sprintf(`
CAPACITY PLANNING REQUEST:
Query: %s
Current Usage: %s
Growth Projections: %s
Budget Constraints: %s
Environment: %s`,
		req.UserQuery,
		currentUsage,
		growthProjections,
		budgetConstraints,
		req.Environment)

	responseFormat := `
RESPONSE FORMAT:
## 📊 CURRENT CAPACITY ANALYSIS
- Resource utilization assessment
- Performance headroom evaluation
- Constraint identification

## 📈 GROWTH PROJECTION ANALYSIS
- Scaling requirements calculation
- Timeline planning
- Resource demand forecasting

## 🏗️ SCALING RECOMMENDATIONS
1. **Horizontal Scaling**
   - Node addition requirements
   - Multi-zone distribution
   - Load balancing strategies

2. **Vertical Scaling**
   - Node size optimization
   - Resource allocation tuning
   - Performance optimization

## 💰 COST OPTIMIZATION
- Right-sizing recommendations
- Resource efficiency improvements
- Cost-effective scaling strategies

## 🔧 IMPLEMENTATION ROADMAP
- Phased scaling approach
- Risk mitigation strategies
- Performance validation plan

Provide specific numbers, timelines, and actionable capacity plans.`

	return fmt.Sprintf("%s\n\n%s\n\n%s",
		basePrompt,
		contextInfo,
		responseFormat)
}

// generateConfigReviewPrompt creates a configuration review prompt
func (pm *PromptManager) generateConfigReviewPrompt(req *PromptRequest) string {
	configType := req.Context["config_type"]
	yamlContent := req.Context["yaml"]

	basePrompt := `You are an OpenShift configuration expert with expertise in:
- YAML manifest best practices
- Resource configuration optimization
- Security configuration hardening
- Operational excellence patterns
- Configuration validation and testing`

	contextInfo := fmt.Sprintf(`
CONFIGURATION REVIEW REQUEST:
Query: %s
Configuration Type: %s
Environment: %s`,
		req.UserQuery,
		configType,
		req.Environment)

	if yamlContent != "" {
		contextInfo += fmt.Sprintf(`

CONFIGURATION:
%s`, yamlContent)
	}

	responseFormat := `
RESPONSE FORMAT:
## 🔍 CONFIGURATION ANALYSIS
- Configuration structure review
- Best practices compliance
- Resource specification validation

## ⚠️ ISSUES IDENTIFIED
- Configuration problems
- Security vulnerabilities
- Performance concerns
- Operational risks

## ✅ RECOMMENDATIONS
- Configuration improvements
- Security enhancements
- Performance optimizations
- Operational best practices

## 📋 IMPROVED CONFIGURATION
[Provide corrected, optimized YAML configuration]

## 🧪 VALIDATION STEPS
- Testing procedures
- Deployment strategies
- Rollback plans

Focus on production-ready, enterprise-grade configurations.`

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		basePrompt,
		pm.knowledgeBase.CoreConcepts,
		contextInfo,
		responseFormat)
}

// generateResourceCreationPrompt creates prompts for resource creation requests
func (pm *PromptManager) generateResourceCreationPrompt(req *PromptRequest) string {
	resourceType := pm.extractResourceType(req.UserQuery)
	environment := req.Environment
	if environment == "" {
		environment = "cluster"
	}

	basePrompt := `You are a senior OpenShift administrator and SRE with expertise in:
- OpenShift resource creation and management
- YAML manifest generation and best practices
- Resource relationships and dependencies
- Security-first configuration
- Production-ready deployments`

	contextInfo := fmt.Sprintf(`
RESOURCE CREATION REQUEST:
Task: %s
Resource Type: %s
Environment: %s`,
		req.UserQuery,
		resourceType,
		environment)

	responseFormat := `
RESPONSE FORMAT:
## 🛠️ RESOURCE CREATION GUIDE

### 1. Step-by-Step Commands
[Provide exact oc commands in sequence]

### 2. YAML Manifests
[Include complete, production-ready YAML configurations]

### 3. Verification Steps
[Commands to verify successful creation]

### 4. Security Considerations
[Security best practices and compliance notes]

### 5. Common Issues & Troubleshooting
[Potential problems and their solutions]

## 📋 COMPLETE EXAMPLE
[Full working example with all commands and YAML]`

	return fmt.Sprintf(`%s

%s

%s

%s`, pm.knowledgeInjector.InjectOpenShiftKnowledge(""), basePrompt, contextInfo, responseFormat)
}

// generateConfigurationPrompt creates prompts for configuration and RBAC requests
func (pm *PromptManager) generateConfigurationPrompt(req *PromptRequest) string {
	configType := pm.extractConfigurationType(req.UserQuery)
	environment := req.Environment
	if environment == "" {
		environment = "production"
	}

	basePrompt := `You are a senior OpenShift security and RBAC expert with deep knowledge of:
- Role-Based Access Control (RBAC) design and implementation
- Security Context Constraints (SCCs)
- Service Account configuration and security
- Namespace isolation and multi-tenancy
- Least-privilege security principles
- Enterprise compliance and governance`

	contextInfo := fmt.Sprintf(`
CONFIGURATION REQUEST:
Task: %s
Configuration Type: %s
Environment: %s
Security Requirements: Least-privilege, production-ready`,
		req.UserQuery,
		configType,
		environment)

	responseFormat := `
RESPONSE FORMAT:
## 🔐 RBAC CONFIGURATION GUIDE

### 1. Resource Creation Commands
[Exact oc commands with proper syntax]

### 2. RBAC Manifests
[Complete Role, RoleBinding, ServiceAccount YAML]

### 3. Security Validation
[Commands to verify permissions and security]

### 4. Principle of Least Privilege
[Explanation of granted permissions and why]

### 5. Testing & Verification
[How to test the configuration works correctly]

## ✅ COMPLETE WORKING CONFIGURATION
[Full example with all YAML and commands]`

	return fmt.Sprintf(`%s

%s

%s

%s`, pm.knowledgeInjector.InjectOpenShiftKnowledge(""), basePrompt, contextInfo, responseFormat)
}

// Helper methods for resource type detection
func (pm *PromptManager) extractResourceType(query string) string {
	query = strings.ToLower(query)

	resourceTypes := map[string]string{
		"namespace":       "Namespace",
		"service account": "ServiceAccount",
		"deployment":      "Deployment",
		"service":         "Service",
		"route":           "Route",
		"configmap":       "ConfigMap",
		"secret":          "Secret",
		"pvc":             "PersistentVolumeClaim",
		"pod":             "Pod",
		"job":             "Job",
		"cronjob":         "CronJob",
	}

	for keyword, resourceType := range resourceTypes {
		if strings.Contains(query, keyword) {
			return resourceType
		}
	}

	return "OpenShift Resource"
}

func (pm *PromptManager) extractConfigurationType(query string) string {
	query = strings.ToLower(query)

	if strings.Contains(query, "rbac") || strings.Contains(query, "role") {
		return "RBAC"
	}
	if strings.Contains(query, "access") || strings.Contains(query, "permission") {
		return "Access Control"
	}
	if strings.Contains(query, "security") {
		return "Security Configuration"
	}
	if strings.Contains(query, "admin") {
		return "Administrative Access"
	}

	return "General Configuration"
}

// InjectGeneralKnowledge injects comprehensive OpenShift knowledge for general queries
func (pm *PromptManager) InjectGeneralKnowledge(userQuery string) string {
	return pm.knowledgeInjector.InjectOpenShiftKnowledge(userQuery)
}

// InjectSpecializedKnowledge injects specialized knowledge based on request type
func (pm *PromptManager) InjectSpecializedKnowledge(req *PromptRequest) string {
	return pm.knowledgeInjector.InjectSpecializedKnowledge(req.UserQuery, req.Type, req.Context)
}
