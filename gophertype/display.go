package main

import (
	"fmt"
	"strings"
)

// ANSI escape codes for terminal colors and formatting
const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiBlue   = "\033[34m"
	ansiYellow = "\033[33m"
	ansiBold   = "\033[1m"
)

// wrappedLine represents a single display line with word-wrapping applied.
type wrappedLine struct {
	content        []rune // Display characters
	charIndices    []int  // Maps display position to original text index
	hasCursor      bool
	cursorPosition int
}

// displayProgress renders color-coded typing progress with word-wrapping
// in a 3-line scrolling window. Uses double-buffering to prevent flicker.
func displayProgress(state *TypingState) {
	clearPreviousDisplay(state.lastLineCount)

	lines := wrapTextToLines(state.sessionText, state.position, 80)
	startLine, endLine := calculateVisibleWindow(lines, 3)
	output := renderLines(lines[startLine:endLine], state)

	fmt.Print(output)
	state.lastLineCount = endLine - startLine
}

// clearPreviousDisplay moves cursor up and clears previous output.
func clearPreviousDisplay(lineCount int) {
	if lineCount > 0 {
		for i := 0; i < lineCount-1; i++ {
			fmt.Print("\033[A") // Move up one line
		}
		fmt.Print("\r\033[J") // Clear from cursor down
	}
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

		// Add newline between lines
		if lineIdx < len(lines)-1 {
			output.WriteString("\r\n")
		}
	}

	return output.String()
}
