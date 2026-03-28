package parser

import (
	"testing"
)

func TestOpenAIParser_BasicConversation(t *testing.T) {
	data := []byte(`{
		"model": "gpt-4o",
		"messages": [
			{"role": "system", "content": "You are helpful"},
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi!"}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conv.Format != "openai" {
		t.Errorf("expected openai format, got %s", conv.Format)
	}
	if conv.Provider != "openai" {
		t.Errorf("expected openai provider, got %s", conv.Provider)
	}
	if conv.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", conv.Model)
	}
	if len(conv.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(conv.Messages))
	}
}

func TestOpenAIParser_ToolCalls(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "assistant", "content": "Let me check.", "tool_calls": [
				{"id": "call_1", "type": "function", "function": {"name": "get_weather", "arguments": "{\"city\": \"SF\"}"}}
			]},
			{"role": "tool", "tool_call_id": "call_1", "content": "72F sunny"}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(conv.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(conv.Messages))
	}

	// Assistant message should have text + tool_call
	assistantMsg := conv.Messages[0]
	hasText := false
	hasToolUse := false
	for _, block := range assistantMsg.Content {
		if block.Type == ContentText {
			hasText = true
		}
		if block.Type == ContentToolUse {
			hasToolUse = true
			if block.Name != "get_weather" {
				t.Errorf("expected tool name get_weather, got %s", block.Name)
			}
		}
	}
	if !hasText {
		t.Error("expected text in assistant message")
	}
	if !hasToolUse {
		t.Error("expected tool_use in assistant message")
	}

	// Tool result
	toolMsg := conv.Messages[1]
	if toolMsg.Role != RoleTool {
		t.Errorf("expected tool role, got %s", toolMsg.Role)
	}
	if len(toolMsg.Content) == 0 {
		t.Fatal("expected content in tool message")
	}
	if toolMsg.Content[0].Type != ContentToolResult {
		t.Errorf("expected tool_result, got %s", toolMsg.Content[0].Type)
	}
}

func TestOpenAIParser_SystemMessage(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "system", "content": "You are a system"}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Messages[0].Role != RoleSystem {
		t.Errorf("expected system role, got %s", conv.Messages[0].Role)
	}
}

func TestOpenAIParser_CodeInContent(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "assistant", "content": "Here:\n` + "```" + `js\nconsole.log('hi')\n` + "```" + `"}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hasCode := false
	for _, block := range conv.Messages[0].Content {
		if block.Type == ContentCodeBlock {
			hasCode = true
		}
	}
	if !hasCode {
		t.Error("expected code block")
	}
}

func TestOpenAIParser_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid}`)
	p := NewOpenAIParser()
	_, err := p.Parse(data)
	if err == nil {
		t.Error("expected error")
	}
}

func TestOpenAIParser_EmptyMessages(t *testing.T) {
	data := []byte(`{"messages": []}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(conv.Messages))
	}
}

func TestOpenAIParser_NullContent(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "assistant", "content": null, "tool_calls": [
				{"id": "c1", "type": "function", "function": {"name": "test", "arguments": "{}"}}
			]}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(conv.Messages))
	}
	hasToolUse := false
	for _, block := range conv.Messages[0].Content {
		if block.Type == ContentToolUse {
			hasToolUse = true
		}
	}
	if !hasToolUse {
		t.Error("expected tool_use block")
	}
}

func TestOpenAIParser_MultiPartContent(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "user", "content": [
				{"type": "text", "text": "What is this?"}
			]}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(conv.Messages))
	}
}

func TestOpenAIParser_MultipleToolCalls(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "assistant", "content": "Using tools", "tool_calls": [
				{"id": "c1", "type": "function", "function": {"name": "tool_a", "arguments": "{}"}},
				{"id": "c2", "type": "function", "function": {"name": "tool_b", "arguments": "{}"}}
			]}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	toolCalls := 0
	for _, block := range conv.Messages[0].Content {
		if block.Type == ContentToolUse {
			toolCalls++
		}
	}
	if toolCalls != 2 {
		t.Errorf("expected 2 tool calls, got %d", toolCalls)
	}
}

func TestOpenAIParser_ToolNameInToolMessage(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "tool", "tool_call_id": "c1", "name": "my_tool", "content": "result data"}
		]
	}`)
	p := NewOpenAIParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Messages[0].Content[0].Name != "my_tool" {
		t.Errorf("expected tool name my_tool, got %s", conv.Messages[0].Content[0].Name)
	}
}
