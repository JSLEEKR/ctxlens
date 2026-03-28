// Package analyzer provides context window decomposition and analysis.
package analyzer

// Category represents a token usage category.
type Category string

const (
	CategorySystem    Category = "System Prompt"
	CategoryUser      Category = "User Messages"
	CategoryAssistant Category = "Assistant Msgs"
	CategoryToolCall  Category = "Tool Calls"
	CategoryToolResult Category = "Tool Results"
	CategoryCodeBlock Category = "Code Blocks"
)

// AllCategories returns all categories in display order.
func AllCategories() []Category {
	return []Category{
		CategorySystem,
		CategoryUser,
		CategoryAssistant,
		CategoryToolCall,
		CategoryToolResult,
		CategoryCodeBlock,
	}
}

// Segment represents a decomposed piece of a conversation.
type Segment struct {
	// Label is a human-readable identifier (e.g., "system_prompt (main)", "user_message #3").
	Label string
	// Category classifies this segment.
	Category Category
	// Tokens is the token count for this segment.
	Tokens int
	// Content is the raw text content (not echoed in output, used for counting).
	Content string
}

// AnalysisResult holds the complete decomposition of a conversation.
type AnalysisResult struct {
	// Source is the file or input source name.
	Source string
	// TotalTokens is the sum of all segment tokens.
	TotalTokens int
	// Segments contains all decomposed segments.
	Segments []Segment
	// ByCategory maps category to total token count.
	ByCategory map[Category]int
	// TopSegments contains the top N segments by token count.
	TopSegments []Segment
	// ModelLimit is the context window size for the target model.
	ModelLimit int
	// EstimatedCost is the total estimated cost in USD.
	EstimatedCost float64
	// CostByCategory maps category to estimated cost.
	CostByCategory map[Category]float64
	// Provider is the detected or specified provider.
	Provider string
	// Model is the detected or specified model.
	Model string
	// IsEstimate indicates whether token counts are estimates (not tiktoken).
	IsEstimate bool
}
