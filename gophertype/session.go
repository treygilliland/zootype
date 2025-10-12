package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

// Terminal control constants
const (
	keyCtrlC     = 3   // ASCII code for Ctrl+C (interrupt)
	keyDelete    = 127 // DEL key - commonly sent by backspace on Unix/Mac
	keyBackspace = 8   // BS key - sent by backspace on some terminals
)

// TypingState tracks typing session state. Maintains two accuracy metrics:
// current (decremented on backspace) and raw (never decremented).
type TypingState struct {
	sessionText       string
	position          int
	currentErrors     int    // Decremented on backspace correction
	totalErrors       int    // Never decremented (raw accuracy)
	currentCharsTyped int    // Decremented on backspace
	totalKeystrokes   int    // Never decremented
	charCorrectness   []bool // Per-character correctness for coloring
	backspaceCount    int
	startTime         time.Time
	lastLineCount     int // Lines in previous display (for clearing)
}

// newTypingState initializes a new typing session with the given target text.
func newTypingState(target string) *TypingState {
	return &TypingState{
		sessionText:     target,
		charCorrectness: make([]bool, len(target)),
		startTime:       time.Now(),
	}
}

// enableRawMode enables raw terminal mode for immediate keystroke capture.
// Returns a restore function that must be deferred to restore normal mode.
func enableRawMode() (func(), error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	restore := func() {
		_ = term.Restore(fd, oldState)
		fmt.Print("\r\n")
	}

	return restore, nil
}

// runTypingSession is the main event loop, reading keystrokes and updating display.
func runTypingSession(state *TypingState) {
	displayProgress(state)

	buf := make([]byte, 1)
	for state.position < len(state.sessionText) {
		if _, err := os.Stdin.Read(buf); err != nil {
			fmt.Fprintf(os.Stderr, "\nRead error: %v\n", err)
			return
		}

		key := buf[0]

		if isInterrupt(key) {
			fmt.Print("\n")
			results := NewResults(state)
			results.Print()
			return
		}

		if isBackspace(key) {
			handleBackspace(state)
			displayProgress(state)
			continue
		}

		handleKeystroke(state, key)
		displayProgress(state)
	}

	results := NewResults(state)
	results.Print()
}

func isInterrupt(key byte) bool {
	return key == keyCtrlC
}

func isBackspace(key byte) bool {
	return key == keyDelete || key == keyBackspace
}

// handleBackspace moves cursor back and updates metrics.
// currentErrors decrements, but totalErrors doesn't (used for raw accuracy).
func handleBackspace(state *TypingState) {
	if state.position == 0 {
		return
	}

	state.position--
	state.currentCharsTyped--
	state.backspaceCount++

	if !state.charCorrectness[state.position] {
		state.currentErrors--
	}
}

// handleKeystroke processes character input. Space enables word-skipping.
func handleKeystroke(state *TypingState, key byte) {
	char := string(key)
	state.currentCharsTyped++
	state.totalKeystrokes++

	if char == " " {
		handleSpace(state)
		return
	}

	if state.position >= len(state.sessionText) {
		return
	}

	if char == string(state.sessionText[state.position]) {
		state.charCorrectness[state.position] = true
	} else {
		state.charCorrectness[state.position] = false
		state.currentErrors++
		state.totalErrors++
	}
	state.position++
}

// handleSpace advances on correct space or skips to next word mid-word.
func handleSpace(state *TypingState) {
	if state.position < len(state.sessionText) && state.sessionText[state.position] == ' ' {
		state.charCorrectness[state.position] = true
		state.position++
	} else {
		if canSkipWord(state) {
			skipToNextWord(state)
		}
	}
}

// canSkipWord returns true if currently mid-word.
func canSkipWord(state *TypingState) bool {
	if state.position == 0 {
		return false
	}
	return state.sessionText[state.position-1] != ' '
}

// skipToNextWord marks remaining chars in word as incorrect and advances.
func skipToNextWord(state *TypingState) {
	start := state.position

	for i := state.position; i < len(state.sessionText); i++ {
		if state.sessionText[i] == ' ' {
			markRangeIncorrect(state, start, i)
			state.position = i + 1
			return
		}
	}

	markRangeIncorrect(state, start, len(state.sessionText))
	state.position = len(state.sessionText)
}

// markRangeIncorrect marks range [start, end) as incorrect and updates counters.
func markRangeIncorrect(state *TypingState, start, end int) {
	for j := start; j < end; j++ {
		state.charCorrectness[j] = false
		state.currentErrors++
		state.totalErrors++
	}
}
