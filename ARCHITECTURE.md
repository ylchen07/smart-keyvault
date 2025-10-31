# Architecture Documentation

## Overview

Smart KeyVault is a tmux plugin that provides a simple, interactive interface for Azure KeyVault. It consists of two main components:
1. **Go binary**: Fetches and formats data from Azure KeyVault
2. **Shell scripts**: Tmux plugin that orchestrates fzf-tmux UI and clipboard operations

The architecture is intentionally simple with clear separation: Go handles data, shell handles UI.

## System Architecture

### High-Level Components

```
┌─────────────────────────────────────────────────────────┐
│                    User (Tmux)                          │
│              Presses keybinding (prefix + K)            │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Tmux Plugin (Shell Scripts)                │
│                                                         │
│  ┌──────────────────────────────────────────────┐      │
│  │  smart-keyvault.tmux (TPM entry point)       │      │
│  │  - Sets up keybindings                       │      │
│  │  - Loads configuration                       │      │
│  └──────────────────┬───────────────────────────┘      │
│                     │                                   │
│  ┌──────────────────▼───────────────────────────┐      │
│  │  scripts/browse-secrets.sh                   │      │
│  │  - Calls Go binary to get vault list         │      │
│  │  - Pipes output to fzf-tmux                  │      │
│  │  - Gets user selection                       │      │
│  │  - Calls Go binary to get secret list        │      │
│  │  - Pipes to fzf-tmux                         │      │
│  │  - Gets secret value                         │      │
│  │  - Copies to clipboard                       │      │
│  └──────────────────┬───────────────────────────┘      │
└────────────────────┼────────────────────────────────────┘
                     │
                     │ (executes commands)
                     ▼
┌─────────────────────────────────────────────────────────┐
│           Go Binary (smart-keyvault)                    │
│                                                         │
│  ┌──────────────┐                                      │
│  │  CLI (Cobra) │                                      │
│  │              │                                      │
│  │  Commands:   │                                      │
│  │  - list-vaults                                      │
│  │  - list-secrets --vault <name>                     │
│  │  - get-secret --vault <name> --name <secret>       │
│  └──────┬───────┘                                      │
│         │                                               │
│  ┌──────▼────────────────────────────────────┐         │
│  │    Azure Provider                        │         │
│  │  ┌────────────────────────────────────┐  │         │
│  │  │   Azure Client (SDK-based)         │  │         │
│  │  │  - DefaultAzureCredential          │  │         │
│  │  │  - VaultsClient (list vaults)      │  │         │
│  │  │  - SecretsClient (per-vault cache) │  │         │
│  │  └────────────┬───────────────────────┘  │         │
│  └───────────────┼──────────────────────────┘         │
│                  │                                      │
│  ┌───────────────▼──────────────────────┐              │
│  │    Output Formatters                 │              │
│  │  - Plain text (for fzf)              │              │
│  │  - JSON (for scripting)              │              │
│  │  - Walk secrets (grouped by vault)   │              │
│  └──────────────────────────────────────┘              │
└─────────────────────┬────────────────────────────────────┘
                      │
                      ▼
┌──────────────────────────────────────────────────────────┐
│              Azure SDK for Go                            │
│  - azidentity.DefaultAzureCredential                     │
│  - armkeyvault.VaultsClient (management plane)           │
│  - azsecrets.Client (data plane, per vault)              │
└──────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Tmux Plugin Scripts

**Location**: `smart-keyvault.tmux`, `scripts/`

**Responsibility**: User interface, orchestration, clipboard handling

**Key Files**:
- `smart-keyvault.tmux`: TPM plugin entry point, sets up keybindings
- `scripts/browse-secrets.sh`: Main workflow (vault → secret → copy)
- `scripts/utils.sh`: Helper functions

**Shell Script Flow**:
```bash
#!/usr/bin/env bash
# browse-secrets.sh

# Step 1: Select vault
vault=$(smart-keyvault list-vaults | fzf-tmux -p 80%,60% --prompt="Select Vault: ")
[[ -z "$vault" ]] && exit 0

# Step 2: Select secret
secret=$(smart-keyvault list-secrets --vault "$vault" | \
         fzf-tmux -p 80%,60% --prompt="Select Secret ($vault): ")
[[ -z "$secret" ]] && exit 0

# Step 3: Get secret value and copy
value=$(smart-keyvault get-secret --vault "$vault" --name "$secret")
echo "$value" | tmux load-buffer -
tmux display-message "Secret '$secret' copied to clipboard!"
```

### 2. Go CLI Binary

**Location**: `cmd/smart-keyvault/`

**Responsibility**: Data fetching, formatting

**Technology**: Cobra CLI framework

**Commands**:
```
smart-keyvault list-vaults              # Output: one vault name per line
smart-keyvault list-secrets --vault X   # Output: one secret name per line
smart-keyvault get-secret --vault X --name Y   # Output: secret value only
smart-keyvault get-secret --vault X --name Y --format json  # JSON output
```

### 3. Azure Provider (internal/azure)

**Responsibility**: Interact with Azure Key Vault using Azure SDK for Go

**Files**:
- `provider.go`: Implements Provider interface
- `client.go`: Azure SDK client wrapper with authentication and caching

**Key Implementation**:

```go
// Client implements Azure Key Vault operations using Azure SDK
type Client struct {
    credential     *azidentity.DefaultAzureCredential
    subscriptionID string
    vaultsClient   *armkeyvault.VaultsClient
    secretClients  map[string]*azsecrets.Client // cached per vault
    mu             sync.RWMutex                 // protects secretClients map
}

func NewClient(subscriptionID string) (*Client, error) {
    // Use DefaultAzureCredential - supports:
    // - Azure CLI (az login)
    // - Managed Identity
    // - Environment variables
    // - Service Principal
    cred, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        return nil, err
    }

    vaultsClient, err := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)
    if err != nil {
        return nil, err
    }

    return &Client{
        credential:     cred,
        subscriptionID: subscriptionID,
        vaultsClient:   vaultsClient,
        secretClients:  make(map[string]*azsecrets.Client),
    }, nil
}

func (c *Client) ListVaults(ctx context.Context) ([]*models.Vault, error) {
    // Use SDK pager for pagination
    pager := c.vaultsClient.NewListBySubscriptionPager(nil)
    // Iterate through pages and collect vaults
}

func (c *Client) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
    client, _ := c.getSecretsClient(vaultName) // Cached
    pager := client.NewListSecretPropertiesPager(nil)
    // Iterate through pages, filter enabled secrets
}

func (c *Client) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
    client, _ := c.getSecretsClient(vaultName)
    resp, _ := client.GetSecret(ctx, secretName, "", nil) // Empty version = latest
    return &models.SecretValue{Value: *resp.Value}, nil
}
```

**Data Flow**:
1. Authenticate using DefaultAzureCredential
2. Create management client (VaultsClient) and data clients (SecretsClient per vault)
3. Use SDK pagination for listing operations
4. Cache SecretsClient instances per vault for performance
5. Return structured Go data models

**Key Benefits over CLI**:
- No process spawning (faster)
- Native Go types and error handling
- Connection pooling and reuse
- Thread-safe client caching

### 4. Output Formatters (internal/output)

**Responsibility**: Format data for different output types

**Files**:
- `plain.go`: Plain text output (for fzf)
- `json.go`: JSON output (for scripting)
- `formatter.go`: Formatter interface

**Implementation**:

```go
package output

// Formatter interface
type Formatter interface {
    FormatVaults(vaults []*models.Vault) (string, error)
    FormatSecrets(secrets []*models.Secret) (string, error)
    FormatProviders(providers []string) (string, error)
    FormatWalkSecrets(secretsByVault map[string][]*models.SecretValue) (string, error)
}

// Plain text formatter - one item per line
func (f *PlainFormatter) FormatVaultsPlain(vaults []*models.Vault) (string, error) {
    var lines []string
    for _, v := range vaults {
        lines = append(lines, v.Name)
    }
    return strings.Join(lines, "\n"), nil
}

func (f *PlainFormatter) FormatWalkSecrets(secretsByVault map[string][]*models.SecretValue) (string, error) {
    // Format: vault:secret_name=secret_value
    var lines []string
    for vaultName, secrets := range secretsByVault {
        for _, secret := range secrets {
            lines = append(lines, vaultName+":"+secret.Name+"="+secret.Value)
        }
    }
    return strings.Join(lines, "\n"), nil
}

// JSON formatter - for scripting/parsing
func (f *JSONFormatter) FormatVaultsJSON(vaults []*models.Vault) (string, error) {
    data, err := json.MarshalIndent(vaults, "", "  ")
    return string(data), err
}

func (f *JSONFormatter) FormatWalkSecrets(secretsByVault map[string][]*models.SecretValue) (string, error) {
    // Groups secrets by vault: {"vault1": [...], "vault2": [...]}
    data, err := json.MarshalIndent(secretsByVault, "", "  ")
    return string(data), err
}
```

**Usage in CLI**:
```go
vaults, _ := vaultService.ListVaults()

if format == "json" {
    output, _ := output.FormatVaultsJSON(vaults)
    fmt.Println(output)
} else {
    output := output.FormatVaultsPlain(vaults)
    fmt.Println(output)  // One vault name per line
}
```

## Data Models (pkg/models)

### Core Types

```go
package models

import "time"

// Vault represents an Azure Key Vault (minimal info)
type Vault struct {
    Name          string `json:"name"`
    Location      string `json:"location,omitempty"`
    ResourceGroup string `json:"resourceGroup,omitempty"`
}

// Secret represents a key vault secret (minimal info)
type Secret struct {
    Name    string `json:"name"`
    Enabled bool   `json:"enabled"`
}

// SecretValue includes the actual secret data
type SecretValue struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}
```

Note: We keep models minimal since the Go binary only needs to output names for fzf selection.

## Execution Flow Examples

### Complete Workflow: Browse and Copy Secret

```
User: <prefix> + K (in tmux)

1. Tmux triggers keybinding
   └─> Executes: ~/.tmux/plugins/smart-keyvault/scripts/browse-secrets.sh

2. Shell Script: Get vault list
   ├─> Executes: smart-keyvault list-vaults
   │   ├─> Go: VaultService.ListVaults()
   │   ├─> Go: Execute `az keyvault list --output json`
   │   ├─> Go: Parse JSON response
   │   └─> Go: Output plain text (one vault per line)
   └─> Pipes to: fzf-tmux -p 80%,60% --prompt="Select Vault: "

3. User selects vault in fzf
   └─> Returns: "my-prod-vault"

4. Shell Script: Get secret list
   ├─> Executes: smart-keyvault list-secrets --vault my-prod-vault
   │   ├─> Go: SecretService.ListSecrets("my-prod-vault")
   │   ├─> Go: Execute `az keyvault secret list --vault-name my-prod-vault --output json`
   │   ├─> Go: Parse JSON, filter enabled secrets
   │   └─> Go: Output plain text (one secret per line)
   └─> Pipes to: fzf-tmux -p 80%,60% --prompt="Select Secret (my-prod-vault): "

5. User selects secret in fzf
   └─> Returns: "database-password"

6. Shell Script: Get secret value
   ├─> Executes: smart-keyvault get-secret --vault my-prod-vault --name database-password
   │   ├─> Go: SecretService.GetSecret("my-prod-vault", "database-password")
   │   ├─> Go: Execute `az keyvault secret show --vault-name my-prod-vault --name database-password --output json`
   │   ├─> Go: Parse JSON response
   │   └─> Go: Output only the secret value
   └─> Returns: "my-secret-value-123"

7. Shell Script: Copy to clipboard
   ├─> Executes: echo "$value" | tmux load-buffer -
   └─> Displays: tmux display-message "Secret 'database-password' copied!"

8. User pastes secret wherever needed
```

### Simple Command: List Vaults

```bash
$ smart-keyvault list-vaults

Go Binary Flow:
1. Parse CLI flags (none in this case)
2. Create VaultService with Azure client
3. VaultService.ListVaults()
   ├─> Build command: ["keyvault", "list", "--output", "json"]
   ├─> Execute: exec.Command("az", args...)
   ├─> Read stdout (JSON array of vaults)
   ├─> Unmarshal JSON into []models.Vault
   └─> Return vault list
4. Format output as plain text (one name per line)
5. Print to stdout:
   my-prod-vault
   my-dev-vault
   shared-vault
```

## Error Handling Strategy

### Go Binary Error Handling

```go
// Clear error messages to stderr
if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}

// Azure CLI errors - parse and simplify
if exitErr, ok := err.(*exec.ExitError); ok {
    stderr := string(exitErr.Stderr)
    if strings.Contains(stderr, "az login") {
        fmt.Fprintln(os.Stderr, "Error: Not logged in to Azure. Run 'az login'")
    } else if strings.Contains(stderr, "does not exist") {
        fmt.Fprintln(os.Stderr, "Error: Vault or secret not found")
    } else {
        fmt.Fprintf(os.Stderr, "Azure CLI error: %s\n", stderr)
    }
    os.Exit(1)
}
```

### Shell Script Error Handling

```bash
#!/usr/bin/env bash
set -e  # Exit on error

# Handle user cancellation (ESC in fzf)
vault=$(smart-keyvault list-vaults | fzf-tmux -p 80%,60% --prompt="Select Vault: " || true)
if [[ -z "$vault" ]]; then
    exit 0  # Silent exit on cancellation
fi

# Check if binary exists
if ! command -v smart-keyvault &> /dev/null; then
    tmux display-message "Error: smart-keyvault binary not found. Run 'make install'"
    exit 1
fi
```

### Common Error Scenarios

1. **Not logged in to Azure**: Clear message to run `az login`
2. **No vaults found**: Message explaining permissions or subscription
3. **fzf not found**: Installation instructions
4. **User cancellation**: Silent exit (no error)
5. **Vault doesn't exist**: Clear message with vault name

## Security Considerations

1. **Secret Handling**:
   - Never log secret values to stdout/files
   - Secret values only printed to stdout (for piping)
   - No secrets stored in shell history (avoid echo)
   - Tmux buffer can be cleared manually

2. **Azure Authentication**:
   - Relies entirely on `az login` (no credential storage)
   - Respects Azure CLI authentication (service principals, managed identities)
   - No additional auth configuration needed

3. **Minimal Disk I/O**:
   - No caching of secrets to disk
   - No config file with sensitive data
   - Temporary data only in memory

## TPM Plugin Installation

The `smart-keyvault.tmux` file is the TPM entry point:

```bash
#!/usr/bin/env bash
# smart-keyvault.tmux

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Build Go binary if not exists
if [[ ! -f "$CURRENT_DIR/bin/smart-keyvault" ]]; then
    echo "Building smart-keyvault..."
    cd "$CURRENT_DIR" && make build
fi

# Add binary to PATH
export PATH="$CURRENT_DIR/bin:$PATH"

# Get user-configured keybindings or use defaults
keybind=$(tmux show-option -gv @smart-keyvault-key 2>/dev/null || echo "K")
quick_keybind=$(tmux show-option -gv @smart-keyvault-quick-key 2>/dev/null || echo "k")

# Set keybindings
tmux bind-key "$keybind" run-shell "$CURRENT_DIR/scripts/browse-secrets.sh"
tmux bind-key "$quick_keybind" run-shell "$CURRENT_DIR/scripts/browse-secrets.sh --quick"
```

## Testing Strategy

### Go Binary Tests
```bash
# Unit tests with mocked az CLI
go test ./internal/azure/...
go test ./internal/output/...

# Integration test with real Azure (requires az login)
go test -tags=integration ./...
```

### Manual Testing
```bash
# Test binary directly
./bin/smart-keyvault list-vaults
./bin/smart-keyvault list-secrets --vault my-vault

# Test with fzf
./bin/smart-keyvault list-vaults | fzf

# Test full workflow
./scripts/browse-secrets.sh
```

## Project Structure Summary (Multi-Provider Design)

```
smart-keyvault/
├── cmd/smart-keyvault/
│   └── main.go                      # CLI entry point (Cobra)
├── internal/
│   ├── provider/                    # Provider abstraction
│   │   ├── provider.go              # Provider interface
│   │   ├── registry.go              # Provider registry & factory
│   │   └── config.go                # Provider configuration
│   ├── azure/                       # Azure KeyVault provider
│   │   ├── provider.go              # Implements Provider interface
│   │   ├── client.go                # Azure CLI executor
│   │   └── parser.go                # JSON response parser
│   ├── vault/                       # Hashicorp Vault provider
│   │   ├── provider.go              # Implements Provider interface
│   │   ├── client.go                # Vault API client
│   │   └── auth.go                  # Vault authentication
│   ├── output/                      # Output formatters
│   │   ├── formatter.go             # Formatter interface
│   │   ├── plain.go                 # Plain text (for fzf)
│   │   └── json.go                  # JSON (for scripting)
│   └── clipboard/                   # Clipboard operations
│       └── clipboard.go             # Wrapper for gopasspw/clipboard
├── pkg/
│   └── models/                      # Shared data models
│       ├── vault.go                 # Vault model
│       ├── secret.go                # Secret model
│       └── provider.go              # Provider metadata
├── scripts/                         # Tmux plugin shell scripts
│   ├── browse-secrets.sh            # Main workflow
│   ├── select-provider.sh           # Provider selection
│   └── utils.sh                     # Helper functions
├── smart-keyvault.tmux              # TPM entry point
├── Makefile                         # Build tasks
├── go.mod
├── go.sum
├── README.md
└── ARCHITECTURE.md
```

## Provider Interface Design

### Core Provider Interface

```go
package provider

import (
    "context"
    "github.com/yourusername/smart-keyvault/pkg/models"
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
    FeatureVersioning Feature = iota  // Secret versioning
    FeatureMetadata                   // Rich metadata
    FeatureTags                       // Tag support
)

// Config holds provider-specific configuration
type Config struct {
    Name       string                 // Provider name
    Enabled    bool                   // Whether provider is enabled
    Default    bool                   // Default provider
    Settings   map[string]interface{} // Provider-specific settings
}
```

### Azure Provider Implementation

```go
package azure

import (
    "context"
    "fmt"
    "os"
    "strings"
    "sync"

    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
    "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

    "github.com/yourusername/smart-keyvault/pkg/models"
    "github.com/yourusername/smart-keyvault/internal/provider"
)

type Provider struct {
    client *Client
}

type Client struct {
    credential     *azidentity.DefaultAzureCredential
    subscriptionID string
    vaultsClient   *armkeyvault.VaultsClient
    secretClients  map[string]*azsecrets.Client
    mu             sync.RWMutex
}

func NewProvider(cfg *provider.Config) (provider.Provider, error) {
    subscriptionID := ""

    // Try config first, then environment variable
    if cfg != nil && cfg.Settings != nil {
        if v, ok := cfg.Settings["subscription_id"].(string); ok {
            subscriptionID = v
        }
    }
    if subscriptionID == "" {
        subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
    }

    if subscriptionID == "" {
        return nil, fmt.Errorf("subscription_id required (config or AZURE_SUBSCRIPTION_ID)")
    }

    client, err := NewClient(subscriptionID)
    if err != nil {
        return nil, err
    }

    return &Provider{client: client}, nil
}

func NewClient(subscriptionID string) (*Client, error) {
    cred, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        return nil, err
    }

    vaultsClient, err := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)
    if err != nil {
        return nil, err
    }

    return &Client{
        credential:     cred,
        subscriptionID: subscriptionID,
        vaultsClient:   vaultsClient,
        secretClients:  make(map[string]*azsecrets.Client),
    }, nil
}

func (p *Provider) ListVaults(ctx context.Context) ([]*models.Vault, error) {
    pager := p.client.vaultsClient.NewListBySubscriptionPager(nil)

    var vaults []*models.Vault
    for pager.More() {
        page, err := pager.NextPage(ctx)
        if err != nil {
            return nil, err
        }

        for _, vault := range page.Value {
            if vault.Name != nil && vault.Location != nil && vault.ID != nil {
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
    }
    return vaults, nil
}

func (p *Provider) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
    client, err := p.client.getSecretsClient(vaultName)
    if err != nil {
        return nil, err
    }

    pager := client.NewListSecretPropertiesPager(nil)
    var secrets []*models.Secret

    for pager.More() {
        page, err := pager.NextPage(ctx)
        if err != nil {
            return nil, err
        }

        for _, props := range page.Value {
            if props.ID != nil {
                enabled := true
                if props.Attributes != nil && props.Attributes.Enabled != nil {
                    enabled = *props.Attributes.Enabled
                }

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
    }
    return secrets, nil
}

func (p *Provider) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
    client, err := p.client.getSecretsClient(vaultName)
    if err != nil {
        return nil, err
    }

    resp, err := client.GetSecret(ctx, secretName, "", nil)
    if err != nil {
        return nil, err
    }

    return &models.SecretValue{
        Name:      secretName,
        Value:     *resp.Value,
        VaultName: vaultName,
        Provider:  "azure",
    }, nil
}

func (c *Client) getSecretsClient(vaultName string) (*azsecrets.Client, error) {
    c.mu.RLock()
    client, exists := c.secretClients[vaultName]
    c.mu.RUnlock()

    if exists {
        return client, nil
    }

    vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", vaultName)
    client, err := azsecrets.NewClient(vaultURL, c.credential, nil)
    if err != nil {
        return nil, err
    }

    c.mu.Lock()
    c.secretClients[vaultName] = client
    c.mu.Unlock()

    return client, nil
}

func extractResourceGroup(resourceID string) string {
    parts := strings.Split(resourceID, "/")
    for i, part := range parts {
        if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
            return parts[i+1]
        }
    }
    return ""
}

func (p *Provider) SupportsFeature(feature provider.Feature) bool {
    switch feature {
    case provider.FeatureVersioning, provider.FeatureTags:
        return true
    default:
        return false
    }
}
```

### Hashicorp Vault Provider Implementation

```go
package vault

import (
    "context"
    "fmt"
    vault "github.com/hashicorp/vault/api"
    "github.com/yourusername/smart-keyvault/pkg/models"
    "github.com/yourusername/smart-keyvault/internal/provider"
)

type Provider struct {
    client *vault.Client
}

func NewProvider(cfg *provider.Config) (*Provider, error) {
    config := vault.DefaultConfig()

    // Get Vault address from config or environment
    if addr, ok := cfg.Settings["address"].(string); ok {
        config.Address = addr
    }

    client, err := vault.NewClient(config)
    if err != nil {
        return nil, err
    }

    // Set token from config or environment
    if token, ok := cfg.Settings["token"].(string); ok {
        client.SetToken(token)
    }

    return &Provider{
        client: client,
    }, nil
}

func (p *Provider) Name() string {
    return "hashicorp"
}

func (p *Provider) ListVaults(ctx context.Context) ([]*models.Vault, error) {
    // List KV mounts
    mounts, err := p.client.Sys().ListMounts()
    if err != nil {
        return nil, err
    }

    vaults := make([]*models.Vault, 0)
    for path, mount := range mounts {
        // Only include KV v2 mounts
        if mount.Type == "kv" && mount.Options["version"] == "2" {
            vaults = append(vaults, &models.Vault{
                Name:     path,
                Provider: "hashicorp",
                Metadata: map[string]string{
                    "type":    mount.Type,
                    "version": mount.Options["version"],
                },
            })
        }
    }

    return vaults, nil
}

func (p *Provider) ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error) {
    // List secrets in KV v2 mount
    path := fmt.Sprintf("%s/metadata", vaultName)
    secret, err := p.client.Logical().List(path)
    if err != nil {
        return nil, err
    }

    if secret == nil || secret.Data["keys"] == nil {
        return []*models.Secret{}, nil
    }

    keys := secret.Data["keys"].([]interface{})
    secrets := make([]*models.Secret, len(keys))

    for i, key := range keys {
        secrets[i] = &models.Secret{
            Name:      key.(string),
            VaultName: vaultName,
            Provider:  "hashicorp",
        }
    }

    return secrets, nil
}

func (p *Provider) GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error) {
    path := fmt.Sprintf("%s/data/%s", vaultName, secretName)
    secret, err := p.client.Logical().Read(path)
    if err != nil {
        return nil, err
    }

    if secret == nil {
        return nil, fmt.Errorf("secret not found: %s", secretName)
    }

    // KV v2 stores data under "data" key
    data := secret.Data["data"].(map[string]interface{})

    // For now, return first value found (or support key selection in future)
    var value string
    for _, v := range data {
        value = fmt.Sprintf("%v", v)
        break
    }

    return &models.SecretValue{
        Name:      secretName,
        Value:     value,
        VaultName: vaultName,
        Provider:  "hashicorp",
    }, nil
}

func (p *Provider) SupportsFeature(feature provider.Feature) bool {
    switch feature {
    case provider.FeatureVersioning, provider.FeatureMetadata:
        return true
    default:
        return false
    }
}
```

### Provider Registry

```go
package provider

import (
    "fmt"
    "sync"
)

// ProviderFactory creates a new provider instance
type ProviderFactory func(cfg *Config) (Provider, error)

// Registry manages available providers
type Registry struct {
    mu        sync.RWMutex
    factories map[string]ProviderFactory
}

var defaultRegistry = &Registry{
    factories: make(map[string]ProviderFactory),
}

// Register adds a provider factory to the registry
func Register(name string, factory ProviderFactory) {
    defaultRegistry.mu.Lock()
    defer defaultRegistry.mu.Unlock()
    defaultRegistry.factories[name] = factory
}

// GetProvider creates a provider instance by name
func GetProvider(name string, cfg *Config) (Provider, error) {
    defaultRegistry.mu.RLock()
    factory, exists := defaultRegistry.factories[name]
    defaultRegistry.mu.RUnlock()

    if !exists {
        return nil, fmt.Errorf("provider not found: %s", name)
    }

    return factory(cfg)
}

// ListProviders returns all registered provider names
func ListProviders() []string {
    defaultRegistry.mu.RLock()
    defer defaultRegistry.mu.RUnlock()

    names := make([]string, 0, len(defaultRegistry.factories))
    for name := range defaultRegistry.factories {
        names = append(names, name)
    }
    return names
}
```

### Data Models (pkg/models)

```go
package models

// Vault represents a secrets vault/backend
type Vault struct {
    Name     string            `json:"name"`
    Provider string            `json:"provider"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

// Secret represents a secret (without value)
type Secret struct {
    Name      string `json:"name"`
    VaultName string `json:"vault"`
    Provider  string `json:"provider"`
    Enabled   bool   `json:"enabled,omitempty"`
}

// SecretValue includes the actual secret value
type SecretValue struct {
    Name      string `json:"name"`
    Value     string `json:"value"`
    VaultName string `json:"vault"`
    Provider  string `json:"provider"`
}
```

### Clipboard Integration

```go
package clipboard

import (
    "fmt"
    "github.com/atotto/clipboard"  // or gopasspw/clipboard
)

// Copy copies text to system clipboard
func Copy(text string) error {
    if err := clipboard.WriteAll(text); err != nil {
        return fmt.Errorf("failed to copy to clipboard: %w", err)
    }
    return nil
}

// Read reads text from system clipboard
func Read() (string, error) {
    text, err := clipboard.ReadAll()
    if err != nil {
        return "", fmt.Errorf("failed to read from clipboard: %w", err)
    }
    return text, nil
}
```

### CLI Commands with Clipboard Support

```bash
# List all providers
smart-keyvault list-providers

# List vaults from specific provider
smart-keyvault list-vaults --provider azure
smart-keyvault list-vaults --provider hashicorp

# List secrets
smart-keyvault list-secrets --provider azure --vault my-vault

# Get secret and copy to clipboard
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret --copy

# Get secret (just output value)
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret
```

## Updated Workflow with Multi-Provider

```
User: <prefix> + K

1. Select Provider (if multiple enabled)
   ├─> smart-keyvault list-providers
   └─> fzf-tmux: azure, hashicorp

2. Select Vault
   ├─> smart-keyvault list-vaults --provider azure
   └─> fzf-tmux: vault selection

3. Select Secret
   ├─> smart-keyvault list-secrets --provider azure --vault X
   └─> fzf-tmux: secret selection

4. Copy to Clipboard
   ├─> smart-keyvault get-secret --provider azure --vault X --name Y --copy
   └─> Display: "Secret 'Y' copied to clipboard!"
```

## Next Steps

1. Initialize Go module with dependencies (cobra, clipboard, hashicorp vault SDK)
2. Implement provider interface and registry
3. Implement Azure provider
4. Implement Hashicorp Vault provider
5. Create Cobra CLI with multi-provider commands
6. Write shell scripts for tmux plugin with provider selection
7. Create Makefile for building
8. Test with both Azure KeyVault and Hashicorp Vault
