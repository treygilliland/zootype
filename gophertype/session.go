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

const (
	keyCtrlC     = 3
	keyBackspace = 8  // BS - sent by backspace on some terminals
	keyEnter     = 10 // LF - Unix/Mac enter
	keyReturn    = 13 // CR - Windows enter
	keyEscape    = 27
	keyUpperN    = 78
	keyUpperQ    = 81
	keyUpperR    = 82
	keyLowerN    = 110
	keyLowerQ    = 113
	keyLowerR    = 114
	keyDelete    = 127 // DEL - commonly sent by backspace on Unix/Mac

	textExtensionWords = 100
	escapeSeqTimeout   = 10 * time.Millisecond
	inputDrainTimeout  = 500 * time.Millisecond
)

// TypingState tracks both corrected metrics (accounting for backspaces)
// and raw metrics (all keystrokes) for calculating accuracy statistics.
type TypingState struct {
	sessionText     string
	position        int    // Current cursor position in sessionText
	errors          int    // Decremented on backspace correction
	rawErrors       int    // Never decremented (for raw accuracy)
	charsTyped      int    // Decremented on backspace
	rawKeystrokes   int    // Never decremented
	charCorrectness []bool // Per-character correctness for coloring
	charTyped       []bool // Tracks which chars were actually typed (not skipped)
	backspaceCount  int
	startTime       time.Time
	lastLineCount   int           // Lines in previous display (for clearing)
	timeLimit       time.Duration // 0 = no limit
	isTimedMode     bool
	displayMutex    sync.Mutex // Synchronizes display updates
	terminalWidth   int
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

// setupTerminal enables alternate screen buffer and raw mode for character-by-character input.
// Returns a restore function that must be deferred.
func setupTerminal() (func(), error) {
	fmt.Print(ansiAltScreenEnable + ansiCursorHide)

	stdinFd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(stdinFd)
	if err != nil {
		fmt.Print(ansiCursorShow + ansiAltScreenDisable)
		return nil, err
	}

	restore := func() {
		_ = term.Restore(stdinFd, oldState)
		fmt.Print(ansiCursorShow + ansiAltScreenDisable)
		fmt.Print("\r\n")
	}

	return restore, nil
}

// getTerminalWidth returns the current terminal width.
func getTerminalWidth() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}
	return width, nil
}

// getAndValidateTerminalWidth validates terminal width (min 25, capped at 80).
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

// startKeyboardReader spawns a goroutine that reads keyboard input for the duration of the program.
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

const (
	ActionExit      = 0
	ActionRetry     = 1
	ActionNext      = 2
	ActionInterrupt = 3
)

// promptToContinue asks the user what to do next: (n)ext, (r)etry, or (q)uit.
func promptToContinue(keyChan <-chan byte) int {
	// drain any buffered keypresses to prevent accidental triggering post-test.
	drainChannel(keyChan, inputDrainTimeout)

	fmt.Printf("\n%s(n)ext, (r)etry, (q)uit%s", ansiBlue, ansiReset)

	for {
		key := <-keyChan

		switch key {
		case keyCtrlC:
			return ActionExit
		case keyEnter, keyReturn, keyLowerN, keyUpperN:
			return ActionNext
		case keyLowerR, keyUpperR:
			return ActionRetry
		case keyLowerQ, keyUpperQ:
			return ActionExit
		}
	}
}

// runTypingSession is the main event loop for a typing session.
func runTypingSession(state *TypingState, keyChan <-chan byte) int {
	fmt.Print(ansiClearScreen + ansiCursorHome)
	fmt.Printf("%sgophertype%s\n\n", ansiBlue, ansiReset)

	state.startTime = time.Now()
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
					timeUp <- true
					return
				case <-displayTicker.C:
					state.displayMutex.Lock()
					displayProgress(state)
					state.displayMutex.Unlock()
				}
			}
		}()
	}

	displayProgress(state)

	for {
		// Check if we've reached the end of text (only matters in non-timed mode)
		if !state.isTimedMode && state.position >= len(state.sessionText) {
			fmt.Print("\r\n\r\n")
			results := NewResults(state)
			results.Print()
			return ActionNext
		}

		// In timed mode, extend text if user reaches the end
		if state.isTimedMode && state.position >= len(state.sessionText) {
			extendTextForTimedMode(state)
		}

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
	return key == keyCtrlC
}

// drainEscapeSequence consumes escape sequence bytes (arrow keys, function keys, etc).
// Most sequences are 2-3 bytes (ESC [ X), but some can be longer.
func drainEscapeSequence(keyChan <-chan byte) {
	const maxEscapeLen = 10
	timeout := time.After(escapeSeqTimeout)

	for i := 0; i < maxEscapeLen; i++ {
		select {
		case <-keyChan:
		case <-timeout:
			return
		}
	}
}

// drainChannel consumes buffered keypresses to prevent accidental input.
func drainChannel(keyChan <-chan byte, timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case <-keyChan:
		case <-deadline:
			return
		}
	}
}

func isBackspace(key byte) bool {
	return key == keyDelete || key == keyBackspace
}

// handleBackspace moves cursor back and updates metrics.
// Only decrements errors if the character was actually typed incorrectly (not just skipped).
func handleBackspace(state *TypingState) {
	if state.position == 0 {
		return
	}

	state.position--
	state.charsTyped--
	state.backspaceCount++

	if !state.charCorrectness[state.position] && state.charTyped[state.position] {
		state.errors--
	}

	state.charCorrectness[state.position] = false
	state.charTyped[state.position] = false
}

// handleKeystroke processes character input and updates correctness tracking.
func handleKeystroke(state *TypingState, key byte) {
	char := string(key)
	state.charsTyped++
	state.rawKeystrokes++

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
		state.errors++
		state.rawErrors++
	}
	state.position++
}

// handleSpace advances on correct space or skips to next word if mid-word.
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

func canSkipWord(state *TypingState) bool {
	if state.position == 0 {
		return false
	}
	return state.sessionText[state.position-1] != ' '
}

// skipToNextWord marks remaining characters in the current word as incorrect.
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

// markRangeIncorrect marks a range of characters as incorrect (only affects raw accuracy).
func markRangeIncorrect(state *TypingState, start, end int) {
	for j := start; j < end; j++ {
		state.charCorrectness[j] = false
		state.rawErrors++
	}
}

// extendTextForTimedMode appends more words when the user reaches the end in timed mode.
func extendTextForTimedMode(state *TypingState) {
	words, err := loadTopWords()
	if err != nil {
		return
	}

	var newWords []string
	for i := 0; i < textExtensionWords; i++ {
		newWords = append(newWords, words[rand.Intn(len(words))])
	}

	newText := " " + strings.Join(newWords, " ")
	state.sessionText += newText

	oldLen := len(state.charCorrectness)
	newLen := oldLen + len(newText)

	newCorrectness := make([]bool, newLen)
	copy(newCorrectness, state.charCorrectness)
	state.charCorrectness = newCorrectness

	newTyped := make([]bool, newLen)
	copy(newTyped, state.charTyped)
	state.charTyped = newTyped
}
