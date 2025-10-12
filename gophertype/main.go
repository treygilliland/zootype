// gophertype is a terminal-based typing practice tool with real-time WPM and accuracy tracking.
package main

import (
	"fmt"
	"os"
)

// main entry point
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run orchestrates a typing session:
// load config, generate text, enable raw terminal mode, run session loop, and display results.
func run() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	target, err := getSessionText(config)
	if err != nil {
		return fmt.Errorf("failed to get session text: %w", err)
	}

	state := newTypingState(target)

	fmt.Printf("Hello from %sgophertype%s\n\n", ansiBlue, ansiReset)

	// Enable raw mode for character-by-character input; defer ensures cleanup
	restore, err := enableRawMode()
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	defer restore()

	runTypingSession(state)
	return nil
}
