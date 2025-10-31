package azure

import (
	"context"
	"fmt"
	"os"

	"github.com/ylchen07/smart-keyvault/internal/provider"
	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// Provider implements the provider.Provider interface for Azure KeyVault
type Provider struct {
	client *Client
}

// NewProvider creates a new Azure KeyVault provider
// Configuration options:
//   - "subscription_id" (string): Azure subscription ID
//
// If subscription_id is not provided in config, it will attempt to read from:
//   - AZURE_SUBSCRIPTION_ID environment variable
//   - Default Azure CLI subscription (via `az account show`)
func NewProvider(cfg *provider.Config) (provider.Provider, error) {
	subscriptionID := ""

	// Try to get subscription ID from config
	if cfg != nil && cfg.Settings != nil {
		if v, ok := cfg.Settings["subscription_id"].(string); ok {
			subscriptionID = v
		}
	}

	// Fallback to environment variable
	if subscriptionID == "" {
		subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	}

	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required for Azure provider (set via config or AZURE_SUBSCRIPTION_ID env var)")
	}

	client, err := NewClient(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}

	return &Provider{
		client: client,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "azure"
}

// ListVaults returns all accessible Azure Key Vaults
func (p *Provider) ListVaults(ctx context.Context) ([]*models.Vault, error) {
	return p.client.ListVaults(ctx)
}

// ListSecrets returns all secrets in a specific vault
func (p *Provider) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
	return p.client.ListSecrets(ctx, vaultName)
}

// GetSecret retrieves a specific secret value
func (p *Provider) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
	return p.client.GetSecret(ctx, vaultName, secretName)
}

// SupportsFeature checks if the provider supports a specific feature
func (p *Provider) SupportsFeature(feature provider.Feature) bool {
	switch feature {
	case provider.FeatureVersioning, provider.FeatureTags:
		return true
	default:
		return false
	}
}
