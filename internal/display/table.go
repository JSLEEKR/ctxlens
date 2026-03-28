// Package display provides output formatting for analysis results.
package display

import (
	"fmt"
	"io"
	"strings"

	"github.com/JSLEEKR/ctxlens/internal/analyzer"
)

const barWidth = 20

// RenderTable writes a table-formatted analysis result to the writer.
func RenderTable(w io.Writer, result *analyzer.AnalysisResult) {
	utilization := analyzer.Utilization(result)

	fmt.Fprintf(w, "\nContext Breakdown -- %s\n", result.Source)
	fmt.Fprintf(w, "=====================================\n\n")

	estimateLabel := ""
	if result.IsEstimate {
		estimateLabel = " (estimated)"
	}

	fmt.Fprintf(w, "Total: %s tokens%s (%.1f%% of %s limit)\n",
		formatNumber(result.TotalTokens),
		estimateLabel,
		utilization,
		formatLimit(result.ModelLimit),
	)

	if result.EstimatedCost > 0 {
		modelLabel := result.Model
		if modelLabel == "" {
			modelLabel = "default"
		}
		providerLabel := result.Provider
		if providerLabel == "" {
			providerLabel = "unknown"
		}
		fmt.Fprintf(w, "Est. Cost: $%.3f (%s %s @ input pricing)\n",
			result.EstimatedCost, providerLabel, modelLabel)
	}

	fmt.Fprintln(w)

	// Header
	fmt.Fprintf(w, "%-18s %8s %6s %8s   %-20s\n",
		"Category", "Tokens", "%", "Cost", "Bar")
	fmt.Fprintf(w, "%-18s %8s %6s %8s   %-20s\n",
		strings.Repeat("-", 18),
		strings.Repeat("-", 8),
		strings.Repeat("-", 6),
		strings.Repeat("-", 8),
		strings.Repeat("-", 20),
	)

	// Category rows
	for _, cat := range analyzer.AllCategories() {
		tokens, ok := result.ByCategory[cat]
		if !ok || tokens == 0 {
			continue
		}
		pct := analyzer.CategoryPercentage(result, cat)
		cost := result.CostByCategory[cat]
		bar := renderBar(pct)

		fmt.Fprintf(w, "%-18s %8s %5.1f%% $%7.3f   %s\n",
			string(cat),
			formatNumber(tokens),
			pct,
			cost,
			bar,
		)
	}

	// Top Segments
	if len(result.TopSegments) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Top Segments:")
		for i, seg := range result.TopSegments {
			fmt.Fprintf(w, "  %d. %-35s %s tokens\n",
				i+1,
				seg.Label,
				formatNumber(seg.Tokens),
			)
		}
	}

	fmt.Fprintln(w)
}

func renderBar(pct float64) string {
	filled := int(pct / 100 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	empty := barWidth - filled
	return strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", empty)
}

// FormatNumber formats an integer with comma separators.
func FormatNumber(n int) string {
	return formatNumber(n)
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func formatLimit(limit int) string {
	if limit >= 1000000 {
		return fmt.Sprintf("%.0fM", float64(limit)/1000000)
	}
	if limit >= 1000 {
		return fmt.Sprintf("%dK", limit/1000)
	}
	return fmt.Sprintf("%d", limit)
}
