# Design Decisions

This document outlines key architectural decisions for the Smart KeyVault project.

## Multi-Provider Architecture

### Decision: Provider Interface Pattern

**Why**: To support multiple secret management backends (Azure KeyVault, Hashicorp Vault, and future providers) with a unified interface.

**Benefits**:
- **Extensibility**: Easy to add new providers (AWS Secrets Manager, GCP Secret Manager, etc.)
- **Consistency**: All providers implement the same interface
- **Testing**: Easy to mock providers for testing
- **Separation of Concerns**: Each provider is self-contained

### Provider Interface

```go
type Provider interface {
    Name() string
    ListVaults(ctx context.Context) ([]*models.Vault, error)
    ListSecrets(ctx context.Context, vaultName string) ([]*models.Secret, error)
    GetSecret(ctx context.Context, vaultName, secretName string) (*models.SecretValue, error)
    SupportsFeature(feature Feature) bool
}
```

### Provider Registry Pattern

**Why**: Allows dynamic provider registration and instantiation.

**Benefits**:
- Providers can self-register at init time
- Factory pattern for creating provider instances
- Easy to enable/disable providers via configuration

```go
func init() {
    provider.Register("azure", azure.NewProvider)
    provider.Register("hashicorp", vault.NewProvider)
}
```

## Clipboard Integration

### Decision: Direct Clipboard Integration in Go Binary

**Why**: Using `gopasspw/clipboard` library for direct clipboard operations.

**Benefits**:
- **Cross-platform**: Works on Linux, macOS, Windows
- **No shell dependency**: Don't need to rely on `xclip`, `pbcopy`, etc.
- **Cleaner UX**: Single command copies secret directly
- **Tmux-independent**: Works outside tmux as well

**Usage**:
```bash
# Binary handles clipboard directly
smart-keyvault get-secret --provider azure --vault X --name Y --copy
```

## Separation: Go Binary vs Shell Scripts

### Decision: Keep UI/Orchestration in Shell, Data in Go

**Go Binary Responsibilities**:
- Execute provider operations (list vaults, secrets, get secret)
- Parse responses and format output
- Handle clipboard operations
- Output plain text (for fzf) or JSON (for scripting)

**Shell Script Responsibilities**:
- Tmux keybinding management
- fzf-tmux UI presentation
- Workflow orchestration (provider → vault → secret selection)
- User feedback messages

**Benefits**:
- **Simplicity**: Shell is great for orchestration and UI
- **Performance**: Go is fast for data fetching
- **Portability**: Go binary works standalone (can be used outside tmux)
- **Testability**: Easy to test Go binary independently

## Azure Provider: CLI Wrapper vs SDK

### Decision: Use Azure CLI (`az`) instead of Azure SDK

**Why**:
- Simpler authentication (relies on `az login`)
- No need to manage credentials in code
- Respects user's existing Azure authentication
- Smaller binary size (no SDK dependencies)
- Users already have `az` installed

**Trade-offs**:
- Slightly slower (subprocess execution)
- Depends on `az` being installed
- Limited by CLI capabilities

**Accepted because**: This is a developer tool, users already have `az` CLI.

## Hashicorp Vault Provider: API SDK

### Decision: Use Hashicorp Vault API SDK

**Why**:
- No official Vault CLI with structured output
- SDK provides native Go integration
- Better performance (direct API calls)
- More features available

**Authentication**:
- Reads from `VAULT_ADDR` and `VAULT_TOKEN` environment variables
- Can be extended to support other auth methods (AppRole, Kubernetes, etc.)

## Data Models

### Decision: Minimal, Provider-Agnostic Models

**Why**: Different providers have different metadata, we only need common fields.

```go
type Vault struct {
    Name     string            // Common across all providers
    Provider string            // Which provider this vault belongs to
    Metadata map[string]string // Provider-specific fields
}
```

**Benefits**:
- Extensible: Metadata map holds provider-specific data
- Consistent: All providers return same structure
- Minimal: Only essential fields

## Output Formatting

### Decision: Two Output Modes (Plain Text + JSON)

**Plain Text**:
- One item per line
- For piping to fzf
- Human-readable

**JSON**:
- Structured data
- For scripting/automation
- Machine-readable

**Example**:
```bash
# Plain (default for fzf)
$ smart-keyvault list-vaults --provider azure
my-prod-vault
my-dev-vault

# JSON (for scripts)
$ smart-keyvault list-vaults --provider azure --format json
[
  {"name": "my-prod-vault", "provider": "azure", "metadata": {...}},
  {"name": "my-dev-vault", "provider": "azure", "metadata": {...}}
]
```

## Configuration Strategy

### Decision: Minimal Configuration, Environment-Based

**Why**: Keep it simple, rely on existing tool configurations.

**Azure**:
- Uses `az` CLI's existing authentication
- No additional config needed

**Hashicorp Vault**:
- Uses `VAULT_ADDR` and `VAULT_TOKEN` environment variables
- Standard Vault client configuration

**Optional Config** (future):
```yaml
# ~/.config/smart-keyvault/config.yaml
providers:
  azure:
    enabled: true
    default: true
  hashicorp:
    enabled: true
    address: "https://vault.example.com"
```

## Error Handling

### Decision: Fail Fast, Clear Messages

**Principles**:
- Write errors to stderr
- Exit with non-zero code on error
- Provide actionable error messages
- Parse provider errors and simplify

**Examples**:
```
Error: Not logged in to Azure. Run 'az login'
Error: Vault 'my-vault' not found
Error: No VAULT_TOKEN set. Export VAULT_TOKEN environment variable
```

## Future Extensibility

### Easy to Add New Providers

To add a new provider (e.g., AWS Secrets Manager):

1. Create `internal/aws/provider.go` implementing `Provider` interface
2. Register in `main.go`: `provider.Register("aws", aws.NewProvider)`
3. Update docs

That's it! No changes needed to:
- CLI commands
- Shell scripts
- Output formatters
- Data models

### Feature Flags

Providers can declare supported features:

```go
func (p *AzureProvider) SupportsFeature(feature provider.Feature) bool {
    switch feature {
    case provider.FeatureVersioning:
        return true  // Azure supports secret versioning
    case provider.FeatureTags:
        return true  // Azure supports tags
    default:
        return false
    }
}
```

Future UI can adapt based on features (e.g., show version selection if supported).

## Security Considerations

### Secret Handling
- Secrets only printed to stdout or clipboard
- No secrets in logs
- No secrets written to disk
- No secrets in command history (avoid echo)

### Authentication
- No credential storage in code
- Rely on provider's native auth (az CLI, Vault tokens)
- Respect environment variables and provider configs

### Clipboard
- Secret stays in clipboard until user pastes or copies something else
- Could add auto-clear in future (configurable timeout)

## Testing Strategy

### Unit Tests
- Mock provider interface for testing
- Test each provider implementation with mock backends
- Test output formatters

### Integration Tests
- Real provider tests (with test accounts)
- Tag with build flags: `go test -tags=integration`

### Manual Testing
- Test with real Azure KeyVault
- Test with real Hashicorp Vault
- Test tmux plugin workflow end-to-end

## Performance Considerations

### No Caching (Initially)
- Keep it simple
- Providers are already fast enough
- Can add caching later if needed

### Concurrent Requests (Future)
- Could fetch secrets concurrently when listing
- Use goroutines with worker pools
- Implement when needed for large vaults

## Summary

The architecture is designed to be:
- **Simple**: Easy to understand and maintain
- **Extensible**: Easy to add new providers
- **Practical**: Uses existing tools (az CLI, Vault SDK)
- **User-friendly**: Direct clipboard integration, clear errors
- **Testable**: Provider interface allows mocking
- **Secure**: No credential storage, respects provider auth

The provider pattern is the key architectural decision that enables all of these benefits.
