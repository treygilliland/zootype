# gophertype

Terminal-based typing practice tool with real-time WPM and accuracy tracking.

## Installation

### Homebrew (macOS)

```bash
brew install treygilliland/tap/gophertype
```

### From Source

```bash
cd gophertype
go build -o gophertype .
```

## Usage

```bash
gophertype              # start typing practice (default: 30 second timed mode)
gophertype -w 100       # practice with 100 words (untimed)
gophertype -s sentences # practice with sentences
gophertype -t 60        # timed mode (60 seconds)
gophertype --version    # show version information
```

### Flags

**Mode Selection (mutually exclusive):**

- `-t`, `--time <N>`: Timed mode - type as many words as possible in N seconds (default: 30, takes precedence)
- `-w`, `--words <N>`: Word count mode - complete N words, untimed

**Text Options:**

- `-s`, `--source <TYPE>`: Text source - `words` or `sentences` (default: words)

**Other:**

- `--version`: Print version information

## Configuration

**gophertype works out-of-the-box with sensible defaults** - no configuration needed!

Default behavior:

- 30 second timed mode
- Random words from top 1000 English words
- 50 words (when using word count mode with `-w`)

### Optional Config File

To customize defaults, create `~/.config/gophertype/config.json`:

```bash
mkdir -p ~/.config/gophertype
cat > ~/.config/gophertype/config.json << 'EOF'
{
  "text_source": "words",
  "word_count": 50,
  "time_seconds": 30
}
EOF
```

**Available options:**

- `text_source`: `"words"` or `"sentences"`
- `word_count`: Number of words for untimed mode
- `time_seconds`: Duration for timed mode (0 disables timer)

**Note:** If both `time_seconds` and `word_count` are set, `time_seconds` takes precedence. CLI flags override config file settings.

## Features

- **Real-time feedback**: Color-coded input (green=correct, red=error, yellow=cursor)
- **Accuracy tracking**: Corrected and raw accuracy metrics
- **WPM calculation**: Words per minute based on standard 5-character words
- **Timed mode**: Practice against the clock
- **Backspace support**: Correct mistakes as you type
- **Interrupt with Ctrl+C**: View results before completion

## Development

```bash
# Build
go build -o bin/gophertype .

# Run
./bin/gophertype

# Test
go test ./...
```

## Implementation Notes

- Raw terminal mode via `golang.org/x/term`
- Word list embedded at compile time via `go:embed`
- Minimal dependencies for fast builds and small binaries
- Goroutine-based keyboard input handling
