package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// initArgocdTools initializes ArgoCD-specific tools
func initArgocdTools(s *Server) []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("create_argocd_application",
			mcp.WithDescription("Create an ArgoCD Application resource for GitOps deployment"),
			mcp.WithString("app_name", mcp.Description("Name of the ArgoCD application"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace where the ArgoCD application will be created"), mcp.Required()),
			mcp.WithString("repo_url", mcp.Description("Git repository URL containing the application manifests"), mcp.Required()),
			mcp.WithString("path", mcp.Description("Path within the repository to the application manifests"), mcp.Required()),
			mcp.WithString("target_revision", mcp.Description("Target revision/branch to deploy (default: HEAD)")),
			mcp.WithString("destination_server", mcp.Description("Destination Kubernetes server URL (default: https://kubernetes.default.svc)")),
			mcp.WithString("destination_namespace", mcp.Description("Destination namespace for the application resources"), mcp.Required()),
			mcp.WithString("automated", mcp.Description("Enable automated sync (true/false, default: false)")),
			mcp.WithTitleAnnotation("ArgoCD: Create Application"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.createArgocdApplicationHandler)},

		{Tool: mcp.NewTool("create_argocd_manifest_bundle",
			mcp.WithDescription("Create a complete ArgoCD-compatible manifest bundle for an application"),
			mcp.WithString("app_name", mcp.Description("Name of the application"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Kubernetes namespace for the application"), mcp.Required()),
			mcp.WithString("environment", mcp.Description("Target environment (dev/staging/prod/base)"), mcp.Required()),
			mcp.WithString("image", mcp.Description("Container image for the application"), mcp.Required()),
			mcp.WithString("replicas", mcp.Description("Number of replicas for the deployment"), mcp.Required()),
			mcp.WithString("config_data", mcp.Description("JSON string of configuration data for ConfigMap (optional)")),
			mcp.WithTitleAnnotation("ArgoCD: Create Manifest Bundle"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.createArgocdManifestBundleHandler)},

		{Tool: mcp.NewTool("create_argocd_app_of_apps",
			mcp.WithDescription("Create an ArgoCD App of Apps resource for managing multiple applications"),
			mcp.WithString("environment", mcp.Description("Environment name (dev/staging/prod)"), mcp.Required()),
			mcp.WithString("repo_url", mcp.Description("Git repository URL containing the application definitions"), mcp.Required()),
			mcp.WithString("applications", mcp.Description("JSON array of application names to include"), mcp.Required()),
			mcp.WithTitleAnnotation("ArgoCD: Create App of Apps"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.createArgocdAppOfAppsHandler)},

		{Tool: mcp.NewTool("init_argocd_directory",
			mcp.WithDescription("Initialize ArgoCD-compatible directory structure in the Git repository"),
			mcp.WithTitleAnnotation("ArgoCD: Initialize Directory"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.initArgocdDirectoryHandler)},

		{Tool: mcp.NewTool("list_argocd_applications",
			mcp.WithDescription("List all ArgoCD applications in the repository"),
			mcp.WithTitleAnnotation("ArgoCD: List Applications"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.listArgocdApplicationsHandler)},

		{Tool: mcp.NewTool("get_argocd_application_manifests",
			mcp.WithDescription("Get all manifests for a specific ArgoCD application"),
			mcp.WithString("app_name", mcp.Description("Name of the application"), mcp.Required()),
			mcp.WithString("environment", mcp.Description("Environment (dev/staging/prod/base)"), mcp.Required()),
			mcp.WithTitleAnnotation("ArgoCD: Get Manifests"),
			mcp.WithReadOnlyHintAnnotation(true),
		), Handler: server.ToolHandlerFunc(s.getArgocdApplicationManifestsHandler)},

		{Tool: mcp.NewTool("commit_argocd_changes",
			mcp.WithDescription("Commit ArgoCD changes with structured commit message"),
			mcp.WithString("app_name", mcp.Description("Name of the application"), mcp.Required()),
			mcp.WithString("environment", mcp.Description("Environment (dev/staging/prod)"), mcp.Required()),
			mcp.WithString("action", mcp.Description("Action performed (create/update/delete/scale)"), mcp.Required()),
			mcp.WithString("message", mcp.Description("Additional commit message details"), mcp.Required()),
			mcp.WithTitleAnnotation("ArgoCD: Commit Changes"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.commitArgocdChangesHandler)},
	}
}

// Enhanced tools that replace existing ones for ArgoCD compatibility
func initEnhancedArgocdTools(s *Server) []server.ServerTool {
	return []server.ServerTool{
		{Tool: mcp.NewTool("scale_deployment_argocd",
			mcp.WithDescription("Scale deployment and create ArgoCD-compatible YAML record"),
			mcp.WithString("deployment_name", mcp.Description("Name of the deployment to scale"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Namespace of the deployment"), mcp.Required()),
			mcp.WithString("replicas", mcp.Description("Number of replicas to scale to"), mcp.Required()),
			mcp.WithTitleAnnotation("ArgoCD: Scale Deployment"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.enhancedScaleDeploymentHandler)},

		{Tool: mcp.NewTool("create_resource_argocd",
			mcp.WithDescription("Create Kubernetes resource and generate ArgoCD-compatible manifest"),
			mcp.WithString("resource_type", mcp.Description("Type of resource (deployment, service, configmap, namespace)"), mcp.Required()),
			mcp.WithString("name", mcp.Description("Name of the resource"), mcp.Required()),
			mcp.WithString("namespace", mcp.Description("Kubernetes namespace"), mcp.Required()),
			mcp.WithString("image", mcp.Description("Container image (for deployments)")),
			mcp.WithString("replicas", mcp.Description("Number of replicas (for deployments)")),
			mcp.WithString("data", mcp.Description("JSON string of data (for configmaps)")),
			mcp.WithTitleAnnotation("ArgoCD: Create Resource"),
			mcp.WithDestructiveHintAnnotation(false),
		), Handler: server.ToolHandlerFunc(s.enhancedCreateResourceHandler)},
	}
}
