# Smart KeyVault

A multi-provider tmux plugin for secret management that simplifies browsing and retrieving secrets through interactive selection with fzf-tmux.

## Overview

Smart KeyVault is a tmux plugin that provides a unified interface for browsing and retrieving secrets from multiple secret management systems. It uses a Go binary with a provider architecture to fetch data and presents it through fzf-tmux for interactive selection.

**Supported Providers:**
- **Azure KeyVault** - via Azure CLI (`az keyvault`)
- **Hashicorp Vault** - via Vault API client

No need to remember complex commands or vault names anymore!

## Features

- 🔌 **Multi-Provider**: Support for Azure KeyVault and Hashicorp Vault
- 🔍 **Interactive Selection**: Browse vaults and secrets using fzf-tmux
- 🚀 **Simple**: Just press a keybinding and select from the menu
- 📋 **Copy to Clipboard**: Direct clipboard integration (no manual copy needed)
- ⚡ **Fast**: Go binary for quick data fetching
- 🏗️ **Extensible**: Provider architecture for easy addition of new backends
- 🎯 **User-Friendly**: No parameters to remember

## Architecture

```
┌──────────────────────────────────────────┐
│  Tmux TPM Plugin (Shell)                 │
│  - Keybindings                           │
│  - fzf-tmux menu display                 │
└─────────────┬────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────┐
│  Go Binary (smart-keyvault)              │
│  - Provider Registry                     │
│  - Clipboard Integration                 │
│  - Output Formatters                     │
└─────────────┬────────────────────────────┘
              │
       ┌──────┴───────┐
       ▼              ▼
┌─────────────┐  ┌──────────────┐
│   Azure     │  │  Hashicorp   │
│  Provider   │  │    Vault     │
│             │  │   Provider   │
└──────┬──────┘  └──────┬───────┘
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────┐
│  az CLI     │  │  Vault API   │
└─────────────┘  └──────────────┘
```

**Flow**:
1. User presses tmux keybinding
2. Plugin prompts for provider selection (if multiple enabled)
3. Go binary lists vaults from selected provider
4. fzf-tmux displays vault list
5. User selects vault
6. Go binary lists secrets from that vault
7. fzf-tmux displays secret list
8. User selects secret
9. Go binary retrieves secret and copies to clipboard via gopasspw/clipboard
10. Confirmation message displayed

## Technology Stack

- **Go 1.21+**: Core binary with provider architecture
  - [Cobra](https://github.com/spf13/cobra): CLI framework
  - [gopasspw/clipboard](https://github.com/gopasspw/clipboard): Clipboard integration
  - [Hashicorp Vault SDK](https://github.com/hashicorp/vault): Vault provider
- **Shell Scripts**: Tmux plugin implementation
- **fzf-tmux**: Interactive fuzzy finder UI
- **tmux**: Terminal multiplexer with TPM

## Prerequisites

### For Azure KeyVault Provider
- Azure CLI (`az`) installed and authenticated (`az login`)

### For Hashicorp Vault Provider
- Vault server accessible
- `VAULT_ADDR` and `VAULT_TOKEN` environment variables set (or configured)

### Common Requirements
- fzf installed
- tmux with TPM (Tmux Plugin Manager)
- Go 1.21+ (for building from source)

## Installation

### Via TPM (Recommended)

Add to your `~/.tmux.conf`:

```bash
set -g @plugin 'yourusername/smart-keyvault'
```

Then install:
- Press `prefix + I` to fetch and install the plugin
- The Go binary will be built automatically on first install

### Manual Installation

```bash
git clone https://github.com/yourusername/smart-keyvault.git ~/.tmux/plugins/smart-keyvault
cd ~/.tmux/plugins/smart-keyvault
make install
```

## Usage

### Keybindings

Default keybindings (customizable):

- `prefix + K` - Browse and copy secrets (full workflow)
- `prefix + k` - Quick secret lookup from default vault

### Workflow Example

```
User: <prefix> + K

┌─────────────────────────────────────────┐
│ Select Azure KeyVault:                  │
│ > my-prod-vault                         │
│   my-dev-vault                          │
│   shared-vault                          │
└─────────────────────────────────────────┘
        ↓ (user selects)
┌─────────────────────────────────────────┐
│ Select Secret (my-prod-vault):          │
│ > database-password                     │
│   api-key                               │
│   oauth-secret                          │
│   stripe-webhook-secret                 │
└─────────────────────────────────────────┘
        ↓ (user selects)
Secret 'database-password' copied to clipboard!
```

## Configuration

Optional configuration file: `~/.config/smart-keyvault/config.yaml`

```yaml
# Default vault (skip vault selection if set)
default_vault: "my-prod-vault"

# fzf-tmux options
fzf:
  height: "40%"
  border: "rounded"
  preview: false

# Filters
filters:
  enabled_only: true  # Only show enabled secrets
```

## Project Structure

```
smart-keyvault/
├── cmd/
│   └── smart-keyvault/      # Go binary entry point
│       └── main.go
├── internal/
│   ├── azure/               # Azure KeyVault wrapper
│   │   ├── client.go        # az CLI executor
│   │   ├── vault.go         # Vault operations
│   │   └── secret.go        # Secret operations
│   └── output/              # Output formatters
│       ├── json.go
│       └── plain.go
├── pkg/
│   └── models/              # Data models
│       └── types.go
├── scripts/                 # Tmux plugin shell scripts
│   ├── browse-secrets.sh    # Main workflow script
│   └── utils.sh             # Helper functions
├── smart-keyvault.tmux      # TPM plugin entry point
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── ARCHITECTURE.md
```

## Go Binary Commands

The binary provides simple commands that output data for fzf:

```bash
# List available providers
smart-keyvault list-providers
# Output: azure, hashicorp

# List vaults from a specific provider (one per line)
smart-keyvault list-vaults --provider azure
smart-keyvault list-vaults --provider hashicorp

# List secrets in a vault (one per line)
smart-keyvault list-secrets --provider azure --vault my-prod-vault

# Get secret value (outputs to stdout)
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret

# Get secret and copy to clipboard directly
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret --copy

# JSON output (for scripting/parsing)
smart-keyvault list-vaults --provider azure --format json
smart-keyvault list-secrets --provider azure --vault my-vault --format json
```

## Development

### Build the Go binary

```bash
make build
```

### Test

```bash
make test
```

### Run locally

```bash
# Test the binary
go run cmd/smart-keyvault/main.go list-vaults

# Test with fzf
go run cmd/smart-keyvault/main.go list-vaults | fzf

# Test the full workflow
./scripts/browse-secrets.sh
```

## Customization

### Custom Keybindings

In your `~/.tmux.conf`:

```bash
# Change default keybindings
set -g @smart-keyvault-key 'C-k'        # Ctrl+k instead of prefix+K
set -g @smart-keyvault-quick-key 'M-k'  # Alt+k for quick access

# fzf-tmux options
set -g @smart-keyvault-fzf-height '50%'
set -g @smart-keyvault-fzf-border 'rounded'
```

## Roadmap

- [x] Basic vault and secret listing
- [x] Copy secret to clipboard
- [ ] Support for certificates and keys
- [ ] Secret metadata preview
- [ ] Multiple output formats (JSON, YAML)
- [ ] Secret version history
- [ ] Batch operations
- [ ] Configuration file support

## License

MIT License

## Support

For issues and questions, please open an issue on GitHub.
