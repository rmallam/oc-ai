package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`

	// Logging
	Debug bool `mapstructure:"debug"`

	// LLM configuration
	GeminiAPIKey string `mapstructure:"gemini-api-key"`
	Model        string `mapstructure:"model"`
	LLMProvider  string `mapstructure:"llm-provider"`

	// Kubernetes configuration
	Kubeconfig string `mapstructure:"kubeconfig"`

	// Database configuration
	DatabasePath string `mapstructure:"database-path"`

	// Plugin configuration
	PluginsDir string `mapstructure:"plugins-dir"`

	// Decision engine configuration
	ConfidenceThreshold float64 `mapstructure:"confidence-threshold"`
	EvidenceLimit       int     `mapstructure:"evidence-limit"`
}

// Load loads configuration from various sources
func Load(cmd *cobra.Command) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Bind command-line flags
	if err := v.BindPFlags(cmd.PersistentFlags()); err != nil {
		return nil, err
	}

	// Set up config file search paths
	configFile, _ := cmd.PersistentFlags().GetString("config")
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		configDir := filepath.Join(home, ".config", "openshift-mcp")
		v.AddConfigPath(configDir)
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Read environment variables
	v.SetEnvPrefix("OPENSHIFT_MCP")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Override with environment variables
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		cfg.GeminiAPIKey = apiKey
	}

	if model := os.Getenv("GEMINI_MODEL"); model != "" {
		cfg.Model = model
	}

	// Set default kubeconfig if not specified
	if cfg.Kubeconfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.Kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Set default database path if not specified
	if cfg.DatabasePath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfg.DatabasePath = filepath.Join(home, ".config", "openshift-mcp", "memory.db")
		}
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", "8080")
	v.SetDefault("debug", false)
	v.SetDefault("model", "gemini-2.0-flash-001")
	v.SetDefault("llm-provider", "gemini")
	v.SetDefault("confidence-threshold", 0.7)
	v.SetDefault("evidence-limit", 10)
}
