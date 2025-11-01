# Design Decisions

Key architectural decisions for Smart KeyVault.

## Multi-Provider Architecture

### Provider Interface Pattern

**Why**: Support multiple secret management backends with a unified interface.

```go
type Provider interface {
    Name() string
    ListVaults(ctx context.Context) ([]*models.Vault, error)
    ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error)
    GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error)
    SupportsFeature(feature Feature) bool
}
```

**Benefits**:
- Easy to add new providers (AWS, GCP, etc.)
- Consistent interface across all providers
- Mockable for testing
- Self-contained provider implementations

### Provider Registry

Providers self-register at init time:

```go
func init() {
    provider.Register("azure", azure.NewProvider)
    provider.Register("hashicorp", hashicorp.NewProvider)
}
```

## Configuration System

### Viper-Based Multi-Instance Config

**Why**: Users need to manage multiple Azure subscriptions and Vault servers.

**Features**:
- Config file at `~/.config/smart-keyvault/config.yaml`
- Environment variable substitution (`${VAR}` pattern)
- Multiple instances per provider
- Default instances to skip selection prompts

**Config Structure**:
```yaml
defaults:
  provider: "azure"

providers:
  azure:
    instances:
      - name: "prod"
        subscription_id: "${AZURE_SUBSCRIPTION_ID}"
        default: true

  hashicorp:
    instances:
      - name: "prod-vault"
        address: "https://vault.example.com"
        token: "${VAULT_TOKEN}"
```

**Precedence**: CLI flags > Env vars > Config file > Defaults

## Azure Provider: SDK vs CLI

### Decision: Use Azure SDK for Go

**Why**:
- Native Go integration, no subprocess spawning
- Better performance and error handling
- Thread-safe client caching
- Connection pooling and reuse

**Authentication**: `DefaultAzureCredential` supports:
- Azure CLI (`az login`)
- Managed Identity
- Environment variables
- Service Principal

## Hashicorp Vault Provider: API SDK

**Why**: No official CLI, SDK provides native Go integration.

**Authentication**: Environment variables (`VAULT_ADDR`, `VAULT_TOKEN`, `VAULT_NAMESPACE`)

**Supports**: KV v2 secret engines only

## Clipboard Integration

### Direct Clipboard in Go Binary

**Why**: Using `gopasspw/clipboard` for direct integration.

**Benefits**:
- Cross-platform (Linux, macOS, Windows)
- No shell dependency (xclip, pbcopy)
- Works outside tmux too

**Usage**:
```bash
smart-keyvault get-secret --provider azure --vault X --name Y --copy
```

## Architecture Separation

### Go Binary vs Shell Scripts

**Go Binary**:
- Data fetching from providers
- Config loading and instance selection
- Clipboard operations
- Output formatting (plain/JSON)

**Shell Scripts**:
- Tmux keybinding management
- fzf-tmux UI orchestration
- Workflow coordination
- User feedback messages

**Benefits**: Simple separation, Go handles data, shell handles UI

## Data Models

### Minimal, Provider-Agnostic Models

```go
type Vault struct {
    Name     string
    Provider string
    Metadata map[string]string // Provider-specific fields
}
```

**Why**: Different providers have different metadata, only store common fields + extensible metadata map.

## Output Formatting

**Plain Text** (default):
- One item per line
- For piping to fzf
- Human-readable

**JSON** (`--format json`):
- Structured data
- For scripting/automation
- Machine-readable

## Error Handling

**Principles**:
- Fail fast with clear messages
- Errors to stderr, data to stdout
- Non-zero exit codes on error
- Actionable error messages

**Examples**:
```
Error: subscription_id required (set via config or AZURE_SUBSCRIPTION_ID)
Error: vault token not set (provide via config or VAULT_TOKEN)
Error: vault 'my-vault' not found
```

## Security

**Secret Handling**:
- Secrets only to stdout or clipboard
- No secrets in logs or disk
- No secrets in shell history

**Authentication**:
- No credential storage in code
- Rely on provider native auth
- Config file supports `${VAR}` for sensitive data

**Clipboard**:
- Secret persists until next copy/paste
- Could add auto-clear timeout (future)

## Extensibility

### Adding New Providers

Three steps to add a new provider (e.g., AWS):

1. Create `internal/aws/provider.go` implementing `Provider` interface
2. Register: `provider.Register("aws", aws.NewProvider)`
3. Update docs

No changes needed to CLI, shell scripts, formatters, or models.

### Feature Flags

Providers declare supported features:

```go
func (p *Provider) SupportsFeature(feature provider.Feature) bool {
    switch feature {
    case provider.FeatureVersioning: return true  // Azure supports versioning
    case provider.FeatureTags:       return true  // Azure supports tags
    default:                         return false
    }
}
```

UI can adapt based on features (e.g., show version selection if supported).

## Summary

Core principles:
- **Simple**: Clear separation of concerns
- **Extensible**: Provider pattern enables easy additions
- **Practical**: Use existing tools and native SDKs
- **User-friendly**: Config file, clipboard integration, clear errors
- **Secure**: No credential storage, respects provider auth
- **Fast**: Native Go SDKs, client caching, no subprocess overhead
