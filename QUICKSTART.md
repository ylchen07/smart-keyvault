# Quick Start Guide

This guide will help you get started with Smart KeyVault.

## Prerequisites

Make sure you have the following installed:
- Go 1.21+
- fzf
- tmux with TPM
- Azure CLI (`az`) - for Azure KeyVault provider

## Installation

### 1. Clone and Build

```bash
git clone https://github.com/ylchen07/smart-keyvault.git
cd smart-keyvault
make build
```

### 2. Test the Binary

```bash
# List available providers
./bin/smart-keyvault list-providers

# List vaults (requires az login)
./bin/smart-keyvault list-vaults --provider azure

# List secrets in a vault
./bin/smart-keyvault list-secrets --provider azure --vault <vault-name>

# Get a secret (copy to clipboard)
./bin/smart-keyvault get-secret --provider azure --vault <vault-name> --name <secret-name> --copy
```

### 3. Install as Tmux Plugin

Add to your `~/.tmux.conf`:

```bash
set -g @plugin 'ylchen07/smart-keyvault'
```

Then reload tmux:
- Press `prefix + I` to install the plugin (TPM)
- Or manually: `tmux source-file ~/.tmux.conf`

### 4. Configure (Optional)

In your `~/.tmux.conf`, you can customize keybindings:

```bash
# Custom keybindings (optional)
set -g @smart-keyvault-key 'K'           # Default: K
set -g @smart-keyvault-quick-key 'k'     # Default: k

# fzf-tmux options (optional)
set -g @smart-keyvault-fzf-height '60%'  # Default: 60%
set -g @smart-keyvault-fzf-width '80%'   # Default: 80%
```

## Usage

### Using the Tmux Plugin

1. Press `prefix + K` (default) in tmux
2. Select provider (if multiple available)
3. Select vault from the list
4. Select secret from the list
5. Secret is automatically copied to clipboard!

### Using the CLI Directly

```bash
# List all providers
smart-keyvault list-providers

# List vaults
smart-keyvault list-vaults --provider azure

# List secrets
smart-keyvault list-secrets --provider azure --vault my-vault

# Get secret value (output to stdout)
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret

# Get secret and copy to clipboard
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret --copy

# JSON output (for scripting)
smart-keyvault list-vaults --provider azure --format json
```

## Azure Setup

1. **Install Azure CLI**:
   ```bash
   # macOS
   brew install azure-cli

   # Linux (Debian/Ubuntu)
   curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
   ```

2. **Login to Azure**:
   ```bash
   az login
   ```

3. **Verify access to KeyVaults**:
   ```bash
   az keyvault list
   ```

## Hashicorp Vault Setup (Coming Soon)

The Hashicorp Vault provider is planned but not yet implemented. To add support:

1. Implement `internal/vault/provider.go`
2. Register in `cmd/smart-keyvault/main.go`
3. Set environment variables:
   ```bash
   export VAULT_ADDR="https://vault.example.com"
   export VAULT_TOKEN="your-token"
   ```

## Troubleshooting

### Binary not found
```bash
# Make sure binary is built
make build

# Or install to PATH
make install
```

### Azure CLI errors
```bash
# Make sure you're logged in
az login

# Check your subscriptions
az account list

# Set default subscription if needed
az account set --subscription <subscription-id>
```

### fzf not found
```bash
# macOS
brew install fzf

# Linux (Debian/Ubuntu)
sudo apt-get install fzf

# Or from source
git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
~/.fzf/install
```

### Clipboard not working
The clipboard functionality uses the system clipboard:
- **Linux**: Requires `xclip` or `xsel`
- **macOS**: Works out of the box
- **Windows**: Works with WSL

Install clipboard tools on Linux:
```bash
sudo apt-get install xclip  # or xsel
```

## Next Steps

- Read [ARCHITECTURE.md](ARCHITECTURE.md) for technical details
- Read [DESIGN_DECISIONS.md](DESIGN_DECISIONS.md) for architectural rationale
- Check [README.md](README.md) for full documentation

## Contributing

Contributions are welcome! Areas to contribute:
1. Implement Hashicorp Vault provider (`internal/vault/`)
2. Add AWS Secrets Manager provider
3. Add GCP Secret Manager provider
4. Improve error messages
5. Add tests
6. Improve documentation

See the provider interface in `internal/provider/provider.go` for implementation guidance.
