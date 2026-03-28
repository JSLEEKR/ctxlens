package parser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	dir := testdataDir()
	if dir == "" {
		t.Skip("could not determine testdata directory")
	}
	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading testdata %s: %v", name, err)
	}
	return data
}

func TestIntegration_AnthropicFile(t *testing.T) {
	data := readTestdata(t, "anthropic.json")

	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	if conv.Format != "anthropic" {
		t.Errorf("expected anthropic format, got %s", conv.Format)
	}
	if conv.Model != "claude-sonnet-4" {
		t.Errorf("expected model claude-sonnet-4, got %s", conv.Model)
	}
	if len(conv.Messages) < 3 {
		t.Errorf("expected at least 3 messages, got %d", len(conv.Messages))
	}

	// Check for tool_use
	hasToolUse := false
	hasToolResult := false
	for _, msg := range conv.Messages {
		for _, block := range msg.Content {
			if block.Type == ContentToolUse {
				hasToolUse = true
			}
			if block.Type == ContentToolResult {
				hasToolResult = true
			}
		}
	}
	if !hasToolUse {
		t.Error("expected tool_use in anthropic test data")
	}
	if !hasToolResult {
		t.Error("expected tool_result in anthropic test data")
	}
}

func TestIntegration_OpenAIFile(t *testing.T) {
	data := readTestdata(t, "openai.json")

	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	if conv.Format != "openai" {
		t.Errorf("expected openai format, got %s", conv.Format)
	}
	if conv.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", conv.Model)
	}
	if len(conv.Messages) < 3 {
		t.Errorf("expected at least 3 messages, got %d", len(conv.Messages))
	}
}

func TestIntegration_JSONLFile(t *testing.T) {
	data := readTestdata(t, "conversation.jsonl")

	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	if conv.Format != "jsonl" {
		t.Errorf("expected jsonl format, got %s", conv.Format)
	}
	if len(conv.Messages) < 3 {
		t.Errorf("expected at least 3 messages, got %d", len(conv.Messages))
	}
}
