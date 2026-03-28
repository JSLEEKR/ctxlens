package parser

import (
	"testing"
)

func TestAnthropicParser_BasicConversation(t *testing.T) {
	data := []byte(`{
		"model": "claude-sonnet-4",
		"system": "You are helpful",
		"messages": [
			{"role": "user", "content": "Hello"},
			{"role": "assistant", "content": "Hi there!"}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conv.Format != "anthropic" {
		t.Errorf("expected format anthropic, got %s", conv.Format)
	}
	if conv.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", conv.Provider)
	}
	if conv.Model != "claude-sonnet-4" {
		t.Errorf("expected model claude-sonnet-4, got %s", conv.Model)
	}
	// 1 system + 2 messages = 3
	if len(conv.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(conv.Messages))
	}
	if conv.Messages[0].Role != RoleSystem {
		t.Errorf("first message should be system, got %s", conv.Messages[0].Role)
	}
}

func TestAnthropicParser_SystemAsArray(t *testing.T) {
	data := []byte(`{
		"system": [{"type": "text", "text": "System prompt here"}],
		"messages": [
			{"role": "user", "content": "Hello"}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(conv.Messages))
	}
	if conv.Messages[0].Role != RoleSystem {
		t.Errorf("expected system role, got %s", conv.Messages[0].Role)
	}
	if len(conv.Messages[0].Content) == 0 {
		t.Error("system message should have content")
	}
}

func TestAnthropicParser_ToolUseAndResult(t *testing.T) {
	data := []byte(`{
		"system": "You can use tools",
		"messages": [
			{"role": "user", "content": "Search for info"},
			{"role": "assistant", "content": [
				{"type": "text", "text": "Let me search."},
				{"type": "tool_use", "id": "t1", "name": "web_search", "input": {"query": "test"}}
			]},
			{"role": "user", "content": [
				{"type": "tool_result", "tool_use_id": "t1", "content": "Search results here"}
			]}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// system + user + assistant + user(tool_result) = 4
	if len(conv.Messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(conv.Messages))
	}

	// Check assistant message has text + tool_use
	assistantMsg := conv.Messages[2]
	if len(assistantMsg.Content) != 2 {
		t.Fatalf("expected 2 content blocks in assistant message, got %d", len(assistantMsg.Content))
	}
	if assistantMsg.Content[0].Type != ContentText {
		t.Errorf("expected text content, got %s", assistantMsg.Content[0].Type)
	}
	if assistantMsg.Content[1].Type != ContentToolUse {
		t.Errorf("expected tool_use content, got %s", assistantMsg.Content[1].Type)
	}
	if assistantMsg.Content[1].Name != "web_search" {
		t.Errorf("expected tool name web_search, got %s", assistantMsg.Content[1].Name)
	}

	// Check tool result
	toolResultMsg := conv.Messages[3]
	if len(toolResultMsg.Content) != 1 {
		t.Fatalf("expected 1 content block in tool result, got %d", len(toolResultMsg.Content))
	}
	if toolResultMsg.Content[0].Type != ContentToolResult {
		t.Errorf("expected tool_result content, got %s", toolResultMsg.Content[0].Type)
	}
}

func TestAnthropicParser_CodeBlocks(t *testing.T) {
	data := []byte(`{
		"system": "You write code",
		"messages": [
			{"role": "assistant", "content": "Here is code:\n` + "```" + `python\nprint('hello')\n` + "```" + `"}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assistantMsg := conv.Messages[1]
	hasCode := false
	for _, block := range assistantMsg.Content {
		if block.Type == ContentCodeBlock {
			hasCode = true
		}
	}
	if !hasCode {
		t.Error("expected code block in assistant message")
	}
}

func TestAnthropicParser_EmptySystem(t *testing.T) {
	data := []byte(`{
		"system": "",
		"messages": [
			{"role": "user", "content": "Hello"}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty system should not create a system message
	if len(conv.Messages) != 1 {
		t.Errorf("expected 1 message (no empty system), got %d", len(conv.Messages))
	}
}

func TestAnthropicParser_InvalidJSON(t *testing.T) {
	data := []byte(`not valid json`)
	p := NewAnthropicParser()
	_, err := p.Parse(data)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAnthropicParser_NoMessages(t *testing.T) {
	data := []byte(`{"system": "hi"}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 1 { // just system
		t.Errorf("expected 1 message, got %d", len(conv.Messages))
	}
}

func TestAnthropicParser_ContentStringMessage(t *testing.T) {
	data := []byte(`{
		"system": "sys",
		"messages": [
			{"role": "user", "content": "simple string message"}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	userMsg := conv.Messages[1]
	if len(userMsg.Content) == 0 {
		t.Fatal("expected content blocks")
	}
	if userMsg.Content[0].Type != ContentText {
		t.Errorf("expected text content, got %s", userMsg.Content[0].Type)
	}
	if userMsg.Content[0].Text != "simple string message" {
		t.Errorf("expected 'simple string message', got %q", userMsg.Content[0].Text)
	}
}

func TestAnthropicParser_MultipleToolCalls(t *testing.T) {
	data := []byte(`{
		"system": "tools",
		"messages": [
			{"role": "assistant", "content": [
				{"type": "tool_use", "id": "t1", "name": "read_file", "input": {"path": "a.txt"}},
				{"type": "tool_use", "id": "t2", "name": "write_file", "input": {"path": "b.txt"}}
			]}
		]
	}`)
	p := NewAnthropicParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msg := conv.Messages[1]
	toolCalls := 0
	for _, block := range msg.Content {
		if block.Type == ContentToolUse {
			toolCalls++
		}
	}
	if toolCalls != 2 {
		t.Errorf("expected 2 tool calls, got %d", toolCalls)
	}
}
