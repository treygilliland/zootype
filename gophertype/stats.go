package main

import (
	"fmt"
	"time"
)

const charsPerWord = 5.0

type Results struct {
	Duration       time.Duration
	Accuracy       float64
	RawAccuracy    float64
	WPM            float64
	CorrectChars   int
	BackspaceCount int
	Errors         int
	RawErrors      int
}

// NewResults calculates final statistics from a completed typing session.
func NewResults(state *TypingState) *Results {
	duration := time.Since(state.startTime)
	correctChars := countCorrectChars(state)

	return &Results{
		Duration:       duration,
		Accuracy:       calculateAccuracy(state.charsTyped, state.errors),
		RawAccuracy:    calculateAccuracy(state.rawKeystrokes, state.rawErrors),
		WPM:            calculateWPM(correctChars, duration),
		CorrectChars:   correctChars,
		BackspaceCount: state.backspaceCount,
		Errors:         state.errors,
		RawErrors:      state.rawErrors,
	}
}

func (r *Results) Print() {
	fmt.Printf("%sResults:%s\r\n", ansiBlue, ansiReset)
	fmt.Printf("WPM:          %.1f\r\n", r.WPM)
	fmt.Printf("Duration:     %ds\r\n", int(r.Duration.Seconds()))
	fmt.Printf("Accuracy:     %.1f%%\r\n", r.Accuracy)
	fmt.Printf("Errors:       %d\r\n", r.Errors)
	fmt.Printf("Raw Accuracy: %.1f%%\r\n", r.RawAccuracy)
	fmt.Printf("Raw Errors:   %d\r\n", r.RawErrors)
	fmt.Printf("Backspaces:   %d\r\n", r.BackspaceCount)
}

func calculateAccuracy(typed, errors int) float64 {
	if typed == 0 {
		return 0.0
	}
	accuracy := float64(typed-errors) / float64(typed) * 100
	return clamp(accuracy, 0, 100)
}

func calculateWPM(correctChars int, duration time.Duration) float64 {
	if correctChars == 0 || duration.Minutes() == 0 {
		return 0.0
	}
	return float64(correctChars) / charsPerWord / duration.Minutes()
}

func countCorrectChars(state *TypingState) int {
	count := 0
	for i := 0; i < state.position; i++ {
		if state.charCorrectness[i] {
			count++
		}
	}
	return count
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
