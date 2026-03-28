// Package parser provides format detection and message parsing for LLM conversation payloads.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Role represents the role of a message sender.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ContentType represents the type of content within a message.
type ContentType string

const (
	ContentText       ContentType = "text"
	ContentToolUse    ContentType = "tool_use"
	ContentToolResult ContentType = "tool_result"
	ContentCodeBlock  ContentType = "code_block"
)

// ContentBlock represents a piece of content within a message.
type ContentBlock struct {
	Type     ContentType
	Text     string
	Name     string // for tool_use: tool name
	ToolID   string // for tool_use/tool_result: tool call ID
	Input    string // for tool_use: serialized input
}

// Message represents a parsed conversation message.
type Message struct {
	Role    Role
	Content []ContentBlock
	Index   int // position in conversation
}

// Conversation represents a fully parsed conversation.
type Conversation struct {
	Messages []Message
	Format   string // "anthropic", "openai", "jsonl"
	Model    string // if detected from payload
	Provider string // if detected from payload
}

// Format represents a conversation format.
type Format string

const (
	FormatAnthropic Format = "anthropic"
	FormatOpenAI    Format = "openai"
	FormatJSONL     Format = "jsonl"
	FormatUnknown   Format = "unknown"
)

// Parser defines the interface for conversation parsers.
type Parser interface {
	Parse(data []byte) (*Conversation, error)
}

// DetectFormat detects the format of a conversation payload.
func DetectFormat(data []byte) Format {
	data = trimBOM(data)
	trimmed := strings.TrimSpace(string(data))

	// First, try parsing as a single JSON object
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
		// Anthropic: has "system" field (string or array) and "messages"
		if _, hasSystem := obj["system"]; hasSystem {
			if _, hasMessages := obj["messages"]; hasMessages {
				return FormatAnthropic
			}
		}

		// OpenAI: has "messages" but no "system" at top level
		if _, hasMessages := obj["messages"]; hasMessages {
			return FormatOpenAI
		}

		// If it has "role" at top level, treat as single message in OpenAI format
		if _, hasRole := obj["role"]; hasRole {
			return FormatOpenAI
		}
	}

	// JSONL: multiple lines, each a JSON object
	if strings.Contains(trimmed, "\n") {
		lines := strings.Split(trimmed, "\n")
		jsonLines := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
				jsonLines++
			}
		}
		if jsonLines >= 2 {
			return FormatJSONL
		}
	}

	return FormatUnknown
}

// Parse parses conversation data, auto-detecting the format.
func Parse(data []byte) (*Conversation, error) {
	data = trimBOM(data)
	format := DetectFormat(data)
	switch format {
	case FormatAnthropic:
		return NewAnthropicParser().Parse(data)
	case FormatOpenAI:
		return NewOpenAIParser().Parse(data)
	case FormatJSONL:
		return NewJSONLParser().Parse(data)
	default:
		return nil, fmt.Errorf("unknown conversation format")
	}
}

// trimBOM removes UTF-8 BOM if present.
func trimBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}

// extractCodeBlocks splits text into code blocks and non-code parts.
func extractCodeBlocks(text string) (nonCode string, codeBlocks []string) {
	parts := strings.Split(text, "```")
	for i, part := range parts {
		if i%2 == 0 {
			nonCode += part
		} else {
			codeBlocks = append(codeBlocks, part)
		}
	}
	return nonCode, codeBlocks
}
