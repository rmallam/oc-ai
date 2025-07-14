package decision

import (
	"testing"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/rakeshkumarmallam/openshift-mcp-go/pkg/models"
)

func TestEngine_isDiagnosticQuery(t *testing.T) {
	cfg := &config.Config{}
	engine := &Engine{config: cfg}

	tests := []struct {
		prompt   string
		expected bool
	}{
		{"check why pod is crashlooping", true},
		{"pod is failing to start", true},
		{"troubleshoot nginx deployment", true},
		{"list all pods", false},
		{"create a new service", false},
		{"debug application logs", true},
	}

	for _, test := range tests {
		result := engine.isDiagnosticQuery(test.prompt)
		if result != test.expected {
			t.Errorf("isDiagnosticQuery(%q) = %v, expected %v", test.prompt, result, test.expected)
		}
	}
}

func TestEngine_extractResourceInfo(t *testing.T) {
	cfg := &config.Config{}
	engine := &Engine{config: cfg}

	tests := []struct {
		prompt   string
		expected map[string]string
	}{
		{
			"check pod nginx-123 in namespace default",
			map[string]string{"pod_name": "nginx-123", "namespace": "default"},
		},
		{
			"deployment webapp-prod is failing",
			map[string]string{"deployment": "webapp-prod"},
		},
		{
			"troubleshoot pod app-server",
			map[string]string{"pod_name": "app-server"},
		},
	}

	for _, test := range tests {
		result := engine.extractResourceInfo(test.prompt)
		for key, expectedValue := range test.expected {
			if result[key] != expectedValue {
				t.Errorf("extractResourceInfo(%q)[%s] = %v, expected %v",
					test.prompt, key, result[key], expectedValue)
			}
		}
	}
}

func TestEngine_calculateConfidence(t *testing.T) {
	cfg := &config.Config{}
	engine := &Engine{config: cfg}

	rootCauses := []models.RootCause{
		{Description: "Test cause 1", Confidence: 0.8},
		{Description: "Test cause 2", Confidence: 0.9},
	}

	evidence := []models.Evidence{
		{Type: "logs", Content: "Error message"},
		{Type: "status", Content: "Pod status"},
		{Type: "events", Content: "Event data"},
	}

	confidence := engine.calculateConfidence(rootCauses, evidence)
	expectedMin := 0.8 // Should be at least the average of root cause confidences

	if confidence < expectedMin {
		t.Errorf("calculateConfidence() = %v, expected >= %v", confidence, expectedMin)
	}

	if confidence > 1.0 {
		t.Errorf("calculateConfidence() = %v, expected <= 1.0", confidence)
	}
}

func TestEngine_calculateSeverity(t *testing.T) {
	cfg := &config.Config{}
	engine := &Engine{config: cfg}

	tests := []struct {
		evidence []models.Evidence
		expected string
	}{
		{
			[]models.Evidence{
				{Content: "Pod is in CrashLoopBackOff state"},
			},
			"High",
		},
		{
			[]models.Evidence{
				{Content: "ImagePullBackOff error"},
			},
			"High",
		},
		{
			[]models.Evidence{
				{Content: "Normal pod operation"},
			},
			"Medium",
		},
	}

	for _, test := range tests {
		rootCauses := []models.RootCause{{Description: "Test cause"}}
		result := engine.calculateSeverity(rootCauses, test.evidence)
		if result != test.expected {
			t.Errorf("calculateSeverity() = %v, expected %v", result, test.expected)
		}
	}
}
