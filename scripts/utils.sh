#!/usr/bin/env bash

# Utility functions for smart-keyvault tmux plugin

# Display error message in tmux
tmux_error() {
    local message="$1"
    tmux display-message "Error: $message"
}

# Display success message in tmux
tmux_success() {
    local message="$1"
    tmux display-message "✓ $message"
}

# Check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Validate prerequisites
validate_prerequisites() {
    if ! command_exists fzf; then
        tmux_error "fzf not found. Please install fzf"
        return 1
    fi

    if ! command_exists smart-keyvault && ! command_exists "$SMART_KEYVAULT_BIN"; then
        tmux_error "smart-keyvault binary not found"
        return 1
    fi

    return 0
}
