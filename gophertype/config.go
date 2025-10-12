package main

import (
	_ "embed" // Used for embedding the word list file at compile time
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// Configuration defaults
const (
	defaultTextSource = TextSourceWords
	defaultWordCount  = 50
	configFileName    = "zootype.json"
)

// topWordsData is embedded at compile time via go:embed.
//
//go:embed data/top-1000-words.txt
var topWordsData string

// TextSource represents the type of text used for typing practice.
type TextSource string

const (
	TextSourceSentences TextSource = "sentences"
	TextSourceWords     TextSource = "words"
)

// Config holds runtime configuration with priority: CLI flags > config file > defaults.
type Config struct {
	TextSource TextSource `json:"text_source"`
	WordCount  int        `json:"word_count"`
}

// JSONConfig represents the zootype.json file structure.
type JSONConfig struct {
	Binary     string `json:"binary"`
	TextSource string `json:"text_source"`
	WordCount  int    `json:"word_count"`
}

// loadConfig loads configuration with priority: CLI flags > config file > defaults.
func loadConfig() (Config, error) {
	var textSource string
	var wordCount int

	flag.StringVar(&textSource, "source", "", "Text source: words or sentences")
	flag.StringVar(&textSource, "s", "", "Text source: words or sentences (shorthand)")
	flag.IntVar(&wordCount, "words", 0, "Number of words to practice")
	flag.IntVar(&wordCount, "w", 0, "Number of words to practice (shorthand)")
	flag.Parse()

	config := Config{
		TextSource: defaultTextSource,
		WordCount:  defaultWordCount,
	}

	// Override with config file if present
	configPath, err := findConfigFile()
	if err == nil {
		if fileConfig, err := loadJSONConfig(configPath); err == nil {
			if fileConfig.TextSource != "" {
				config.TextSource = TextSource(fileConfig.TextSource)
			}
			if fileConfig.WordCount > 0 {
				config.WordCount = fileConfig.WordCount
			}
		}
	}

	// CLI flags take highest priority
	if textSource != "" {
		config.TextSource = TextSource(textSource)
	}
	if wordCount > 0 {
		config.WordCount = wordCount
	}

	return config, nil
}

// findConfigFile searches for zootype.json relative to the executable
// (two dirs up) or by walking up from the current directory.
func findConfigFile() (string, error) {
	// Check project root (executable is at zootype/bin/gophertype)
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	projectRoot := filepath.Dir(filepath.Dir(exePath))
	configPath := filepath.Join(projectRoot, configFileName)

	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// Walk up from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for dir := cwd; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, configFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("config file not found")
}

// loadJSONConfig reads and parses zootype.json.
func loadJSONConfig(path string) (*JSONConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config JSONConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// getSessionText generates practice text based on configured source.
func getSessionText(config Config) (string, error) {
	switch config.TextSource {
	case TextSourceSentences:
		sentences := defaultSentences()
		return sentences[rand.Intn(len(sentences))], nil
	case TextSourceWords:
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
