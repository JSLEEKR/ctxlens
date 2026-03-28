package display

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/JSLEEKR/ctxlens/internal/analyzer"
)

const flameWidth = 60

// RenderFlame writes an ASCII flamegraph-style output to the writer.
func RenderFlame(w io.Writer, result *analyzer.AnalysisResult) {
	fmt.Fprintf(w, "\nContext Flamegraph -- %s\n", result.Source)
	fmt.Fprintf(w, "Total: %s tokens\n\n", formatNumber(result.TotalTokens))

	if result.TotalTokens == 0 {
		fmt.Fprintln(w, "(empty)")
		return
	}

	// Sort segments by token count descending
	segments := make([]analyzer.Segment, len(result.Segments))
	copy(segments, result.Segments)
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].Tokens > segments[j].Tokens
	})

	// Render each segment as a bar
	for _, seg := range segments {
		if seg.Tokens == 0 {
			continue
		}
		pct := float64(seg.Tokens) / float64(result.TotalTokens) * 100
		barLen := int(pct / 100 * float64(flameWidth))
		if barLen < 1 {
			barLen = 1
		}

		bar := strings.Repeat("\u2588", barLen)
		fmt.Fprintf(w, "%-35s %8s (%5.1f%%) %s\n",
			truncate(seg.Label, 35),
			formatNumber(seg.Tokens),
			pct,
			bar,
		)
	}

	fmt.Fprintln(w)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
