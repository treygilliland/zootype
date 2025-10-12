package main

import (
	"fmt"
	"time"
)

// Standard WPM calculation: 5 characters = 1 word
const charsPerWord = 5.0

// Results contains the final statistics from a typing session.
type Results struct {
	Duration       time.Duration
	Accuracy       float64 // Corrected accuracy (accounts for backspaces)
	RawAccuracy    float64 // Raw accuracy (all keystrokes, no corrections)
	WPM            float64 // Words per minute (5 chars = 1 word)
	CorrectChars   int
	BackspaceCount int
	CurrentErrors  int
	TotalErrors    int
}

// NewResults calculates final statistics from a completed typing session.
func NewResults(state *TypingState) *Results {
	duration := time.Since(state.startTime)
	correctChars := countCorrectChars(state)

	return &Results{
		Duration:       duration,
		Accuracy:       calculateAccuracy(state.currentCharsTyped, state.currentErrors),
		RawAccuracy:    calculateAccuracy(state.totalKeystrokes, state.totalErrors),
		WPM:            calculateWPM(correctChars, duration),
		CorrectChars:   correctChars,
		BackspaceCount: state.backspaceCount,
		CurrentErrors:  state.currentErrors,
		TotalErrors:    state.totalErrors,
	}
}

// Print displays results. Uses \r\n for proper line breaks in raw mode.
func (r *Results) Print() {
	fmt.Printf("\r\n%sResults:%s\r\n", ansiBlue, ansiReset)
	fmt.Printf("WPM:          %.1f\r\n", r.WPM)
	fmt.Printf("Duration:     %ds\r\n", int(r.Duration.Seconds()))
	fmt.Printf("Accuracy:     %.1f%%\r\n", r.Accuracy)
	fmt.Printf("Errors:       %d\r\n", r.CurrentErrors)
	fmt.Printf("Raw Accuracy: %.1f%%\r\n", r.RawAccuracy)
	fmt.Printf("Raw Errors:   %d\r\n", r.TotalErrors)
	fmt.Printf("Backspaces:   %d\r\n", r.BackspaceCount)
}

// calculateAccuracy computes accuracy as a percentage of correct keystrokes.
func calculateAccuracy(typed, errors int) float64 {
	if typed == 0 {
		return 0.0
	}
	accuracy := float64(typed-errors) / float64(typed) * 100
	return clamp(accuracy, 0, 100)
}

// calculateWPM computes words per minute (1 word = 5 chars, correct chars only).
func calculateWPM(correctChars int, duration time.Duration) float64 {
	if correctChars == 0 || duration.Minutes() == 0 {
		return 0.0
	}
	return float64(correctChars) / charsPerWord / duration.Minutes()
}

// countCorrectChars returns the number of correctly typed characters.
func countCorrectChars(state *TypingState) int {
	count := 0
	for i := 0; i < state.position; i++ {
		if state.charCorrectness[i] {
			count++
		}
	}
	return count
}

// clamp restricts a value to [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
