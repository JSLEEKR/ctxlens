package pricing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil pricer")
	}
}

func TestCalculateCost_Anthropic(t *testing.T) {
	p := New()
	cost := p.CalculateCost("anthropic", "claude-sonnet-4", 1000000, true)
	// $3.0 per 1M input tokens
	if cost != 3.0 {
		t.Errorf("expected $3.0, got $%.2f", cost)
	}
}

func TestCalculateCost_AnthropicOutput(t *testing.T) {
	p := New()
	cost := p.CalculateCost("anthropic", "claude-sonnet-4", 1000000, false)
	// $15.0 per 1M output tokens
	if cost != 15.0 {
		t.Errorf("expected $15.0, got $%.2f", cost)
	}
}

func TestCalculateCost_OpenAI(t *testing.T) {
	p := New()
	cost := p.CalculateCost("openai", "gpt-4o", 1000000, true)
	// $2.5 per 1M input tokens
	if cost != 2.5 {
		t.Errorf("expected $2.5, got $%.2f", cost)
	}
}

func TestCalculateCost_SmallTokens(t *testing.T) {
	p := New()
	cost := p.CalculateCost("anthropic", "claude-sonnet-4", 1000, true)
	// $3.0 * 1000 / 1M = $0.003
	expected := 0.003
	if cost != expected {
		t.Errorf("expected $%.4f, got $%.4f", expected, cost)
	}
}

func TestCalculateCost_UnknownProvider(t *testing.T) {
	p := New()
	cost := p.CalculateCost("unknown", "model", 1000, true)
	if cost != 0 {
		t.Errorf("expected $0 for unknown provider, got $%.2f", cost)
	}
}

func TestCalculateCost_UnknownModel(t *testing.T) {
	p := New()
	cost := p.CalculateCost("anthropic", "unknown-model", 1000, true)
	if cost != 0 {
		t.Errorf("expected $0 for unknown model, got $%.2f", cost)
	}
}

func TestCalculateCost_ZeroTokens(t *testing.T) {
	p := New()
	cost := p.CalculateCost("anthropic", "claude-sonnet-4", 0, true)
	if cost != 0 {
		t.Errorf("expected $0 for 0 tokens, got $%.2f", cost)
	}
}

func TestCalculateCost_PartialMatch(t *testing.T) {
	p := New()
	// "claude-sonnet" should partially match "claude-sonnet-4"
	cost := p.CalculateCost("anthropic", "claude-sonnet", 1000000, true)
	if cost == 0 {
		t.Error("expected non-zero cost for partial model match")
	}
}

func TestCalculateCost_Opus(t *testing.T) {
	p := New()
	cost := p.CalculateCost("anthropic", "claude-opus-4", 1000000, true)
	if cost != 15.0 {
		t.Errorf("expected $15.0 for opus, got $%.2f", cost)
	}
}

func TestGetProviders(t *testing.T) {
	p := New()
	providers := p.GetProviders()
	if len(providers) < 2 {
		t.Errorf("expected at least 2 providers, got %d", len(providers))
	}

	providerMap := make(map[string]bool)
	for _, prov := range providers {
		providerMap[prov] = true
	}
	if !providerMap["anthropic"] {
		t.Error("expected anthropic in providers")
	}
	if !providerMap["openai"] {
		t.Error("expected openai in providers")
	}
}

func TestGetModels(t *testing.T) {
	p := New()
	models := p.GetModels("anthropic")
	if models == nil {
		t.Fatal("expected non-nil models for anthropic")
	}
	if _, ok := models["claude-sonnet-4"]; !ok {
		t.Error("expected claude-sonnet-4 in anthropic models")
	}
}

func TestGetModels_Unknown(t *testing.T) {
	p := New()
	models := p.GetModels("unknown")
	if models != nil {
		t.Error("expected nil for unknown provider")
	}
}

func TestNewWithConfig(t *testing.T) {
	cfg := &Config{
		Providers: map[string]map[string]ModelPricing{
			"custom": {
				"custom-model": {Input: 5.0, Output: 25.0},
			},
		},
	}
	p := NewWithConfig(cfg)

	// Custom model should work
	cost := p.CalculateCost("custom", "custom-model", 1000000, true)
	if cost != 5.0 {
		t.Errorf("expected $5.0 for custom model, got $%.2f", cost)
	}

	// Default models should still work
	cost = p.CalculateCost("anthropic", "claude-sonnet-4", 1000000, true)
	if cost != 3.0 {
		t.Errorf("expected $3.0 for default model, got $%.2f", cost)
	}
}

func TestNewWithConfig_Override(t *testing.T) {
	cfg := &Config{
		Providers: map[string]map[string]ModelPricing{
			"anthropic": {
				"claude-sonnet-4": {Input: 10.0, Output: 50.0},
			},
		},
	}
	p := NewWithConfig(cfg)

	cost := p.CalculateCost("anthropic", "claude-sonnet-4", 1000000, true)
	if cost != 10.0 {
		t.Errorf("expected $10.0 for overridden model, got $%.2f", cost)
	}
}

func TestNewWithConfig_Nil(t *testing.T) {
	p := NewWithConfig(nil)
	cost := p.CalculateCost("anthropic", "claude-sonnet-4", 1000000, true)
	if cost != 3.0 {
		t.Errorf("expected $3.0 for default, got $%.2f", cost)
	}
}

func TestLoadConfig_NonExistent(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("expected no error for missing file, got %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config for missing file")
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := `providers:
  test:
    test-model:
      input: 1.5
      output: 7.5
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Providers["test"]["test-model"].Input != 1.5 {
		t.Errorf("expected input 1.5, got %f", cfg.Providers["test"]["test-model"].Input)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfig_EmptyPath(t *testing.T) {
	// Empty path should use default (which likely doesn't exist)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// cfg may be nil (no default config file) - that's OK
	_ = cfg
}

func TestDefaultConfigDir(t *testing.T) {
	dir := DefaultConfigDir()
	if dir == "" {
		t.Skip("could not determine home directory")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("expected absolute path, got %s", dir)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	if path == "" {
		t.Skip("could not determine home directory")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		want     bool
	}{
		{"hello world", "world", true},
		{"hello", "xyz", false},
		{"", "test", false},
		{"test", "", false},
		{"ab", "abc", false},
	}
	for _, tt := range tests {
		got := contains(tt.haystack, tt.needle)
		if got != tt.want {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.haystack, tt.needle, got, tt.want)
		}
	}
}

func TestCalculateCost_Google(t *testing.T) {
	p := New()
	cost := p.CalculateCost("google", "gemini-2.0-flash", 1000000, true)
	if cost != 0.10 {
		t.Errorf("expected $0.10 for gemini flash, got $%.2f", cost)
	}
}
