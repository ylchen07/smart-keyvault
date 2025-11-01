package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigDir is the default directory for config files
	DefaultConfigDir = ".config/smart-keyvault"
	// DefaultConfigName is the default config file name (without extension)
	DefaultConfigName = "config"
)

var (
	// envVarPattern matches ${VAR_NAME} or $VAR_NAME patterns
	envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Z_][A-Z0-9_]*)`)
)

// Load loads configuration from file, environment variables, and defaults
// Configuration precedence (highest to lowest):
// 1. Environment variables (prefixed with SMART_KEYVAULT_)
// 2. Config file (~/.config/smart-keyvault/config.yaml)
// 3. Default values
func Load() (*Config, error) {
	v := viper.New()

	// Set config file location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, DefaultConfigDir)
	v.SetConfigName(DefaultConfigName)
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	// Set environment variable prefix and automatic binding
	v.SetEnvPrefix("SMART_KEYVAULT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Read config file (optional - don't error if it doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file found but has error
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found - continue with defaults and env vars
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Perform environment variable substitution
	if err := substituteEnvVars(&cfg); err != nil {
		return nil, fmt.Errorf("failed to substitute environment variables: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from a specific file path
func LoadFromFile(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)

	// Set environment variable prefix
	v.SetEnvPrefix("SMART_KEYVAULT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Perform environment variable substitution
	if err := substituteEnvVars(&cfg); err != nil {
		return nil, fmt.Errorf("failed to substitute environment variables: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// FZF defaults
	v.SetDefault("fzf.height", "40%")
	v.SetDefault("fzf.border", "rounded")
	v.SetDefault("fzf.preview", false)

	// Filters defaults
	v.SetDefault("filters.enabled_only", true)

	// Provider defaults
	v.SetDefault("providers.azure.enabled", true)
	v.SetDefault("providers.hashicorp.enabled", true)
}

// substituteEnvVars replaces ${VAR} or $VAR patterns with environment variable values
func substituteEnvVars(cfg *Config) error {
	// Substitute in Azure instances
	if cfg.Providers.Azure != nil {
		for i := range cfg.Providers.Azure.Instances {
			cfg.Providers.Azure.Instances[i].SubscriptionID = expandEnvVars(cfg.Providers.Azure.Instances[i].SubscriptionID)
		}
	}

	// Substitute in Hashicorp instances
	if cfg.Providers.Hashicorp != nil {
		for i := range cfg.Providers.Hashicorp.Instances {
			inst := &cfg.Providers.Hashicorp.Instances[i]
			inst.Address = expandEnvVars(inst.Address)
			inst.Token = expandEnvVars(inst.Token)
			inst.Namespace = expandEnvVars(inst.Namespace)
		}
	}

	return nil
}

// expandEnvVars expands environment variables in a string
// Supports both ${VAR_NAME} and $VAR_NAME formats
func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name (handle both ${VAR} and $VAR)
		var varName string
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1] // Remove ${ and }
		} else {
			varName = match[1:] // Remove $
		}

		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return original if not found
		return match
	})
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Validate Azure provider instances
	if cfg.Providers.Azure != nil && cfg.Providers.Azure.Enabled {
		if len(cfg.Providers.Azure.Instances) == 0 {
			return fmt.Errorf("azure provider is enabled but has no instances configured")
		}

		for i, inst := range cfg.Providers.Azure.Instances {
			if inst.Name == "" {
				return fmt.Errorf("azure instance at index %d has no name", i)
			}
			if inst.SubscriptionID == "" {
				return fmt.Errorf("azure instance '%s' has no subscription_id", inst.Name)
			}
		}
	}

	// Validate Hashicorp provider instances
	if cfg.Providers.Hashicorp != nil && cfg.Providers.Hashicorp.Enabled {
		if len(cfg.Providers.Hashicorp.Instances) == 0 {
			return fmt.Errorf("hashicorp provider is enabled but has no instances configured")
		}

		for i, inst := range cfg.Providers.Hashicorp.Instances {
			if inst.Name == "" {
				return fmt.Errorf("hashicorp instance at index %d has no name", i)
			}
			if inst.Address == "" {
				return fmt.Errorf("hashicorp instance '%s' has no address", inst.Name)
			}
			if inst.Token == "" {
				return fmt.Errorf("hashicorp instance '%s' has no token", inst.Name)
			}
		}
	}

	return nil
}
