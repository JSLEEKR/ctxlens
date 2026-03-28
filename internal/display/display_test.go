package display

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JSLEEKR/ctxlens/internal/analyzer"
)

func makeTestResult() *analyzer.AnalysisResult {
	return &analyzer.AnalysisResult{
		Source:      "test.json",
		TotalTokens: 10000,
		ModelLimit:  200000,
		IsEstimate: true,
		Provider:   "anthropic",
		Model:      "claude-sonnet-4",
		EstimatedCost: 0.030,
		ByCategory: map[analyzer.Category]int{
			analyzer.CategorySystem:     2000,
			analyzer.CategoryUser:       3000,
			analyzer.CategoryAssistant:  2500,
			analyzer.CategoryToolCall:   500,
			analyzer.CategoryToolResult: 1500,
			analyzer.CategoryCodeBlock:  500,
		},
		CostByCategory: map[analyzer.Category]float64{
			analyzer.CategorySystem:     0.006,
			analyzer.CategoryUser:       0.009,
			analyzer.CategoryAssistant:  0.0075,
			analyzer.CategoryToolCall:   0.0015,
			analyzer.CategoryToolResult: 0.0045,
			analyzer.CategoryCodeBlock:  0.0015,
		},
		TopSegments: []analyzer.Segment{
			{Label: "user_message #3", Category: analyzer.CategoryUser, Tokens: 2000},
			{Label: "system_prompt (main)", Category: analyzer.CategorySystem, Tokens: 1500},
			{Label: "tool_result (search)", Category: analyzer.CategoryToolResult, Tokens: 1000},
		},
		Segments: []analyzer.Segment{
			{Label: "system_prompt (main)", Category: analyzer.CategorySystem, Tokens: 2000},
			{Label: "user_message #1", Category: analyzer.CategoryUser, Tokens: 1000},
			{Label: "user_message #3", Category: analyzer.CategoryUser, Tokens: 2000},
			{Label: "assistant_message #2", Category: analyzer.CategoryAssistant, Tokens: 2500},
			{Label: "tool_call (search)", Category: analyzer.CategoryToolCall, Tokens: 500},
			{Label: "tool_result (search)", Category: analyzer.CategoryToolResult, Tokens: 1500},
			{Label: "code_block #2", Category: analyzer.CategoryCodeBlock, Tokens: 500},
		},
	}
}

func TestRenderTable_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "Context Breakdown") {
		t.Error("expected 'Context Breakdown' in output")
	}
	if !strings.Contains(output, "test.json") {
		t.Error("expected source filename in output")
	}
}

func TestRenderTable_ContainsTokens(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "10,000") {
		t.Error("expected formatted total tokens")
	}
}

func TestRenderTable_ContainsCategories(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	for _, cat := range analyzer.AllCategories() {
		if !strings.Contains(output, string(cat)) {
			t.Errorf("expected category %s in output", cat)
		}
	}
}

func TestRenderTable_ContainsPercentages(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "%") {
		t.Error("expected percentages in output")
	}
}

func TestRenderTable_ContainsCost(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "$") {
		t.Error("expected cost in output")
	}
}

func TestRenderTable_ContainsTopSegments(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "Top Segments") {
		t.Error("expected 'Top Segments' section")
	}
	if !strings.Contains(output, "user_message #3") {
		t.Error("expected top segment label")
	}
}

func TestRenderTable_EstimateLabel(t *testing.T) {
	var buf bytes.Buffer
	result := makeTestResult()
	result.IsEstimate = true
	RenderTable(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "estimated") {
		t.Error("expected 'estimated' label")
	}
}

func TestRenderTable_NoEstimateLabel(t *testing.T) {
	var buf bytes.Buffer
	result := makeTestResult()
	result.IsEstimate = false
	RenderTable(&buf, result)
	output := buf.String()

	if strings.Contains(output, "estimated") {
		t.Error("should not show 'estimated' when not estimate")
	}
}

func TestRenderTable_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	result := &analyzer.AnalysisResult{
		Source:         "empty.json",
		TotalTokens:   0,
		ModelLimit:     200000,
		ByCategory:     make(map[analyzer.Category]int),
		CostByCategory: make(map[analyzer.Category]float64),
	}
	RenderTable(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "empty.json") {
		t.Error("expected source in empty result")
	}
}

func TestRenderTable_UtilizationPercentage(t *testing.T) {
	var buf bytes.Buffer
	RenderTable(&buf, makeTestResult())
	output := buf.String()

	// 10000/200000 = 5.0%
	if !strings.Contains(output, "5.0%") {
		t.Error("expected 5.0% utilization")
	}
}

func TestRenderJSON_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	RenderJSON(&buf, makeTestResult())

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
}

func TestRenderJSON_Fields(t *testing.T) {
	var buf bytes.Buffer
	RenderJSON(&buf, makeTestResult())

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if output.Source != "test.json" {
		t.Errorf("expected source test.json, got %s", output.Source)
	}
	if output.TotalTokens != 10000 {
		t.Errorf("expected 10000 tokens, got %d", output.TotalTokens)
	}
	if output.ModelLimit != 200000 {
		t.Errorf("expected model limit 200000, got %d", output.ModelLimit)
	}
	if output.Provider != "anthropic" {
		t.Errorf("expected anthropic provider, got %s", output.Provider)
	}
	if !output.IsEstimate {
		t.Error("expected IsEstimate to be true")
	}
}

func TestRenderJSON_Categories(t *testing.T) {
	var buf bytes.Buffer
	RenderJSON(&buf, makeTestResult())

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(output.Categories) != 6 {
		t.Errorf("expected 6 categories, got %d", len(output.Categories))
	}
}

func TestRenderJSON_TopSegments(t *testing.T) {
	var buf bytes.Buffer
	RenderJSON(&buf, makeTestResult())

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(output.TopSegments) != 3 {
		t.Errorf("expected 3 top segments, got %d", len(output.TopSegments))
	}
}

func TestRenderJSON_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	result := &analyzer.AnalysisResult{
		Source:         "empty.json",
		ByCategory:     make(map[analyzer.Category]int),
		CostByCategory: make(map[analyzer.Category]float64),
	}
	RenderJSON(&buf, result)

	var output JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("invalid JSON for empty result: %v", err)
	}
}

func TestRenderFlame_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	RenderFlame(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "Context Flamegraph") {
		t.Error("expected 'Context Flamegraph' header")
	}
}

func TestRenderFlame_ContainsSegments(t *testing.T) {
	var buf bytes.Buffer
	RenderFlame(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "system_prompt") {
		t.Error("expected segment labels in flame output")
	}
}

func TestRenderFlame_ContainsBars(t *testing.T) {
	var buf bytes.Buffer
	RenderFlame(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "\u2588") {
		t.Error("expected bar characters in flame output")
	}
}

func TestRenderFlame_EmptyResult(t *testing.T) {
	var buf bytes.Buffer
	result := &analyzer.AnalysisResult{
		Source:         "empty.json",
		ByCategory:     make(map[analyzer.Category]int),
		CostByCategory: make(map[analyzer.Category]float64),
	}
	RenderFlame(&buf, result)
	output := buf.String()

	if !strings.Contains(output, "empty") {
		t.Error("expected (empty) for zero-token result")
	}
}

func TestRenderFlame_TotalTokens(t *testing.T) {
	var buf bytes.Buffer
	RenderFlame(&buf, makeTestResult())
	output := buf.String()

	if !strings.Contains(output, "10,000") {
		t.Error("expected formatted total tokens in flame output")
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{10000, "10,000"},
		{100000, "100,000"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatNumber(tt.n)
			if got != tt.want {
				t.Errorf("formatNumber(%d) = %s, want %s", tt.n, got, tt.want)
			}
		})
	}
}

func TestFormatLimit(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{200000, "200K"},
		{1000000, "1M"},
		{500, "500"},
		{128000, "128K"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatLimit(tt.n)
			if got != tt.want {
				t.Errorf("formatLimit(%d) = %s, want %s", tt.n, got, tt.want)
			}
		})
	}
}

func TestRenderBar(t *testing.T) {
	bar := renderBar(50)
	if len([]rune(bar)) != barWidth {
		t.Errorf("bar should be %d runes, got %d", barWidth, len([]rune(bar)))
	}
}

func TestRenderBar_Zero(t *testing.T) {
	bar := renderBar(0)
	// Should be all empty
	for _, r := range bar {
		if r == '\u2588' {
			t.Error("expected no filled chars for 0%")
			break
		}
	}
}

func TestRenderBar_Full(t *testing.T) {
	bar := renderBar(100)
	for _, r := range bar {
		if r == '\u2591' {
			t.Error("expected all filled chars for 100%")
			break
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"this is very long text", 10, "this is..."},
		{"exactly10!", 10, "exactly10!"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}
