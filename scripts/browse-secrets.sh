#!/usr/bin/env bash
set -e

# Get binary path from tmux environment
BINARY="${SMART_KEYVAULT_BIN:-smart-keyvault}"
FZF_HEIGHT="${SMART_KEYVAULT_FZF_HEIGHT:-60%}"
FZF_WIDTH="${SMART_KEYVAULT_FZF_WIDTH:-80%}"

# Check if binary exists
if ! command -v "$BINARY" &> /dev/null; then
    tmux display-message "Error: smart-keyvault binary not found"
    exit 1
fi

# Check if fzf is available
if ! command -v fzf &> /dev/null; then
    tmux display-message "Error: fzf not found. Please install fzf"
    exit 1
fi

# Step 1: Select provider
provider=$("$BINARY" list-providers | fzf-tmux -p "$FZF_WIDTH,$FZF_HEIGHT" --prompt="Select Provider: " --border=rounded || true)

if [[ -z "$provider" ]]; then
    exit 0  # User cancelled
fi

# Step 2: Select vault
vault=$("$BINARY" list-vaults --provider "$provider" | fzf-tmux -p "$FZF_WIDTH,$FZF_HEIGHT" --prompt="Select Vault ($provider): " --border=rounded || true)

if [[ -z "$vault" ]]; then
    exit 0  # User cancelled
fi

# Step 3: Select secret
secret=$("$BINARY" list-secrets --provider "$provider" --vault "$vault" | fzf-tmux -p "$FZF_WIDTH,$FZF_HEIGHT" --prompt="Select Secret ($vault): " --border=rounded || true)

if [[ -z "$secret" ]]; then
    exit 0  # User cancelled
fi

# Step 4: Get secret and copy to clipboard
if "$BINARY" get-secret --provider "$provider" --vault "$vault" --name "$secret" --copy 2>&1; then
    tmux display-message "✓ Secret '$secret' copied to clipboard!"
else
    tmux display-message "✗ Failed to retrieve secret '$secret'"
    exit 1
fi
