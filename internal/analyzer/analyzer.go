package analyzer

import (
	"fmt"
	"sort"

	"github.com/JSLEEKR/ctxlens/internal/parser"
	"github.com/JSLEEKR/ctxlens/internal/pricing"
	"github.com/JSLEEKR/ctxlens/internal/tokenizer"
)

// DefaultModelLimit is the default context window size (200K tokens).
const DefaultModelLimit = 200000

// TopSegmentCount is how many top segments to return.
const TopSegmentCount = 5

// Analyzer decomposes conversations into categorized segments.
type Analyzer struct {
	tokenizer *tokenizer.Tokenizer
	pricer    *pricing.Pricer
}

// New creates a new Analyzer.
func New(tok *tokenizer.Tokenizer, pricer *pricing.Pricer) *Analyzer {
	return &Analyzer{
		tokenizer: tok,
		pricer:    pricer,
	}
}

// Analyze decomposes a conversation into an AnalysisResult.
func (a *Analyzer) Analyze(conv *parser.Conversation, source string) *AnalysisResult {
	result := &AnalysisResult{
		Source:         source,
		ByCategory:     make(map[Category]int),
		CostByCategory: make(map[Category]float64),
		ModelLimit:     DefaultModelLimit,
		IsEstimate:    a.tokenizer.IsEstimate(),
		Provider:      conv.Provider,
		Model:         conv.Model,
	}

	// Decompose messages into segments
	for _, msg := range conv.Messages {
		segments := a.decomposeMessage(msg)
		result.Segments = append(result.Segments, segments...)
	}

	// Calculate totals
	for i := range result.Segments {
		seg := &result.Segments[i]
		result.TotalTokens += seg.Tokens
		result.ByCategory[seg.Category] += seg.Tokens
	}

	// Calculate costs
	if a.pricer != nil && conv.Provider != "" && conv.Model != "" {
		for cat, tokens := range result.ByCategory {
			cost := a.pricer.CalculateCost(conv.Provider, conv.Model, tokens, true)
			result.CostByCategory[cat] = cost
			result.EstimatedCost += cost
		}
	} else if a.pricer != nil {
		// Default to anthropic/claude-sonnet-4
		for cat, tokens := range result.ByCategory {
			cost := a.pricer.CalculateCost("anthropic", "claude-sonnet-4", tokens, true)
			result.CostByCategory[cat] = cost
			result.EstimatedCost += cost
		}
		if result.Provider == "" {
			result.Provider = "anthropic"
		}
		if result.Model == "" {
			result.Model = "claude-sonnet-4"
		}
	}

	// Get top segments
	result.TopSegments = a.topSegments(result.Segments, TopSegmentCount)

	return result
}

func (a *Analyzer) decomposeMessage(msg parser.Message) []Segment {
	var segments []Segment

	for _, block := range msg.Content {
		seg := Segment{}

		switch block.Type {
		case parser.ContentText:
			seg.Content = block.Text
			seg.Tokens = a.tokenizer.Count(block.Text)
			switch msg.Role {
			case parser.RoleSystem:
				seg.Category = CategorySystem
				seg.Label = fmt.Sprintf("system_prompt (main)")
			case parser.RoleUser:
				seg.Category = CategoryUser
				seg.Label = fmt.Sprintf("user_message #%d", msg.Index)
			case parser.RoleAssistant:
				seg.Category = CategoryAssistant
				seg.Label = fmt.Sprintf("assistant_message #%d", msg.Index)
			default:
				seg.Category = CategoryUser
				seg.Label = fmt.Sprintf("message #%d", msg.Index)
			}

		case parser.ContentToolUse:
			seg.Content = block.Name + " " + block.Input
			seg.Tokens = a.tokenizer.Count(seg.Content)
			seg.Category = CategoryToolCall
			seg.Label = fmt.Sprintf("tool_call (%s)", block.Name)

		case parser.ContentToolResult:
			seg.Content = block.Text
			seg.Tokens = a.tokenizer.Count(block.Text)
			seg.Category = CategoryToolResult
			name := block.Name
			if name == "" {
				name = "result"
			}
			seg.Label = fmt.Sprintf("tool_result (%s)", name)

		case parser.ContentCodeBlock:
			seg.Content = block.Text
			seg.Tokens = a.tokenizer.Count(block.Text)
			seg.Category = CategoryCodeBlock
			seg.Label = fmt.Sprintf("code_block #%d", msg.Index)
		}

		segments = append(segments, seg)
	}

	return segments
}

func (a *Analyzer) topSegments(segments []Segment, n int) []Segment {
	sorted := make([]Segment, len(segments))
	copy(sorted, segments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Tokens > sorted[j].Tokens
	})
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

// Utilization returns the percentage of context window used.
func Utilization(result *AnalysisResult) float64 {
	if result.ModelLimit == 0 {
		return 0
	}
	return float64(result.TotalTokens) / float64(result.ModelLimit) * 100
}

// CategoryPercentage returns the percentage of total tokens for a category.
func CategoryPercentage(result *AnalysisResult, cat Category) float64 {
	if result.TotalTokens == 0 {
		return 0
	}
	return float64(result.ByCategory[cat]) / float64(result.TotalTokens) * 100
}
