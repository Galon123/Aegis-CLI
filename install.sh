#!/usr/bin/env bash
set -euo pipefail

#Config details
readonly BIN_NAME="aegis"
readonly DEFAULT_INSTALL_DIR="$HOME/.local/bin"

# Script variables
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
DRY_RUN=false

#Colors
if [[ -t 1 ]]; then
    readonly RED='\033[0;31m'
    readonly GREEN='\033[0;32m'
    readonly YELLOW='\033[1;33m'
    readonly NC='\033[0m'
else    
    readonly RED=''; readonly GREEN=''; readonly YELLOW=''; readonly NC=''
fi

#Logging Functions
log_info() { echo -e "${GREEN}✓ ${NC}$*" >&2; }

log_warn() { echo -e "${YELLOW}⚠ ${NC}$*" >&2; }

log_error() { echo -e "${RED}✗ ${NC}$*" >&2; }

usage() {
cat <<EOF
Usage: install.sh [options]

Compiles and installs Aegis from the current directory.

Options:
-d, --dir       Install directory (default: $DEFAULT_INSTALL_DIR)
--dry-run             Show what would be done without making changes
-h, --help            Show this help message and exit
EOF
}

while [[ $# -gt 0 ]]; do
case "$1" in
    -d|--dir) INSTALL_DIR="$2"; shift 2 ;;
    --dry-run) DRY_RUN=true; shift ;;
    -h|--help) usage; exit 0 ;; 
    *) log_error "Unknown option: $1"; usage; exit 1 ;;
esac
done

#Ensuring Correct Compile
if [[ "$DRY_RUN" == false ]]; then
    log_info "Compiling Aegis..."
    if ! go build -o "$BIN_NAME"; then
        log_error "Compilation failed. Ensure Go is installed."
        exit 1
    fi
fi  

validate_install_dir() {
if [[ ! -d "$INSTALL_DIR" ]]; then
    mkdir -p "$INSTALL_DIR"
    log_info "Created install directory: $INSTALL_DIR"
fi
}

if [[ "$DRY_RUN" == false ]]; then
    validate_install_dir
    mv "$BIN_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BIN_NAME"
    log_info "Binary installed to: $INSTALL_DIR/$BIN_NAME"
fi

detect_current_shell() { basename "${SHELL:-/bin/bash}"; }

get_shell_config_path() {
case "$(detect_current_shell)" in
    bash) echo "$HOME/.bashrc" ;;
    zsh) echo "$HOME/.zshrc" ;;
    fish) echo "$HOME/.config/fish/config.fish" ;;
    *) echo "" ;;
esac
}

update_shell_config() {
local config_file
config_file=$(get_shell_config_path)

if [[ -n "$config_file" && -f "$config_file" ]]; then
    if ! grep -q "$INSTALL_DIR" "$config_file" 2>/dev/null; then
        if [[ "$DRY_RUN" == false ]]; then
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$config_file"
        fi
        log_info "Added $INSTALL_DIR to $config_file"
    fi
fi


}

update_shell_config

if [[ "$DRY_RUN" == false ]]; then
echo -e "\n${GREEN}Installation complete!${NC}"
echo "Next steps:"
echo "  1. Reload your shell: source $(get_shell_config_path)"
echo "  2. Run: aegis -help"
else
log_info "Dry run complete - no changes were made."
fi