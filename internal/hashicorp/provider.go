package hashicorp

import (
	"context"
	"fmt"
	"strings"

	"github.com/ylchen07/smart-keyvault/internal/provider"
	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// Provider implements the provider.Provider interface for HashiCorp Vault
type Provider struct {
	client *Client
}

// NewProvider creates a new HashiCorp Vault provider
// Configuration options:
//   - "address" (string): Vault server address
//   - "token" (string): Vault authentication token
//   - "namespace" (string): Vault namespace (optional, for Enterprise)
func NewProvider(cfg *provider.Config) (provider.Provider, error) {
	var address, token, namespace string

	// Try to get config from Settings
	if cfg != nil && cfg.Settings != nil {
		if v, ok := cfg.Settings["address"].(string); ok {
			address = v
		}
		if v, ok := cfg.Settings["token"].(string); ok {
			token = v
		}
		if v, ok := cfg.Settings["namespace"].(string); ok {
			namespace = v
		}
	}

	client, err := NewClient(address, token, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	return &Provider{
		client: client,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "hashicorp"
}

// ListVaults returns all KV v2 secret engine mounts
func (p *Provider) ListVaults(ctx context.Context) ([]*models.Vault, error) {
	mounts, err := p.client.ListMounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list vaults: %w", err)
	}

	vaults := make([]*models.Vault, 0)
	for path, mount := range mounts {
		// Only include KV v2 mounts
		if mount.Type == "kv" {
			version := "1"
			if mount.Options != nil {
				if v, ok := mount.Options["version"]; ok {
					version = v
				}
			}

			// Only include KV v2
			if version == "2" {
				// Remove trailing slash from path
				vaultName := strings.TrimSuffix(path, "/")

				vaults = append(vaults, &models.Vault{
					Name:     vaultName,
					Provider: "hashicorp",
					Metadata: map[string]string{
						"type":        mount.Type,
						"version":     version,
						"description": mount.Description,
					},
				})
			}
		}
	}

	return vaults, nil
}

// ListSecrets returns all secrets in a specific KV v2 mount
func (p *Provider) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
	// Ensure vaultName ends with /
	if !strings.HasSuffix(vaultName, "/") {
		vaultName = vaultName + "/"
	}

	// List secrets at the root of the mount
	keys, err := p.client.ListSecrets(ctx, vaultName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	secrets := make([]*models.Secret, 0)
	for _, key := range keys {
		keyStr, ok := key.(string)
		if !ok {
			continue
		}

		// Skip directories (they end with /)
		if strings.HasSuffix(keyStr, "/") {
			continue
		}

		secrets = append(secrets, &models.Secret{
			Name:      keyStr,
			VaultName: strings.TrimSuffix(vaultName, "/"),
			Provider:  "hashicorp",
			Enabled:   true,
		})
	}

	return secrets, nil
}

// GetSecret retrieves a specific secret value from a KV v2 mount
func (p *Provider) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
	// Ensure vaultName ends with /
	if !strings.HasSuffix(vaultName, "/") {
		vaultName = vaultName + "/"
	}

	data, err := p.client.GetSecret(ctx, vaultName, secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// For KV v2, secrets can have multiple key-value pairs
	// We'll return the first value found, or a specific key if it exists
	// Priority: "value" > "password" > first key found
	var value string

	if v, ok := data["value"]; ok {
		value = fmt.Sprintf("%v", v)
	} else if v, ok := data["password"]; ok {
		value = fmt.Sprintf("%v", v)
	} else {
		// Return the first value found
		for _, v := range data {
			value = fmt.Sprintf("%v", v)
			break
		}
	}

	return &models.SecretValue{
		Name:      secretName,
		Value:     value,
		VaultName: strings.TrimSuffix(vaultName, "/"),
		Provider:  "hashicorp",
	}, nil
}

// SupportsFeature checks if the provider supports a specific feature
func (p *Provider) SupportsFeature(feature provider.Feature) bool {
	switch feature {
	case provider.FeatureVersioning, provider.FeatureMetadata:
		return true
	default:
		return false
	}
}
