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
	keyBackspace = 8   // BS key - sent by backspace on some terminals
	keyEnter     = 10  // Line feed (Unix/Mac Enter key)
	keyReturn    = 13  // Carriage return (Windows Enter key)
	keyEscape    = 27  // ESC key
	keyUpperN    = 78  // 'N' key
	keyUpperQ    = 81  // 'Q' key
	keyUpperR    = 82  // 'R' key
	keyLowerN    = 110 // 'n' key
	keyLowerQ    = 113 // 'q' key
	keyLowerR    = 114 // 'r' key
	keyDelete    = 127 // DEL key - commonly sent by backspace on Unix/Mac
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
	charTyped         []bool // Tracks which characters were actually typed (not skipped)
	backspaceCount    int
	startTime         time.Time
	lastLineCount     int           // Lines in previous display (for clearing)
	timeLimit         time.Duration // Time limit for timed mode (0 = no limit)
	isTimedMode       bool          // Whether this is a timed session
	displayMutex      sync.Mutex    // Synchronizes display updates
	terminalWidth     int           // Terminal width for text wrapping
}

// newTypingState initializes a new typing session with the given target text.
func newTypingState(target string, config Config, termWidth int) *TypingState {
	isTimedMode := config.TimeSeconds > 0
	var timeLimit time.Duration
	if isTimedMode {
		timeLimit = time.Duration(config.TimeSeconds) * time.Second
	}

	return &TypingState{
		sessionText:     target,
		charCorrectness: make([]bool, len(target)),
		charTyped:       make([]bool, len(target)),
		startTime:       time.Time{},
		timeLimit:       timeLimit,
		isTimedMode:     isTimedMode,
		terminalWidth:   termWidth,
	}
}

// setupTerminal enables alternate screen buffer and raw terminal mode.
// Returns a restore function that must be deferred to restore normal terminal state.
func setupTerminal() (func(), error) {
	// Enable alternate screen buffer (doesn't affect scrollback)
	fmt.Print(ansiAltScreenEnable)

	// Enable raw mode for character-by-character input
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Print(ansiAltScreenDisable) // Restore screen on error
		return nil, err
	}

	restore := func() {
		_ = term.Restore(fd, oldState)
		fmt.Print(ansiAltScreenDisable) // Restore original screen
		fmt.Print("\r\n")
	}

	return restore, nil
}

// getTerminalWidth returns the terminal width, or error if it can't be determined.
func getTerminalWidth() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}
	return width, nil
}

// getAndValidateTerminalWidth detects and validates terminal width for display.
// Returns the width to use (capped at 80, minimum 25), or error if terminal too narrow.
func getAndValidateTerminalWidth() (int, error) {
	const (
		minWidth = 25
		maxWidth = 80
	)

	termWidth, err := getTerminalWidth()
	if err != nil {
		return 0, fmt.Errorf("failed to get terminal size: %w", err)
	}

	if termWidth < minWidth {
		return 0, fmt.Errorf("terminal too narrow: %d chars (minimum %d chars required)", termWidth, minWidth)
	}

	if termWidth > maxWidth {
		termWidth = maxWidth
	}

	return termWidth, nil
}

// startKeyboardReader starts a goroutine that reads keyboard input byte-by-byte
// and sends it to the returned channel. This goroutine runs for the program's lifetime.
func startKeyboardReader() <-chan byte {
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

	return keyChan
}

// Action constants for session control
const (
	ActionExit      = 0
	ActionRetry     = 1
	ActionNext      = 2
	ActionInterrupt = 3 // Ctrl+C pressed during typing
)

// promptToContinue asks the user what to do next.
// Returns ActionNext (n/Enter), ActionRetry (r), or ActionExit (q).
// Takes a channel that provides keyboard input.
func promptToContinue(keyChan <-chan byte) int {
	// Drain any buffered keypresses to prevent accidental triggering
	drainChannel(keyChan, 500*time.Millisecond)

	fmt.Printf("\n%s(n)ext, (r)etry, (q)uit%s", ansiBlue, ansiReset)

	for {
		key := <-keyChan

		switch key {
		case keyEnter, keyReturn, keyLowerN, keyUpperN:
			return ActionNext
		case keyLowerR, keyUpperR:
			return ActionRetry
		case keyLowerQ, keyUpperQ:
			return ActionExit
		}
		// Ignore other keys (including Esc)
	}
}

// runTypingSession is the main event loop, reading keystrokes and updating display.
// Takes a channel that provides keyboard input.
// Returns an action to take based on user input.
func runTypingSession(state *TypingState, keyChan <-chan byte) int {
	// Clear screen and display header
	fmt.Print(ansiClearScreen + ansiCursorHome)
	fmt.Printf("%sgophertype%s\n\n", ansiBlue, ansiReset)

	// Start the timer when typing actually begins
	state.startTime = time.Now()

	// Channel to signal when time is up
	timeUp := make(chan bool)

	// Start timer goroutine for timed mode
	if state.isTimedMode {
		go func() {
			displayTicker := time.NewTicker(1 * time.Second) // Update display every second
			defer displayTicker.Stop()

			timeUpTimer := time.After(state.timeLimit)

			for {
				select {
				case <-timeUpTimer:
					// Time is up (fires once at exact duration)
					timeUp <- true
					return
				case <-displayTicker.C:
					// Update display every second (when countdown value changes)
					state.displayMutex.Lock()
					displayProgress(state)
					state.displayMutex.Unlock()
				}
			}
		}()
	}

	displayProgress(state)

	for {
		// Check if we've reached the end of text (only in non-timed mode)
		if !state.isTimedMode && state.position >= len(state.sessionText) {
			fmt.Print("\r\n\r\n")
			results := NewResults(state)
			results.Print()
			return ActionNext
		}

		// In timed mode, if we've reached the end of text, extend it
		if state.isTimedMode && state.position >= len(state.sessionText) {
			extendTextForTimedMode(state)
		}

		// Wait for either keyboard input or timer expiration
		var key byte
		select {
		case <-timeUp:
			fmt.Print("\r\n\r\n")
			results := NewResults(state)
			results.Print()
			return ActionNext
		case key = <-keyChan:
			// Process the key below
		}

		if isInterrupt(key) {
			fmt.Print("\r\n\r\n")
			results := NewResults(state)
			results.Print()
			return ActionInterrupt
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
	const (
		escapeTimeout = 10 * time.Millisecond
		maxEscapeLen  = 10
	)

	timeout := time.After(escapeTimeout)

	// Consume up to maxEscapeLen bytes or until timeout
	for i := 0; i < maxEscapeLen; i++ {
		select {
		case <-keyChan:
			// Consumed one byte of the sequence
		case <-timeout:
			// No more bytes available, sequence is complete
			return
		}
	}
}

// drainChannel consumes all buffered input from the channel for a given timeout.
// Used to clear accidental keypresses before prompting user.
func drainChannel(keyChan <-chan byte, timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case <-keyChan:
			// swallow the keypress
		case <-deadline:
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

	// Only decrement currentErrors if this character was actually typed incorrectly
	// (not just marked incorrect from skipping)
	if !state.charCorrectness[state.position] && state.charTyped[state.position] {
		state.currentErrors--
	}

	// Reset flags so they're recalculated on next keystroke
	state.charCorrectness[state.position] = false
	state.charTyped[state.position] = false
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

	state.charTyped[state.position] = true
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
		state.charTyped[state.position] = true
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
// Only increments totalErrors (raw accuracy), not currentErrors (correctable errors).
func markRangeIncorrect(state *TypingState, start, end int) {
	for j := start; j < end; j++ {
		state.charCorrectness[j] = false
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

	// Extend the charCorrectness and charTyped slices
	oldLen := len(state.charCorrectness)
	newLen := oldLen + len(newText)

	newCorrectness := make([]bool, newLen)
	copy(newCorrectness, state.charCorrectness)
	state.charCorrectness = newCorrectness

	newTyped := make([]bool, newLen)
	copy(newTyped, state.charTyped)
	state.charTyped = newTyped
}
