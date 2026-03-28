package parser

import (
	"testing"
)

func TestDetectFormat_Anthropic(t *testing.T) {
	data := []byte(`{"system": "hello", "messages": []}`)
	got := DetectFormat(data)
	if got != FormatAnthropic {
		t.Errorf("expected anthropic, got %s", got)
	}
}

func TestDetectFormat_OpenAI(t *testing.T) {
	data := []byte(`{"messages": [{"role": "user", "content": "hi"}]}`)
	got := DetectFormat(data)
	if got != FormatOpenAI {
		t.Errorf("expected openai, got %s", got)
	}
}

func TestDetectFormat_JSONL(t *testing.T) {
	data := []byte("{\"role\": \"user\", \"content\": \"hi\"}\n{\"role\": \"assistant\", \"content\": \"hello\"}")
	got := DetectFormat(data)
	if got != FormatJSONL {
		t.Errorf("expected jsonl, got %s", got)
	}
}

func TestDetectFormat_Unknown(t *testing.T) {
	data := []byte(`not json at all`)
	got := DetectFormat(data)
	if got != FormatUnknown {
		t.Errorf("expected unknown, got %s", got)
	}
}

func TestDetectFormat_EmptyJSON(t *testing.T) {
	data := []byte(`{}`)
	got := DetectFormat(data)
	if got != FormatUnknown {
		t.Errorf("expected unknown, got %s", got)
	}
}

func TestDetectFormat_SingleRole(t *testing.T) {
	data := []byte(`{"role": "user", "content": "test"}`)
	got := DetectFormat(data)
	if got != FormatOpenAI {
		t.Errorf("expected openai, got %s", got)
	}
}

func TestDetectFormat_BOM(t *testing.T) {
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"system": "hi", "messages": []}`)...)
	got := DetectFormat(data)
	if got != FormatAnthropic {
		t.Errorf("expected anthropic, got %s", got)
	}
}

func TestParse_AutoDetect_Anthropic(t *testing.T) {
	data := []byte(`{
		"system": "You are helpful",
		"messages": [
			{"role": "user", "content": "Hello"}
		]
	}`)
	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Format != "anthropic" {
		t.Errorf("expected anthropic, got %s", conv.Format)
	}
	if len(conv.Messages) != 2 { // system + user
		t.Errorf("expected 2 messages, got %d", len(conv.Messages))
	}
}

func TestParse_AutoDetect_OpenAI(t *testing.T) {
	data := []byte(`{
		"messages": [
			{"role": "system", "content": "You are helpful"},
			{"role": "user", "content": "Hello"}
		]
	}`)
	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Format != "openai" {
		t.Errorf("expected openai, got %s", conv.Format)
	}
}

func TestParse_AutoDetect_JSONL(t *testing.T) {
	data := []byte("{\"role\": \"user\", \"content\": \"hi\"}\n{\"role\": \"assistant\", \"content\": \"hello\"}")
	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Format != "jsonl" {
		t.Errorf("expected jsonl, got %s", conv.Format)
	}
}

func TestParse_Unknown(t *testing.T) {
	data := []byte("not json")
	_, err := Parse(data)
	if err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestTrimBOM(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"no bom", []byte("hello"), []byte("hello")},
		{"with bom", append([]byte{0xEF, 0xBB, 0xBF}, []byte("hello")...), []byte("hello")},
		{"empty", []byte{}, []byte{}},
		{"bom only", []byte{0xEF, 0xBB, 0xBF}, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimBOM(tt.in)
			if string(got) != string(tt.want) {
				t.Errorf("trimBOM(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantText  string
		wantCode  int
	}{
		{"no code", "just text here", "just text here", 0},
		{"one block", "before ```code here``` after", "before  after", 1},
		{"two blocks", "a ```b``` c ```d``` e", "a  c  e", 2},
		{"empty", "", "", 0},
		{"only code", "```all code```", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonCode, codeBlocks := extractCodeBlocks(tt.text)
			if nonCode != tt.wantText {
				t.Errorf("nonCode = %q, want %q", nonCode, tt.wantText)
			}
			if len(codeBlocks) != tt.wantCode {
				t.Errorf("codeBlocks count = %d, want %d", len(codeBlocks), tt.wantCode)
			}
		})
	}
}

func TestParse_EmptyMessages(t *testing.T) {
	data := []byte(`{"system": "sys", "messages": []}`)
	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have system message only
	if len(conv.Messages) != 1 {
		t.Errorf("expected 1 message (system), got %d", len(conv.Messages))
	}
}

func TestParse_Whitespace(t *testing.T) {
	data := []byte(`  { "system" : "hi" , "messages" : [] }  `)
	conv, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Format != "anthropic" {
		t.Errorf("expected anthropic, got %s", conv.Format)
	}
}
