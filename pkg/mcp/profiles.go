package mcp

import (
	"slices"

	"github.com/mark3labs/mcp-go/server"
)

type Profile interface {
	GetName() string
	GetDescription() string
	GetTools(s *Server) []server.ServerTool
}

var Profiles = []Profile{
	&OpenShiftSREProfile{},
	&OpenShiftDeveloperProfile{},
	&OpenShiftAdminProfile{},
}

var ProfileNames []string

func ProfileFromString(name string) Profile {
	for _, profile := range Profiles {
		if profile.GetName() == name {
			return profile
		}
	}
	return &OpenShiftSREProfile{} // Default
}

// OpenShift SRE Profile - Focus on diagnostics and troubleshooting
type OpenShiftSREProfile struct{}

func (p *OpenShiftSREProfile) GetName() string {
	return "sre"
}

func (p *OpenShiftSREProfile) GetDescription() string {
	return "OpenShift SRE profile focused on cluster diagnostics, troubleshooting, and monitoring"
}

func (p *OpenShiftSREProfile) GetTools(s *Server) []server.ServerTool {
	return slices.Concat(
		s.initConfiguration(),
		s.initOpenShiftTools(),
		s.initPods(),
		s.initResources(),
		s.initEvents(),
		s.initNamespaces(),
		s.initWriteOperations(),
		s.initGitTools(),
		s.initArgocdTools(), // Add ArgoCD tools
		s.initDiagnostics(),
		s.initMonitoring(),
		s.initWriteOperations(), // Add write operations for SRE
	)
}

// OpenShift Developer Profile - Focus on development and deployment
type OpenShiftDeveloperProfile struct{}

func (p *OpenShiftDeveloperProfile) GetName() string {
	return "developer"
}

func (p *OpenShiftDeveloperProfile) GetDescription() string {
	return "OpenShift Developer profile focused on application deployment and development workflows"
}

func (p *OpenShiftDeveloperProfile) GetTools(s *Server) []server.ServerTool {
	return slices.Concat(
		s.initConfiguration(),
		s.initPods(),
		s.initResources(),
		s.initWriteOperations(),
		s.initGitTools(),
		s.initArgocdTools(), // Add ArgoCD tools
		s.initHelm(),
		s.initImageStreams(),
		s.initBuildConfigs(),
		s.initDeploymentConfigs(),
	)
}

// OpenShift Admin Profile - Focus on cluster administration
type OpenShiftAdminProfile struct{}

func (p *OpenShiftAdminProfile) GetName() string {
	return "admin"
}

func (p *OpenShiftAdminProfile) GetDescription() string {
	return "OpenShift Administrator profile with full cluster management capabilities"
}

func (p *OpenShiftAdminProfile) GetTools(s *Server) []server.ServerTool {
	return slices.Concat(
		s.initConfiguration(),
		s.initOpenShiftTools(),
		s.initPods(),
		s.initResources(),
		s.initEvents(),
		s.initNamespaces(),
		s.initWriteOperations(),
		s.initGitTools(),
		s.initArgocdTools(), // Add ArgoCD tools
		s.initHelm(),
		s.initDiagnostics(),
		s.initMonitoring(),
		s.initImageStreams(),
		s.initBuildConfigs(),
		s.initDeploymentConfigs(),
		s.initClusterAdmin(),
	)
}

// Additional OpenShift-specific tool initializers
func (s *Server) initDiagnostics() []server.ServerTool {
	// Diagnostic tools implementation
	return []server.ServerTool{}
}

func (s *Server) initMonitoring() []server.ServerTool {
	// Monitoring tools implementation
	return []server.ServerTool{}
}

func (s *Server) initImageStreams() []server.ServerTool {
	// ImageStream tools implementation
	return []server.ServerTool{}
}

func (s *Server) initBuildConfigs() []server.ServerTool {
	// BuildConfig tools implementation
	return []server.ServerTool{}
}

func (s *Server) initDeploymentConfigs() []server.ServerTool {
	// DeploymentConfig tools implementation
	return []server.ServerTool{}
}

func (s *Server) initClusterAdmin() []server.ServerTool {
	// Cluster admin tools implementation
	return []server.ServerTool{}
}

func init() {
	ProfileNames = make([]string, 0)
	for _, profile := range Profiles {
		ProfileNames = append(ProfileNames, profile.GetName())
	}
}
