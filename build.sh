#!/bin/sh
set -e

# Usage: ./build.sh [build|install|uninstall] [languages...]
#
# To add a new language:
# 1. Add it to LANGUAGES list below
# 2. Create a build_<name>() function below that

LANGUAGES="gophertype pythontype"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"

build_gophertype() {
    echo "Building gophertype..."
    mkdir -p "$SCRIPT_DIR/gophertype/data"
    cp "$SCRIPT_DIR/data/top-1000-words.txt" "$SCRIPT_DIR/gophertype/data/top-1000-words.txt"
    cd "$SCRIPT_DIR/gophertype"
    go build -o "$BIN_DIR/gophertype" *.go
}

build_pythontype() {
    echo "Building pythontype..."
    {
        echo '#!/bin/sh'
        echo "exec uv --directory \"$SCRIPT_DIR/pythontype\" run pythontype \"\$@\""
    } > "$BIN_DIR/pythontype"
    chmod +x "$BIN_DIR/pythontype"
}

build_binary() {
    lang="$1"
    func="build_$lang"
    
    if ! command -v "$func" >/dev/null 2>&1; then
        echo "Unknown binary: $lang"
        return 1
    fi
    
    "$func"
}

install_binaries() {
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR="$HOME/.config/zootype"
    
    echo "Installing zootype..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    
    if [ ! -d "$BIN_DIR" ] || [ -z "$(ls -A "$BIN_DIR" 2>/dev/null | grep -v '\.whl$')" ]; then
        echo "Error: No binaries found in bin/. Run '$0' to build first."
        exit 1
    fi
    
    for binary in "$BIN_DIR"/*; do
        [ -f "$binary" ] || continue
        [ -x "$binary" ] || continue
        basename=$(basename "$binary")
        cp "$binary" "$INSTALL_DIR/$basename"
        echo "Installed $basename"
    done
    
    # Write list of available binaries for error messages (one per line)
    for lang in $LANGUAGES; do
        echo "$lang"
    done > "$CONFIG_DIR/binaries"
    
    if [ ! -f "$CONFIG_DIR/zootype.toml" ]; then
        cp "$SCRIPT_DIR/zootype.toml" "$CONFIG_DIR/zootype.toml"
        echo "Created config at $CONFIG_DIR/zootype.toml"
    fi
    
    cp "$SCRIPT_DIR/zootype.sh" "$INSTALL_DIR/zootype"
    chmod +x "$INSTALL_DIR/zootype"
    
    # Show which binary is currently configured
    current_binary=$(grep "^binary" "$CONFIG_DIR/zootype.toml" | cut -d"=" -f2 | tr -d ' "')
    
    echo "Installed successfully!"
    echo ""
    echo "Run 'zootype' from anywhere to practice typing"
    echo ""
    echo "Currently using: $current_binary"
    echo "To change: zootype config"
}

uninstall_binaries() {
    INSTALL_DIR="$HOME/.local/bin"
    CONFIG_DIR="$HOME/.config/zootype"
    
    echo "Uninstalling zootype..."
    rm -f "$INSTALL_DIR/zootype"
    
    for lang in $LANGUAGES; do
        rm -f "$INSTALL_DIR/$lang"
    done
    
    echo "Uninstalled. Config preserved at $CONFIG_DIR/zootype.toml"
}

build_all() {
    mkdir -p "$BIN_DIR"
    
    if [ $# -eq 0 ]; then
        for lang in $LANGUAGES; do
            if [ -d "$SCRIPT_DIR/$lang" ]; then
                build_binary "$lang"
            fi
        done
    else
        for lang in "$@"; do
            build_binary "$lang"
        done
    fi
    
    echo "Build complete!"
    echo ""
}

case "${1:-build}" in
    build)
        shift
        build_all "$@"
        ;;
    install)
        build_all
        install_binaries
        ;;
    uninstall)
        uninstall_binaries
        ;;
    *)
        echo "Usage: $0 {build|install|uninstall} [languages...]"
        exit 1
        ;;
esac

