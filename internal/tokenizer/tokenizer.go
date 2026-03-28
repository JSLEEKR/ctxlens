// Package tokenizer provides token counting for text content.
package tokenizer

import (
	"strings"
	"unicode/utf8"
)

// Tokenizer counts tokens in text.
type Tokenizer struct {
	// isEstimate indicates the tokenizer is using character-based estimation.
	isEstimate bool
}

// New creates a new Tokenizer.
// It uses a 4-chars-per-token estimator (labeled as estimate).
func New() *Tokenizer {
	return &Tokenizer{
		isEstimate: true,
	}
}

// IsEstimate returns true if the tokenizer uses estimation rather than exact counting.
func (t *Tokenizer) IsEstimate() bool {
	return t.isEstimate
}

// Count returns the number of tokens in the given text.
func (t *Tokenizer) Count(text string) int {
	if text == "" {
		return 0
	}
	return estimateTokens(text)
}

// CountAll returns the total tokens across multiple text strings.
func (t *Tokenizer) CountAll(texts ...string) int {
	total := 0
	for _, text := range texts {
		total += t.Count(text)
	}
	return total
}

// estimateTokens estimates tokens at approximately 4 characters per token.
// This is a widely-used heuristic for English text with cl100k_base encoding.
// For code and mixed content, it tends to slightly overcount which is conservative.
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}

	// Count characters (not bytes) for proper Unicode handling
	charCount := utf8.RuneCountInString(text)

	// Base estimate: ~4 chars per token
	tokens := charCount / 4

	// Adjust for whitespace-heavy content (whitespace is often its own token)
	words := len(strings.Fields(text))
	if words > 0 {
		// Average: max(char-based, word-based) to avoid undercount
		wordBased := words + (words / 4) // words + some overhead for punctuation
		if wordBased > tokens {
			tokens = wordBased
		}
	}

	// Minimum 1 token for non-empty text
	if tokens == 0 {
		tokens = 1
	}

	return tokens
}
