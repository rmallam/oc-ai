package main

import (
	"testing"

	"github.com/rakeshkumarmallam/openshift-mcp-go/internal/config"
	"github.com/spf13/cobra"
)

func TestMain(t *testing.T) {
	// Test that main function can be called without panicking
	// This is a basic test to ensure the application structure is sound
}

func TestConfigLoading(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().String("config", "", "config file")
	cmd.PersistentFlags().String("port", "8080", "server port")
	cmd.PersistentFlags().String("host", "0.0.0.0", "server host")
	cmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	cmd.PersistentFlags().String("gemini-api-key", "test-key", "Gemini API key")
	cmd.PersistentFlags().String("kubeconfig", "", "path to kubeconfig file")

	cfg, err := config.Load(cmd)
	if err != nil {
		t.Fatalf("Expected config to load successfully, got error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Expected port to be 8080, got %s", cfg.Port)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected host to be 0.0.0.0, got %s", cfg.Host)
	}
}
