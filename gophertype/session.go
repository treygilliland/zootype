package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Terminal control constants
const (
	keyCtrlC     = 3   // ASCII code for Ctrl+C (interrupt)
	keyDelete    = 127 // DEL key - commonly sent by backspace on Unix/Mac
	keyBackspace = 8   // BS key - sent by backspace on some terminals
	keyEscape    = 27  // ESC key
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
	lastLineCount     int           // Lines in previous display (for clearing)
	timeLimit         time.Duration // Time limit for timed mode (0 = no limit)
	isTimedMode       bool          // Whether this is a timed session
	displayMutex      sync.Mutex    // Synchronizes display updates
}

// newTypingState initializes a new typing session with the given target text.
func newTypingState(target string, config Config) *TypingState {
	isTimedMode := config.TimeSeconds > 0
	var timeLimit time.Duration
	if isTimedMode {
		timeLimit = time.Duration(config.TimeSeconds) * time.Second
	}

	return &TypingState{
		sessionText:     target,
		charCorrectness: make([]bool, len(target)),
		startTime:       time.Time{}, // Will be set when typing actually starts
		timeLimit:       timeLimit,
		isTimedMode:     isTimedMode,
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

// promptToContinue asks the user to press Enter to play again or q/Esc to exit.
// Returns true to play again, false to exit.
// Takes a channel that provides keyboard input.
func promptToContinue(keyChan <-chan byte) bool {
	fmt.Printf("\n%sPress Enter to play again or q/Esc to exit...%s", ansiBlue, ansiReset)

	for {
		key := <-keyChan

		if key == 13 || key == 10 { // Enter key
			return true
		}
		if key == 27 || key == 113 || key == 81 { // Esc, 'q', or 'Q'
			return false
		}
	}
}

// runTypingSession is the main event loop, reading keystrokes and updating display.
// Takes a channel that provides keyboard input.
func runTypingSession(state *TypingState, keyChan <-chan byte) {
	// Start the timer when typing actually begins
	state.startTime = time.Now()

	// Channel to signal when time is up
	timeUp := make(chan bool)

	// Start timer goroutine for timed mode
	if state.isTimedMode {
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond) // Update every 100ms
			defer ticker.Stop()

			for {
				<-ticker.C
				// Check if time is up
				if time.Since(state.startTime) >= state.timeLimit {
					timeUp <- true
					return
				}
				// Update display with mutex to prevent flicker
				state.displayMutex.Lock()
				displayProgress(state)
				state.displayMutex.Unlock()
			}
		}()
	}

	displayProgress(state)

	for {
		// Check if we've reached the end of text (only in non-timed mode)
		if !state.isTimedMode && state.position >= len(state.sessionText) {
			results := NewResults(state)
			results.Print()
			return
		}

		// In timed mode, if we've reached the end of text, extend it
		if state.isTimedMode && state.position >= len(state.sessionText) {
			extendTextForTimedMode(state)
		}

		// Wait for either keyboard input or timer expiration
		var key byte
		select {
		case <-timeUp:
			fmt.Print("\n")
			results := NewResults(state)
			results.Print()
			return
		case key = <-keyChan:
			// Process the key below
		}

		if isInterrupt(key) {
			fmt.Print("\n")
			results := NewResults(state)
			results.Print()
			return
		}

		// Ignore escape sequences (arrow keys, function keys, etc.)
		if key == keyEscape {
			drainEscapeSequence(keyChan)
			continue
		}

		if isBackspace(key) {
			handleBackspace(state)
			state.displayMutex.Lock()
			displayProgress(state)
			state.displayMutex.Unlock()
			continue
		}

		handleKeystroke(state, key)
		state.displayMutex.Lock()
		displayProgress(state)
		state.displayMutex.Unlock()
	}
}

func isInterrupt(key byte) bool {
	return key == keyCtrlC // Only Ctrl+C during typing
}

// drainEscapeSequence consumes remaining bytes from an escape sequence.
// Most escape sequences are 2-3 bytes (ESC [ X), but some can be longer.
func drainEscapeSequence(keyChan <-chan byte) {
	// Use a timeout to avoid blocking indefinitely
	timeout := time.After(10 * time.Millisecond)

	// Consume up to 10 bytes or until timeout
	for i := 0; i < 10; i++ {
		select {
		case <-keyChan:
		case <-timeout:
			return
		}
	}
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

	// Reset the correctness flag so it's recalculated on next keystroke
	state.charCorrectness[state.position] = false
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

// extendTextForTimedMode adds more text to the session when in timed mode.
func extendTextForTimedMode(state *TypingState) {
	// Generate more words to extend the text
	words, err := loadTopWords()
	if err != nil {
		return // If we can't load words, just return
	}

	// Add 100 more words
	var newWords []string
	for i := 0; i < 100; i++ {
		newWords = append(newWords, words[rand.Intn(len(words))])
	}

	newText := " " + strings.Join(newWords, " ")
	state.sessionText += newText

	// Extend the charCorrectness slice
	oldLen := len(state.charCorrectness)
	newLen := oldLen + len(newText)
	newCorrectness := make([]bool, newLen)
	copy(newCorrectness, state.charCorrectness)
	state.charCorrectness = newCorrectness
}
