// gophertype is a terminal-based typing practice tool with real-time WPM and accuracy tracking.
package main

import (
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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

	// Set up terminal (alternate screen buffer + raw mode); defer ensures cleanup
	restore, err := setupTerminal()
	if err != nil {
		return fmt.Errorf("failed to setup terminal: %w", err)
	}
	defer restore()

	// Use terminal width for display, erroring if it's too narrow
	termWidth, err := getAndValidateTerminalWidth()
	if err != nil {
		return err
	}

	// Start keyboard reader goroutine
	keyChan := startKeyboardReader()

	for {
		target, err := getSessionText(config)
		if err != nil {
			return fmt.Errorf("failed to load session text: %w", err)
		}

		// Run sessions with current target until user wants new text or exits
		if !runSessionLoop(target, config, termWidth, keyChan) {
			return nil
		}
	}
}

// runSessionLoop runs typing sessions with the given target text until the user
// wants new text or exits. Returns true if user wants next text, false if exiting.
func runSessionLoop(target string, config Config, termWidth int, keyChan <-chan byte) bool {
	for {
		state := newTypingState(target, config, termWidth)

		action := runTypingSession(state, keyChan)

		// If interrupted with Ctrl+C, exit immediately
		if action == ActionInterrupt {
			return false
		}

		action = promptToContinue(keyChan)

		switch action {
		case ActionExit:
			return false
		case ActionRetry:
			continue
		case ActionNext:
			return true
		}
	}
}
