// Package pricing provides cost estimation for LLM API usage.
package pricing

// ModelPricing contains per-million-token pricing for a model.
type ModelPricing struct {
	Input  float64 `yaml:"input"`  // USD per 1M input tokens
	Output float64 `yaml:"output"` // USD per 1M output tokens
}

// Pricer calculates costs based on provider pricing.
type Pricer struct {
	providers map[string]map[string]ModelPricing
}

// New creates a new Pricer with default pricing.
func New() *Pricer {
	return &Pricer{
		providers: defaultPricing(),
	}
}

// NewWithConfig creates a Pricer with custom config merged over defaults.
func NewWithConfig(cfg *Config) *Pricer {
	p := New()
	if cfg != nil {
		for provider, models := range cfg.Providers {
			if _, ok := p.providers[provider]; !ok {
				p.providers[provider] = make(map[string]ModelPricing)
			}
			for model, pricing := range models {
				p.providers[provider][model] = pricing
			}
		}
	}
	return p
}

// CalculateCost returns the estimated cost in USD.
func (p *Pricer) CalculateCost(provider, model string, tokens int, isInput bool) float64 {
	models, ok := p.providers[provider]
	if !ok {
		return 0
	}
	pricing, ok := models[model]
	if !ok {
		// Try partial match
		pricing, ok = p.findPartialMatch(provider, model)
		if !ok {
			return 0
		}
	}
	rate := pricing.Input
	if !isInput {
		rate = pricing.Output
	}
	return float64(tokens) * rate / 1_000_000
}

// findPartialMatch tries to find a model by partial name match.
func (p *Pricer) findPartialMatch(provider, model string) (ModelPricing, bool) {
	models, ok := p.providers[provider]
	if !ok {
		return ModelPricing{}, false
	}
	for name, pricing := range models {
		if contains(model, name) || contains(name, model) {
			return pricing, true
		}
	}
	return ModelPricing{}, false
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && searchString(haystack, needle)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetProviders returns all known provider names.
func (p *Pricer) GetProviders() []string {
	var providers []string
	for name := range p.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetModels returns all models for a provider.
func (p *Pricer) GetModels(provider string) map[string]ModelPricing {
	if models, ok := p.providers[provider]; ok {
		return models
	}
	return nil
}

func defaultPricing() map[string]map[string]ModelPricing {
	return map[string]map[string]ModelPricing{
		"anthropic": {
			"claude-sonnet-4": {Input: 3.0, Output: 15.0},
			"claude-opus-4":   {Input: 15.0, Output: 75.0},
			"claude-haiku-3.5":  {Input: 0.80, Output: 4.0},
		},
		"openai": {
			"gpt-4o":      {Input: 2.5, Output: 10.0},
			"gpt-4o-mini": {Input: 0.15, Output: 0.60},
			"gpt-4-turbo": {Input: 10.0, Output: 30.0},
			"o1":          {Input: 15.0, Output: 60.0},
			"o1-mini":     {Input: 3.0, Output: 12.0},
		},
		"google": {
			"gemini-2.0-flash": {Input: 0.10, Output: 0.40},
			"gemini-2.0-pro":   {Input: 1.25, Output: 10.0},
		},
	}
}
