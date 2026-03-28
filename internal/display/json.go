package display

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/JSLEEKR/ctxlens/internal/analyzer"
)

// JSONOutput represents the JSON output structure.
type JSONOutput struct {
	Source      string                    `json:"source"`
	TotalTokens int                      `json:"total_tokens"`
	ModelLimit  int                       `json:"model_limit"`
	Utilization float64                   `json:"utilization_pct"`
	EstCost     float64                   `json:"estimated_cost_usd"`
	Provider    string                    `json:"provider"`
	Model       string                    `json:"model"`
	IsEstimate  bool                      `json:"is_estimate"`
	Categories  []JSONCategory            `json:"categories"`
	TopSegments []JSONSegment             `json:"top_segments"`
}

// JSONCategory represents a category in JSON output.
type JSONCategory struct {
	Name    string  `json:"name"`
	Tokens  int     `json:"tokens"`
	Percent float64 `json:"percent"`
	Cost    float64 `json:"cost_usd"`
}

// JSONSegment represents a segment in JSON output.
type JSONSegment struct {
	Label    string `json:"label"`
	Category string `json:"category"`
	Tokens   int    `json:"tokens"`
}

// RenderJSON writes a JSON-formatted analysis result to the writer.
func RenderJSON(w io.Writer, result *analyzer.AnalysisResult) {
	output := JSONOutput{
		Source:      result.Source,
		TotalTokens: result.TotalTokens,
		ModelLimit:  result.ModelLimit,
		Utilization: analyzer.Utilization(result),
		EstCost:     result.EstimatedCost,
		Provider:    result.Provider,
		Model:       result.Model,
		IsEstimate:  result.IsEstimate,
	}

	for _, cat := range analyzer.AllCategories() {
		tokens := result.ByCategory[cat]
		if tokens == 0 {
			continue
		}
		output.Categories = append(output.Categories, JSONCategory{
			Name:    string(cat),
			Tokens:  tokens,
			Percent: analyzer.CategoryPercentage(result, cat),
			Cost:    result.CostByCategory[cat],
		})
	}

	for _, seg := range result.TopSegments {
		output.TopSegments = append(output.TopSegments, JSONSegment{
			Label:    seg.Label,
			Category: string(seg.Category),
			Tokens:   seg.Tokens,
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
		return
	}
	fmt.Fprintln(w, string(data))
}
