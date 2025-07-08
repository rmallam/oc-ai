package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ExtractPodName extracts pod name from various input formats
func ExtractPodName(input string) string {
	// Remove common prefixes
	input = strings.TrimSpace(input)
	
	// Handle "pod/<name>" format
	if strings.HasPrefix(input, "pod/") {
		return strings.TrimPrefix(input, "pod/")
	}
	
	// Handle direct pod name
	return input
}

// ExtractNamespace extracts namespace from input or returns default
func ExtractNamespace(input, defaultNS string) string {
	// Look for namespace patterns in input
	nsRegex := regexp.MustCompile(`(?i)namespace[:\s]+([a-zA-Z0-9\-]+)`)
	if matches := nsRegex.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}
	
	// Look for -n flag pattern
	flagRegex := regexp.MustCompile(`-n\s+([a-zA-Z0-9\-]+)`)
	if matches := flagRegex.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}
	
	return defaultNS
}

// ValidateResourceName validates Kubernetes resource name
func ValidateResourceName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}
	
	// Kubernetes naming rules
	validName := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid resource name: %s", name)
	}
	
	if len(name) > 63 {
		return fmt.Errorf("resource name too long: %s", name)
	}
	
	return nil
}

// FormatConfidence formats confidence score as percentage
func FormatConfidence(confidence float64) string {
	return fmt.Sprintf("%.0f%%", confidence*100)
}

// FormatSeverity formats severity with appropriate emoji
func FormatSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴 CRITICAL"
	case "high":
		return "🟠 HIGH"
	case "medium":
		return "🟡 MEDIUM"
	case "low":
		return "🟢 LOW"
	default:
		return "⚪ UNKNOWN"
	}
}

// SanitizeInput sanitizes user input for logging
func SanitizeInput(input string) string {
	// Remove potential sensitive information
	input = regexp.MustCompile(`(?i)(password|token|key|secret)[\s:=]+\S+`).ReplaceAllString(input, "$1=***")
	return input
}

// ParseDuration parses duration strings with units
func ParseDuration(duration string) (string, error) {
	// Simple duration parsing - can be enhanced
	if duration == "" {
		return "5m", nil
	}
	
	validDuration := regexp.MustCompile(`^\d+[smhd]$`)
	if !validDuration.MatchString(duration) {
		return "", fmt.Errorf("invalid duration format: %s", duration)
	}
	
	return duration, nil
}
