#!/bin/sh
# eggtype - Shell implementation of zootype typing test

VERSION="dev"

# Parse arguments
if [ "$1" = "--version" ] || [ "$1" = "-v" ]; then
    echo "eggtype $VERSION"
    exit 0
fi

echo "Hello from eggtype"

