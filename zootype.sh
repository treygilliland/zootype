#!/bin/sh
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Allow overriding paths for local development
CONFIG_FILE="${ZOOTYPE_CONFIG:-$HOME/.config/zootype/zootype.toml}"
BIN_DIR="${ZOOTYPE_BIN_DIR:-$HOME/.local/bin}"

# Handle special commands
if [ "$1" = "config" ]; then
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "Config file not found at $CONFIG_FILE"
        exit 1
    fi
    
    # Use EDITOR if set, otherwise try common editors
    EDITOR="${EDITOR:-${VISUAL:-$(command -v nano || command -v vim || command -v vi)}}"
    
    if [ -z "$EDITOR" ]; then
        echo "No editor found. Set EDITOR environment variable or install nano/vim"
        exit 1
    fi
    
    # Use eval to handle editors with arguments (e.g., "cursor --wait")
    eval "$EDITOR" '"$CONFIG_FILE"'
    exit $?
fi

if [ "$1" = "build" ]; then
    # Find the zootype source directory by looking for build.sh
    # Check if we're in the source directory
    if [ -f "$SCRIPT_DIR/build.sh" ]; then
        SOURCE_DIR="$SCRIPT_DIR"
    # Check common locations
    elif [ -f "$HOME/code/zootype/build.sh" ]; then
        SOURCE_DIR="$HOME/code/zootype"
    else
        echo "Error: Cannot find zootype source directory with build.sh"
        echo "Please run './build.sh install' from the source directory"
        exit 1
    fi
    
    echo "Rebuilding zootype binaries..."
    cd "$SOURCE_DIR"
    ./build.sh install
    exit $?
fi

if [ -f "$CONFIG_FILE" ]; then
    BINARY=$(grep "^binary" "$CONFIG_FILE" | cut -d"=" -f2 | tr -d ' "')
else
    BINARY="gophertype"
fi

BINARY_PATH="$BIN_DIR/$BINARY"

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary '$BINARY' not found"
    echo ""
    echo "Available binaries:"
    
    # Read list of installed binaries
    config_dir=$(dirname "$CONFIG_FILE")
    if [ -f "$config_dir/binaries" ]; then
        while read -r lang; do
            echo "  - $lang"
        done < "$config_dir/binaries"
    else
        echo "  (none found - did you run 'make install'?)"
    fi
    
    echo ""
    echo "To fix: zootype config"
    exit 1
fi

exec "$BINARY_PATH" "$@"

