package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ylchen07/smart-keyvault/internal/provider"
	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// Provider implements the provider.Provider interface for Azure KeyVault
type Provider struct {
	client *Client
}

// NewProvider creates a new Azure KeyVault provider
func NewProvider(cfg *provider.Config) (provider.Provider, error) {
	return &Provider{
		client: NewClient(),
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "azure"
}

// ListVaults returns all accessible Azure Key Vaults
func (p *Provider) ListVaults(ctx context.Context) ([]*models.Vault, error) {
	output, err := p.client.Execute(ctx, "keyvault", "list", "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list vaults: %w", err)
	}

	// Parse Azure response
	var azVaults []struct {
		Name          string `json:"name"`
		Location      string `json:"location"`
		ResourceGroup string `json:"resourceGroup"`
	}

	if err := json.Unmarshal(output, &azVaults); err != nil {
		return nil, fmt.Errorf("failed to parse vault list: %w", err)
	}

	// Convert to common models
	vaults := make([]*models.Vault, len(azVaults))
	for i, v := range azVaults {
		vaults[i] = &models.Vault{
			Name:     v.Name,
			Provider: "azure",
			Metadata: map[string]string{
				"location":      v.Location,
				"resourceGroup": v.ResourceGroup,
			},
		}
	}

	return vaults, nil
}

// ListSecrets returns all secrets in a specific vault
func (p *Provider) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
	output, err := p.client.Execute(ctx, "keyvault", "secret", "list",
		"--vault-name", vaultName,
		"--output", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// Parse Azure response
	var azSecrets []struct {
		Name       string `json:"name"`
		Attributes struct {
			Enabled bool `json:"enabled"`
		} `json:"attributes"`
	}

	if err := json.Unmarshal(output, &azSecrets); err != nil {
		return nil, fmt.Errorf("failed to parse secret list: %w", err)
	}

	// Convert to common models, filter only enabled secrets
	secrets := make([]*models.Secret, 0)
	for _, s := range azSecrets {
		if s.Attributes.Enabled {
			secrets = append(secrets, &models.Secret{
				Name:      s.Name,
				VaultName: vaultName,
				Provider:  "azure",
				Enabled:   true,
			})
		}
	}

	return secrets, nil
}

// GetSecret retrieves a specific secret value
func (p *Provider) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
	output, err := p.client.Execute(ctx, "keyvault", "secret", "show",
		"--vault-name", vaultName,
		"--name", secretName,
		"--output", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Parse Azure response
	var azSecret struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := json.Unmarshal(output, &azSecret); err != nil {
		return nil, fmt.Errorf("failed to parse secret: %w", err)
	}

	return &models.SecretValue{
		Name:      azSecret.Name,
		Value:     azSecret.Value,
		VaultName: vaultName,
		Provider:  "azure",
	}, nil
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
