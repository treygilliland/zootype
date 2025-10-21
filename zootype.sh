#!/bin/sh
set -e

# Usage:
#   ./zootype.sh [build|install|uninstall] [languages...]
#   zootype [-b backend] [args...]        (when installed)
#   zootype build                          (rebuild from source)
#
# To add a new language:
# 1. Add it to LANGUAGES list below
# 2. Create a build_<name>() function below that

LANGUAGES="gophertype pythontype cameltype rattype crabtype"
DEFAULT_BINARY="gophertype"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"

# Determine if we're being called as the installed 'zootype' wrapper
SCRIPT_NAME="$(basename "$0")"
if [ "$SCRIPT_NAME" = "zootype" ]; then
    IS_WRAPPER=1
elif [ -n "$ZOOTYPE_BIN_DIR" ]; then
    # Development mode: ZOOTYPE_BIN_DIR is set (e.g., from make run)
    IS_WRAPPER=1
else
    IS_WRAPPER=0
fi

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

build_cameltype() {
    echo "Building cameltype..."
    cd "$SCRIPT_DIR/cameltype"
    opam exec -- dune build
    rm -f "$BIN_DIR/cameltype"
    cp "_build/default/main.exe" "$BIN_DIR/cameltype"
    chmod +x "$BIN_DIR/cameltype"
}

build_rattype() {
    echo "Building rattype..."
    cd "$SCRIPT_DIR/rattype"
    make clean
    make
    rm -f "$BIN_DIR/rattype"
    cp "rattype" "$BIN_DIR/rattype"
    chmod +x "$BIN_DIR/rattype"
}

build_crabtype() {
    echo "Building crabtype..."
    cd "$SCRIPT_DIR/crabtype"
    cargo build --release
    rm -f "$BIN_DIR/crabtype"
    cp "target/release/crabtype" "$BIN_DIR/crabtype"
    chmod +x "$BIN_DIR/crabtype"
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
    
    echo "Installing zootype development wrapper..."
    
    mkdir -p "$INSTALL_DIR"
    
    if [ ! -d "$BIN_DIR" ] || [ -z "$(ls -A "$BIN_DIR" 2>/dev/null | grep -v '\.whl$')" ]; then
        echo "Error: No binaries found in bin/. Run '$0' to build first."
        exit 1
    fi
    
    # Install zootype wrapper for development
    cp "$SCRIPT_DIR/zootype.sh" "$INSTALL_DIR/zootype"
    chmod +x "$INSTALL_DIR/zootype"
    
    echo "Installed successfully!"
    echo ""
    echo "Development wrapper installed to $INSTALL_DIR/zootype"
    echo ""
    echo "Usage:"
    echo "  zootype -b gophertype    # run dev build of gophertype"
    echo "  zootype -b pythontype    # run dev build of pythontype"
    echo ""
    echo "Note: This uses binaries from $SCRIPT_DIR/bin/"
    echo "      Your Homebrew installations remain unchanged."
}

uninstall_binaries() {
    INSTALL_DIR="$HOME/.local/bin"
    
    echo "Uninstalling zootype development wrapper..."
    rm -f "$INSTALL_DIR/zootype"
    
    echo "Uninstalled successfully!"
    echo "Note: Your Homebrew installations remain unchanged."
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

# Wrapper mode: when installed as 'zootype'
if [ "$IS_WRAPPER" = "1" ]; then
    # Find the source directory to locate bin/
    if [ -n "$ZOOTYPE_BIN_DIR" ]; then
        # Development mode with ZOOTYPE_BIN_DIR set
        SOURCE_BIN_DIR="$ZOOTYPE_BIN_DIR"
    else
        # Installed mode - find source directory
        if [ -d "$HOME/code/zootype/bin" ]; then
            SOURCE_BIN_DIR="$HOME/code/zootype/bin"
        elif [ -d "$SCRIPT_DIR/bin" ]; then
            SOURCE_BIN_DIR="$SCRIPT_DIR/bin"
        else
            echo "Error: Cannot find zootype bin/ directory"
            echo ""
            echo "The zootype wrapper looks for development builds in:"
            echo "  - \$HOME/code/zootype/bin/"
            echo "  - Current directory"
            echo ""
            echo "Run 'make' from the zootype source directory to build binaries."
            exit 1
        fi
    fi
    
    # Handle special 'build' command
    if [ "$1" = "build" ]; then
        SOURCE_DIR=$(dirname "$SOURCE_BIN_DIR")
        if [ ! -f "$SOURCE_DIR/zootype.sh" ]; then
            echo "Error: Cannot find zootype.sh in $SOURCE_DIR"
            exit 1
        fi
        
        echo "Rebuilding zootype binaries..."
        cd "$SOURCE_DIR"
        ./zootype.sh
        exit $?
    fi
    
    # Parse -b flag for binary selection
    BINARY="$DEFAULT_BINARY"
    while [ $# -gt 0 ]; do
        case "$1" in
            -b)
                if [ -z "$2" ]; then
                    echo "Error: -b requires an argument"
                    echo "Available backends: $LANGUAGES"
                    exit 1
                fi
                BINARY="$2"
                shift 2
                ;;
            *)
                break
                ;;
        esac
    done
    
    BINARY_PATH="$SOURCE_BIN_DIR/$BINARY"
    
    if [ ! -f "$BINARY_PATH" ]; then
        echo "Error: Binary '$BINARY' not found at $BINARY_PATH"
        echo ""
        echo "Available backends: $LANGUAGES"
        echo ""
        echo "Run 'make' from the zootype source directory to build binaries."
        exit 1
    fi
    
    # Execute the selected binary with remaining arguments
    exec "$BINARY_PATH" "$@"
fi

# Build mode: when run as './zootype.sh'
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
