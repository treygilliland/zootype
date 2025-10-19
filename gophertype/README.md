# gophertype

Go implementation of zootype - a terminal-based typing test.

Part of the [zootype project](https://github.com/treygilliland/zootype).

## Installation

### Homebrew (macOS)

```bash
brew install treygilliland/tap/gophertype
gophertype
```

### From Source

```bash
cd gophertype
go build -o gophertype .
./gophertype
```

## Usage

See the [main zootype README](../README.md#usage) for complete CLI documentation and usage examples.

Quick start:

```bash
gophertype              # 30 second timed mode with random words
gophertype -t 60        # 60 second timed mode
gophertype -w 50        # type 50 words, untimed
gophertype -v           # show version
```

## Implementation Details

**Language:** Go 1.21+

**Key Features:**

- Raw terminal I/O via `golang.org/x/term`
- Concurrent goroutines for input/display management
- Word list embedded at compile time via `go:embed`
- Zero external dependencies (only Go stdlib + x/term)
- Alternate screen buffer to preserve terminal history
- ANSI escape sequences for color and cursor control

**Architecture:**

- `config.go` - CLI flags and configuration
- `session.go` - Main typing test loop and terminal handling
- `stats.go` - WPM and accuracy calculations
- `display.go` - Terminal rendering and ANSI codes
- `main.go` - Entry point

## Development

```bash
# Build
go build -o bin/gophertype .

# Run
./bin/gophertype

# Test
go test ./...

# Run from zootype root with wrapper
cd ..
make
zootype -b gophertype
```

## Dependencies

```
golang.org/x/term  # Terminal control and raw mode
```
