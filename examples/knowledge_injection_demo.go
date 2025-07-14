package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/llm"
)

// Examples of Knowledge Injection Usage
// This demonstrates how the Go implementation provides the same advanced knowledge injection
// as the Python version, making Gemini an OpenShift expert through comprehensive context.

func main() {
	// Load configuration
	cfg := &config.Config{
		LLMProvider:  "gemini",
		GeminiAPIKey: "your-api-key-here", // Set via environment variable
		Model:        "gemini-1.5-pro",
		Debug:        true,
	}

	// Create enhanced Gemini client with OpenShift knowledge
	client, err := llm.NewGeminiClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Demonstrate different knowledge injection scenarios
	runKnowledgeInjectionExamples(client)
}

func runKnowledgeInjectionExamples(client *llm.GeminiClient) {
	fmt.Println("🚀 OpenShift MCP Go - Knowledge Injection Examples")
	fmt.Println(strings.Repeat("=", 60))

	// Example 1: General OpenShift Query with Knowledge Injection
	fmt.Println("\n📋 Example 1: General OpenShift Knowledge Injection")
	fmt.Println(strings.Repeat("-", 50))

	generalQuery := "My pods are stuck in CrashLoopBackOff state. What should I do?"
	response, err := client.GenerateResponse(generalQuery)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Query: %s\n", generalQuery)
		fmt.Printf("Enhanced Response with Knowledge Injection:\n%s\n", response)
	}
	// Example 2: Specialized Troubleshooting with Context
	fmt.Println("\n🔧 Example 2: Specialized Troubleshooting Response")
	fmt.Println(strings.Repeat("-", 50))

	issue := "Application deployment failing with ImagePullBackOff"
	symptoms := "Pods stuck in Pending state, events show image pull errors"
	logs := "Failed to pull image: rpc error: code = NotFound desc = image not found"

	troubleshootingResponse, err := client.GenerateTroubleshootingResponse(issue, symptoms, logs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Issue: %s\n", issue)
		fmt.Printf("Specialized Troubleshooting Response:\n%s\n", troubleshootingResponse)
	}

	// Example 3: Security Review with Comprehensive Analysis
	fmt.Println("\n🔒 Example 3: Security Review with Domain Expertise")
	fmt.Println(strings.Repeat("-", 50))

	yamlConfig := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: app
    image: nginx:latest
    securityContext:
      runAsUser: 0
      privileged: true
      allowPrivilegeEscalation: true
    ports:
    - containerPort: 80
  serviceAccount: default`

	securityResponse, err := client.GenerateSecurityReview(yamlConfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("YAML Configuration Security Review:\n%s\n", securityResponse)
	}

	// Example 4: Incident Response with Severity Classification
	fmt.Println("\n🚨 Example 4: Critical Incident Response")
	fmt.Println(strings.Repeat("-", 50))

	incidentType := "Complete cluster API server outage"
	severity := "P1"
	affectedServices := "All user-facing applications, monitoring, logging"

	incidentResponse, err := client.GenerateIncidentResponse(incidentType, severity, affectedServices)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Incident: %s\n", incidentType)
		fmt.Printf("Incident Response Guidance:\n%s\n", incidentResponse)
	}

	// Example 5: Performance Analysis with Metrics Context
	fmt.Println("\n📊 Example 5: Performance Analysis")
	fmt.Println(strings.Repeat("-", 50))

	metrics := "CPU usage: 95%, Memory usage: 80%, Pod restart count: 15"
	issues := "High latency, frequent timeouts, OOMKilled events"

	performanceResponse, err := client.GeneratePerformanceAnalysis(metrics, issues)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Performance Analysis:\n%s\n", performanceResponse)
	}

	// Example 6: Capacity Planning
	fmt.Println("\n📈 Example 6: Capacity Planning Guidance")
	fmt.Println(strings.Repeat("-", 50))

	currentUsage := "Nodes: 85% CPU, 70% Memory, Storage: 60% full"
	projectedGrowth := "Expected 30% traffic increase over next 6 months"

	capacityResponse, err := client.GenerateCapacityPlanningGuidance(currentUsage, projectedGrowth)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Capacity Planning:\n%s\n", capacityResponse)
	}
}

// demonstrateKnowledgeInjectionStrategies shows how different strategies work
func demonstrateKnowledgeInjectionStrategies() {
	fmt.Println("\n🧠 Knowledge Injection Strategies")
	fmt.Println(strings.Repeat("=", 60))

	// Strategy 1: Documentation Context Injection
	fmt.Println("\n📚 Strategy 1: Documentation Context Injection")
	fmt.Println("Injects comprehensive OpenShift core concepts, troubleshooting patterns,")
	fmt.Println("and command references directly into the prompt context.")

	// Strategy 2: Specialized Domain Patterns
	fmt.Println("\n🎯 Strategy 2: Specialized Domain Patterns")
	fmt.Println("Uses request classification to inject domain-specific knowledge:")
	fmt.Println("- Security: RBAC, SCCs, compliance frameworks")
	fmt.Println("- Performance: Resource monitoring, bottleneck analysis")
	fmt.Println("- Incident: Response procedures, communication templates")

	// Strategy 3: Context-Aware Enhancement
	fmt.Println("\n🔄 Strategy 3: Context-Aware Enhancement")
	fmt.Println("Dynamically enhances prompts based on:")
	fmt.Println("- Symptoms and error logs")
	fmt.Println("- Environment (production, staging, development)")
	fmt.Println("- Severity and impact classification")

	// Strategy 4: Example-Driven Learning
	fmt.Println("\n💡 Strategy 4: Example-Driven Learning")
	fmt.Println("Includes real-world troubleshooting examples and command patterns")
	fmt.Println("to guide the model toward practical, actionable responses.")
}

// compareWithXPRRApproach shows how our approach compares to XPRR
func compareWithXPRRApproach() {
	fmt.Println("\n🔄 Comparison with XPRR Approach")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("XPRR (X-Pull-Request-Reviewer):")
	fmt.Println("✓ Uses fine-tuned CodeLlama for code review")
	fmt.Println("✓ Manual provider switching (Ollama/Gemini)")
	fmt.Println("✓ Specialized code review prompts")
	fmt.Println("✗ No automatic fallback mechanism")

	fmt.Println("\nOpenShift MCP Go:")
	fmt.Println("✓ Uses generic Gemini with advanced prompt engineering")
	fmt.Println("✓ Comprehensive OpenShift domain knowledge injection")
	fmt.Println("✓ Context-aware specialized prompt generation")
	fmt.Println("✓ Multiple specialized response modes (troubleshooting, security, incident)")
	fmt.Println("✓ Systematic knowledge base with troubleshooting methodologies")
	fmt.Println("✓ Real-time context injection based on request classification")

	fmt.Println("\nKey Advantage:")
	fmt.Println("Instead of fine-tuning a local model, we make a generic cloud model (Gemini)")
	fmt.Println("into an OpenShift expert through comprehensive context injection and")
	fmt.Println("specialized prompt engineering. This provides:")
	fmt.Println("- Better scalability (cloud-based)")
	fmt.Println("- Always up-to-date model capabilities")
	fmt.Println("- Comprehensive domain expertise without training overhead")
}

func init() {
	// Demonstrate the knowledge injection approach
	demonstrateKnowledgeInjectionStrategies()
	compareWithXPRRApproach()
}
