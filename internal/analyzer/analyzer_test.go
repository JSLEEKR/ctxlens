package analyzer

import (
	"testing"

	"github.com/JSLEEKR/ctxlens/internal/parser"
	"github.com/JSLEEKR/ctxlens/internal/pricing"
	"github.com/JSLEEKR/ctxlens/internal/tokenizer"
)

func newTestAnalyzer() *Analyzer {
	tok := tokenizer.New()
	pricer := pricing.New()
	return New(tok, pricer)
}

func TestNew(t *testing.T) {
	a := newTestAnalyzer()
	if a == nil {
		t.Fatal("expected non-nil analyzer")
	}
}

func TestAnalyze_BasicConversation(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "anthropic",
		Provider: "anthropic",
		Model:    "claude-sonnet-4",
		Messages: []parser.Message{
			{
				Role:  parser.RoleSystem,
				Index: 0,
				Content: []parser.ContentBlock{
					{Type: parser.ContentText, Text: "You are a helpful assistant."},
				},
			},
			{
				Role:  parser.RoleUser,
				Index: 1,
				Content: []parser.ContentBlock{
					{Type: parser.ContentText, Text: "Hello, how are you?"},
				},
			},
			{
				Role:  parser.RoleAssistant,
				Index: 2,
				Content: []parser.ContentBlock{
					{Type: parser.ContentText, Text: "I'm doing well, thank you for asking!"},
				},
			},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if result.Source != "test.json" {
		t.Errorf("expected source test.json, got %s", result.Source)
	}
	if result.TotalTokens == 0 {
		t.Error("expected non-zero total tokens")
	}
	if result.ByCategory[CategorySystem] == 0 {
		t.Error("expected non-zero system tokens")
	}
	if result.ByCategory[CategoryUser] == 0 {
		t.Error("expected non-zero user tokens")
	}
	if result.ByCategory[CategoryAssistant] == 0 {
		t.Error("expected non-zero assistant tokens")
	}
}

func TestAnalyze_WithToolCalls(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "anthropic",
		Provider: "anthropic",
		Model:    "claude-sonnet-4",
		Messages: []parser.Message{
			{
				Role:  parser.RoleAssistant,
				Index: 0,
				Content: []parser.ContentBlock{
					{Type: parser.ContentText, Text: "Let me search for that."},
					{Type: parser.ContentToolUse, Name: "web_search", Input: `{"query": "test"}`},
				},
			},
			{
				Role:  parser.RoleTool,
				Index: 1,
				Content: []parser.ContentBlock{
					{Type: parser.ContentToolResult, Text: "Search results here with some content."},
				},
			},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if result.ByCategory[CategoryToolCall] == 0 {
		t.Error("expected non-zero tool call tokens")
	}
	if result.ByCategory[CategoryToolResult] == 0 {
		t.Error("expected non-zero tool result tokens")
	}
}

func TestAnalyze_WithCodeBlocks(t *testing.T) {
	conv := &parser.Conversation{
		Format: "openai",
		Messages: []parser.Message{
			{
				Role:  parser.RoleAssistant,
				Index: 0,
				Content: []parser.ContentBlock{
					{Type: parser.ContentText, Text: "Here is the code:"},
					{Type: parser.ContentCodeBlock, Text: "def hello():\n    print('hello')"},
				},
			},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if result.ByCategory[CategoryCodeBlock] == 0 {
		t.Error("expected non-zero code block tokens")
	}
}

func TestAnalyze_EmptyConversation(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "openai",
		Messages: []parser.Message{},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "empty.json")

	if result.TotalTokens != 0 {
		t.Errorf("expected 0 tokens for empty conversation, got %d", result.TotalTokens)
	}
	if len(result.Segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(result.Segments))
	}
}

func TestAnalyze_CostCalculation(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "anthropic",
		Provider: "anthropic",
		Model:    "claude-sonnet-4",
		Messages: []parser.Message{
			{
				Role:  parser.RoleUser,
				Index: 0,
				Content: []parser.ContentBlock{
					{Type: parser.ContentText, Text: "Hello world, this is a test message."},
				},
			},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if result.EstimatedCost == 0 {
		t.Error("expected non-zero estimated cost")
	}
	if result.Provider != "anthropic" {
		t.Errorf("expected anthropic provider, got %s", result.Provider)
	}
}

func TestAnalyze_TopSegments(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "openai",
		Provider: "openai",
		Model:    "gpt-4o",
		Messages: []parser.Message{
			{Role: parser.RoleUser, Index: 0, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "Short"},
			}},
			{Role: parser.RoleUser, Index: 1, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "This is a much longer message with more content that should rank higher."},
			}},
			{Role: parser.RoleUser, Index: 2, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "Medium length message here."},
			}},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if len(result.TopSegments) == 0 {
		t.Fatal("expected top segments")
	}
	// Top segment should be the longest
	if result.TopSegments[0].Tokens < result.TopSegments[len(result.TopSegments)-1].Tokens {
		t.Error("top segments should be sorted by token count descending")
	}
}

func TestAnalyze_ModelLimit(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "openai",
		Messages: []parser.Message{},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if result.ModelLimit != DefaultModelLimit {
		t.Errorf("expected default model limit %d, got %d", DefaultModelLimit, result.ModelLimit)
	}
}

func TestAnalyze_IsEstimate(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "openai",
		Messages: []parser.Message{},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if !result.IsEstimate {
		t.Error("expected IsEstimate to be true")
	}
}

func TestAnalyze_DefaultProvider(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "openai",
		Messages: []parser.Message{
			{Role: parser.RoleUser, Index: 0, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "Hello"},
			}},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	// Should default to anthropic/claude-sonnet-4
	if result.Provider != "anthropic" {
		t.Errorf("expected default provider anthropic, got %s", result.Provider)
	}
	if result.Model != "claude-sonnet-4" {
		t.Errorf("expected default model claude-sonnet-4, got %s", result.Model)
	}
}

func TestUtilization(t *testing.T) {
	result := &AnalysisResult{
		TotalTokens: 50000,
		ModelLimit:  200000,
	}
	util := Utilization(result)
	if util != 25.0 {
		t.Errorf("expected 25.0%%, got %.1f%%", util)
	}
}

func TestUtilization_Zero(t *testing.T) {
	result := &AnalysisResult{
		TotalTokens: 0,
		ModelLimit:  200000,
	}
	util := Utilization(result)
	if util != 0 {
		t.Errorf("expected 0%%, got %.1f%%", util)
	}
}

func TestUtilization_ZeroLimit(t *testing.T) {
	result := &AnalysisResult{
		TotalTokens: 100,
		ModelLimit:  0,
	}
	util := Utilization(result)
	if util != 0 {
		t.Errorf("expected 0%% for zero limit, got %.1f%%", util)
	}
}

func TestCategoryPercentage(t *testing.T) {
	result := &AnalysisResult{
		TotalTokens: 1000,
		ByCategory: map[Category]int{
			CategorySystem: 250,
			CategoryUser:   750,
		},
	}
	pct := CategoryPercentage(result, CategorySystem)
	if pct != 25.0 {
		t.Errorf("expected 25.0%%, got %.1f%%", pct)
	}
}

func TestCategoryPercentage_ZeroTotal(t *testing.T) {
	result := &AnalysisResult{
		TotalTokens: 0,
		ByCategory:  map[Category]int{},
	}
	pct := CategoryPercentage(result, CategorySystem)
	if pct != 0 {
		t.Errorf("expected 0%%, got %.1f%%", pct)
	}
}

func TestCategoryPercentage_Missing(t *testing.T) {
	result := &AnalysisResult{
		TotalTokens: 1000,
		ByCategory: map[Category]int{
			CategoryUser: 1000,
		},
	}
	pct := CategoryPercentage(result, CategoryToolCall)
	if pct != 0 {
		t.Errorf("expected 0%% for missing category, got %.1f%%", pct)
	}
}

func TestAllCategories(t *testing.T) {
	cats := AllCategories()
	if len(cats) != 6 {
		t.Errorf("expected 6 categories, got %d", len(cats))
	}
}

func TestAnalyze_SegmentLabels(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "anthropic",
		Provider: "anthropic",
		Model:    "claude-sonnet-4",
		Messages: []parser.Message{
			{Role: parser.RoleSystem, Index: 0, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "System prompt"},
			}},
			{Role: parser.RoleUser, Index: 1, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "User message"},
			}},
			{Role: parser.RoleAssistant, Index: 2, Content: []parser.ContentBlock{
				{Type: parser.ContentToolUse, Name: "search", Input: "{}"},
			}},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	labels := make(map[string]bool)
	for _, seg := range result.Segments {
		labels[seg.Label] = true
	}

	if !labels["system_prompt (main)"] {
		t.Error("expected system_prompt label")
	}
	if !labels["user_message #1"] {
		t.Error("expected user_message label")
	}
	if !labels["tool_call (search)"] {
		t.Error("expected tool_call label")
	}
}

func TestAnalyze_CostByCategory(t *testing.T) {
	conv := &parser.Conversation{
		Format:   "anthropic",
		Provider: "anthropic",
		Model:    "claude-sonnet-4",
		Messages: []parser.Message{
			{Role: parser.RoleSystem, Index: 0, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "System prompt here"},
			}},
			{Role: parser.RoleUser, Index: 1, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "User message here"},
			}},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	totalCategoryCost := 0.0
	for _, cost := range result.CostByCategory {
		totalCategoryCost += cost
	}

	// Total should approximately match
	diff := result.EstimatedCost - totalCategoryCost
	if diff < -0.0001 || diff > 0.0001 {
		t.Errorf("category costs (%.6f) should sum to total (%.6f)", totalCategoryCost, result.EstimatedCost)
	}
}

func TestAnalyze_NilPricer(t *testing.T) {
	tok := tokenizer.New()
	a := New(tok, nil)

	conv := &parser.Conversation{
		Format: "openai",
		Messages: []parser.Message{
			{Role: parser.RoleUser, Index: 0, Content: []parser.ContentBlock{
				{Type: parser.ContentText, Text: "Hello"},
			}},
		},
	}

	result := a.Analyze(conv, "test.json")

	if result.EstimatedCost != 0 {
		t.Errorf("expected 0 cost with nil pricer, got %f", result.EstimatedCost)
	}
}

func TestAnalyze_ToolResultLabel(t *testing.T) {
	conv := &parser.Conversation{
		Format: "openai",
		Messages: []parser.Message{
			{Role: parser.RoleTool, Index: 0, Content: []parser.ContentBlock{
				{Type: parser.ContentToolResult, Text: "Some result", Name: "my_tool"},
			}},
		},
	}

	a := newTestAnalyzer()
	result := a.Analyze(conv, "test.json")

	if len(result.Segments) == 0 {
		t.Fatal("expected segments")
	}
	if result.Segments[0].Label != "tool_result (my_tool)" {
		t.Errorf("expected label 'tool_result (my_tool)', got %q", result.Segments[0].Label)
	}
}
