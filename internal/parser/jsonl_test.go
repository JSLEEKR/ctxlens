package parser

import (
	"testing"
)

func TestJSONLParser_BasicConversation(t *testing.T) {
	data := []byte(`{"role": "user", "content": "Hello"}
{"role": "assistant", "content": "Hi there!"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conv.Format != "jsonl" {
		t.Errorf("expected jsonl format, got %s", conv.Format)
	}
	if len(conv.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(conv.Messages))
	}
	if conv.Messages[0].Role != RoleUser {
		t.Errorf("expected user role, got %s", conv.Messages[0].Role)
	}
	if conv.Messages[1].Role != RoleAssistant {
		t.Errorf("expected assistant role, got %s", conv.Messages[1].Role)
	}
}

func TestJSONLParser_WithModel(t *testing.T) {
	data := []byte(`{"role": "system", "content": "You are helpful", "model": "claude-sonnet-4"}
{"role": "user", "content": "Hello"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conv.Model != "claude-sonnet-4" {
		t.Errorf("expected model claude-sonnet-4, got %s", conv.Model)
	}
	if conv.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", conv.Provider)
	}
}

func TestJSONLParser_EmptyLines(t *testing.T) {
	data := []byte(`{"role": "user", "content": "Hello"}

{"role": "assistant", "content": "Hi"}

`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(conv.Messages))
	}
}

func TestJSONLParser_InvalidLines(t *testing.T) {
	data := []byte(`{"role": "user", "content": "Hello"}
not valid json
{"role": "assistant", "content": "Hi"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 2 {
		t.Errorf("expected 2 messages (skipping invalid), got %d", len(conv.Messages))
	}
}

func TestJSONLParser_AllInvalid(t *testing.T) {
	data := []byte(`not json
also not json`)

	p := NewJSONLParser()
	_, err := p.Parse(data)
	if err == nil {
		t.Error("expected error for all-invalid JSONL")
	}
}

func TestJSONLParser_ToolMessages(t *testing.T) {
	data := []byte(`{"role": "user", "content": "Search for it"}
{"role": "assistant", "content": [{"type": "tool_use", "id": "t1", "name": "search", "input": {"q": "test"}}]}
{"role": "tool", "content": "search results here"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conv.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(conv.Messages))
	}

	// Check tool result message
	toolMsg := conv.Messages[2]
	if toolMsg.Role != RoleTool {
		t.Errorf("expected tool role, got %s", toolMsg.Role)
	}
	if len(toolMsg.Content) == 0 {
		t.Fatal("expected content in tool message")
	}
	if toolMsg.Content[0].Type != ContentToolResult {
		t.Errorf("expected tool_result type, got %s", toolMsg.Content[0].Type)
	}
}

func TestJSONLParser_SystemMessage(t *testing.T) {
	data := []byte(`{"role": "system", "content": "You are helpful"}
{"role": "user", "content": "Hi"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Messages[0].Role != RoleSystem {
		t.Errorf("expected system role, got %s", conv.Messages[0].Role)
	}
}

func TestJSONLParser_GPTModel(t *testing.T) {
	data := []byte(`{"role": "user", "content": "Hi", "model": "gpt-4o"}
{"role": "assistant", "content": "Hello"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Provider != "openai" {
		t.Errorf("expected openai provider, got %s", conv.Provider)
	}
}

func TestJSONLParser_EmptyContent(t *testing.T) {
	data := []byte(`{"role": "user", "content": ""}
{"role": "assistant", "content": "Hello"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty content message should be skipped
	if len(conv.Messages) != 1 {
		t.Errorf("expected 1 message (skip empty), got %d", len(conv.Messages))
	}
}

func TestJSONLParser_CodeBlocks(t *testing.T) {
	data := []byte(`{"role": "assistant", "content": "Code:\n` + "```" + `go\nfmt.Println(\"hi\")\n` + "```" + `"}
{"role": "user", "content": "Thanks"}`)

	p := NewJSONLParser()
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
		t.Error("expected code block in assistant message")
	}
}

func TestJSONLParser_TypeField(t *testing.T) {
	data := []byte(`{"type": "system", "content": "System prompt"}
{"role": "user", "content": "Hello"}`)

	p := NewJSONLParser()
	conv, err := p.Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Messages[0].Role != RoleSystem {
		t.Errorf("expected system role from type field, got %s", conv.Messages[0].Role)
	}
}
