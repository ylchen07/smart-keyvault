#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Build Go binary if not exists
if [[ ! -f "$CURRENT_DIR/bin/smart-keyvault" ]]; then
    echo "Building smart-keyvault..."
    cd "$CURRENT_DIR" && make build
fi

# Get user-configured keybindings or use defaults
keybind=$(tmux show-option -gqv @smart-keyvault-key)
keybind=${keybind:-K}

quick_keybind=$(tmux show-option -gqv @smart-keyvault-quick-key)
quick_keybind=${quick_keybind:-k}

# Get fzf options
fzf_height=$(tmux show-option -gqv @smart-keyvault-fzf-height)
fzf_height=${fzf_height:-60%}

fzf_width=$(tmux show-option -gqv @smart-keyvault-fzf-width)
fzf_width=${fzf_width:-80%}

# Export variables for scripts to use
tmux set-environment -g SMART_KEYVAULT_BIN "$CURRENT_DIR/bin/smart-keyvault"
tmux set-environment -g SMART_KEYVAULT_FZF_HEIGHT "$fzf_height"
tmux set-environment -g SMART_KEYVAULT_FZF_WIDTH "$fzf_width"

# Set keybindings
tmux bind-key "$keybind" run-shell "$CURRENT_DIR/scripts/browse-secrets.sh"
tmux bind-key "$quick_keybind" run-shell "$CURRENT_DIR/scripts/browse-secrets.sh --quick"
