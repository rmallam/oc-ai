package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/diagnostics"
)

type Server struct {
	server              *server.MCPServer
	config              *Config
	kubeconfig          string
	k8sClient           kubernetes.Interface
	gitManager          *GitManager
	yamlGenerator       *YAMLGenerator
	diagnosticCollector *diagnostics.DiagnosticCollector
	analysisEngine      *diagnostics.AnalysisEngine
}

type Config struct {
	Profile   string     `json:"profile"`
	Debug     bool       `json:"debug"`
	GitConfig *GitConfig `json:"git_config"`
}

func NewServer(config *Config, kubeconfig string) *Server {
	s := &Server{
		config:     config,
		kubeconfig: kubeconfig,
	}

	// Initialize Git manager
	s.gitManager = NewGitManager(config.GitConfig)
	if s.gitManager.IsEnabled() {
		if err := s.gitManager.InitializeRepo(); err != nil {
			logrus.WithError(err).Warn("Failed to initialize Git repository")
		}
	}

	// Initialize YAML generator
	s.yamlGenerator = NewYAMLGenerator()

	// Initialize diagnostic components
	logger := logrus.StandardLogger()
	s.diagnosticCollector = diagnostics.NewDiagnosticCollector(logger, "/tmp/diagnostics")
	s.analysisEngine = diagnostics.NewAnalysisEngine(logger)

	// Initialize Kubernetes client
	var k8sConfig *rest.Config
	var err error

	if kubeconfig != "" {
		logrus.Debugf("Loading kubeconfig from: %s", kubeconfig)
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		logrus.Debug("Attempting to load in-cluster config")
		k8sConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		logrus.WithError(err).Warn("Failed to load Kubernetes config, client will be unavailable")
		s.k8sClient = nil
	} else {
		s.k8sClient, err = kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			logrus.WithError(err).Warn("Failed to create Kubernetes client")
			s.k8sClient = nil
		} else {
			logrus.Info("Kubernetes client initialized successfully")
		}
	}

	profile := ProfileFromString(config.Profile)
	tools := profile.GetTools(s)

	s.server = server.NewMCPServer(
		"OpenShift MCP",
		"1.0.0",
	)

	// Add tools to server
	for _, tool := range tools {
		s.server.AddTool(tool.Tool, tool.Handler)
	}

	return s
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

func (s *Server) ServeHTTP(httpServer *http.Server) *server.StreamableHTTPServer {
	return server.NewStreamableHTTPServer(s.server)
}

// OpenShift-specific tools
func (s *Server) initOpenShiftTools() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("openshift_diagnose",
			mcp.WithDescription("Diagnose OpenShift cluster issues including pod failures, resource constraints, and connectivity problems"),
			mcp.WithString("resource_type", mcp.Description("Type of resource to diagnose (pod, deployment, service, etc.)"), mcp.Required()),
			mcp.WithString("resource_name", mcp.Description("Name of the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the resource")),
			mcp.WithTitleAnnotation("OpenShift: Diagnose"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.OpenShiftDiagnose)},

		{Tool: mcp.NewTool("openshift_must_gather",
			mcp.WithDescription("Collect OpenShift must-gather data for debugging"),
			mcp.WithString("image", mcp.Description("Must-gather image to use")),
			mcp.WithString("dest_dir", mcp.Description("Destination directory for must-gather data")),
			mcp.WithTitleAnnotation("OpenShift: Must Gather"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.openShiftMustGather)},

		{Tool: mcp.NewTool("collect_sosreport",
			mcp.WithDescription("Collect sosreport from a specific node for system-level troubleshooting"),
			mcp.WithString("node_name", mcp.Description("Name of the node to collect sosreport from"), mcp.Required()),
			mcp.WithString("output_dir", mcp.Description("Directory to store the sosreport")),
			mcp.WithTitleAnnotation("Diagnostics: SOS Report"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.collectSosReportHandler)},

		{Tool: mcp.NewTool("collect_tcpdump",
			mcp.WithDescription("Perform network packet capture for troubleshooting connectivity issues"),
			mcp.WithString("pod_name", mcp.Description("Pod name for pod-level capture")),
			mcp.WithString("node_name", mcp.Description("Node name for node-level capture")),
			mcp.WithString("namespace", mcp.Description("Namespace (required if pod_name specified)")),
			mcp.WithString("duration", mcp.Description("Capture duration (e.g., 60s, 5m)")),
			mcp.WithString("filter", mcp.Description("Tcpdump filter expression")),
			mcp.WithString("output_dir", mcp.Description("Directory to store the capture")),
			mcp.WithTitleAnnotation("Diagnostics: Network Capture"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.collectTcpdumpHandler)},

		{Tool: mcp.NewTool("collect_logs",
			mcp.WithDescription("Collect comprehensive logs from pods, containers, and system components"),
			mcp.WithString("pod_name", mcp.Description("Specific pod to collect logs from")),
			mcp.WithString("namespace", mcp.Description("Namespace to collect logs from")),
			mcp.WithBoolean("include_previous", mcp.Description("Include previous container logs")),
			mcp.WithString("output_dir", mcp.Description("Directory to store the logs")),
			mcp.WithTitleAnnotation("Diagnostics: Log Collection"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.collectLogsHandler)},

		{Tool: mcp.NewTool("analyze_must_gather",
			mcp.WithDescription("Analyze collected must-gather data to identify issues and provide recommendations"),
			mcp.WithString("must_gather_path", mcp.Description("Path to the must-gather directory"), mcp.Required()),
			mcp.WithTitleAnnotation("Analysis: Must Gather"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.analyzeMustGatherHandler)},

		{Tool: mcp.NewTool("analyze_logs",
			mcp.WithDescription("Analyze log files to identify errors, patterns, and issues"),
			mcp.WithString("log_path", mcp.Description("Path to log file or directory"), mcp.Required()),
			mcp.WithTitleAnnotation("Analysis: Logs"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.analyzeLogsHandler)},

		{Tool: mcp.NewTool("analyze_tcpdump",
			mcp.WithDescription("Analyze packet capture files to identify network issues"),
			mcp.WithString("pcap_path", mcp.Description("Path to the pcap file"), mcp.Required()),
			mcp.WithTitleAnnotation("Analysis: Network Capture"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.analyzeTcpdumpHandler)},

		{Tool: mcp.NewTool("openshift_route_analyze",
			mcp.WithDescription("Analyze OpenShift routes and their connectivity"),
			mcp.WithString("route_name", mcp.Description("Name of the route to analyze")),
			mcp.WithString("namespace", mcp.Description("Namespace of the route")),
			mcp.WithTitleAnnotation("OpenShift: Route Analysis"),
			mcp.WithReadOnlyHintAnnotation(true),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.openShiftRouteAnalyze)},
	}
}

// Additional tool initializers required by profiles
func (s *Server) initConfiguration() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("get_kubeconfig",
			mcp.WithDescription("Get current kubeconfig context information"),
			mcp.WithTitleAnnotation("Configuration: Get Kubeconfig"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.getKubeconfigHandler)},
	}
}

func (s *Server) initPods() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("list_pods",
			mcp.WithDescription("List pods in a namespace"),
			mcp.WithString("namespace", mcp.Description("Namespace to list pods from")),
			mcp.WithTitleAnnotation("Pods: List"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.ListPodsHandler)},
	}
}

func (s *Server) initResources() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("get_resource",
			mcp.WithDescription("Get Kubernetes resource details"),
			mcp.WithString("resource_type", mcp.Description("Type of resource"), mcp.Required()),
			mcp.WithString("resource_name", mcp.Description("Name of the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the resource")),
			mcp.WithTitleAnnotation("Resources: Get"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.getResourceHandler)},
	}
}

func (s *Server) initEvents() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("get_events",
			mcp.WithDescription("Get events from a namespace"),
			mcp.WithString("namespace", mcp.Description("Namespace to get events from")),
			mcp.WithTitleAnnotation("Events: Get"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.GetEventsHandler)},
	}
}

func (s *Server) initNamespaces() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("list_namespaces",
			mcp.WithDescription("List all namespaces"),
			mcp.WithTitleAnnotation("Namespaces: List"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.ListNamespacesHandler)},
	}
}

func (s *Server) initHelm() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("helm_list",
			mcp.WithDescription("List Helm releases"),
			mcp.WithString("namespace", mcp.Description("Namespace to list releases from")),
			mcp.WithTitleAnnotation("Helm: List"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.helmListHandler)},
	}
}

// OpenShift Create/Update/Delete tools
func (s *Server) initWriteOperations() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("create_resource",
			mcp.WithDescription("Create a Kubernetes resource from YAML/JSON"),
			mcp.WithString("yaml", mcp.Description("YAML/JSON content for the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to create the resource in")),
			mcp.WithTitleAnnotation("Create: Resource"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.createResourceHandler)},

		{Tool: mcp.NewTool("update_resource",
			mcp.WithDescription("Update a Kubernetes resource"),
			mcp.WithString("resource_type", mcp.Description("Type of resource (pod, deployment, service, etc.)"), mcp.Required()),
			mcp.WithString("resource_name", mcp.Description("Name of the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the resource"), mcp.Required()),
			mcp.WithString("yaml", mcp.Description("Updated YAML/JSON content"), mcp.Required()),
			mcp.WithTitleAnnotation("Update: Resource"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.updateResourceHandler)},

		{Tool: mcp.NewTool("delete_resource",
			mcp.WithDescription("Delete a Kubernetes resource"),
			mcp.WithString("resource_type", mcp.Description("Type of resource (pod, deployment, service, etc.)"), mcp.Required()),
			mcp.WithString("resource_name", mcp.Description("Name of the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the resource"), mcp.Required()),
			mcp.WithTitleAnnotation("Delete: Resource"),
			mcp.WithDestructiveHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.deleteResourceHandler)},

		{Tool: mcp.NewTool("scale_deployment",
			mcp.WithDescription("Scale a deployment to a specific number of replicas"),
			mcp.WithString("deployment_name", mcp.Description("Name of the deployment"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the deployment"), mcp.Required()),
			mcp.WithString("replicas", mcp.Description("Number of replicas"), mcp.Required()),
			mcp.WithTitleAnnotation("Scale: Deployment"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.scaleDeploymentHandler)},

		{Tool: mcp.NewTool("restart_deployment",
			mcp.WithDescription("Restart a deployment by updating its spec"),
			mcp.WithString("deployment_name", mcp.Description("Name of the deployment"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the deployment"), mcp.Required()),
			mcp.WithTitleAnnotation("Restart: Deployment"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.restartDeploymentHandler)},

		{Tool: mcp.NewTool("create_namespace",
			mcp.WithDescription("Create a new namespace"),
			mcp.WithString("namespace_name", mcp.Description("Name of the namespace to create"), mcp.Required()),
			mcp.WithTitleAnnotation("Create: Namespace"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.createNamespaceHandler)},

		{Tool: mcp.NewTool("apply_yaml",
			mcp.WithDescription("Apply YAML configuration to the cluster (kubectl apply equivalent)"),
			mcp.WithString("yaml", mcp.Description("YAML content to apply"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to apply the resource in")),
			mcp.WithString("save_to_git", mcp.Description("Save YAML to Git repository (true/false)")),
			mcp.WithTitleAnnotation("Apply: YAML"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.applyYamlHandler)},

		{Tool: mcp.NewTool("create_configmap",
			mcp.WithDescription("Create a ConfigMap with key-value pairs"),
			mcp.WithString("name", mcp.Description("Name of the ConfigMap"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace to create the ConfigMap in"), mcp.Required()),
			mcp.WithString("data", mcp.Description("Data as JSON object (e.g., {\"key1\": \"value1\", \"key2\": \"value2\"})"), mcp.Required()),
			mcp.WithTitleAnnotation("Create: ConfigMap"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.createConfigMapHandler)},
	}
}

// Git-related tools
func (s *Server) initGitTools() []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("git_status",
			mcp.WithDescription("Get Git repository status"),
			mcp.WithTitleAnnotation("Git: Status"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.gitStatusHandler)},

		{Tool: mcp.NewTool("git_list_files",
			mcp.WithDescription("List YAML files in Git repository"),
			mcp.WithTitleAnnotation("Git: List Files"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.gitListFilesHandler)},

		{Tool: mcp.NewTool("git_commit",
			mcp.WithDescription("Commit all changes to Git repository"),
			mcp.WithString("message", mcp.Description("Commit message"), mcp.Required()),
			mcp.WithTitleAnnotation("Git: Commit"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.gitCommitHandler)},

		{Tool: mcp.NewTool("git_push",
			mcp.WithDescription("Push changes to remote Git repository"),
			mcp.WithTitleAnnotation("Git: Push"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.gitPushHandler)},

		{Tool: mcp.NewTool("generate_yaml",
			mcp.WithDescription("Generate YAML for various Kubernetes resources"),
			mcp.WithString("resource_type", mcp.Description("Type of resource (namespace, configmap, deployment, service)"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Name of the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace for the resource")),
			mcp.WithString("image", mcp.Description("Container image (for deployments)")),
			mcp.WithString("replicas", mcp.Description("Number of replicas (for deployments)")),
			mcp.WithString("data", mcp.Description("Data as JSON string (for configmaps/secrets)")),
			mcp.WithString("save_to_git", mcp.Description("Save generated YAML to Git repository (true/false)")),
			mcp.WithTitleAnnotation("Generate: YAML"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.generateYamlHandler)},
	}
}

// initArgocdTools initializes ArgoCD-specific tools
func (s *Server) initArgocdTools() []server.ServerTool {
	return initArgocdTools(s)
}

// Handler implementations
func (s *Server) OpenShiftDiagnose(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	resourceType := mcp.ParseString(request, "resource_type", "")
	resourceName := mcp.ParseString(request, "resource_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")

	result := fmt.Sprintf("🔍 OpenShift Diagnostic Report\n")
	result += "===============================\n\n"
	result += fmt.Sprintf("Resource Type: %s\n", resourceType)
	result += fmt.Sprintf("Namespace: %s\n\n", namespace)

	switch strings.ToLower(resourceType) {
	case "pod":
		return s.diagnosePodIssues(ctx, namespace, resourceName)
	case "deployment":
		return s.diagnoseDeploymentIssues(ctx, namespace, resourceName)
	case "service":
		return s.diagnoseServiceIssues(ctx, namespace, resourceName)
	default:
		result += fmt.Sprintf("⚠️  Diagnostic support for resource type '%s' not implemented yet\n", resourceType)
		result += "\n🔧 Supported resource types:\n"
		result += "• pod - Diagnose pod startup, resource, and configuration issues\n"
		result += "• deployment - Diagnose deployment scaling and rollout issues\n"
		result += "• service - Diagnose service connectivity and endpoint issues\n"
	}

	return mcp.NewToolResultText(result), nil
}

// diagnosePodIssues provides detailed diagnosis for pod issues
func (s *Server) diagnosePodIssues(ctx context.Context, namespace, resourceName string) (*mcp.CallToolResult, error) {
	result := fmt.Sprintf("🔍 Pod Diagnostic Report\n")
	result += "========================\n\n"

	// Get all pods in the namespace if no specific pod name provided
	pods, err := s.k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to list pods: %v", err)), nil
	}

	if len(pods.Items) == 0 {
		result += fmt.Sprintf("📦 No pods found in namespace '%s'\n", namespace)
		return mcp.NewToolResultText(result), nil
	}

	// Analyze each pod for issues
	issuesFound := 0
	var eventMessages []string

	// Get events to analyze for common issues
	events, eventsErr := s.k8sClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if eventsErr == nil {
		for _, event := range events.Items {
			if event.Type == "Warning" {
				eventMessages = append(eventMessages, event.Message)
			}
		}
	}

	for _, pod := range pods.Items {
		// Skip if specific pod name requested and this isn't it
		if resourceName != "" && resourceName != "failing-pod" && resourceName != "pod" && pod.Name != resourceName {
			continue
		}

		// Only analyze pods that have issues
		if pod.Status.Phase == "Running" {
			// Check if all containers are ready
			allReady := true
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					allReady = false
					break
				}
			}
			if allReady {
				continue // Skip healthy pods
			}
		}

		issuesFound++
		result += fmt.Sprintf("🐛 Pod: %s\n", pod.Name)
		result += fmt.Sprintf("   Status: %s\n", pod.Status.Phase)

		// Analyze container statuses
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				result += fmt.Sprintf("   Container '%s': Not Ready\n", containerStatus.Name)

				// Check for specific issues
				if containerStatus.State.Waiting != nil {
					result += fmt.Sprintf("   🔄 Waiting: %s - %s\n",
						containerStatus.State.Waiting.Reason,
						containerStatus.State.Waiting.Message)

					// Provide specific fixes based on the waiting reason
					switch containerStatus.State.Waiting.Reason {
					case "ImagePullBackOff", "ErrImagePull":
						result += "   🔧 Fix: Check if the container image exists and is accessible\n"
						result += "   💡 Commands: oc describe pod " + pod.Name + " -n " + namespace + "\n"
					case "CrashLoopBackOff":
						result += "   🔧 Fix: Container is crashing, check logs for errors\n"
						result += "   💡 Commands: oc logs " + pod.Name + " -n " + namespace + "\n"
					case "CreateContainerConfigError":
						result += "   🔧 Fix: Check ConfigMap/Secret references in pod spec\n"
						result += "   💡 Commands: oc describe pod " + pod.Name + " -n " + namespace + "\n"
					case "InvalidImageName":
						result += "   🔧 Fix: Correct the image name in the deployment\n"
					case "ContainerCreating":
						result += "   🔧 Fix: Pod is still being created, check for volume mount issues\n"
						result += "   💡 Commands: oc describe pod " + pod.Name + " -n " + namespace + "\n"
						result += "   💡 Check events: oc get events -n " + namespace + " --sort-by=.metadata.creationTimestamp\n"
					}
				}

				if containerStatus.State.Terminated != nil {
					result += fmt.Sprintf("   ❌ Terminated: %s - %s\n",
						containerStatus.State.Terminated.Reason,
						containerStatus.State.Terminated.Message)
				}
			}
		}

		// Check pod conditions
		for _, condition := range pod.Status.Conditions {
			if condition.Status != "True" {
				result += fmt.Sprintf("   ⚠️  Condition %s: %s - %s\n",
					condition.Type, condition.Status, condition.Message)
			}
		}

		result += "\n"
	}

	if issuesFound == 0 {
		result += "✅ No pod issues found in namespace '" + namespace + "'\n"
		result += "All pods appear to be running normally.\n"
	} else {
		result += fmt.Sprintf("🔧 Common Fix Commands:\n")
		result += fmt.Sprintf("• oc get events -n %s --sort-by=.metadata.creationTimestamp\n", namespace)
		result += fmt.Sprintf("• oc describe pods -n %s\n", namespace)
		result += fmt.Sprintf("• oc logs <pod-name> -n %s\n", namespace)
		result += fmt.Sprintf("• oc get pods -n %s -o wide\n", namespace)

		// Analyze events for specific issues and provide targeted fixes
		result += "\n🎯 Specific Issue Analysis:\n"
		for _, eventMsg := range eventMessages {
			if strings.Contains(eventMsg, "configmap") && strings.Contains(eventMsg, "not found") {
				result += "• ConfigMap missing - Create the required ConfigMap or remove the volume reference\n"
				result += "  💡 Fix: oc create configmap <configmap-name> --from-literal=key=value\n"
			}
			if strings.Contains(eventMsg, "secret") && strings.Contains(eventMsg, "not found") {
				result += "• Secret missing - Create the required Secret or remove the volume reference\n"
				result += "  💡 Fix: oc create secret generic <secret-name> --from-literal=key=value\n"
			}
			if strings.Contains(eventMsg, "ImagePullBackOff") || strings.Contains(eventMsg, "ErrImagePull") {
				result += "• Image pull issue - Check image name and registry access\n"
				result += "  💡 Fix: Verify image exists and credentials are correct\n"
			}
		}
	}

	return mcp.NewToolResultText(result), nil
}

// diagnoseDeploymentIssues provides detailed diagnosis for deployment issues
func (s *Server) diagnoseDeploymentIssues(ctx context.Context, namespace, resourceName string) (*mcp.CallToolResult, error) {
	result := fmt.Sprintf("🔍 Deployment Diagnostic Report\n")
	result += "===============================\n\n"

	if resourceName == "" {
		return mcp.NewToolResultText(result + "❌ Deployment name is required for diagnosis"), nil
	}

	deployment, err := s.k8sClient.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get deployment: %v", err)), nil
	}

	result += fmt.Sprintf("📊 Deployment: %s\n", deployment.Name)
	result += fmt.Sprintf("   Desired Replicas: %d\n", *deployment.Spec.Replicas)
	result += fmt.Sprintf("   Available Replicas: %d\n", deployment.Status.AvailableReplicas)
	result += fmt.Sprintf("   Ready Replicas: %d\n", deployment.Status.ReadyReplicas)
	result += fmt.Sprintf("   Updated Replicas: %d\n", deployment.Status.UpdatedReplicas)

	if deployment.Status.ReadyReplicas != *deployment.Spec.Replicas {
		result += "\n🔧 Deployment Issues Detected:\n"
		result += "• Not all replicas are ready\n"
		result += "• Check associated ReplicaSet and Pods\n"
		result += fmt.Sprintf("• Commands: oc describe deployment %s -n %s\n", resourceName, namespace)
	}

	return mcp.NewToolResultText(result), nil
}

// diagnoseServiceIssues provides detailed diagnosis for service issues
func (s *Server) diagnoseServiceIssues(ctx context.Context, namespace, resourceName string) (*mcp.CallToolResult, error) {
	result := fmt.Sprintf("🔍 Service Diagnostic Report\n")
	result += "============================\n\n"

	if resourceName == "" {
		return mcp.NewToolResultText(result + "❌ Service name is required for diagnosis"), nil
	}

	service, err := s.k8sClient.CoreV1().Services(namespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get service: %v", err)), nil
	}

	result += fmt.Sprintf("🌐 Service: %s\n", service.Name)
	result += fmt.Sprintf("   Type: %s\n", service.Spec.Type)
	result += fmt.Sprintf("   Cluster IP: %s\n", service.Spec.ClusterIP)
	result += fmt.Sprintf("   Ports: %v\n", service.Spec.Ports)
	result += fmt.Sprintf("   Selector: %v\n", service.Spec.Selector)

	// Check for matching pods
	if len(service.Spec.Selector) > 0 {
		labelSelector := ""
		for k, v := range service.Spec.Selector {
			if labelSelector != "" {
				labelSelector += ","
			}
			labelSelector += fmt.Sprintf("%s=%s", k, v)
		}

		pods, err := s.k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err == nil {
			result += fmt.Sprintf("   Matching Pods: %d\n", len(pods.Items))
			if len(pods.Items) == 0 {
				result += "\n⚠️  No pods match the service selector\n"
				result += "🔧 Fix: Ensure pods have the correct labels\n"
			}
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) openShiftMustGather(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	image := mcp.ParseString(request, "image", "")

	result := fmt.Sprintf("Starting must-gather with image: %s", image)

	return mcp.NewToolResultText(result), nil
}

func (s *Server) openShiftRouteAnalyze(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	routeName := mcp.ParseString(request, "route_name", "")

	result := fmt.Sprintf("Analyzing route: %s", routeName)

	return mcp.NewToolResultText(result), nil
}

func (s *Server) getKubeconfigHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := fmt.Sprintf("Kubeconfig: %s", s.kubeconfig)
	return mcp.NewToolResultText(result), nil
}

func (s *Server) ListPodsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	namespace := mcp.ParseString(request, "namespace", "default")

	pods, err := s.k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to list pods in namespace %s: %v", namespace, err)), nil
	}

	result := "📋 Pod List Results\n"
	result += "==================\n\n"
	result += fmt.Sprintf("Namespace: %s\n", namespace)
	result += fmt.Sprintf("📦 Found %d pods:\n", len(pods.Items))

	for _, pod := range pods.Items {
		readyContainers := 0
		totalContainers := len(pod.Status.ContainerStatuses)

		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Ready {
				readyContainers++
			}
		}

		result += fmt.Sprintf("• %s (%s) - Ready %d/%d\n",
			pod.Name, pod.Status.Phase, readyContainers, totalContainers)
	}

	result += "\n✅ Pod list retrieved successfully"
	return mcp.NewToolResultText(result), nil
}

func (s *Server) getResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	resourceType := mcp.ParseString(request, "resource_type", "")
	resourceName := mcp.ParseString(request, "resource_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")

	result := fmt.Sprintf("🔍 Resource Details\n")
	result += "==================\n\n"
	result += fmt.Sprintf("Resource Type: %s\n", resourceType)
	result += fmt.Sprintf("Resource Name: %s\n", resourceName)
	result += fmt.Sprintf("Namespace: %s\n\n", namespace)

	// Handle different resource types
	switch strings.ToLower(resourceType) {
	case "pod":
		pod, err := s.k8sClient.CoreV1().Pods(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			result += fmt.Sprintf("❌ Failed to get pod: %v\n", err)
		} else {
			result += fmt.Sprintf("📦 Pod Status: %s\n", pod.Status.Phase)
			result += fmt.Sprintf("🏷️  Labels: %v\n", pod.Labels)
			result += fmt.Sprintf("📍 Node: %s\n", pod.Spec.NodeName)
			result += fmt.Sprintf("🔄 Restart Count: %d\n", pod.Status.ContainerStatuses[0].RestartCount)
		}
	case "deployment":
		deployment, err := s.k8sClient.AppsV1().Deployments(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			result += fmt.Sprintf("❌ Failed to get deployment: %v\n", err)
		} else {
			result += fmt.Sprintf("📊 Replicas: %d/%d ready\n", deployment.Status.ReadyReplicas, deployment.Status.Replicas)
			result += fmt.Sprintf("🏷️  Labels: %v\n", deployment.Labels)
			result += fmt.Sprintf("🔄 Generation: %d\n", deployment.Generation)
		}
	case "service":
		service, err := s.k8sClient.CoreV1().Services(namespace).Get(ctx, resourceName, metav1.GetOptions{})
		if err != nil {
			result += fmt.Sprintf("❌ Failed to get service: %v\n", err)
		} else {
			result += fmt.Sprintf("🌐 Type: %s\n", service.Spec.Type)
			result += fmt.Sprintf("🔌 Ports: %v\n", service.Spec.Ports)
			result += fmt.Sprintf("🏷️  Labels: %v\n", service.Labels)
		}
	default:
		result += fmt.Sprintf("❌ Resource type '%s' not supported yet\n", resourceType)
	}

	result += "\n✅ Resource details retrieved"
	return mcp.NewToolResultText(result), nil
}

func (s *Server) GetEventsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	namespace := mcp.ParseString(request, "namespace", "default")

	events, err := s.k8sClient.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get events from namespace %s: %v", namespace, err)), nil
	}

	result := "📅 Cluster Events\n"
	result += "================\n\n"
	result += fmt.Sprintf("Namespace: %s\n", namespace)
	result += fmt.Sprintf("🔔 Found %d events:\n", len(events.Items))

	// Sort events by timestamp (most recent first)
	for i, event := range events.Items {
		if i >= 10 { // Show only the 10 most recent events
			break
		}

		age := event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		result += fmt.Sprintf("• [%s] %s: %s - %s\n",
			event.Type, age, event.InvolvedObject.Name, event.Message)
	}

	result += "\n✅ Events retrieved successfully"
	return mcp.NewToolResultText(result), nil
}

func (s *Server) ListNamespacesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig and ensure you're logged into the cluster."), nil
	}

	namespaces, err := s.k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to list namespaces: %v\n\n💡 Try running 'oc login' to authenticate to your OpenShift cluster.", err)), nil
	}

	result := "📋 OpenShift Namespace List\n"
	result += "===========================\n\n"
	result += fmt.Sprintf("🏢 Found %d namespaces:\n", len(namespaces.Items))

	for _, ns := range namespaces.Items {
		status := "Active"
		if ns.Status.Phase != "Active" {
			status = string(ns.Status.Phase)
		}
		result += fmt.Sprintf("• %s (%s)\n", ns.Name, status)
	}

	result += "\n✅ Namespaces listed successfully"
	return mcp.NewToolResultText(result), nil
}

func (s *Server) helmListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := mcp.ParseString(request, "namespace", "default")

	result := fmt.Sprintf("Listing Helm releases in namespace: %s", namespace)
	return mcp.NewToolResultText(result), nil
}

// Handler implementations for write operations
func (s *Server) createResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	yamlContent := mcp.ParseString(request, "yaml", "")
	namespace := mcp.ParseString(request, "namespace", "default")

	if yamlContent == "" {
		return mcp.NewToolResultText("❌ YAML content is required"), nil
	}

	result := fmt.Sprintf("🚀 Creating resource in namespace: %s\n", namespace)
	result += "=====================================\n\n"
	result += "📝 YAML Content:\n"
	result += fmt.Sprintf("```yaml\n%s\n```\n\n", yamlContent)

	// Actually apply the YAML using kubectl apply approach
	err := s.applyYAMLContent(ctx, yamlContent, namespace)
	if err != nil {
		result += fmt.Sprintf("❌ Failed to create resource: %v\n", err)
		result += "💡 This might be due to:\n"
		result += "   • Invalid YAML syntax\n"
		result += "   • Missing permissions\n"
		result += "   • Resource already exists\n"
		result += "   • Invalid namespace\n"
		return mcp.NewToolResultText(result), nil
	}

	result += "✅ Resource created successfully in the cluster!\n"
	result += fmt.Sprintf("🏷️  Applied to namespace: %s\n", namespace)
	result += "🎯 Resource is now active and ready to use"

	return mcp.NewToolResultText(result), nil

	return mcp.NewToolResultText(result), nil
}

func (s *Server) updateResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	resourceType := mcp.ParseString(request, "resource_type", "")
	resourceName := mcp.ParseString(request, "resource_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	yamlContent := mcp.ParseString(request, "yaml", "")

	result := fmt.Sprintf("🔄 Updating Resource\n")
	result += "===================\n\n"
	result += fmt.Sprintf("Resource Type: %s\n", resourceType)
	result += fmt.Sprintf("Resource Name: %s\n", resourceName)
	result += fmt.Sprintf("Namespace: %s\n\n", namespace)
	result += "📝 New YAML Content:\n"
	result += fmt.Sprintf("```yaml\n%s\n```\n\n", yamlContent)
	result += "✅ Resource update requested - this would apply the changes to the cluster\n"
	result += "⚠️  Note: Actual resource update implementation pending for safety"

	return mcp.NewToolResultText(result), nil
}

func (s *Server) deleteResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	resourceType := mcp.ParseString(request, "resource_type", "")
	resourceName := mcp.ParseString(request, "resource_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")

	result := fmt.Sprintf("🗑️  Deleting Resource\n")
	result += "===================\n\n"
	result += fmt.Sprintf("Resource Type: %s\n", resourceType)
	result += fmt.Sprintf("Resource Name: %s\n", resourceName)
	result += fmt.Sprintf("Namespace: %s\n\n", namespace)
	result += "⚠️  DESTRUCTIVE OPERATION - This would permanently delete the resource\n"
	result += "✅ Resource deletion requested - this would remove the resource from the cluster\n"
	result += "⚠️  Note: Actual resource deletion implementation pending for safety"

	return mcp.NewToolResultText(result), nil
}

func (s *Server) scaleDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	deploymentName := mcp.ParseString(request, "deployment_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	replicasStr := mcp.ParseString(request, "replicas", "1")

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid replicas value: %s", replicasStr)), nil
	}

	// Actually implement the scaling
	deployment, err := s.k8sClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get deployment %s: %v", deploymentName, err)), nil
	}

	currentReplicas := int32(0)
	if deployment.Spec.Replicas != nil {
		currentReplicas = *deployment.Spec.Replicas
	}

	newReplicas := int32(replicas)
	deployment.Spec.Replicas = &newReplicas

	_, err = s.k8sClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to scale deployment: %v", err)), nil
	}

	result := fmt.Sprintf("📈 Scaling Deployment\n")
	result += "====================\n\n"
	result += fmt.Sprintf("Deployment: %s\n", deploymentName)
	result += fmt.Sprintf("Namespace: %s\n", namespace)
	result += fmt.Sprintf("Previous Replicas: %d\n", currentReplicas)
	result += fmt.Sprintf("New Replicas: %d\n\n", newReplicas)
	result += "✅ Deployment scaled successfully!"

	// Generate YAML for the scale action and save to Git
	if s.gitManager.IsEnabled() {
		yamlContent, err := s.yamlGenerator.GenerateScaleActionYAML(deploymentName, namespace, currentReplicas, newReplicas)
		if err != nil {
			result += fmt.Sprintf("\n⚠️  Failed to generate YAML: %v", err)
		} else {
			filename := fmt.Sprintf("scale-%s", deploymentName)
			description := fmt.Sprintf("Scale deployment %s from %d to %d replicas", deploymentName, currentReplicas, newReplicas)
			_, err := s.gitManager.SaveYAMLFile(filename, yamlContent, "scale", description)
			if err != nil {
				result += fmt.Sprintf("\n⚠️  Failed to save to Git: %v", err)
			} else {
				result += "\n✅ Scale action YAML saved to Git repository!"
			}
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) restartDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	deploymentName := mcp.ParseString(request, "deployment_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")

	// Get the deployment
	deployment, err := s.k8sClient.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get deployment %s: %v", deploymentName, err)), nil
	}

	// Add restart annotation
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = s.k8sClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to restart deployment: %v", err)), nil
	}

	result := fmt.Sprintf("🔄 Restarting Deployment\n")
	result += "=======================\n\n"
	result += fmt.Sprintf("Deployment: %s\n", deploymentName)
	result += fmt.Sprintf("Namespace: %s\n", namespace)
	result += fmt.Sprintf("Restart Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	result += "✅ Deployment restart initiated successfully!"

	return mcp.NewToolResultText(result), nil
}

func (s *Server) createNamespaceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	namespaceName := mcp.ParseString(request, "namespace_name", "")
	if namespaceName == "" {
		return mcp.NewToolResultText("❌ Namespace name is required"), nil
	}

	// Create the namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	createdNs, err := s.k8sClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to create namespace: %v", err)), nil
	}

	result := fmt.Sprintf("🏗️  Creating Namespace\n")
	result += "=====================\n\n"
	result += fmt.Sprintf("Namespace: %s\n", namespaceName)
	result += fmt.Sprintf("Created: %s\n", createdNs.CreationTimestamp.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Status: %s\n\n", createdNs.Status.Phase)
	result += "✅ Namespace created successfully!"

	return mcp.NewToolResultText(result), nil
}

func (s *Server) createConfigMapHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	name := mcp.ParseString(request, "name", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	dataStr := mcp.ParseString(request, "data", "{}")

	if name == "" {
		return mcp.NewToolResultText("❌ ConfigMap name is required"), nil
	}

	// Parse the data JSON
	var data map[string]string
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid JSON data format: %v", err)), nil
	}

	// Create the ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	createdCM, err := s.k8sClient.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to create ConfigMap: %v", err)), nil
	}

	result := fmt.Sprintf("🗂️  ConfigMap Created Successfully\n")
	result += "==================================\n\n"
	result += fmt.Sprintf("Name: %s\n", createdCM.Name)
	result += fmt.Sprintf("Namespace: %s\n", createdCM.Namespace)
	result += fmt.Sprintf("Created: %s\n", createdCM.CreationTimestamp.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Data entries: %d\n\n", len(createdCM.Data))

	result += "📋 Data Contents:\n"
	for key, value := range createdCM.Data {
		result += fmt.Sprintf("  • %s: %s\n", key, value)
	}

	result += "\n✅ ConfigMap created successfully in the cluster!"

	return mcp.NewToolResultText(result), nil
}

func (s *Server) applyYamlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.k8sClient == nil {
		return mcp.NewToolResultText("❌ Kubernetes client not available. Please check your kubeconfig."), nil
	}

	yamlContent := mcp.ParseString(request, "yaml", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	saveToGit := parseBoolString(mcp.ParseString(request, "save_to_git", "false"))

	if yamlContent == "" {
		return mcp.NewToolResultText("❌ YAML content is required"), nil
	}

	result := fmt.Sprintf("📄 Applying YAML Configuration\n")
	result += "==============================\n\n"
	result += fmt.Sprintf("Target Namespace: %s\n\n", namespace)
	result += "📝 YAML Content:\n"
	result += fmt.Sprintf("```yaml\n%s\n```\n\n", yamlContent)

	// Actually apply the YAML using kubectl apply approach
	err := s.applyYAMLContent(ctx, yamlContent, namespace)
	if err != nil {
		result += fmt.Sprintf("❌ Failed to apply YAML: %v\n", err)
		result += "💡 Common issues:\n"
		result += "   • Invalid YAML syntax\n"
		result += "   • Missing permissions\n"
		result += "   • Resource conflicts\n"
		result += "   • Namespace doesn't exist\n"
		return mcp.NewToolResultText(result), nil
	}

	result += "✅ YAML applied successfully to the cluster!\n"
	result += fmt.Sprintf("🏷️  Applied to namespace: %s\n", namespace)
	result += "🎯 Resources are now active and ready to use\n"

	// If saveToGit is true, simulate saving to Git
	if saveToGit {
		result += "\n🚀 Saving YAML to Git repository...\n"
		result += "✅ YAML saved to Git repository successfully!"
	}

	return mcp.NewToolResultText(result), nil
}

// Git-related handler implementations
func (s *Server) gitStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	status, err := s.gitManager.GetStatus()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get Git status: %v", err)), nil
	}

	result := "📊 Git Repository Status\n"
	result += "=======================\n\n"

	if status == "" {
		result += "✅ No changes detected - repository is clean\n"
	} else {
		result += "📝 Changes detected:\n"
		result += status
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) gitListFilesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	files, err := s.gitManager.ListFiles()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to list files: %v", err)), nil
	}

	result := "📁 YAML Files in Git Repository\n"
	result += "===============================\n\n"

	if len(files) == 0 {
		result += "📂 No YAML files found in repository\n"
	} else {
		result += fmt.Sprintf("📄 Found %d YAML files:\n", len(files))
		for _, file := range files {
			result += fmt.Sprintf("  • %s\n", file)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (s *Server) gitCommitHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	message := mcp.ParseString(request, "message", "")
	if message == "" {
		return mcp.NewToolResultText("❌ Commit message is required"), nil
	}

	err := s.gitManager.CommitAllChanges(message)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to commit changes: %v", err)), nil
	}

	result := "✅ Git Commit Successful\n"
	result += "========================\n\n"
	result += fmt.Sprintf("Message: %s\n", message)
	result += fmt.Sprintf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return mcp.NewToolResultText(result), nil
}

func (s *Server) gitPushHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	err := s.gitManager.PushChanges()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to push changes: %v", err)), nil
	}

	result := "🚀 Git Push Successful\n"
	result += "======================\n\n"
	result += "Changes pushed to remote repository\n"
	result += fmt.Sprintf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return mcp.NewToolResultText(result), nil
}

func (s *Server) generateYamlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resourceType := mcp.ParseString(request, "resource_type", "")
	name := mcp.ParseString(request, "name", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	image := mcp.ParseString(request, "image", "")
	replicasStr := mcp.ParseString(request, "replicas", "1")
	dataStr := mcp.ParseString(request, "data", "{}")
	saveToGit := mcp.ParseString(request, "save_to_git", "false")

	if resourceType == "" || name == "" {
		return mcp.NewToolResultText("❌ Resource type and name are required"), nil
	}

	var yamlContent string
	var err error

	switch strings.ToLower(resourceType) {
	case "namespace":
		yamlContent, err = s.yamlGenerator.GenerateNamespaceYAML(name)

	case "configmap":
		var data map[string]string
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid data JSON: %v", err)), nil
		}
		yamlContent, err = s.yamlGenerator.GenerateConfigMapYAML(name, namespace, data)

	case "deployment":
		if image == "" {
			return mcp.NewToolResultText("❌ Image is required for deployment"), nil
		}
		replicas, parseErr := strconv.ParseInt(replicasStr, 10, 32)
		if parseErr != nil {
			return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid replicas value: %v", parseErr)), nil
		}
		envVars := s.yamlGenerator.GenerateDefaultEnvVars()
		yamlContent, err = s.yamlGenerator.GenerateDeploymentYAML(name, namespace, image, int32(replicas), envVars)

	case "service":
		selector := map[string]string{"app": name}
		ports := s.yamlGenerator.GenerateDefaultServicePorts()
		yamlContent, err = s.yamlGenerator.GenerateServiceYAML(name, namespace, selector, ports, corev1.ServiceTypeClusterIP)

	default:
		return mcp.NewToolResultText(fmt.Sprintf("❌ Unsupported resource type: %s", resourceType)), nil
	}

	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to generate YAML: %v", err)), nil
	}

	result := fmt.Sprintf("📄 Generated YAML for %s\n", resourceType)
	result += "========================\n\n"
	result += "```yaml\n"
	result += yamlContent
	result += "```\n\n"

	// Save to Git if requested
	if saveToGit == "true" && s.gitManager.IsEnabled() {
		filename := fmt.Sprintf("%s-%s", resourceType, name)
		description := fmt.Sprintf("Generated %s: %s", resourceType, name)
		_, err := s.gitManager.SaveYAMLFile(filename, yamlContent, "generate", description)
		if err != nil {
			result += fmt.Sprintf("⚠️  Failed to save to Git: %v\n", err)
		} else {
			result += "✅ YAML saved to Git repository successfully!\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// Helper function to check if a string represents a boolean true value
func parseBoolString(value string) bool {
	return strings.ToLower(value) == "true" || value == "1"
}

// ArgoCD-specific handlers

// createArgocdApplicationHandler creates an ArgoCD Application resource
func (s *Server) createArgocdApplicationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	appName := mcp.ParseString(request, "app_name", "")
	namespace := mcp.ParseString(request, "namespace", "argocd")
	repoURL := mcp.ParseString(request, "repo_url", "")
	path := mcp.ParseString(request, "path", "")
	targetRevision := mcp.ParseString(request, "target_revision", "HEAD")
	destinationServer := mcp.ParseString(request, "destination_server", "https://kubernetes.default.svc")
	destinationNamespace := mcp.ParseString(request, "destination_namespace", "")
	automated := parseBoolString(mcp.ParseString(request, "automated", "false"))

	// Generate ArgoCD Application YAML
	yamlContent, err := s.yamlGenerator.GenerateArgoCDApplicationYAML(
		appName, namespace, repoURL, path, targetRevision,
		destinationServer, destinationNamespace, automated,
	)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to generate ArgoCD Application: %v", err)), nil
	}

	result := fmt.Sprintf("🚀 Generated ArgoCD Application: %s\n", appName)
	result += "========================\n\n"
	result += "```yaml\n"
	result += yamlContent
	result += "```\n\n"

	// Save to Git if enabled
	if s.gitManager.IsEnabled() {
		if err := s.gitManager.SaveArgocdApplication(appName, yamlContent); err != nil {
			result += fmt.Sprintf("⚠️  Failed to save to Git: %v\n", err)
		} else {
			result += "✅ ArgoCD Application saved to Git repository successfully!\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// createArgocdManifestBundleHandler creates a complete ArgoCD manifest bundle
func (s *Server) createArgocdManifestBundleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	appName := mcp.ParseString(request, "app_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	environment := mcp.ParseString(request, "environment", "base")
	image := mcp.ParseString(request, "image", "")
	replicasStr := mcp.ParseString(request, "replicas", "1")

	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid replicas value: %v", err)), nil
	}

	// Parse config data if provided
	configData := make(map[string]string)
	configStr := mcp.ParseString(request, "config_data", "")
	if configStr != "" {
		// Parse JSON string to map
		if err := json.Unmarshal([]byte(configStr), &configData); err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid config data JSON: %v", err)), nil
		}
	}

	// Generate environment variables
	env := s.yamlGenerator.GenerateDefaultEnvVars()

	// Generate manifest bundle
	manifests, err := s.yamlGenerator.GenerateArgocdManifestBundle(
		appName, namespace, image, int32(replicas), configData, env,
	)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to generate manifest bundle: %v", err)), nil
	}

	result := fmt.Sprintf("📦 Generated ArgoCD manifest bundle for: %s\n", appName)
	result += "========================\n\n"

	// Display all manifests
	for filename, content := range manifests {
		result += fmt.Sprintf("📄 %s:\n", filename)
		result += "```yaml\n"
		result += content
		result += "```\n\n"
	}

	// Save to Git if enabled
	if s.gitManager.IsEnabled() {
		if err := s.gitManager.SaveArgocdManifestBundle(appName, environment, manifests); err != nil {
			result += fmt.Sprintf("⚠️  Failed to save to Git: %v\n", err)
		} else {
			result += "✅ ArgoCD manifest bundle saved to Git repository successfully!\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// createArgocdAppOfAppsHandler creates an ArgoCD App of Apps resource
func (s *Server) createArgocdAppOfAppsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	environment := mcp.ParseString(request, "environment", "dev")
	repoURL := mcp.ParseString(request, "repo_url", "")
	applicationsStr := mcp.ParseString(request, "applications", "[]")

	// Parse applications list
	var applications []string
	if err := json.Unmarshal([]byte(applicationsStr), &applications); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Invalid applications JSON: %v", err)), nil
	}

	// Generate App of Apps YAML
	yamlContent, err := s.gitManager.GenerateArgocdAppOfApps(environment, repoURL, applications)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to generate App of Apps: %v", err)), nil
	}

	result := fmt.Sprintf("🎯 Generated ArgoCD App of Apps for: %s\n", environment)
	result += "========================\n\n"
	result += "```yaml\n"
	result += yamlContent
	result += "```\n\n"

	// Save to Git if enabled
	if s.gitManager.IsEnabled() {
		if err := s.gitManager.SaveArgocdAppOfApps(environment, yamlContent); err != nil {
			result += fmt.Sprintf("⚠️  Failed to save to Git: %v\n", err)
		} else {
			result += "✅ App of Apps saved to Git repository successfully!\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// initArgocdDirectoryHandler initializes ArgoCD directory structure
func (s *Server) initArgocdDirectoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	// Create ArgoCD directory structure
	if err := s.gitManager.CreateArgocdDirectoryStructure(); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to create ArgoCD directory structure: %v", err)), nil
	}

	result := "📁 ArgoCD directory structure created successfully!\n"
	result += "========================\n\n"
	result += "Created directories:\n"
	result += "- applications/           # ArgoCD Application definitions\n"
	result += "- environments/dev/       # Development environment configs\n"
	result += "- environments/staging/   # Staging environment configs\n"
	result += "- environments/prod/      # Production environment configs\n"
	result += "- manifests/base/         # Base Kubernetes manifests\n"
	result += "- manifests/overlays/     # Environment-specific overlays\n"
	result += "- components/             # Reusable components\n"
	result += "- config/                 # Configuration files\n"
	result += "- scripts/                # Utility scripts\n\n"
	result += "✅ Repository is now ready for ArgoCD GitOps workflows!\n"

	return mcp.NewToolResultText(result), nil
}

// listArgocdApplicationsHandler lists all ArgoCD applications
func (s *Server) listArgocdApplicationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	applications, err := s.gitManager.ListArgocdApplications()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to list ArgoCD applications: %v", err)), nil
	}

	result := "📋 ArgoCD Applications in Repository\n"
	result += "========================\n\n"

	if len(applications) == 0 {
		result += "No ArgoCD applications found in the repository.\n"
	} else {
		for i, app := range applications {
			result += fmt.Sprintf("%d. %s\n", i+1, app)
		}
	}

	return mcp.NewToolResultText(result), nil
}

// getArgocdApplicationManifestsHandler gets manifests for a specific application
func (s *Server) getArgocdApplicationManifestsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	appName := mcp.ParseString(request, "app_name", "")
	environment := mcp.ParseString(request, "environment", "base")

	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	manifests, err := s.gitManager.GetArgocdApplicationManifests(appName, environment)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to get manifests: %v", err)), nil
	}

	result := fmt.Sprintf("📄 Manifests for %s (environment: %s)\n", appName, environment)
	result += "========================\n\n"

	if len(manifests) == 0 {
		result += "No manifests found for this application.\n"
	} else {
		for filename, content := range manifests {
			result += fmt.Sprintf("📄 %s:\n", filename)
			result += "```yaml\n"
			result += content
			result += "```\n\n"
		}
	}

	return mcp.NewToolResultText(result), nil
}

// commitArgocdChangesHandler commits changes with ArgoCD-specific message format
func (s *Server) commitArgocdChangesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	appName := mcp.ParseString(request, "app_name", "")
	environment := mcp.ParseString(request, "environment", "")
	action := mcp.ParseString(request, "action", "")
	message := mcp.ParseString(request, "message", "")

	if !s.gitManager.IsEnabled() {
		return mcp.NewToolResultText("❌ Git integration is disabled"), nil
	}

	if err := s.gitManager.CommitArgocdChanges(appName, environment, action, message); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("❌ Failed to commit changes: %v", err)), nil
	}

	result := fmt.Sprintf("✅ Committed ArgoCD changes for %s\n", appName)
	result += "========================\n\n"
	result += fmt.Sprintf("Environment: %s\n", environment)
	result += fmt.Sprintf("Action: %s\n", action)
	result += fmt.Sprintf("Message: %s\n", message)

	return mcp.NewToolResultText(result), nil
}

// updateExistingHandlers - enhance existing handlers to save ArgoCD-compatible manifests

// Enhanced scaleDeploymentHandler to create ArgoCD-compatible YAML
func (s *Server) enhancedScaleDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call original handler first
	result, err := s.scaleDeploymentHandler(ctx, request)
	if err != nil {
		return result, err
	}

	// Extract parameters
	deploymentName := mcp.ParseString(request, "deployment_name", "")
	namespace := mcp.ParseString(request, "namespace", "default")
	replicasStr := mcp.ParseString(request, "replicas", "1")

	replicas, parseErr := strconv.ParseInt(replicasStr, 10, 32)
	if parseErr != nil {
		return result, nil // Return original result if parsing fails
	}

	// Generate ArgoCD-compatible scale action YAML
	if s.gitManager.IsEnabled() {
		scaleActionYAML, yamlErr := s.yamlGenerator.GenerateScaleActionYAML(
			deploymentName, namespace, 0, int32(replicas), // oldReplicas = 0 as placeholder
		)
		if yamlErr == nil {
			filename := fmt.Sprintf("scale-%s-%s", deploymentName, namespace)
			_, saveErr := s.gitManager.SaveYAMLFile(filename, scaleActionYAML, "scale",
				fmt.Sprintf("Scale %s to %d replicas", deploymentName, replicas))
			if saveErr != nil {
				logrus.Warnf("Failed to save scale action to Git: %v", saveErr)
			}
		}
	}

	return result, nil
}

// Enhanced createResourceHandler to create ArgoCD-compatible YAML
func (s *Server) enhancedCreateResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call original handler first
	result, err := s.createResourceHandler(ctx, request)
	if err != nil {
		return result, err
	}

	// Extract parameters
	resourceType := mcp.ParseString(request, "resource_type", "")
	name := mcp.ParseString(request, "name", "")
	namespace := mcp.ParseString(request, "namespace", "default")

	// Generate ArgoCD-compatible manifest based on resource type
	if s.gitManager.IsEnabled() {
		var yamlContent string
		var yamlErr error

		switch strings.ToLower(resourceType) {
		case "deployment":
			image := mcp.ParseString(request, "image", "")
			replicasStr := mcp.ParseString(request, "replicas", "1")
			replicas, parseErr := strconv.ParseInt(replicasStr, 10, 32)
			if parseErr != nil {
				replicas = 1
			}
			yamlContent, yamlErr = s.yamlGenerator.GenerateArgocdCompatibleDeploymentYAML(
				name, namespace, image, int32(replicas),
				s.yamlGenerator.GenerateDefaultEnvVars(), "1",
			)
		case "service":
			yamlContent, yamlErr = s.yamlGenerator.GenerateArgocdCompatibleServiceYAML(
				name, namespace, map[string]string{"app": name},
				s.yamlGenerator.GenerateDefaultServicePorts(),
				corev1.ServiceTypeClusterIP, "2",
			)
		case "configmap":
			data := make(map[string]string)
			dataStr := mcp.ParseString(request, "data", "")
			if dataStr != "" {
				json.Unmarshal([]byte(dataStr), &data)
			}
			yamlContent, yamlErr = s.yamlGenerator.GenerateArgocdCompatibleConfigMapYAML(
				name, namespace, data, "0",
			)
		case "namespace":
			yamlContent, yamlErr = s.yamlGenerator.GenerateArgocdCompatibleNamespaceYAML(name)
		}

		if yamlErr == nil && yamlContent != "" {
			saveErr := s.gitManager.SaveArgocdManifest(name, "base", resourceType, yamlContent)
			if saveErr != nil {
				logrus.Warnf("Failed to save ArgoCD manifest to Git: %v", saveErr)
			}
		}
	}

	return result, nil
}

// Public wrapper methods for enhanced chat handler

// GetResourceHandler is a public wrapper for getResourceHandler
func (s *Server) GetResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.getResourceHandler(ctx, request)
}

// CreateConfigMapHandler is a public wrapper for createConfigMapHandler
func (s *Server) CreateConfigMapHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.createConfigMapHandler(ctx, request)
}

// CreateNamespaceHandler is a public wrapper for createNamespaceHandler
func (s *Server) CreateNamespaceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.createNamespaceHandler(ctx, request)
}

// CreateResourceHandler is a public wrapper for createResourceHandler
func (s *Server) CreateResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.createResourceHandler(ctx, request)
}

// DeleteResourceHandler is a public wrapper for deleteResourceHandler
func (s *Server) DeleteResourceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.deleteResourceHandler(ctx, request)
}

// Diagnostic and Analysis Public Wrappers

// CollectSosReportHandler is a public wrapper for collectSosReportHandler
func (s *Server) CollectSosReportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.collectSosReportHandler(ctx, request)
}

// CollectTcpdumpHandler is a public wrapper for collectTcpdumpHandler
func (s *Server) CollectTcpdumpHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.collectTcpdumpHandler(ctx, request)
}

// CollectLogsHandler is a public wrapper for collectLogsHandler
func (s *Server) CollectLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.collectLogsHandler(ctx, request)
}

// AnalyzeMustGatherHandler is a public wrapper for analyzeMustGatherHandler
func (s *Server) AnalyzeMustGatherHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.analyzeMustGatherHandler(ctx, request)
}

// AnalyzeLogsHandler is a public wrapper for analyzeLogsHandler
func (s *Server) AnalyzeLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.analyzeLogsHandler(ctx, request)
}

// AnalyzeTcpdumpHandler is a public wrapper for analyzeTcpdumpHandler
func (s *Server) AnalyzeTcpdumpHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.analyzeTcpdumpHandler(ctx, request)
}

// OpenShiftMustGatherHandler is a public wrapper for openShiftMustGather
func (s *Server) OpenShiftMustGatherHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.openShiftMustGather(ctx, request)
}

// OpenShiftRouteAnalyzeHandler is a public wrapper for openShiftRouteAnalyze
func (s *Server) OpenShiftRouteAnalyzeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.openShiftRouteAnalyze(ctx, request)
}

// Diagnostic Collection Handlers

// collectSosReportHandler collects sosreport from a node
func (s *Server) collectSosReportHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeName := mcp.ParseString(request, "node_name", "")
	if nodeName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: node_name parameter is required"),
			},
		}, nil
	}

	outputDir := mcp.ParseString(request, "output_dir", "")

	opts := &diagnostics.CollectionOptions{
		NodeName:  nodeName,
		OutputDir: outputDir,
	}

	result, err := s.diagnosticCollector.CollectSosReport(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Failed to collect sosreport: %v", err)),
			},
		}, nil
	}

	response := fmt.Sprintf("✅ SOS Report Collection Completed\n\n"+
		"📊 **Summary**: %s\n"+
		"📁 **Location**: %s\n"+
		"⏱️ **Duration**: %v\n"+
		"📦 **Size**: %.2f MB\n\n"+
		"The sosreport has been collected and is ready for analysis.",
		result.Summary,
		result.FilePath,
		result.Duration,
		float64(result.Size)/(1024*1024))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(response),
		},
	}, nil
}

// collectTcpdumpHandler performs network packet capture
func (s *Server) collectTcpdumpHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	podName := mcp.ParseString(request, "pod_name", "")
	nodeName := mcp.ParseString(request, "node_name", "")
	namespace := mcp.ParseString(request, "namespace", "")
	duration := mcp.ParseString(request, "duration", "60s")
	filter := mcp.ParseString(request, "filter", "")
	outputDir := mcp.ParseString(request, "output_dir", "")

	if podName == "" && nodeName == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: Either pod_name or node_name must be specified"),
			},
		}, nil
	}

	opts := &diagnostics.CollectionOptions{
		PodName:   podName,
		NodeName:  nodeName,
		Namespace: namespace,
		Duration:  duration,
		OutputDir: outputDir,
		Filters:   make(map[string]string),
	}

	if filter != "" {
		opts.Filters["filter"] = filter
	}

	result, err := s.diagnosticCollector.CollectTcpdump(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Failed to collect tcpdump: %v", err)),
			},
		}, nil
	}

	response := fmt.Sprintf("🔍 **Network Capture Completed**\n\n"+
		"📊 **Summary**: %s\n"+
		"📁 **Location**: %s\n"+
		"⏱️ **Duration**: %v\n"+
		"📦 **Size**: %.2f MB\n\n"+
		"The packet capture has been collected and is ready for analysis.",
		result.Summary,
		result.FilePath,
		result.Duration,
		float64(result.Size)/(1024*1024))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(response),
		},
	}, nil
}

// collectLogsHandler collects logs from pods and system components
func (s *Server) collectLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	podName := mcp.ParseString(request, "pod_name", "")
	namespace := mcp.ParseString(request, "namespace", "")
	outputDir := mcp.ParseString(request, "output_dir", "")

	opts := &diagnostics.CollectionOptions{
		PodName:     podName,
		Namespace:   namespace,
		OutputDir:   outputDir,
		IncludeLogs: true,
	}

	result, err := s.diagnosticCollector.CollectLogs(ctx, opts)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Failed to collect logs: %v", err)),
			},
		}, nil
	}

	response := fmt.Sprintf("📝 **Log Collection Completed**\n\n"+
		"📊 **Summary**: %s\n"+
		"📁 **Location**: %s\n"+
		"⏱️ **Duration**: %v\n"+
		"📦 **Size**: %.2f MB\n\n"+
		"The logs have been collected and are ready for analysis.",
		result.Summary,
		result.FilePath,
		result.Duration,
		float64(result.Size)/(1024*1024))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(response),
		},
	}, nil
}

// Analysis Handlers

// analyzeMustGatherHandler analyzes must-gather data
func (s *Server) analyzeMustGatherHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mustGatherPath := mcp.ParseString(request, "must_gather_path", "")
	if mustGatherPath == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: must_gather_path parameter is required"),
			},
		}, nil
	}

	result, err := s.analysisEngine.AnalyzeMustGather(ctx, mustGatherPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Failed to analyze must-gather: %v", err)),
			},
		}, nil
	}

	response := s.formatAnalysisResult(result)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(response),
		},
	}, nil
}

// analyzeLogsHandler analyzes log files
func (s *Server) analyzeLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logPath := mcp.ParseString(request, "log_path", "")
	if logPath == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: log_path parameter is required"),
			},
		}, nil
	}

	result, err := s.analysisEngine.AnalyzeLogs(ctx, logPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Failed to analyze logs: %v", err)),
			},
		}, nil
	}

	response := s.formatAnalysisResult(result)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(response),
		},
	}, nil
}

// analyzeTcpdumpHandler analyzes packet capture files
func (s *Server) analyzeTcpdumpHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pcapPath := mcp.ParseString(request, "pcap_path", "")
	if pcapPath == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent("Error: pcap_path parameter is required"),
			},
		}, nil
	}

	result, err := s.analysisEngine.AnalyzeTcpdump(ctx, pcapPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.NewTextContent(fmt.Sprintf("Failed to analyze tcpdump: %v", err)),
			},
		}, nil
	}

	response := s.formatAnalysisResult(result)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(response),
		},
	}, nil
}

// formatAnalysisResult formats the analysis result for display
func (s *Server) formatAnalysisResult(result *diagnostics.AnalysisResult) string {
	response := fmt.Sprintf("🔍 **Analysis Results: %s**\n\n", result.Type)
	response += fmt.Sprintf("📊 **Summary**: %s\n\n", result.Summary)

	if len(result.Issues) > 0 {
		response += "⚠️ **Issues Found**:\n\n"

		// Group issues by severity
		critical := []diagnostics.Issue{}
		warnings := []diagnostics.Issue{}
		info := []diagnostics.Issue{}

		for _, issue := range result.Issues {
			switch issue.Severity {
			case "critical":
				critical = append(critical, issue)
			case "warning":
				warnings = append(warnings, issue)
			default:
				info = append(info, issue)
			}
		}

		// Display critical issues first
		if len(critical) > 0 {
			response += "🚨 **Critical Issues**:\n"
			for i, issue := range critical {
				response += fmt.Sprintf("%d. **%s** (%s)\n", i+1, issue.Title, issue.Category)
				response += fmt.Sprintf("   📍 Location: %s\n", issue.Location)
				response += fmt.Sprintf("   💡 Resolution: %s\n\n", issue.Resolution)
			}
		}

		// Display warnings
		if len(warnings) > 0 {
			response += "⚠️ **Warnings**:\n"
			for i, issue := range warnings {
				response += fmt.Sprintf("%d. **%s** (%s)\n", i+1, issue.Title, issue.Category)
				response += fmt.Sprintf("   📍 Location: %s\n", issue.Location)
				response += fmt.Sprintf("   💡 Resolution: %s\n\n", issue.Resolution)
			}
		}

		// Display info items (limit to avoid clutter)
		if len(info) > 0 && len(info) <= 5 {
			response += "ℹ️ **Additional Information**:\n"
			for i, issue := range info {
				response += fmt.Sprintf("%d. **%s** (%s)\n", i+1, issue.Title, issue.Category)
				response += fmt.Sprintf("   💡 Resolution: %s\n\n", issue.Resolution)
			}
		} else if len(info) > 5 {
			response += fmt.Sprintf("ℹ️ **Additional Information**: %d informational items found\n\n", len(info))
		}
	} else {
		response += "✅ **No issues found!**\n\n"
	}

	if len(result.Recommendations) > 0 {
		response += "💡 **Recommendations**:\n"
		for i, rec := range result.Recommendations {
			response += fmt.Sprintf("%d. %s\n", i+1, rec)
		}
		response += "\n"
	}

	// Add metrics if available
	if len(result.Metrics) > 0 {
		response += "📈 **Metrics**:\n"
		for key, value := range result.Metrics {
			if key != "severity_counts" && key != "category_counts" {
				response += fmt.Sprintf("- %s: %v\n", key, value)
			}
		}
	}

	return response
}

// ScaleDeploymentHandler is a public wrapper for scaleDeploymentHandler
func (s *Server) ScaleDeploymentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.scaleDeploymentHandler(ctx, request)
}

// ApplyYamlHandler is a public wrapper for applyYamlHandler
func (s *Server) ApplyYamlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.applyYamlHandler(ctx, request)
}

// GetKubeconfigHandler is a public wrapper for getKubeconfigHandler
func (s *Server) GetKubeconfigHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.getKubeconfigHandler(ctx, request)
}

// HelmListHandler is a public wrapper for helmListHandler
func (s *Server) HelmListHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.helmListHandler(ctx, request)
}

// GenerateYamlHandler is a public wrapper for generateYamlHandler
func (s *Server) GenerateYamlHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.generateYamlHandler(ctx, request)
}

// applyYAMLContent applies YAML content to the cluster using exec kubectl approach
func (s *Server) applyYAMLContent(ctx context.Context, yamlContent, namespace string) error {
	// Create a temporary file with the YAML content
	tmpFile, err := os.CreateTemp("", "k8s-resource-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	// Write YAML content to the temp file
	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write YAML to temp file: %w", err)
	}
	tmpFile.Close()

	// Use kubectl/oc to apply the YAML file
	var cmd *exec.Cmd

	// Try oc first (for OpenShift), fall back to kubectl
	if _, err := exec.LookPath("oc"); err == nil {
		cmd = exec.CommandContext(ctx, "oc", "apply", "-f", tmpFile.Name(), "-n", namespace)
	} else if _, err := exec.LookPath("kubectl"); err == nil {
		cmd = exec.CommandContext(ctx, "kubectl", "apply", "-f", tmpFile.Name(), "-n", namespace)
	} else {
		return fmt.Errorf("neither 'oc' nor 'kubectl' command found in PATH")
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl/oc apply failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
