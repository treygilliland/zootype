package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

const (
	Reset  = "\033[0m"
	Green  = "\033[32m"
	Red    = "\033[31m"
	Blue   = "\033[34m"
	Gray   = "\033[90m"
	Yellow = "\033[33m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
)

type TypingState struct {
	target      string
	position    int
	errors      int
	totalErrors int
	typed       int
	rawTyped    int
	correct     []bool // Track which characters were typed correctly
	backspaces  int
	startTime   time.Time
}

func main() {
	// main wires the typing session together.
	target := defaultSentences()[0]
	state := newTypingState(target)

	printPrompt(state)

	restore, err := enableRawMode()
	if err != nil {
		panic(err)
	}
	defer restore()

	runTypingSession(state)
}

// defaultSentences returns static practice prompts.
func defaultSentences() []string {
	return []string{
		"The quick brown fox jumps over the lazy dog.",
		"Pack my box with five dozen liquor jugs.",
		"How vexingly quick daft zebras jump!",
		"Sphinx of black quartz, judge my vow.",
	}
}

// newTypingState primes all counters, including both corrected and raw accuracy.
func newTypingState(target string) *TypingState {
	return &TypingState{
		target:      target,
		position:    0,
		errors:      0,
		totalErrors: 0,
		typed:       0,
		rawTyped:    0,
		correct:     make([]bool, len(target)),
		backspaces:  0,
		startTime:   time.Now(),
	}
}

// printPrompt shows the sentence before we drop into raw mode.
func printPrompt(state *TypingState) {
	fmt.Printf("Hello from %sgophertype%s\n\n", Blue, Reset)
}

// enableRawMode switches stdin to raw; caller must defer the restore or the terminal stays weird.
func enableRawMode() (func(), error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	restore := func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}

	return restore, nil
}

// runTypingSession handles the keystroke loop and ANSI feedback.
func runTypingSession(state *TypingState) {
	displayProgress(state)

	for state.position < len(state.target) {
		buf := make([]byte, 1)
		if _, err := os.Stdin.Read(buf); err != nil {
			continue
		}

		if handleInterrupt(buf[0]) {
			fmt.Print("\n")
			printResults(state)
			return
		}

		if handleBackspace(state, buf[0]) {
			displayProgress(state)
			continue
		}

		applyKeystroke(state, buf[0])
		displayProgress(state)
	}

	printResults(state)
}

// handleInterrupt catches Ctrl+C so we can bail gracefully.
func handleInterrupt(b byte) bool {
	return b == 3
}

// handleBackspace rewinds the cursor and error counters when possible.
func handleBackspace(state *TypingState, b byte) bool {
	if b != 127 && b != 8 {
		return false
	}

	if state.position == 0 {
		return true
	}

	state.position--
	state.typed--
	state.backspaces++
	if !state.correct[state.position] {
		state.errors--
	}

	return true
}

// applyKeystroke mutates state for the newly typed character; note the dual accuracy tracking.
func applyKeystroke(state *TypingState, b byte) {
	char := string(b)
	state.typed++
	state.rawTyped++

	if char == " " {
		if state.position < len(state.target) && state.target[state.position] == ' ' {
			state.correct[state.position] = true
			state.position++
		} else {
			skipToNextWord(state)
		}
		return
	}

	if state.position >= len(state.target) {
		return
	}

	if char == string(state.target[state.position]) {
		state.correct[state.position] = true
	} else {
		state.correct[state.position] = false
		state.errors++
		state.totalErrors++
	}
	state.position++
}

// printResults dumps the session summary, including the raw vs corrected metrics.
func printResults(state *TypingState) {
	fmt.Print("\n\r\033[K")

	duration := time.Since(state.startTime)
	accuracy := 0.0
	if state.typed > 0 {
		accuracy = float64(state.typed-state.errors) / float64(state.typed) * 100
		if accuracy < 0 {
			accuracy = 0
		} else if accuracy > 100 {
			accuracy = 100
		}
	}
	rawAccuracy := 0.0
	if state.rawTyped > 0 {
		rawAccuracy = float64(state.rawTyped-state.totalErrors) / float64(state.rawTyped) * 100
		if rawAccuracy < 0 {
			rawAccuracy = 0
		} else if rawAccuracy > 100 {
			rawAccuracy = 100
		}
	}

	// Count correctly typed characters (excluding skipped words)
	correctChars := 0
	for i := 0; i < state.position; i++ {
		if state.correct[i] {
			correctChars++
		}
	}

	// Calculate WPM based on correctly typed characters only
	wpm := 0.0
	if correctChars > 0 && duration.Minutes() > 0 {
		wpm = float64(correctChars) / 5 / duration.Minutes()
	}

	fmt.Print("\r\n")
	fmt.Print("Results:\r\n")
	fmt.Printf("Accuracy: %.1f%%\r\n", accuracy)
	fmt.Printf("Raw Accuracy: %.1f%%\r\n", rawAccuracy)
	fmt.Printf("Duration: %.1fs\r\n", duration.Seconds())
	fmt.Printf("WPM: %.1f\r\n", wpm)
	fmt.Printf("Backspaces: %d\r\n", state.backspaces)
	fmt.Printf("Errors: %d\r\n", state.errors)
	fmt.Printf("Total Errors: %d\r\n", state.totalErrors)
}

// skipToNextWord flags the rest of the current word as wrong and jumps the cursor.
func skipToNextWord(state *TypingState) {
	// Mark skipped characters as incorrect
	start := state.position

	// Find next space starting from current position
	for i := state.position; i < len(state.target); i++ {
		if state.target[i] == ' ' {
			// Mark all skipped characters as incorrect
			for j := start; j < i; j++ {
				state.correct[j] = false
				state.errors++
				state.totalErrors++
			}
			state.position = i + 1
			return
		}
	}

	// If no space found, mark remaining characters as incorrect and go to end
	for j := start; j < len(state.target); j++ {
		state.correct[j] = false
		state.errors++
		state.totalErrors++
	}
	state.position = len(state.target)
}

// displayProgress paints the prompt with ANSI coloring; raw mode makes
// the cursor control a bit twitchy if stdout is redirected.
func displayProgress(state *TypingState) {
	// Move cursor to beginning of line and clear it
	fmt.Print("\r\033[K")

	result := ""

	// Show typed characters (correct in green, incorrect in red)
	for i := 0; i < state.position && i < len(state.target); i++ {
		if state.correct[i] {
			result += Green + string(state.target[i]) + Reset
		} else {
			result += Red + string(state.target[i]) + Reset
		}
	}

	// Show cursor at current position
	if state.position < len(state.target) {
		result += Yellow + Bold + "|" + Reset
		result += Gray + state.target[state.position:] + Reset
	} else {
		// At end of text
		result += Yellow + Bold + "|" + Reset
	}

	fmt.Print(result)
}
