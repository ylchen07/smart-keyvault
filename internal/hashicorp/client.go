package hashicorp

import (
	"context"
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
)

// Client wraps the HashiCorp Vault API client
type Client struct {
	client *vault.Client
}

// NewClient creates a new HashiCorp Vault client
// It reads configuration from standard Vault environment variables:
// - VAULT_ADDR: Vault server address (required)
// - VAULT_TOKEN: Authentication token (required)
// - VAULT_NAMESPACE: Vault namespace (optional, required for Vault Enterprise)
func NewClient() (*Client, error) {
	// Create default config (reads from VAULT_ADDR, VAULT_CACERT, etc.)
	config := vault.DefaultConfig()

	// Check if VAULT_ADDR is set
	if config.Address == "" {
		return nil, fmt.Errorf("VAULT_ADDR environment variable not set")
	}

	// Create client
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Set token from environment
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("VAULT_TOKEN environment variable not set")
	}
	client.SetToken(token)

	// Set namespace if provided (required for Vault Enterprise)
	namespace := os.Getenv("VAULT_NAMESPACE")
	if namespace != "" {
		client.SetNamespace(namespace)
	}

	return &Client{
		client: client,
	}, nil
}

// ListMounts returns all secret engine mounts
func (c *Client) ListMounts(ctx context.Context) (map[string]*vault.MountOutput, error) {
	mounts, err := c.client.Sys().ListMountsWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mounts: %w", err)
	}
	return mounts, nil
}

// ListSecrets lists all secrets at a given path in a KV v2 mount
func (c *Client) ListSecrets(ctx context.Context, mountPath, secretPath string) ([]interface{}, error) {
	// For KV v2, we need to use the metadata path
	path := fmt.Sprintf("%smetadata/%s", mountPath, secretPath)

	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// No secrets found
	if secret == nil || secret.Data == nil {
		return []interface{}{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	return keys, nil
}

// GetSecret retrieves a secret value from a KV v2 mount
func (c *Client) GetSecret(ctx context.Context, mountPath, secretPath string) (map[string]interface{}, error) {
	// For KV v2, we need to use the data path
	path := fmt.Sprintf("%sdata/%s", mountPath, secretPath)

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// KV v2 stores the actual secret data under the "data" key
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid secret data format")
	}

	return data, nil
}

// Health checks the health of the Vault server
func (c *Client) Health(ctx context.Context) error {
	health, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}

	if !health.Initialized {
		return fmt.Errorf("vault is not initialized")
	}

	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}

	return nil
}
