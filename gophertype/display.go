package main

import (
	"fmt"
	"strings"
	"time"
)

// ANSI escape codes for terminal colors and formatting
const (
	ansiReset       = "\033[0m"
	ansiGreen       = "\033[32m"
	ansiRed         = "\033[31m"
	ansiBlue        = "\033[34m"
	ansiYellow      = "\033[33m"
	ansiBold        = "\033[1m"
	ansiClearToEOL  = "\033[K"  // Clear from cursor to end of line
	ansiClearScreen = "\033[2J" // Clear entire screen
	ansiCursorUp    = "\033[A"  // Move cursor up one line
	ansiCursorHome  = "\033[H"  // Move cursor to home position (0,0)
)

// Display configuration constants
const (
	maxVisibleLines = 3 // Number of text lines shown in scrolling window
	cursorWidth     = 1 // Characters reserved for cursor display
)

// wrappedLine represents a single display line with word-wrapping applied.
type wrappedLine struct {
	content        []rune // Display characters
	charIndices    []int  // Maps display position to original text index
	hasCursor      bool
	cursorPosition int
}

// displayProgress renders color-coded typing progress with word-wrapping
// in a 3-line scrolling window. Buffers all output to prevent flicker.
func displayProgress(state *TypingState) {
	var buffer strings.Builder

	// Build cursor positioning sequence
	buffer.WriteString(buildClearSequence(state.lastLineCount))

	// Add timer if in timed mode
	if state.isTimedMode {
		buffer.WriteString(formatTimer(state))
	}

	// Render text lines (reserve space for cursor)
	lineWidth := state.terminalWidth - cursorWidth
	lines := wrapTextToLines(state.sessionText, state.position, lineWidth)
	startLine, endLine := calculateVisibleWindow(lines, maxVisibleLines)
	buffer.WriteString(renderLines(lines[startLine:endLine], state))

	// Single atomic write to terminal
	fmt.Print(buffer.String())

	// Update line count for next render
	state.lastLineCount = endLine - startLine
	if state.isTimedMode {
		state.lastLineCount++ // Account for timer line
	}
}

// buildClearSequence generates cursor positioning commands to prepare for redraw.
func buildClearSequence(lineCount int) string {
	if lineCount == 0 {
		return "\r"
	}

	var output strings.Builder
	// Move cursor to start of line first
	output.WriteString("\r")
	// Move up lineCount-1 lines (we're already on the last line)
	for i := 0; i < lineCount-1; i++ {
		output.WriteString(ansiCursorUp)
	}
	return output.String()
}

// formatTimer returns the countdown timer string for timed mode.
func formatTimer(state *TypingState) string {
	elapsed := time.Since(state.startTime)
	remaining := state.timeLimit - elapsed

	if remaining <= 0 {
		remaining = 0
	}

	// Round up to the next second
	seconds := int(remaining.Seconds())
	if remaining.Milliseconds()%1000 > 0 {
		seconds++
	}

	return fmt.Sprintf("%s%d%s%s\r\n", ansiBlue, seconds, ansiReset, ansiClearToEOL)
}

// wrapTextToLines splits text into display lines with word-boundary wrapping.
// Tracks cursor position and maps display positions back to original text indices.
func wrapTextToLines(text string, cursorPos, lineWidth int) []wrappedLine {
	words := splitIntoWords(text)
	var lines []wrappedLine
	currentLine := wrappedLine{}
	textIndex := 0

	for _, word := range words {
		// Wrap to new line if word doesn't fit
		if len(currentLine.content) > 0 && len(currentLine.content)+len(word) > lineWidth {
			lines = append(lines, currentLine)
			currentLine = wrappedLine{}

			// Skip leading spaces on new lines
			if word == " " {
				textIndex += len(word)
				continue
			}
		}

		// Add word character by character
		for _, char := range word {
			if textIndex == cursorPos {
				currentLine.hasCursor = true
				currentLine.cursorPosition = len(currentLine.content)
			}

			currentLine.content = append(currentLine.content, char)
			currentLine.charIndices = append(currentLine.charIndices, textIndex)
			textIndex++
		}
	}

	// Handle cursor at end of text
	if cursorPos >= len(text) {
		currentLine.hasCursor = true
		currentLine.cursorPosition = len(currentLine.content)
	}

	if len(currentLine.content) > 0 || currentLine.hasCursor {
		lines = append(lines, currentLine)
	}

	return lines
}

// splitIntoWords breaks text into words and spaces as separate tokens.
// This enables word-boundary wrapping (words don't split across lines).
func splitIntoWords(text string) []string {
	var words []string
	currentWord := ""

	for _, char := range text {
		if char == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
			words = append(words, " ")
		} else {
			currentWord += string(char)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// calculateVisibleWindow determines which lines to display in a scrolling window.
// Centers the cursor line when possible, adjusting at text boundaries.
func calculateVisibleWindow(lines []wrappedLine, maxLines int) (start, end int) {
	cursorLine := 0
	for i, line := range lines {
		if line.hasCursor {
			cursorLine = i
			break
		}
	}

	// Try to center cursor with one line of context above
	start = cursorLine - 1
	if start < 0 {
		start = 0
	}

	end = start + maxLines
	if end > len(lines) {
		end = len(lines)
		start = end - maxLines
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

// renderLines generates the color-coded display string for the given lines.
// Characters are colored based on correctness: green (correct), red (incorrect), default (untyped).
func renderLines(lines []wrappedLine, state *TypingState) string {
	var output strings.Builder

	for lineIdx, line := range lines {
		// Render each character with appropriate color
		for pos, char := range line.content {
			// Show cursor before this character if applicable
			if line.hasCursor && pos == line.cursorPosition {
				output.WriteString(ansiYellow + ansiBold + "|" + ansiReset)
			}

			origIdx := line.charIndices[pos]

			if origIdx < state.position {
				// Character has been typed - color by correctness
				if state.charCorrectness[origIdx] {
					output.WriteString(ansiGreen)
				} else {
					output.WriteString(ansiRed)
				}
				output.WriteRune(char)
				output.WriteString(ansiReset)
			} else {
				// Not yet typed - default color
				output.WriteRune(char)
			}
		}

		// Show cursor at end of line if applicable
		if line.hasCursor && line.cursorPosition >= len(line.content) {
			output.WriteString(ansiYellow + ansiBold + "|" + ansiReset)
		}

		// Clear to end of line to remove any leftover characters
		output.WriteString(ansiClearToEOL)

		// Add newline between lines
		if lineIdx < len(lines)-1 {
			output.WriteString("\r\n")
		}
	}

	return output.String()
}
