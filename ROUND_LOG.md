# Round Log

## Round 24 - ctxlens

- **Date:** 2026-03-28
- **Category:** LLM DevTools / Observability
- **Language:** Go
- **Status:** Build Complete

### Concept
CLI tool that decomposes LLM context windows into categories, counts tokens per segment, and estimates costs. Like a flamegraph for your AI context spend.

### Key Metrics
- Tests: 138
- Packages: 6 (parser, tokenizer, analyzer, display, pricing, cmd)
- Formats: 3 (Anthropic, OpenAI, JSONL)
- Output modes: 3 (table, JSON, flame)

### Architecture
- `cmd/ctxlens/` - CLI entry point with cobra
- `internal/parser/` - Format detection + Anthropic/OpenAI/JSONL parsers
- `internal/tokenizer/` - Token counting (4-char estimation)
- `internal/analyzer/` - Decomposition into 6 categories
- `internal/display/` - Table, JSON, flamegraph renderers
- `internal/pricing/` - Provider pricing + YAML config
