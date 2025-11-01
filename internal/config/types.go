package config

// Config represents the complete application configuration
type Config struct {
	Defaults  Defaults            `mapstructure:"defaults"`
	Providers Providers           `mapstructure:"providers"`
	FZF       FZFConfig           `mapstructure:"fzf"`
	Filters   Filters             `mapstructure:"filters"`
}

// Defaults holds default values for provider and vault selection
type Defaults struct {
	Provider string `mapstructure:"provider"`
	Vault    string `mapstructure:"vault"`
}

// Providers holds configuration for all secret providers
type Providers struct {
	Azure     *AzureConfig     `mapstructure:"azure"`
	Hashicorp *HashicorpConfig `mapstructure:"hashicorp"`
}

// AzureConfig holds Azure KeyVault provider configuration
type AzureConfig struct {
	Enabled   bool             `mapstructure:"enabled"`
	Instances []AzureInstance  `mapstructure:"instances"`
}

// AzureInstance represents a single Azure subscription configuration
type AzureInstance struct {
	Name           string `mapstructure:"name"`
	SubscriptionID string `mapstructure:"subscription_id"`
	Default        bool   `mapstructure:"default"`
}

// HashicorpConfig holds Hashicorp Vault provider configuration
type HashicorpConfig struct {
	Enabled   bool                `mapstructure:"enabled"`
	Instances []HashicorpInstance `mapstructure:"instances"`
}

// HashicorpInstance represents a single Vault server configuration
type HashicorpInstance struct {
	Name      string `mapstructure:"name"`
	Address   string `mapstructure:"address"`
	Token     string `mapstructure:"token"`
	Namespace string `mapstructure:"namespace"`
	Default   bool   `mapstructure:"default"`
}

// FZFConfig holds fzf-tmux display configuration
type FZFConfig struct {
	Height  string `mapstructure:"height"`
	Border  string `mapstructure:"border"`
	Preview bool   `mapstructure:"preview"`
}

// Filters holds filtering options for secrets
type Filters struct {
	EnabledOnly bool `mapstructure:"enabled_only"`
}
