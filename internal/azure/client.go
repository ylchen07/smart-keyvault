package azure

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// Client implements Azure Key Vault operations using Azure SDK
type Client struct {
	credential     *azidentity.DefaultAzureCredential
	subscriptionID string
	vaultsClient   *armkeyvault.VaultsClient
	secretClients  map[string]*azsecrets.Client // cached clients per vault
	mu             sync.RWMutex                 // protects secretClients map
}

// NewClient creates a new SDK-based Azure client
func NewClient(subscriptionID string) (*Client, error) {
	// Use DefaultAzureCredential which supports:
	// - Azure CLI (az login)
	// - Managed Identity
	// - Environment variables
	// - Interactive browser
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Create vault management client for listing vaults
	vaultsClient, err := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create vaults client: %w", err)
	}

	return &Client{
		credential:     cred,
		subscriptionID: subscriptionID,
		vaultsClient:   vaultsClient,
		secretClients:  make(map[string]*azsecrets.Client),
	}, nil
}

// ListVaults returns all accessible Azure Key Vaults in the subscription
func (c *Client) ListVaults(ctx context.Context) ([]*models.Vault, error) {
	pager := c.vaultsClient.NewListBySubscriptionPager(nil)

	var vaults []*models.Vault
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list vaults: %w", err)
		}

		for _, vault := range page.Value {
			if vault.Name == nil || vault.Location == nil || vault.ID == nil {
				continue
			}

			vaults = append(vaults, &models.Vault{
				Name:     *vault.Name,
				Provider: "azure",
				Metadata: map[string]string{
					"location":      *vault.Location,
					"resourceGroup": extractResourceGroup(*vault.ID),
				},
			})
		}
	}

	return vaults, nil
}

// ListSecrets returns all secrets in a specific vault
func (c *Client) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
	client, err := c.getSecretsClient(vaultName)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets client: %w", err)
	}

	pager := client.NewListSecretPropertiesPager(nil)

	var secrets []*models.Secret
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, props := range page.Value {
			if props.ID == nil {
				continue
			}

			// Check if secret is enabled
			enabled := true
			if props.Attributes != nil && props.Attributes.Enabled != nil {
				enabled = *props.Attributes.Enabled
			}

			// Only include enabled secrets (matching current CLI behavior)
			if enabled {
				secrets = append(secrets, &models.Secret{
					Name:      props.ID.Name(),
					VaultName: vaultName,
					Provider:  "azure",
					Enabled:   enabled,
				})
			}
		}
	}

	return secrets, nil
}

// GetSecret retrieves a specific secret value
func (c *Client) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
	client, err := c.getSecretsClient(vaultName)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets client: %w", err)
	}

	// Get secret with empty version to get the latest version
	resp, err := client.GetSecret(ctx, secretName, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	if resp.Value == nil {
		return nil, fmt.Errorf("secret value is nil")
	}

	return &models.SecretValue{
		Name:      secretName,
		Value:     *resp.Value,
		VaultName: vaultName,
		Provider:  "azure",
	}, nil
}

// getSecretsClient retrieves or creates a secrets client for a specific vault
func (c *Client) getSecretsClient(vaultName string) (*azsecrets.Client, error) {
	// Check if we already have a client for this vault
	c.mu.RLock()
	client, exists := c.secretClients[vaultName]
	c.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Create new secrets client
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", vaultName)
	client, err := azsecrets.NewClient(vaultURL, c.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets client for vault %s: %w", vaultName, err)
	}

	// Cache the client
	c.mu.Lock()
	c.secretClients[vaultName] = client
	c.mu.Unlock()

	return client, nil
}

// extractResourceGroup extracts the resource group name from an Azure resource ID
// Example: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.KeyVault/vaults/{name}
func extractResourceGroup(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
