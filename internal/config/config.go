package config

import (
	"fmt"
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

	// MCP configuration
	MCP MCPConfig `mapstructure:"mcp"`

	// Server configuration nested struct
	Server ServerConfig `mapstructure:"server"`

	// LLM configuration
	LLM LLMConfig `mapstructure:"llm"`

	// Planning configuration
	Planning PlanningConfig `mapstructure:"planning"`
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	Provider string       `mapstructure:"provider"`
	OpenAI   OpenAIConfig `mapstructure:"openai"`
	Gemini   GeminiConfig `mapstructure:"gemini"`
	Ollama   OllamaConfig `mapstructure:"ollama"`
	Claude   ClaudeConfig `mapstructure:"claude"`
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

// GeminiConfig holds Gemini configuration
type GeminiConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
}

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Model    string `mapstructure:"model"`
}

// ClaudeConfig holds Claude configuration
type ClaudeConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

// PlanningConfig holds planning configuration
type PlanningConfig struct {
	EnableLLMPlanning bool           `mapstructure:"enable_llm_planning"`
	FallbackToStatic  bool           `mapstructure:"fallback_to_static"`
	EnableCaching     bool           `mapstructure:"enable_caching"`
	CacheTTL          string         `mapstructure:"cache_ttl"`
	Templates         TemplateConfig `mapstructure:"templates"`
}

// TemplateConfig holds template configuration
type TemplateConfig struct {
	BasePrompt string `mapstructure:"base_prompt"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// MCPConfig holds MCP-specific configuration
type MCPConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Profile    string `mapstructure:"profile"`
	SSEBaseURL string `mapstructure:"sse-base-url"`
	ReadOnly   bool   `mapstructure:"read-only"`
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

	// Also try to read LLM config file
	llmViper := viper.New()
	llmViper.AddConfigPath("./config")
	llmViper.AddConfigPath(".")
	llmViper.SetConfigName("llm_config")
	llmViper.SetConfigType("yaml")

	if err := llmViper.ReadInConfig(); err == nil {
		// Merge LLM config into main config
		if err := v.MergeConfigMap(llmViper.AllSettings()); err != nil {
			return nil, fmt.Errorf("failed to merge LLM config: %w", err)
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

	// MCP defaults
	v.SetDefault("mcp.enabled", true)
	v.SetDefault("mcp.profile", "sre")
	v.SetDefault("mcp.read-only", false)

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8080")
}
