package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONLParser parses JSONL (Claude Code log) format conversations.
type JSONLParser struct{}

// NewJSONLParser creates a new JSONLParser.
func NewJSONLParser() *JSONLParser {
	return &JSONLParser{}
}

type jsonlEntry struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Type    string          `json:"type"`
	Model   string          `json:"model"`
}

// Parse parses JSONL format conversation data.
func (p *JSONLParser) Parse(data []byte) (*Conversation, error) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	conv := &Conversation{
		Format: "jsonl",
	}

	idx := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry jsonlEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip unparseable lines
			continue
		}

		if entry.Model != "" && conv.Model == "" {
			conv.Model = entry.Model
		}

		msg := Message{
			Index: idx,
		}

		// Determine role
		switch entry.Role {
		case "system":
			msg.Role = RoleSystem
		case "user":
			msg.Role = RoleUser
		case "assistant":
			msg.Role = RoleAssistant
		case "tool":
			msg.Role = RoleTool
		default:
			// Try to infer from type field
			if entry.Type == "system" {
				msg.Role = RoleSystem
			} else if entry.Type == "tool_result" {
				msg.Role = RoleTool
			} else {
				msg.Role = RoleUser
			}
		}

		// Parse content
		blocks := parseJSONLContent(entry)
		if len(blocks) == 0 {
			continue
		}
		msg.Content = blocks

		conv.Messages = append(conv.Messages, msg)
		idx++
	}

	if len(conv.Messages) == 0 {
		return nil, fmt.Errorf("no valid messages found in JSONL data")
	}

	// Detect provider from model name
	if conv.Model != "" {
		if strings.Contains(conv.Model, "claude") {
			conv.Provider = "anthropic"
		} else if strings.Contains(conv.Model, "gpt") {
			conv.Provider = "openai"
		} else if strings.Contains(conv.Model, "gemini") {
			conv.Provider = "google"
		}
	}

	return conv, nil
}

func parseJSONLContent(entry jsonlEntry) []ContentBlock {
	if entry.Content == nil {
		return nil
	}

	// Try as string
	var s string
	if err := json.Unmarshal(entry.Content, &s); err == nil {
		if s == "" {
			return nil
		}
		var blocks []ContentBlock
		if entry.Role == "tool" || entry.Type == "tool_result" {
			blocks = append(blocks, ContentBlock{
				Type: ContentToolResult,
				Text: s,
			})
		} else {
			nonCode, codeBlocks := extractCodeBlocks(s)
			if nonCode != "" {
				blocks = append(blocks, ContentBlock{
					Type: ContentText,
					Text: nonCode,
				})
			}
			for _, cb := range codeBlocks {
				blocks = append(blocks, ContentBlock{
					Type: ContentCodeBlock,
					Text: cb,
				})
			}
		}
		return blocks
	}

	// Try as Anthropic-style content blocks
	return parseAnthropicContent(entry.Content)
}
