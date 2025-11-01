# Quick Start Guide

Get started with Smart KeyVault in minutes.

## Prerequisites

- Go 1.21+
- fzf
- tmux with TPM (Tmux Plugin Manager)
- Azure CLI (`az`) - for Azure KeyVault
- HashiCorp Vault - for Vault provider (optional)

## Installation

### 1. Build from Source

```bash
git clone https://github.com/ylchen07/smart-keyvault.git
cd smart-keyvault
make build
```

### 2. Test the Binary

```bash
# List providers
./bin/smart-keyvault list-providers

# List vaults (requires authentication)
AZURE_SUBSCRIPTION_ID=xxx ./bin/smart-keyvault list-vaults --provider azure

# Get a secret
./bin/smart-keyvault get-secret --provider azure --vault <vault-name> --name <secret-name> --copy
```

### 3. Install as Tmux Plugin

Add to `~/.tmux.conf`:

```bash
set -g @plugin 'ylchen07/smart-keyvault'
```

Reload tmux:
- Press `prefix + I` to install (TPM)
- Or: `tmux source-file ~/.tmux.conf`

## Configuration

### Option 1: Environment Variables Only

```bash
# Azure
export AZURE_SUBSCRIPTION_ID="xxx-xxx-xxx"

# HashiCorp Vault
export VAULT_ADDR="https://vault.example.com:8200"
export VAULT_TOKEN="your-token"
export VAULT_NAMESPACE="admin/production"  # For Vault Enterprise
```

### Option 2: Config File (Recommended)

Create `~/.config/smart-keyvault/config.yaml`:

```yaml
defaults:
  provider: "azure"

providers:
  azure:
    enabled: true
    instances:
      - name: "prod"
        subscription_id: "${AZURE_SUBSCRIPTION_ID}"
        default: true

      - name: "dev"
        subscription_id: "yyy-yyy-yyy"

  hashicorp:
    enabled: true
    instances:
      - name: "prod-vault"
        address: "https://vault-prod.example.com:8200"
        token: "${VAULT_TOKEN_PROD}"
        namespace: "admin/production"
        default: true

      - name: "local"
        address: "http://127.0.0.1:8200"
        token: "${VAULT_TOKEN}"

fzf:
  height: "40%"
  border: "rounded"
```

Copy example config:
```bash
mkdir -p ~/.config/smart-keyvault
cp config.example.yaml ~/.config/smart-keyvault/config.yaml
# Edit and add your instances
```

## Usage

### Tmux Plugin

Press `prefix + K` in tmux:
1. Select provider (if multiple)
2. Select vault
3. Select secret
4. Secret copied to clipboard!

### CLI Direct Usage

```bash
# List providers
smart-keyvault list-providers

# List vaults (uses default instance from config)
smart-keyvault list-vaults --provider azure

# List vaults from specific instance
smart-keyvault list-vaults --provider azure --instance dev

# List secrets
smart-keyvault list-secrets --provider azure --vault my-vault

# Get secret (output to stdout)
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret

# Get secret and copy to clipboard
smart-keyvault get-secret --provider azure --vault my-vault --name my-secret --copy

# JSON output for scripting
smart-keyvault list-vaults --provider azure --format json
```

## Azure Setup

```bash
# Install Azure CLI
brew install azure-cli  # macOS
# or
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash  # Linux

# Login
az login

# Set subscription (if needed)
az account set --subscription <subscription-id>

# Verify access
az keyvault list
```

## HashiCorp Vault Setup

```bash
# Set environment variables
export VAULT_ADDR='https://vault.example.com:8200'
export VAULT_TOKEN='your-token'
export VAULT_NAMESPACE='admin/production'  # For Vault Enterprise

# Test connection
smart-keyvault list-vaults --provider hashicorp
```

## Troubleshooting

### Binary not found
```bash
make build && make install
```

### Azure authentication error
```bash
az login
az account set --subscription <subscription-id>
```

### Vault connection error
```bash
# Check environment variables
echo $VAULT_ADDR
echo $VAULT_TOKEN

# Test vault connection
vault status
```

### fzf not found
```bash
brew install fzf  # macOS
sudo apt-get install fzf  # Linux
```

### Clipboard not working (Linux)
```bash
sudo apt-get install xclip  # or xsel
```

## Next Steps

- Read [ARCHITECTURE.md](ARCHITECTURE.md) for technical details
- Read [DESIGN_DECISIONS.md](DESIGN_DECISIONS.md) for rationale
- See [README.md](README.md) for full documentation
- Check `config.example.yaml` for all config options

## Custom Keybindings

In `~/.tmux.conf`:

```bash
set -g @smart-keyvault-key 'C-k'    # Ctrl+k instead of prefix+K
set -g @smart-keyvault-fzf-height '50%'
```
