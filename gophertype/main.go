// gophertype is a terminal-based typing test with real-time feedback and WPM tracking.
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

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// orchestrates typing sessions in an infinite loop until the user exits.
func run() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	restore, err := setupTerminal()
	if err != nil {
		return fmt.Errorf("failed to setup terminal: %w", err)
	}
	defer restore()

	termWidth, err := getAndValidateTerminalWidth()
	if err != nil {
		return err
	}

	keyChan := startKeyboardReader()

	for {
		target, err := getSessionText(config)
		if err != nil {
			return fmt.Errorf("failed to load session text: %w", err)
		}

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
