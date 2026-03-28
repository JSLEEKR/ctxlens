# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-28

### Added

- Initial release of ctxlens
- `analyze` command: decompose LLM conversation payloads into token categories
- `top` command: rank conversations by token usage or cost across a directory
- `providers` command: list supported providers and pricing
- `version` command: display current version
- Anthropic format parser (system + messages with content blocks)
- OpenAI format parser (messages with tool_calls)
- JSONL format parser (Claude Code conversation logs)
- Auto-format detection across all supported formats
- Token counting via 4-chars-per-token estimation (labeled as estimate)
- ASCII bar chart visualization in table output
- JSON output format for scripting and automation
- ASCII flamegraph-style output for visual profiling
- Cost estimation with embedded default pricing for Anthropic, OpenAI, Google
- User-configurable pricing via ~/.ctxlens/config.yaml
- Stdin pipe support: `cat file.json | ctxlens analyze -`
- Code block extraction and separate categorization
- Top segments ranking within analysis results
- Context utilization percentage (vs model limit)
- 50MB input size limit for safety
- UTF-8 BOM handling
- 138 tests across all packages
