package main

import (
	_ "embed"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// topWordsData is embedded at compile time via go:embed so this must not be moved.
//
//go:embed data/top-1000-words.txt
var topWordsData string

const (
	defaultTextSource     = TextSourceWords
	defaultWordCount      = 50
	defaultTimeSeconds    = 30
	initialTimedWords     = 1000
	initialTimedSentences = 100
)

var (
	timeSeconds int
	wordCount   int
	textSource  string
	showVersion bool
)

func init() {
	flag.IntVar(&timeSeconds, "time", 0, "Timed mode: type for N seconds (default: 30)")
	flag.IntVar(&timeSeconds, "t", 0, "Timed mode: type for N seconds (default: 30)")
	flag.IntVar(&wordCount, "words", 0, "Word count mode: complete N words, untimed")
	flag.IntVar(&wordCount, "w", 0, "Word count mode: complete N words, untimed")
	flag.StringVar(&textSource, "source", "", "Text source: words or sentences")
	flag.StringVar(&textSource, "s", "", "Text source: words or sentences")
	flag.BoolVar(&showVersion, "version", false, "Print version information")
	flag.BoolVar(&showVersion, "v", false, "Print version information")
}

type TextSource string

const (
	TextSourceSentences TextSource = "sentences"
	TextSourceWords     TextSource = "words"
)

type Config struct {
	TextSource  TextSource
	WordCount   int
	TimeSeconds int
}

func loadConfig() (Config, error) {
	flag.Parse()

	if showVersion {
		fmt.Printf("gophertype %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	config := Config{
		TextSource:  defaultTextSource,
		WordCount:   defaultWordCount,
		TimeSeconds: defaultTimeSeconds,
	}

	if textSource != "" {
		config.TextSource = TextSource(textSource)
	}

	if timeSeconds > 0 {
		config.TimeSeconds = timeSeconds
	} else if wordCount > 0 {
		config.WordCount = wordCount
		config.TimeSeconds = 0
	}

	return config, nil
}

// getSessionText generates practice text based on the configured source and mode.
func getSessionText(config Config) (string, error) {
	switch config.TextSource {
	case TextSourceSentences:
		if config.TimeSeconds > 0 {
			return generateInitialSentenceText(), nil
		}
		sentences := defaultSentences()
		return sentences[rand.Intn(len(sentences))], nil
	case TextSourceWords:
		if config.TimeSeconds > 0 {
			return generateInitialWordText()
		}
		return generateWordText(config.WordCount)
	default:
		return "", fmt.Errorf("unknown text source: %s", config.TextSource)
	}
}

// generateWordText creates practice text by randomly selecting N words.
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

// loadTopWords parses the embedded word list.
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

func defaultSentences() []string {
	return []string{
		"The quick brown fox jumps over the lazy dog.",
		"Pack my box with five dozen liquor jugs.",
		"How vexingly quick daft zebras jump!",
		"Sphinx of black quartz, judge my vow.",
	}
}

// generateInitialWordText generates a large initial buffer for timed mode.
func generateInitialWordText() (string, error) {
	words, err := loadTopWords()
	if err != nil {
		return "", err
	}

	if len(words) == 0 {
		return "", fmt.Errorf("no words available")
	}

	var selectedWords []string
	for i := 0; i < initialTimedWords; i++ {
		selectedWords = append(selectedWords, words[rand.Intn(len(words))])
	}

	return strings.Join(selectedWords, " "), nil
}

// generateInitialSentenceText generates a large initial buffer of sentences for timed mode.
func generateInitialSentenceText() string {
	sentences := defaultSentences()
	var selectedSentences []string

	for i := 0; i < initialTimedSentences; i++ {
		selectedSentences = append(selectedSentences, sentences[rand.Intn(len(sentences))])
	}

	return strings.Join(selectedSentences, " ")
}
