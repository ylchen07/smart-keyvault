# Architecture Documentation

## Overview

Smart KeyVault is a tmux plugin for managing secrets across multiple providers (Azure KeyVault, HashiCorp Vault).

**Components**:
- **Go binary**: Data fetching, config management, clipboard operations
- **Shell scripts**: Tmux integration, fzf UI orchestration

**Philosophy**: Go handles data, shell handles UI.

## System Architecture

```
User (Tmux) → Press prefix + K
    ↓
Tmux Plugin (Shell)
    - scripts/browse-secrets.sh
    - fzf-tmux UI, workflow orchestration
    ↓
Go Binary (smart-keyvault)
    ├── Config System (Viper)
    │   - Load ~/.config/smart-keyvault/config.yaml
    │   - Environment variable substitution (${VAR})
    │   - Multi-instance management
    ├── Provider Registry
    │   ├── Azure Provider → Azure SDK
    │   └── HashiCorp Provider → Vault SDK
    └── Output Formatters (plain/json)
```

## Core Components

### 1. Configuration System (`internal/config/`)

**Viper-based config with multi-instance support.**

Files: `types.go`, `loader.go`, `helpers.go`

Config location: `~/.config/smart-keyvault/config.yaml`

**Features**:
- Multi-instance support (multiple Azure subscriptions, multiple Vault servers)
- Environment variable substitution: `${VAR}` → expanded value
- Default instance selection
- Precedence: CLI flags > Env vars > Config file

**Example**:
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
        default: true
```

### 2. Provider System (`internal/provider/`)

**Unified interface for all secret backends:**
```go
type Provider interface {
    Name() string
    ListVaults(ctx) ([]*models.Vault, error)
    ListSecrets(ctx, vaultName) ([]*models.Secret, error)
    GetSecret(ctx, vault, secret) (*models.SecretValue, error)
    SupportsFeature(feature Feature) bool
}
```

Providers self-register: `provider.Register("azure", azure.NewProvider)`

### 3. Azure Provider (`internal/azure/`)

**Uses Azure SDK for Go** (not CLI wrapper).

**Client**: Caches `armkeyvault.VaultsClient` and `azsecrets.Client` per vault.

**Authentication**: `DefaultAzureCredential` (Azure CLI, Managed Identity, env vars, Service Principal)

**Performance**: Client caching, connection pooling, no subprocess overhead.

### 4. HashiCorp Vault Provider (`internal/hashicorp/`)

**Uses Vault API SDK.**

**Configuration**: Address, token, namespace (from config or `VAULT_*` env vars)

**Supports**: KV v2 secret engines only.

### 5. Output Formatters (`internal/output/`)

**Plain** (default): One item per line, for piping to fzf
**JSON** (`--format json`): Structured data for scripting

### 6. Data Models (`pkg/models/`)

**Provider-agnostic structs**: `Vault`, `Secret`, `SecretValue`

Common fields + extensible `Metadata map[string]string` for provider-specific data.

### 7. CLI Commands (`cmd/main.go`)

**Available commands**:
- `list-providers`: Show enabled providers
- `list-vaults --provider azure [--instance prod]`: List vaults
- `list-secrets --vault X`: List secrets
- `get-secret --vault X --name Y [--copy]`: Get secret value
- `walk-secrets [--vault X]`: Interactive tree walk

**Flags**: `--provider`, `--instance`, `--vault`, `--name`, `--copy`, `--format`

### 8. Tmux Plugin (`scripts/`)

**Workflow**: `prefix + K` → Select vault (fzf) → Select secret (fzf) → Copy to clipboard

## Execution Flow

```
User: prefix + K

1. list-vaults → Load config → Get default instance → Call provider SDK → Output
2. fzf selection → "my-prod-vault"
3. list-secrets --vault my-prod-vault → Output
4. fzf selection → "database-password"
5. get-secret --vault my-prod-vault --name database-password --copy
6. Secret copied to clipboard
```

## Project Structure

```
smart-keyvault/
├── cmd/main.go                 # CLI entry point (Cobra)
├── internal/
│   ├── config/                 # Viper config system (types, loader, helpers)
│   ├── provider/               # Provider interface & registry
│   ├── azure/                  # Azure provider (SDK client)
│   ├── hashicorp/              # Vault provider (API client)
│   ├── output/                 # Formatters (plain, json)
│   └── clipboard/              # Clipboard integration
├── pkg/models/                 # Data models (Vault, Secret, SecretValue)
├── scripts/                    # Tmux plugin (browse-secrets.sh)
├── smart-keyvault.tmux         # TPM entry point
└── config.example.yaml
```

## Security

**Secret Handling**: Secrets only to stdout/clipboard, no logging, no disk persistence
**Authentication**: Provider native auth (Azure CLI, Vault token), no credential storage
**Config**: Stores `${VAR}` references, not actual secrets

## Error Handling

- Errors to stderr, non-zero exit codes
- Clear, actionable messages
- Examples: `Error: subscription_id required (set via config or AZURE_SUBSCRIPTION_ID)`

## Performance

**Native SDKs**: No subprocess overhead, connection pooling, client caching
**Thread-safe**: Concurrent operations supported

## Testing

```bash
make test                # Unit tests
make check               # fmt + vet + test
./bin/smart-keyvault list-vaults --provider azure
./scripts/browse-secrets.sh
```

## Adding New Providers

**Three steps** to add AWS Secrets Manager:

1. Create `internal/aws/provider.go` implementing `Provider` interface
2. Register: `provider.Register("aws", aws.NewProvider)` in `cmd/main.go`
3. Update docs

No changes to CLI, shell scripts, formatters, or models!

## Key Design Decisions

1. **Provider Pattern**: Unified interface for all secret backends
2. **Viper Config**: Multi-instance support with `${VAR}` substitution
3. **Azure SDK**: Native Go integration (not CLI wrapper)
4. **Go/Shell Separation**: Go handles data, shell handles UI
5. **Security**: No credential storage, no disk persistence
6. **Performance**: Client caching, connection pooling

See [DESIGN_DECISIONS.md](DESIGN_DECISIONS.md) for details.
