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

// run orchestrates typing sessions in a loop until user chooses to exit.
// Loads config, generates text, enables raw terminal mode, and runs session loop.
func run() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Enable raw mode for character-by-character input; defer ensures cleanup
	restore, err := enableRawMode()
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	defer restore()

	// Start a single keyboard reading goroutine that stays alive for the entire program
	keyChan := make(chan byte)
	go func() {
		buf := make([]byte, 1)
		for {
			if _, err := os.Stdin.Read(buf); err != nil {
				return
			}
			keyChan <- buf[0]
		}
	}()

	for {
		target, err := getSessionText(config)
		if err != nil {
			return fmt.Errorf("failed to get session text: %w", err)
		}

		state := newTypingState(target, config)

		fmt.Printf("Hello from %sgophertype%s\n\n", ansiBlue, ansiReset)

		runTypingSession(state, keyChan)

		playAgain := promptToContinue(keyChan)
		if !playAgain {
			return nil
		}

		// Clear screen for next session
		fmt.Print("\033[2J\033[H")
	}
}
