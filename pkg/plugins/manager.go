package plugins

import (
	"fmt"
	"strings"

	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/types"
)

// Plugin represents a diagnostic plugin interface
type Plugin interface {
	Name() string
	Description() string
	CanHandle(prompt string) bool
	Handle(prompt string, context map[string]interface{}) (*types.Analysis, error)
}

// Manager manages plugins
type Manager struct {
	plugins []Plugin
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make([]Plugin, 0),
	}
}

// Register registers a new plugin
func (m *Manager) Register(plugin Plugin) {
	m.plugins = append(m.plugins, plugin)
}

// GetPlugin finds a plugin that can handle the prompt
func (m *Manager) GetPlugin(prompt string) Plugin {
	for _, plugin := range m.plugins {
		if plugin.CanHandle(prompt) {
			return plugin
		}
	}
	return nil
}

// ListPlugins returns all registered plugins
func (m *Manager) ListPlugins() []Plugin {
	return m.plugins
}

// CrashLoopPlugin handles crashloop diagnostics
type CrashLoopPlugin struct{}

func (p *CrashLoopPlugin) Name() string {
	return "crashloop-handler"
}

func (p *CrashLoopPlugin) Description() string {
	return "Handles pod crashloop diagnostics and troubleshooting"
}

func (p *CrashLoopPlugin) CanHandle(prompt string) bool {
	lowerPrompt := strings.ToLower(prompt)
	return strings.Contains(lowerPrompt, "crashloop") ||
		strings.Contains(lowerPrompt, "crash loop") ||
		strings.Contains(lowerPrompt, "crashing")
}

func (p *CrashLoopPlugin) Handle(prompt string, context map[string]interface{}) (*types.Analysis, error) {
	// Specialized crashloop analysis
	analysis := &types.Analysis{
		Query:      prompt,
		Confidence: 0.85,
		Severity:   "High",
		RootCauses: []types.RootCause{
			{
				Description: "Application container failing to start",
				Confidence:  0.9,
				Evidence:    "Pod is in CrashLoopBackOff state",
			},
		},
		Actions: []types.RecommendedAction{
			{
				Description: "Check container logs for startup errors",
				Priority:    "High",
				Command:     "oc logs <pod-name> --previous",
				Risk:        "Low",
			},
			{
				Description: "Verify container image and dependencies",
				Priority:    "High",
				Command:     "oc describe pod <pod-name>",
				Risk:        "Low",
			},
		},
	}

	analysis.Response = fmt.Sprintf("🔍 **CrashLoop Diagnostic Analysis**\n\n"+
		"Detected pod crashloop issue. This typically indicates:\n"+
		"1. Application startup failures\n"+
		"2. Missing dependencies or configuration\n"+
		"3. Resource constraints\n\n"+
		"**Immediate Actions:**\n"+
		"• Check recent logs: `oc logs <pod-name> --previous`\n"+
		"• Review pod events: `oc describe pod <pod-name>`\n"+
		"• Verify resource limits and requests\n\n"+
		"Confidence: %.0f%% | Severity: %s",
		analysis.Confidence*100, analysis.Severity)

	return analysis, nil
}

// NetworkPlugin handles network-related issues
type NetworkPlugin struct{}

func (p *NetworkPlugin) Name() string {
	return "network-handler"
}

func (p *NetworkPlugin) Description() string {
	return "Handles network connectivity and service discovery issues"
}

func (p *NetworkPlugin) CanHandle(prompt string) bool {
	lowerPrompt := strings.ToLower(prompt)
	return strings.Contains(lowerPrompt, "network") ||
		strings.Contains(lowerPrompt, "connectivity") ||
		strings.Contains(lowerPrompt, "service") ||
		strings.Contains(lowerPrompt, "dns")
}

func (p *NetworkPlugin) Handle(prompt string, context map[string]interface{}) (*types.Analysis, error) {
	analysis := &types.Analysis{
		Query:      prompt,
		Confidence: 0.75,
		Severity:   "Medium",
		RootCauses: []types.RootCause{
			{
				Description: "Network connectivity issue",
				Confidence:  0.8,
				Evidence:    "Service or DNS resolution problems detected",
			},
		},
		Actions: []types.RecommendedAction{
			{
				Description: "Test service connectivity",
				Priority:    "High",
				Command:     "oc get svc && oc get endpoints",
				Risk:        "Low",
			},
			{
				Description: "Check network policies",
				Priority:    "Medium",
				Command:     "oc get networkpolicy",
				Risk:        "Low",
			},
		},
	}

	analysis.Response = fmt.Sprintf("🌐 **Network Diagnostic Analysis**\n\n"+
		"Detected network-related issue. Common causes:\n"+
		"1. Service configuration problems\n"+
		"2. DNS resolution issues\n"+
		"3. Network policy restrictions\n\n"+
		"**Troubleshooting Steps:**\n"+
		"• Verify services and endpoints\n"+
		"• Test DNS resolution\n"+
		"• Check network policies\n\n"+
		"Confidence: %.0f%% | Severity: %s",
		analysis.Confidence*100, analysis.Severity)

	return analysis, nil
}

// InitializeDefaultPlugins initializes default plugins
func InitializeDefaultPlugins(manager *Manager) {
	manager.Register(&CrashLoopPlugin{})
	manager.Register(&NetworkPlugin{})
}
