package tokenizer

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tok := New()
	if tok == nil {
		t.Fatal("expected non-nil tokenizer")
	}
	if !tok.IsEstimate() {
		t.Error("expected IsEstimate to be true")
	}
}

func TestCount_Empty(t *testing.T) {
	tok := New()
	count := tok.Count("")
	if count != 0 {
		t.Errorf("expected 0 for empty string, got %d", count)
	}
}

func TestCount_SingleWord(t *testing.T) {
	tok := New()
	count := tok.Count("hello")
	if count < 1 {
		t.Errorf("expected at least 1 token for 'hello', got %d", count)
	}
}

func TestCount_Sentence(t *testing.T) {
	tok := New()
	text := "The quick brown fox jumps over the lazy dog."
	count := tok.Count(text)
	// ~45 chars, ~10 words -> expect 10-15 tokens
	if count < 5 {
		t.Errorf("expected at least 5 tokens for sentence, got %d", count)
	}
	if count > 50 {
		t.Errorf("expected at most 50 tokens for sentence, got %d", count)
	}
}

func TestCount_LongText(t *testing.T) {
	tok := New()
	text := strings.Repeat("This is a test sentence with several words. ", 100)
	count := tok.Count(text)
	if count < 100 {
		t.Errorf("expected significant tokens for long text, got %d", count)
	}
}

func TestCount_Unicode(t *testing.T) {
	tok := New()
	text := "Hello 世界! 🌍 Привет"
	count := tok.Count(text)
	if count < 1 {
		t.Errorf("expected at least 1 token for unicode text, got %d", count)
	}
}

func TestCount_Code(t *testing.T) {
	tok := New()
	code := `func main() {
	fmt.Println("Hello, World!")
	for i := 0; i < 10; i++ {
		fmt.Printf("Number: %d\n", i)
	}
}`
	count := tok.Count(code)
	if count < 10 {
		t.Errorf("expected at least 10 tokens for code block, got %d", count)
	}
}

func TestCount_WhitespaceOnly(t *testing.T) {
	tok := New()
	count := tok.Count("   \t\n  ")
	// Whitespace-only should give minimal tokens
	if count < 1 {
		t.Errorf("expected at least 1 token for whitespace, got %d", count)
	}
}

func TestCount_SingleChar(t *testing.T) {
	tok := New()
	count := tok.Count("a")
	if count != 1 {
		t.Errorf("expected 1 token for single char, got %d", count)
	}
}

func TestCountAll(t *testing.T) {
	tok := New()
	count := tok.CountAll("hello", "world", "test")
	individual := tok.Count("hello") + tok.Count("world") + tok.Count("test")
	if count != individual {
		t.Errorf("CountAll(%d) != sum of Count(%d)", count, individual)
	}
}

func TestCountAll_Empty(t *testing.T) {
	tok := New()
	count := tok.CountAll()
	if count != 0 {
		t.Errorf("expected 0 for empty CountAll, got %d", count)
	}
}

func TestCountAll_WithEmpty(t *testing.T) {
	tok := New()
	count := tok.CountAll("hello", "", "world")
	countWithout := tok.CountAll("hello", "world")
	if count != countWithout {
		t.Errorf("empty strings should not affect count: %d vs %d", count, countWithout)
	}
}

func TestCount_Proportional(t *testing.T) {
	tok := New()
	short := "Hello world"
	long := strings.Repeat(short+" ", 10)

	shortCount := tok.Count(short)
	longCount := tok.Count(long)

	// Long text should have significantly more tokens
	if longCount <= shortCount {
		t.Errorf("longer text should have more tokens: short=%d, long=%d", shortCount, longCount)
	}
}

func TestCount_Newlines(t *testing.T) {
	tok := New()
	text := "line1\nline2\nline3\nline4\nline5"
	count := tok.Count(text)
	if count < 5 {
		t.Errorf("expected at least 5 tokens for 5 lines, got %d", count)
	}
}

func TestCount_JSON(t *testing.T) {
	tok := New()
	json := `{"key": "value", "nested": {"a": 1, "b": [1, 2, 3]}}`
	count := tok.Count(json)
	if count < 5 {
		t.Errorf("expected significant tokens for JSON, got %d", count)
	}
}

func TestCount_SpecialChars(t *testing.T) {
	tok := New()
	text := "Hello! @#$%^&*() Special {chars} [here]"
	count := tok.Count(text)
	if count < 1 {
		t.Errorf("expected at least 1 token for special chars, got %d", count)
	}
}

func TestCount_VeryLongWord(t *testing.T) {
	tok := New()
	word := strings.Repeat("a", 1000)
	count := tok.Count(word)
	// 1000 chars / 4 = ~250 tokens
	if count < 100 {
		t.Errorf("expected many tokens for 1000-char word, got %d", count)
	}
}

func TestEstimateTokens_Empty(t *testing.T) {
	result := estimateTokens("")
	if result != 0 {
		t.Errorf("expected 0 for empty, got %d", result)
	}
}

func TestEstimateTokens_MinimumOne(t *testing.T) {
	result := estimateTokens("hi")
	if result < 1 {
		t.Errorf("expected at least 1, got %d", result)
	}
}
