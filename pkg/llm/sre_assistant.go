package llm

import (
	"fmt"
	"strings"
)

// SREAssistant handles high-level SRE operations using the enhanced LLM client
type SREAssistant struct {
	client Client
}

// NewSREAssistant creates a new SRE assistant with enhanced capabilities
func NewSREAssistant(client Client) *SREAssistant {
	return &SREAssistant{
		client: client,
	}
}

// AnalyzeIssue provides comprehensive issue analysis based on user input
func (sre *SREAssistant) AnalyzeIssue(userInput string) (string, error) { // Determine the type of request based on keywords
	requestType := sre.classifyRequest(userInput)

	switch requestType {
	case "troubleshooting":
		return sre.handleTroubleshootingRequest(userInput)
	case "security":
		return sre.handleSecurityRequest(userInput)
	case "incident":
		return sre.handleIncidentRequest(userInput)
	case "performance":
		return sre.handlePerformanceRequest(userInput)
	case "resource-creation":
		return sre.handleResourceCreationRequest(userInput)
	case "configuration":
		return sre.handleConfigurationRequest(userInput)
	default:
		// Use general OpenShift knowledge injection
		return sre.client.GenerateResponse(userInput)
	}
}

// classifyRequest determines the type of SRE request based on keywords
func (sre *SREAssistant) classifyRequest(input string) string {
	input = strings.ToLower(input)
	// Troubleshooting keywords
	troubleshootingKeywords := []string{
		"crashloopbackoff", "imagepullbackoff", "pending", "failed", "error",
		"not working", "troubleshoot", "debug", "investigate", "diagnose",
		"pod stuck", "container restart", "connection refused", "logs", "events",
	}

	// Resource creation keywords
	resourceCreationKeywords := []string{
		"create", "deploy", "apply", "provision", "setup", "install",
		"namespace", "service account", "servicce account", "deployment", "service", "route",
		"configmap", "secret", "pvc", "pod", "job", "cronjob",
	}

	// Configuration and RBAC keywords
	configurationKeywords := []string{
		"rbac", "role", "rolebinding", "clusterrole", "clusterrolebinding",
		"permission", "access", "policy", "scc", "security context",
		"admin access", "configure", "bind", "grant", "allow", "authorize",
		"service account", "servicce account", // handle typos
	}

	// Security review keywords
	securityReviewKeywords := []string{
		"security review", "vulnerability", "compliance", "hardening",
		"scan", "audit", "cve", "security assessment", "penetration test",
	}

	// Incident keywords
	incidentKeywords := []string{
		"incident", "outage", "down", "critical", "emergency", "urgent",
		"production issue", "service unavailable", "cluster down",
	}

	// Performance keywords
	performanceKeywords := []string{
		"performance", "slow", "latency", "cpu", "memory", "optimization",
		"capacity", "scaling", "bottleneck", "resource", "monitoring",
	}
	// Check for keyword matches - order matters for priority
	for _, keyword := range incidentKeywords {
		if strings.Contains(input, keyword) {
			return "incident"
		}
	}

	// Check for configuration/RBAC requests first (higher priority than general resource creation)
	for _, keyword := range configurationKeywords {
		if strings.Contains(input, keyword) {
			return "configuration"
		}
	}

	// Check for resource creation
	for _, keyword := range resourceCreationKeywords {
		if strings.Contains(input, keyword) {
			return "resource-creation"
		}
	}

	// Check for security reviews
	for _, keyword := range securityReviewKeywords {
		if strings.Contains(input, keyword) {
			return "security"
		}
	}

	for _, keyword := range performanceKeywords {
		if strings.Contains(input, keyword) {
			return "performance"
		}
	}

	for _, keyword := range troubleshootingKeywords {
		if strings.Contains(input, keyword) {
			return "troubleshooting"
		}
	}

	return "general"
}

// ClassifyRequest exposes the classification logic publicly
func (sre *SREAssistant) ClassifyRequest(input string) string {
	return sre.classifyRequest(input)
}

// handleTroubleshootingRequest processes troubleshooting requests
func (sre *SREAssistant) handleTroubleshootingRequest(input string) (string, error) {
	// Extract symptoms and context from the input
	symptoms := sre.extractSymptoms(input)
	logs := sre.extractLogs(input)

	prompt := fmt.Sprintf("You are an expert OpenShift/Kubernetes troubleshooting assistant. Analyze this issue and provide step-by-step troubleshooting guidance.\n\nIssue: %s\nSymptoms: %s\nLogs: %s", input, symptoms, logs)
	return sre.client.GenerateResponse(prompt)
}

// handleSecurityRequest processes security review requests
func (sre *SREAssistant) handleSecurityRequest(input string) (string, error) {
	// Extract YAML content if present
	yamlContent := sre.extractYAMLContent(input)

	if yamlContent != "" {
		prompt := fmt.Sprintf("You are an expert OpenShift/Kubernetes security reviewer. Analyze this YAML configuration for security issues and provide recommendations.\n\nYAML Content:\n%s", yamlContent)
		return sre.client.GenerateResponse(prompt)
	}

	// General security guidance
	req := &PromptRequest{
		Type:      "security",
		UserQuery: input,
		Context:   map[string]string{},
	}

	if client, ok := sre.client.(*GeminiClient); ok {
		return client.GenerateSpecializedResponse(req)
	}

	return sre.client.GenerateResponse(input)
}

// handleIncidentRequest processes incident response requests
func (sre *SREAssistant) handleIncidentRequest(input string) (string, error) {
	severity := sre.extractSeverity(input)
	incidentType := sre.extractIncidentType(input)
	affectedServices := sre.extractAffectedServices(input)

	prompt := fmt.Sprintf("You are an expert OpenShift/Kubernetes incident response coordinator. Provide a structured incident response plan.\n\nIncident Type: %s\nSeverity: %s\nAffected Services: %s\nDetails: %s", incidentType, severity, affectedServices, input)
	return sre.client.GenerateResponse(prompt)
}

// handlePerformanceRequest processes performance analysis requests
func (sre *SREAssistant) handlePerformanceRequest(input string) (string, error) {
	metrics := sre.extractMetrics(input)
	issues := sre.extractPerformanceIssues(input)

	prompt := fmt.Sprintf("You are an expert OpenShift/Kubernetes performance analyst. Analyze these performance issues and provide optimization recommendations.\n\nMetrics: %s\nIssues: %s\nDetails: %s", metrics, issues, input)
	return sre.client.GenerateResponse(prompt)
}

// handleResourceCreationRequest processes resource creation requests
func (sre *SREAssistant) handleResourceCreationRequest(input string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert OpenShift/Kubernetes resource creation specialist.

For the request: "%s"

Provide a solution with:

1. **Individual Commands**: Separate kubectl commands (avoid chaining with &&, ||, ;)
2. **Resource Manifests**: YAML manifests when beneficial
3. **Best Practices**: Follow Kubernetes best practices

Requirements:
- Use individual kubectl commands on separate lines
- No command chaining (&&, ||, ;)
- Include namespace creation if needed
- Provide clear explanations

User Request: %s`, input, input)

	return sre.client.GenerateResponse(prompt)
}

// handleConfigurationRequest processes configuration and RBAC requests
func (sre *SREAssistant) handleConfigurationRequest(input string) (string, error) {
	// For RBAC/configuration requests, provide structured YAML and commands
	prompt := fmt.Sprintf(`You are an expert OpenShift/Kubernetes RBAC and configuration specialist. 

For the request: "%s"

Provide a comprehensive solution that includes:

1. **YAML Manifests**: Complete, ready-to-apply YAML manifests for all required resources
2. **Step-by-step Commands**: Individual kubectl commands (not chained with &&) 
3. **RBAC Explanation**: Clear explanation of permissions and security implications

Format your response as:

## YAML Manifests

`+"```yaml"+`
# All required YAML here
`+"```"+`

## Commands to Apply

`+"```bash"+`
# Individual commands, one per line
kubectl apply -f namespace.yaml
kubectl apply -f serviceaccount.yaml
kubectl apply -f rbac.yaml
`+"```"+`

## Explanation

[Clear explanation of what was created and the security implications]

Requirements:
- Use individual kubectl commands, not chained commands (avoid &&, ||, ;)
- Include all necessary RBAC bindings for the requested access level
- Follow security best practices
- Provide complete, working manifests

User Request: %s`, input, input)

	return sre.client.GenerateResponse(prompt)
}

// Helper methods for extracting information from user input

func (sre *SREAssistant) extractSymptoms(input string) string {
	// Look for common symptom patterns
	symptoms := []string{}

	if strings.Contains(strings.ToLower(input), "crashloopbackoff") {
		symptoms = append(symptoms, "Pod in CrashLoopBackOff state")
	}
	if strings.Contains(strings.ToLower(input), "imagepullbackoff") {
		symptoms = append(symptoms, "Image pull failures")
	}
	if strings.Contains(strings.ToLower(input), "pending") {
		symptoms = append(symptoms, "Pod stuck in Pending state")
	}
	if strings.Contains(strings.ToLower(input), "connection refused") {
		symptoms = append(symptoms, "Connection refused errors")
	}

	return strings.Join(symptoms, ", ")
}

func (sre *SREAssistant) extractLogs(input string) string {
	// Look for log patterns in the input
	// This is a simplified extraction - in practice, you might want more sophisticated parsing
	if strings.Contains(input, "error:") || strings.Contains(input, "Error:") {
		// Try to extract error messages
		lines := strings.Split(input, "\n")
		var logLines []string
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "error") {
				logLines = append(logLines, strings.TrimSpace(line))
			}
		}
		return strings.Join(logLines, "\n")
	}
	return ""
}

func (sre *SREAssistant) extractYAMLContent(input string) string {
	// Look for YAML content blocks
	if strings.Contains(input, "```yaml") {
		start := strings.Index(input, "```yaml")
		end := strings.Index(input[start+7:], "```")
		if end > 0 {
			return input[start+7 : start+7+end]
		}
	}
	return ""
}

func (sre *SREAssistant) extractSeverity(input string) string {
	input = strings.ToLower(input)

	if strings.Contains(input, "critical") || strings.Contains(input, "p1") {
		return "P1"
	}
	if strings.Contains(input, "high") || strings.Contains(input, "p2") {
		return "P2"
	}
	if strings.Contains(input, "medium") || strings.Contains(input, "p3") {
		return "P3"
	}
	if strings.Contains(input, "low") || strings.Contains(input, "p4") {
		return "P4"
	}

	// Default severity based on keywords
	if strings.Contains(input, "outage") || strings.Contains(input, "down") {
		return "P1"
	}

	return "P2"
}

func (sre *SREAssistant) extractIncidentType(input string) string {
	input = strings.ToLower(input)

	if strings.Contains(input, "cluster") && strings.Contains(input, "down") {
		return "Cluster Outage"
	}
	if strings.Contains(input, "api") && (strings.Contains(input, "down") || strings.Contains(input, "unresponsive")) {
		return "API Server Unresponsive"
	}
	if strings.Contains(input, "network") {
		return "Network Connectivity Issue"
	}
	if strings.Contains(input, "storage") {
		return "Storage Issue"
	}
	if strings.Contains(input, "performance") {
		return "Performance Degradation"
	}

	return "General Incident"
}

func (sre *SREAssistant) extractAffectedServices(input string) string {
	// Look for service mentions in the input
	// This is a simplified extraction
	services := []string{}

	if strings.Contains(strings.ToLower(input), "all services") {
		return "All cluster services"
	}
	if strings.Contains(strings.ToLower(input), "api") {
		services = append(services, "API Server")
	}
	if strings.Contains(strings.ToLower(input), "ingress") {
		services = append(services, "Ingress Controller")
	}
	if strings.Contains(strings.ToLower(input), "dns") {
		services = append(services, "DNS")
	}

	if len(services) > 0 {
		return strings.Join(services, ", ")
	}

	return "Unknown"
}

func (sre *SREAssistant) extractMetrics(input string) string {
	// Extract performance metrics if mentioned
	metrics := []string{}

	if strings.Contains(strings.ToLower(input), "cpu") {
		metrics = append(metrics, "CPU utilization")
	}
	if strings.Contains(strings.ToLower(input), "memory") {
		metrics = append(metrics, "Memory usage")
	}
	if strings.Contains(strings.ToLower(input), "disk") {
		metrics = append(metrics, "Disk I/O")
	}
	if strings.Contains(strings.ToLower(input), "network") {
		metrics = append(metrics, "Network throughput")
	}

	return strings.Join(metrics, ", ")
}

func (sre *SREAssistant) extractPerformanceIssues(input string) string {
	// Extract performance issues if mentioned
	issues := []string{}

	if strings.Contains(strings.ToLower(input), "slow") {
		issues = append(issues, "Slow response times")
	}
	if strings.Contains(strings.ToLower(input), "high latency") {
		issues = append(issues, "High latency")
	}
	if strings.Contains(strings.ToLower(input), "timeout") {
		issues = append(issues, "Request timeouts")
	}
	if strings.Contains(strings.ToLower(input), "bottleneck") {
		issues = append(issues, "Performance bottlenecks")
	}

	return strings.Join(issues, ", ")
}

// Example usage functions

// ExampleTroubleshooting demonstrates troubleshooting capabilities
func ExampleTroubleshooting() {
	fmt.Println(`
// Example: Troubleshooting Usage
assistant := NewSREAssistant(geminiClient)

response, err := assistant.AnalyzeIssue("My pod is in CrashLoopBackOff state. The container keeps restarting every 30 seconds.")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
`)
}

// ExampleSecurityReview demonstrates security review capabilities
func ExampleSecurityReview() {
	fmt.Println(`
// Example: Security Review Usage
yamlConfig := """
apiVersion: v1
kind: Pod
metadata:
  name: insecure-pod
spec:
  containers:
  - name: app
    image: nginx
    securityContext:
      runAsUser: 0
      privileged: true
"""

response, err := geminiClient.GenerateSecurityReview(yamlConfig)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
`)
}

// ExampleIncidentResponse demonstrates incident response capabilities
func ExampleIncidentResponse() {
	fmt.Println(`
// Example: Incident Response Usage
response, err := geminiClient.GenerateIncidentResponse(
    "API Server Unresponsive", 
    "P1", 
    "All cluster API operations"
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
`)
}
