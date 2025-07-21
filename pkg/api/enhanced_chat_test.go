package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLLMIntegration(t *testing.T) {
	handler := &EnhancedChatHandler{}

	testCases := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "Pod troubleshooting",
			query:    "fix failing pods in debugger namespace",
			expected: "troubleshooting",
		},
		{
			name:     "Deployment scaling",
			query:    "scale deployment to 3 replicas",
			expected: "maintenance",
		},
		{
			name:     "General exploration",
			query:    "what's in the cluster",
			expected: "exploration",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test LLM planning (mock response)
			llmResponse, err := handler.callLLMForPlanning(tc.query)
			if err != nil {
				t.Errorf("LLM planning failed: %v", err)
				return
			}

			// Verify response is valid JSON
			var planData map[string]interface{}
			if err := json.Unmarshal([]byte(llmResponse), &planData); err != nil {
				t.Errorf("LLM response is not valid JSON: %v", err)
				return
			}

			// Verify required fields
			if category, ok := planData["category"].(string); !ok || category != tc.expected {
				t.Errorf("Expected category %s, got %v", tc.expected, planData["category"])
			}

			if _, ok := planData["steps"].([]interface{}); !ok {
				t.Errorf("Expected steps array in response")
			}

			// Test static pattern matching fallback
			plan, err := handler.planWithStaticPatterns(tc.query)
			if err != nil {
				t.Errorf("Static pattern matching failed: %v", err)
				return
			}

			if plan.Category != tc.expected {
				t.Errorf("Expected category %s, got %s", tc.expected, plan.Category)
			}
		})
	}
}

func TestLLMPromptGeneration(t *testing.T) {
	handler := &EnhancedChatHandler{}

	query := "fix failing pods"
	prompt := handler.buildPlanningPrompt(query)

	// Verify prompt contains required elements
	if !strings.Contains(prompt, "OpenShift/Kubernetes administrator") {
		t.Error("Prompt should contain role description")
	}

	if !strings.Contains(prompt, query) {
		t.Error("Prompt should contain user query")
	}

	if !strings.Contains(prompt, "Available Tools:") {
		t.Error("Prompt should contain available tools")
	}

	if !strings.Contains(prompt, "JSON response") {
		t.Error("Prompt should specify JSON response format")
	}
}

func TestLLMResponseParsing(t *testing.T) {
	handler := &EnhancedChatHandler{}

	validResponse := `{
  "description": "Test operation",
  "category": "troubleshooting",
  "complexity": "medium",
  "steps": [
    {
      "action": "list_pods",
      "tool": "list_pods",
      "parameters": {"namespace": "debugger"},
      "description": "List pods",
      "required": true
    }
  ]
}`

	plan, err := handler.parseLLMPlanResponse(validResponse)
	if err != nil {
		t.Errorf("Failed to parse valid response: %v", err)
		return
	}

	if plan.Description != "Test operation" {
		t.Errorf("Expected description 'Test operation', got %s", plan.Description)
	}

	if plan.Category != "troubleshooting" {
		t.Errorf("Expected category 'troubleshooting', got %s", plan.Category)
	}

	if len(plan.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(plan.Steps))
	}

	if plan.Steps[0].Action != "list_pods" {
		t.Errorf("Expected action 'list_pods', got %s", plan.Steps[0].Action)
	}
}

func TestLLMFallbackBehavior(t *testing.T) {
	handler := &EnhancedChatHandler{}

	// Test that planExecution falls back to static patterns when LLM fails
	// This would need to be implemented based on your actual planExecution logic
	query := "fix failing pods"

	// Test static pattern directly
	plan, err := handler.planWithStaticPatterns(query)
	if err != nil {
		t.Errorf("Static pattern fallback failed: %v", err)
		return
	}

	if plan.Category != "troubleshooting" {
		t.Errorf("Expected troubleshooting category, got %s", plan.Category)
	}

	if len(plan.Steps) == 0 {
		t.Error("Expected steps in plan")
	}
}

func TestIntelligentMockResponse(t *testing.T) {
	handler := &EnhancedChatHandler{}

	testCases := []struct {
		prompt   string
		expected string
	}{
		{
			prompt:   "User Query: fix failing pods",
			expected: "troubleshooting",
		},
		{
			prompt:   "User Query: scale deployment",
			expected: "maintenance",
		},
		{
			prompt:   "User Query: what's running",
			expected: "exploration",
		},
	}

	for _, tc := range testCases {
		response, err := handler.generateIntelligentMockResponse(tc.prompt)
		if err != nil {
			t.Errorf("Mock response generation failed: %v", err)
			continue
		}

		var planData map[string]interface{}
		if err := json.Unmarshal([]byte(response), &planData); err != nil {
			t.Errorf("Mock response is not valid JSON: %v", err)
			continue
		}

		if category, ok := planData["category"].(string); !ok || category != tc.expected {
			t.Errorf("Expected category %s, got %v", tc.expected, planData["category"])
		}
	}
}
