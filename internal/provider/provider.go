package provider

import (
	"context"

	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// Provider represents a secrets management backend
type Provider interface {
	// Name returns the provider name (e.g., "azure", "hashicorp")
	Name() string

	// ListVaults returns all accessible vaults/backends
	ListVaults(ctx context.Context) ([]*models.Vault, error)

	// ListSecrets returns all secrets in a specific vault
	ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error)

	// GetSecret retrieves a specific secret value
	GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error)

	// SupportsFeature checks if provider supports a feature
	SupportsFeature(feature Feature) bool
}

// Feature represents optional provider capabilities
type Feature int

const (
	// FeatureVersioning indicates the provider supports secret versioning
	FeatureVersioning Feature = iota
	// FeatureMetadata indicates the provider supports rich metadata
	FeatureMetadata
	// FeatureTags indicates the provider supports tagging
	FeatureTags
)

// Config holds provider-specific configuration
type Config struct {
	Name     string                 // Provider name
	Enabled  bool                   // Whether provider is enabled
	Default  bool                   // Default provider
	Settings map[string]interface{} // Provider-specific settings
}
