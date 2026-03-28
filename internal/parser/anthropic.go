package parser

import (
	"encoding/json"
	"fmt"
)

// AnthropicParser parses Anthropic API format conversations.
type AnthropicParser struct{}

// NewAnthropicParser creates a new AnthropicParser.
func NewAnthropicParser() *AnthropicParser {
	return &AnthropicParser{}
}

type anthropicPayload struct {
	System   json.RawMessage `json:"system"`
	Messages []anthropicMsg  `json:"messages"`
	Model    string          `json:"model"`
}

type anthropicMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type anthropicContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	Input     json.RawMessage `json:"input"`
	Content   json.RawMessage `json:"content"`
}

// Parse parses an Anthropic format conversation payload.
func (p *AnthropicParser) Parse(data []byte) (*Conversation, error) {
	var payload anthropicPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parsing anthropic payload: %w", err)
	}

	conv := &Conversation{
		Format:   "anthropic",
		Provider: "anthropic",
		Model:    payload.Model,
	}

	// Parse system prompt
	if payload.System != nil {
		sysText := parseSystemField(payload.System)
		if sysText != "" {
			msg := Message{
				Role:  RoleSystem,
				Index: 0,
			}
			nonCode, codeBlocks := extractCodeBlocks(sysText)
			if nonCode != "" {
				msg.Content = append(msg.Content, ContentBlock{
					Type: ContentText,
					Text: nonCode,
				})
			}
			for _, cb := range codeBlocks {
				msg.Content = append(msg.Content, ContentBlock{
					Type: ContentCodeBlock,
					Text: cb,
				})
			}
			conv.Messages = append(conv.Messages, msg)
		}
	}

	// Parse messages
	for i, amsg := range payload.Messages {
		msg := Message{
			Role:  Role(amsg.Role),
			Index: i + 1,
		}

		blocks := parseAnthropicContent(amsg.Content)
		msg.Content = blocks
		conv.Messages = append(conv.Messages, msg)
	}

	return conv, nil
}

func parseSystemField(raw json.RawMessage) string {
	// Try as string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	// Try as array of content blocks
	var blocks []anthropicContentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var result string
		for _, b := range blocks {
			if b.Text != "" {
				result += b.Text
			}
		}
		return result
	}

	return ""
}

func parseAnthropicContent(raw json.RawMessage) []ContentBlock {
	// Try as string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var blocks []ContentBlock
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
		if len(blocks) == 0 {
			blocks = append(blocks, ContentBlock{Type: ContentText, Text: ""})
		}
		return blocks
	}

	// Try as array of content blocks
	var rawBlocks []anthropicContentBlock
	if err := json.Unmarshal(raw, &rawBlocks); err != nil {
		return []ContentBlock{{Type: ContentText, Text: string(raw)}}
	}

	var blocks []ContentBlock
	for _, rb := range rawBlocks {
		switch rb.Type {
		case "text":
			nonCode, codeBlocks := extractCodeBlocks(rb.Text)
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
		case "tool_use":
			inputStr := ""
			if rb.Input != nil {
				inputStr = string(rb.Input)
			}
			blocks = append(blocks, ContentBlock{
				Type:   ContentToolUse,
				Name:   rb.Name,
				ToolID: rb.ID,
				Input:  inputStr,
			})
		case "tool_result":
			resultText := ""
			if rb.Content != nil {
				// Try as string
				var s string
				if err := json.Unmarshal(rb.Content, &s); err == nil {
					resultText = s
				} else {
					resultText = string(rb.Content)
				}
			}
			blocks = append(blocks, ContentBlock{
				Type:   ContentToolResult,
				Text:   resultText,
				ToolID: rb.ToolUseID,
			})
		default:
			if rb.Text != "" {
				blocks = append(blocks, ContentBlock{
					Type: ContentText,
					Text: rb.Text,
				})
			}
		}
	}

	return blocks
}
