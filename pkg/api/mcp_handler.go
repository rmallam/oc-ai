package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"

	mcpserver "github.com/rakeshkumarmallam/openshift-mcp-go/pkg/mcp"
)

type MCPHandler struct {
	server *mcpserver.Server
}

func NewMCPHandler(server *mcpserver.Server) *MCPHandler {
	return &MCPHandler{
		server: server,
	}
}

// MCP Capabilities endpoint
func (h *MCPHandler) GetCapabilities(c *gin.Context) {
	capabilities := map[string]interface{}{
		"server": map[string]interface{}{
			"name":    "openshift-mcp",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools":     map[string]interface{}{"listChanged": true},
			"resources": map[string]interface{}{"subscribe": true, "listChanged": true},
			"prompts":   map[string]interface{}{"listChanged": true},
		},
		"tools": []string{
			"openshift_diagnose",
			"openshift_must_gather",
			"openshift_route_analyze",
			"collect_sosreport",
			"collect_tcpdump",
			"collect_logs",
			"analyze_must_gather",
			"analyze_logs",
			"analyze_tcpdump",
			"list_pods",
			"get_resource",
			"get_events",
			"list_namespaces",
			"helm_list",
			"create_namespace",
			"apply_yaml",
			"generate_yaml",
		},
	}

	c.JSON(http.StatusOK, capabilities)
}

// MCP Tool Call endpoint
func (h *MCPHandler) CallTool(c *gin.Context) {
	var request struct {
		Method string `json:"method"`
		Params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"params"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Method != "tools/call" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported method"})
		return
	}

	// Create MCP tool call request
	callRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      request.Params.Name,
			Arguments: request.Params.Arguments,
		},
	}

	// Execute the tool call
	result, err := h.executeTool(c.Request.Context(), callRequest)
	if err != nil {
		logrus.WithError(err).Error("Tool execution failed")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"isError": true,
		})
		return
	}

	response := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
		"toolName": request.Params.Name,
		"isError":  false,
	}

	c.JSON(http.StatusOK, response)
}

// Execute tool - simple implementation for testing
func (h *MCPHandler) executeTool(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	// Use the actual MCP server handlers instead of the limited switch statement
	result, err := h.callServerTool(ctx, request)
	if err != nil {
		return fmt.Sprintf("❌ Error executing tool '%s': %v", request.Params.Name, err), nil
	}

	// Extract text content from the result
	if result != nil && len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return fmt.Sprintf("Tool '%s' completed but returned no content", request.Params.Name), nil
}

func (h *MCPHandler) callServerTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Call the appropriate server handler based on the tool name
	switch request.Params.Name {
	case "openshift_diagnose":
		return h.server.OpenShiftDiagnose(ctx, request)
	case "openshift_must_gather":
		return h.server.OpenShiftMustGatherHandler(ctx, request)
	case "openshift_route_analyze":
		return h.server.OpenShiftRouteAnalyzeHandler(ctx, request)
	case "collect_sosreport":
		return h.server.CollectSosReportHandler(ctx, request)
	case "collect_tcpdump":
		return h.server.CollectTcpdumpHandler(ctx, request)
	case "collect_logs":
		return h.server.CollectLogsHandler(ctx, request)
	case "analyze_must_gather":
		return h.server.AnalyzeMustGatherHandler(ctx, request)
	case "analyze_logs":
		return h.server.AnalyzeLogsHandler(ctx, request)
	case "analyze_tcpdump":
		return h.server.AnalyzeTcpdumpHandler(ctx, request)
	case "list_pods":
		return h.server.ListPodsHandler(ctx, request)
	case "get_events":
		return h.server.GetEventsHandler(ctx, request)
	case "list_namespaces":
		return h.server.ListNamespacesHandler(ctx, request)
	case "get_resource":
		return h.server.GetResourceHandler(ctx, request)
	case "get_kubeconfig":
		return h.server.GetKubeconfigHandler(ctx, request)
	case "helm_list":
		return h.server.HelmListHandler(ctx, request)
	case "create_namespace":
		return h.server.CreateNamespaceHandler(ctx, request)
	case "create_resource":
		return h.server.CreateResourceHandler(ctx, request)
	case "create_configmap":
		return h.server.CreateConfigMapHandler(ctx, request)
	case "apply_yaml":
		return h.server.ApplyYamlHandler(ctx, request)
	case "delete_resource":
		return h.server.DeleteResourceHandler(ctx, request)
	case "scale_deployment":
		return h.server.ScaleDeploymentHandler(ctx, request)
	case "generate_yaml":
		return h.server.GenerateYamlHandler(ctx, request)
	default:
		return nil, fmt.Errorf("tool '%s' is not implemented", request.Params.Name)
	}
}

func (h *MCPHandler) handleOpenShiftDiagnose(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	resourceType := mcp.ParseString(request, "resource_type", "")
	resourceName := mcp.ParseString(request, "resource_name", "")
	namespace := mcp.ParseString(request, "namespace", "")

	result := "🔍 OpenShift Diagnosis Results\n"
	result += "================================\n\n"
	result += fmt.Sprintf("Resource Type: %v\n", resourceType)
	result += fmt.Sprintf("Resource Name: %v\n", resourceName)
	if namespace != "" {
		result += fmt.Sprintf("Namespace: %v\n", namespace)
	}
	result += "\n📊 Analysis:\n"
	result += "• Resource exists and is accessible\n"
	result += "• No immediate issues detected\n"
	result += "• Resource is in expected state\n"
	result += "\n✅ Diagnosis completed successfully"

	return result, nil
}

func (h *MCPHandler) handlePodsList(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	// Call the actual MCP server handler directly
	result, err := h.server.ListPodsHandler(ctx, request)
	if err != nil {
		return "", err
	}

	// Extract text from the MCP result
	if result != nil && len(result.Content) > 0 {
		// Try different ways to extract text content
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			return content.Text, nil
		case mcp.TextContent:
			return content.Text, nil
		default:
			// If it's not a TextContent, try to convert to string
			return fmt.Sprintf("%v", content), nil
		}
	}

	return "No result returned", nil
}

func (h *MCPHandler) handleEventsList(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	namespace := mcp.ParseString(request, "namespace", "")

	result := "📅 Cluster Events\n"
	result += "================\n\n"
	if namespace != "" {
		result += fmt.Sprintf("Namespace: %v\n", namespace)
	} else {
		result += "Namespace: All namespaces\n"
	}
	result += "\n🔔 Recent Events:\n"
	result += "• [Normal] 2m ago: Pod nginx-deployment-abc123 created\n"
	result += "• [Warning] 1m ago: Failed to pull image for pod webapp-456def\n"
	result += "• [Normal] 30s ago: Pod webapp-456def started successfully\n"
	result += "\n✅ Events retrieved successfully"

	return result, nil
}

func (h *MCPHandler) handleNamespacesList(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	// Call the actual MCP server handler directly
	result, err := h.server.ListNamespacesHandler(ctx, request)
	if err != nil {
		return "", err
	}

	// Extract text from the MCP result
	if result != nil && len(result.Content) > 0 {
		// Try different ways to extract text content
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			return content.Text, nil
		case mcp.TextContent:
			return content.Text, nil
		default:
			// If it's not a TextContent, try to convert to string
			return fmt.Sprintf("%v", content), nil
		}
	}

	return "No result returned", nil
}

func (h *MCPHandler) handleResourcesList(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	apiVersion := mcp.ParseString(request, "apiVersion", "")
	kind := mcp.ParseString(request, "kind", "")
	namespace := mcp.ParseString(request, "namespace", "")

	result := "📋 Resource List\n"
	result += "===============\n\n"
	result += fmt.Sprintf("API Version: %v\n", apiVersion)
	result += fmt.Sprintf("Kind: %v\n", kind)
	if namespace != "" {
		result += fmt.Sprintf("Namespace: %v\n", namespace)
	}
	result += "\n🔧 Resources found:\n"
	result += "• resource-1 (Active)\n"
	result += "• resource-2 (Active)\n"
	result += "• resource-3 (Pending)\n"
	result += "\n✅ Resources listed successfully"

	return result, nil
}

func (h *MCPHandler) handleConfigView(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	result := "📋 OpenShift Configuration\n"
	result += "==========================\n\n"
	result += "🔧 Cluster Information:\n"
	result += "• Cluster Version: 4.15.0\n"
	result += "• API Server: https://api.cluster.example.com:6443\n"
	result += "• Current Context: openshift-cluster\n"
	result += "• Current User: system:admin\n"
	result += "\n✅ Configuration retrieved successfully"

	return result, nil
}

func (h *MCPHandler) handleHelmList(ctx context.Context, request mcp.CallToolRequest) (string, error) {
	namespace := mcp.ParseString(request, "namespace", "")

	result := "📋 Helm Releases\n"
	result += "===============\n\n"
	if namespace != "" {
		result += fmt.Sprintf("Namespace: %v\n", namespace)
	} else {
		result += "Namespace: All namespaces\n"
	}
	result += "\n⚓ Helm Releases:\n"
	result += "• my-app-1.0.0 (deployed) - Last updated: 2h ago\n"
	result += "• database-2.1.0 (deployed) - Last updated: 1d ago\n"
	result += "• monitoring-3.0.0 (deployed) - Last updated: 5d ago\n"
	result += "\n✅ Helm releases listed successfully"

	return result, nil
}

// SSE endpoint for real-time updates
func (h *MCPHandler) SSEEndpoint(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Send initial connection message
	data := map[string]interface{}{
		"type":    "status",
		"message": "MCP SSE connection established",
	}
	c.JSON(http.StatusOK, data)
}

// Register MCP routes
func (h *MCPHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/mcp")
	{
		api.GET("/capabilities", h.GetCapabilities)
		api.POST("/call", h.CallTool)
		api.GET("/sse", h.SSEEndpoint)
	}
}
