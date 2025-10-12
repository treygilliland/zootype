# zootype

Minimal typing test CLI written in many languages.

Perfect for practicing between builds or agent prompts.

## Overview

Terminal-based typing trainer with raw mode input and real-time feedback.

**Features:**

- Color-coded feedback (green=correct, red=error, yellow cursor, gray=remaining)
- Live accuracy tracking (corrected & raw)
- WPM calculation
- Backspace support
- CTRL-C to view results early

## Goals

- No frameworks and minimal dependencies
  - Lightweight binaries and fast builds
  - Highlights the "zen" of each language
  - Readable and extendable by anyone
- Minimal and customizable out of the box
  - Single configuration file
  - Works in any terminal that supports ASCII + ANSI color codes

## Installation

```bash
make install  # builds and installs zootype to ~/.local/bin
zootype       # run from anywhere!
```

## Configuration

`zootype` uses a simple TOML config file to specify which language implementation to run.

Example configuration:

```toml
# Which binary to run when calling 'zootype'
# Options: gophertype, pythontype
binary = "gophertype"
```

**Config locations:**

- Development: `zootype.toml` in project root (when using `make run`)
- Installed: `~/.config/zootype/zootype.toml` (when using `zootype` command)

**Quick config edit:**

```bash
zootype config  # opens config file in your $EDITOR
```

## Usage

### CLI

After running `make install`, use the `zootype` command from anywhere:

```bash
zootype         # start typing practice
zootype config  # edit config file
zootype build   # rebuild and reinstall binaries from source
```

### Development

```bash
make            # builds all binaries to bin/
make run        # builds and runs binary specified in zootype.toml
make install    # builds and installs to ~/.local/bin
make uninstall  # removes from ~/.local/bin
make clean      # removes bin/

# Low-level: use build.sh directly
./build.sh                    # defaults to 'build'
./build.sh build pythontype   # build specific binary
./build.sh install            # builds and installs
./build.sh uninstall          # uninstalls
```

## Motivation

I love programming languages and keyboards.
I've spent a lot of time on [monkeytype](https://monkeytype.com/) and wanted a way to practice from my terminal.

Writing an interactive CLI showcases a language's functionality and ecosystem well:

- Terminal and File I/O (raw mode, ANSI escape codes)
- String manipulation and character-by-character processing
- Real-time state management
- Time tracking and calculations
- Performance (must be responsive for typing feedback)
- Build systems and runtimes

## Languages

Currently implemented:

- **Go (gophertype)** - Full implementation with all features

Coming soon:

- **Python (pythontype)** - In development

### Adding New Implementations

To add a new language implementation:

1. Create a directory for your implementation (e.g., `crabtype/`)
2. Add the language name to `LANGUAGES` in `build.sh`
3. Create a `build_<name>()` function in `build.sh` with your build commands
4. (Optional) Create a wrapper script if needed (e.g., `crabtype.sh`)

Example for Rust:

```sh
# In build.sh
LANGUAGES="gophertype pythontype crabtype"

build_crabtype() {
    echo "Building crabtype..."
    cd "$SCRIPT_DIR/crabtype"
    cargo build --release
    cp target/release/crabtype "$BIN_DIR/crabtype"
}
```

### Coming Soon

- JavaScript/TypeScript (Dinotype)
- Rust (Crabtype)
- Zig (Iguanatype)
- OCaml (Cameltype)
- C++ ([Rattype](https://news.ycombinator.com/item?id=44631253))

### Maybe One Day

- PHP (Elephanttype)
- Swift (Swifttype)
  - Swift vs Swallow, [see heated debate here](https://github.com/swiftlang/swift/issues/44791)
- Shell (Eggtype)
- Elixir (Phoenixtype)
- SQL (Ducktype)
  - Obviously will have to get creative here
- Dart/Flutter (Hummingbirdtype)

These need a better mascot before I consider them:

- Haskell (Lamb(da)type)
- C (Bugtype)
- Java (Duketype)
  - Bean (Spring), Elephant (Gradle), Raven (Maven), Duke -> Devil
- Lisp (Lisptype)
  - Clojure or Janet
- Assembly (Robottype)
