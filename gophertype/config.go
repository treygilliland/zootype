package main

import (
	_ "embed" // Used for embedding the word list file at compile time
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// topWordsData is embedded at compile time via go:embed.
//
//go:embed data/top-1000-words.txt
var topWordsData string

// Configuration defaults
const (
	defaultTextSource  = TextSourceWords
	defaultWordCount   = 50
	defaultTimeSeconds = 30
)

// Command-line flags
var (
	timeSeconds      = flag.Int("time", 0, "Timed mode: type for N seconds (default: 30, takes precedence)")
	timeSecondsShort = flag.Int("t", 0, "Timed mode: type for N seconds (default: 30, takes precedence)")
	wordCount        = flag.Int("words", 0, "Word count mode: complete N words, untimed")
	wordCountShort   = flag.Int("w", 0, "Word count mode: complete N words, untimed")
	textSource       = flag.String("source", "", "Text source: words or sentences")
	textSourceShort  = flag.String("s", "", "Text source: words or sentences")
	showVersion      = flag.Bool("version", false, "Print version information")
	showVersionShort = flag.Bool("v", false, "Print version information")
)

// TextSource represents the type of text used for typing practice.
type TextSource string

const (
	TextSourceSentences TextSource = "sentences"
	TextSourceWords     TextSource = "words"
)

// Config holds runtime configuration from CLI flags and defaults.
type Config struct {
	TextSource  TextSource
	WordCount   int
	TimeSeconds int
}

// loadConfig loads configuration from CLI flags and defaults.
func loadConfig() (Config, error) {
	flag.Parse()

	if *showVersion || *showVersionShort {
		fmt.Printf("gophertype %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	config := Config{
		TextSource:  defaultTextSource,
		WordCount:   defaultWordCount,
		TimeSeconds: defaultTimeSeconds,
	}

	// Apply CLI flags
	if *textSource != "" {
		config.TextSource = TextSource(*textSource)
	} else if *textSourceShort != "" {
		config.TextSource = TextSource(*textSourceShort)
	}

	// Handle mutually exclusive modes: -t (timed) takes precedence over -w (word count)
	timeFlag := *timeSeconds
	if *timeSecondsShort > 0 {
		timeFlag = *timeSecondsShort
	}
	wordFlag := *wordCount
	if *wordCountShort > 0 {
		wordFlag = *wordCountShort
	}

	if timeFlag > 0 {
		// Timed mode explicitly requested - takes precedence
		config.TimeSeconds = timeFlag
	} else if wordFlag > 0 {
		// Word count mode - disable timer
		config.WordCount = wordFlag
		config.TimeSeconds = 0
	}
	// Otherwise use defaults (30 second timed mode)

	return config, nil
}

// getSessionText generates practice text based on configured source.
func getSessionText(config Config) (string, error) {
	switch config.TextSource {
	case TextSourceSentences:
		if config.TimeSeconds > 0 {
			return generateInfiniteSentences(), nil
		}
		sentences := defaultSentences()
		return sentences[rand.Intn(len(sentences))], nil
	case TextSourceWords:
		if config.TimeSeconds > 0 {
			return generateInfiniteWordText()
		}
		return generateWordText(config.WordCount)
	default:
		return "", fmt.Errorf("unknown text source: %s", config.TextSource)
	}
}

// generateWordText creates practice text by randomly selecting words.
// Words can repeat for more frequent practice of common words.
func generateWordText(count int) (string, error) {
	words, err := loadTopWords()
	if err != nil {
		return "", err
	}

	if len(words) == 0 {
		return "", fmt.Errorf("no words available")
	}

	selectedWords := make([]string, count)
	for i := 0; i < count; i++ {
		selectedWords[i] = words[rand.Intn(len(words))]
	}

	return strings.Join(selectedWords, " "), nil
}

// loadTopWords parses the embedded word list (one word per line).
func loadTopWords() ([]string, error) {
	lines := strings.Split(topWordsData, "\n")
	var words []string

	for _, line := range lines {
		word := strings.TrimSpace(line)
		if word != "" {
			words = append(words, word)
		}
	}

	if len(words) == 0 {
		return nil, fmt.Errorf("no words found in embedded data")
	}

	return words, nil
}

// defaultSentences returns pangrams for sentence-based practice.
func defaultSentences() []string {
	return []string{
		"The quick brown fox jumps over the lazy dog.",
		"Pack my box with five dozen liquor jugs.",
		"How vexingly quick daft zebras jump!",
		"Sphinx of black quartz, judge my vow.",
	}
}

// generateInfiniteWordText creates an infinitely long text by repeating words.
func generateInfiniteWordText() (string, error) {
	words, err := loadTopWords()
	if err != nil {
		return "", err
	}

	if len(words) == 0 {
		return "", fmt.Errorf("no words available")
	}

	// Generate a very long text by repeating words
	var selectedWords []string
	for i := 0; i < 1000; i++ { // Generate 1000 words initially
		selectedWords = append(selectedWords, words[rand.Intn(len(words))])
	}

	return strings.Join(selectedWords, " "), nil
}

// generateInfiniteSentences creates an infinitely long text by repeating sentences.
func generateInfiniteSentences() string {
	sentences := defaultSentences()
	var selectedSentences []string

	// Generate a very long text by repeating sentences
	for i := 0; i < 100; i++ { // Generate 100 sentences initially
		selectedSentences = append(selectedSentences, sentences[rand.Intn(len(sentences))])
	}

	return strings.Join(selectedSentences, " ")
}
