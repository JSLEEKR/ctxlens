package parser

import (
	"encoding/json"
	"fmt"
)

// OpenAIParser parses OpenAI API format conversations.
type OpenAIParser struct{}

// NewOpenAIParser creates a new OpenAIParser.
func NewOpenAIParser() *OpenAIParser {
	return &OpenAIParser{}
}

type openAIPayload struct {
	Messages []openAIMsg `json:"messages"`
	Model    string      `json:"model"`
}

type openAIMsg struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
}

type openAIToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Parse parses an OpenAI format conversation payload.
func (p *OpenAIParser) Parse(data []byte) (*Conversation, error) {
	var payload openAIPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parsing openai payload: %w", err)
	}

	// Handle single message wrapped in object
	if payload.Messages == nil {
		var singleMsg openAIMsg
		if err := json.Unmarshal(data, &singleMsg); err == nil && singleMsg.Role != "" {
			payload.Messages = []openAIMsg{singleMsg}
		} else {
			return nil, fmt.Errorf("no messages found in openai payload")
		}
	}

	conv := &Conversation{
		Format:   "openai",
		Provider: "openai",
		Model:    payload.Model,
	}

	for i, omsg := range payload.Messages {
		msg := Message{
			Role:  Role(omsg.Role),
			Index: i,
		}

		// Parse content
		contentText := parseOpenAIContent(omsg.Content)

		if omsg.Role == "tool" {
			// Tool result message
			msg.Role = RoleTool
			msg.Content = append(msg.Content, ContentBlock{
				Type:   ContentToolResult,
				Text:   contentText,
				ToolID: omsg.ToolCallID,
				Name:   omsg.Name,
			})
		} else {
			// Regular message - extract code blocks
			nonCode, codeBlocks := extractCodeBlocks(contentText)
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
			if len(msg.Content) == 0 {
				msg.Content = append(msg.Content, ContentBlock{Type: ContentText, Text: ""})
			}
		}

		// Parse tool calls
		for _, tc := range omsg.ToolCalls {
			msg.Content = append(msg.Content, ContentBlock{
				Type:   ContentToolUse,
				Name:   tc.Function.Name,
				ToolID: tc.ID,
				Input:  tc.Function.Arguments,
			})
		}

		conv.Messages = append(conv.Messages, msg)
	}

	return conv, nil
}

func parseOpenAIContent(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}
	// Try as string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// Try as array (OpenAI multi-modal)
	var parts []map[string]interface{}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var result string
		for _, part := range parts {
			if t, ok := part["type"].(string); ok && t == "text" {
				if text, ok := part["text"].(string); ok {
					result += text
				}
			}
		}
		return result
	}
	return string(raw)
}
