[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-138-blue?style=for-the-badge)](https://github.com/JSLEEKR/ctxlens)
[![Version](https://img.shields.io/badge/Version-0.1.0-orange?style=for-the-badge)](CHANGELOG.md)

# ctxlens

**LLM context window profiler** -- decompose, count tokens, estimate costs.

Like a flamegraph for your AI context spend. See exactly where your tokens go: system prompts, user messages, tool calls, code blocks, or tool results.

---

## Why This Exists

LLM API costs scale with token count. Anthropic charges 2x above 200K tokens. OpenAI charges per-token. Developers using AI coding tools (Claude Code, Cursor, Copilot) have **zero visibility** into what fills their context window.

Is it system prompts? Pasted code? Conversation history? Tool results? Without knowing what consumes tokens, optimization is guesswork.

**ctxlens** reads LLM API request/response payloads (JSON), tokenizes each section, and produces a visual breakdown showing token allocation by category. No more guessing.

### The Problem

```
You: "Why is my API bill so high?"
Also you: *has no idea 40% of context is tool results*
```

### The Solution

```bash
$ ctxlens analyze conversation.json

Context Breakdown -- conversation.json
=====================================

Total: 45,230 tokens (estimated) (22.6% of 200K limit)
Est. Cost: $0.135 (anthropic claude-sonnet-4 @ input pricing)

Category           Tokens      %      Cost   Bar
------------------  --------  ------  --------   --------------------
System Prompt          8,200  18.1%  $ 0.025   ||||||||............
User Messages          5,400  11.9%  $ 0.016   ||||||..............
Assistant Msgs         7,800  17.2%  $ 0.023   |||||||||...........
Tool Calls             1,200   2.7%  $ 0.004   |...................
Tool Results          18,430  40.8%  $ 0.055   ||||||||||||||||||||.
Code Blocks            4,200   9.3%  $ 0.013   |||||...............

Top Segments:
  1. tool_result (web_search)              6,200 tokens
  2. system_prompt (main)                  5,100 tokens
  3. tool_result (read_file)               4,800 tokens
  4. user_message #3                       3,200 tokens
  5. assistant_message #5                  2,900 tokens
```

Now you know: **tool results eat 40% of your context**. Time to trim those search results.

---

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/JSLEEKR/ctxlens.git
cd ctxlens

# Build
go build -o ctxlens ./cmd/ctxlens/

# Install to $GOPATH/bin
go install ./cmd/ctxlens/
```

### Requirements

- Go 1.22 or later
- No external runtime dependencies (pure Go, no CGO)

---

## Quick Start

### Analyze a Conversation File

```bash
# Anthropic format
ctxlens analyze conversation.json

# OpenAI format
ctxlens analyze openai-payload.json

# Claude Code JSONL logs
ctxlens analyze ~/.claude/projects/myproject/conversations/session.jsonl
```

### Pipe from stdin

```bash
# Pipe any JSON conversation
cat conversation.json | ctxlens analyze -

# Pipe from curl
curl -s https://api.example.com/conversation | ctxlens analyze -

# Pipe from clipboard (macOS)
pbpaste | ctxlens analyze -
```

### Different Output Formats

```bash
# Default: table with ASCII bars
ctxlens analyze conversation.json

# JSON for scripting
ctxlens analyze conversation.json --format json

# ASCII flamegraph
ctxlens analyze conversation.json --format flame
```

### Override Provider/Model for Cost Estimation

```bash
# Use OpenAI pricing
ctxlens analyze conversation.json --provider openai --model gpt-4o

# Use Anthropic Opus pricing
ctxlens analyze conversation.json --provider anthropic --model claude-opus-4
```

---

## CLI Reference

### `ctxlens analyze <file>`

Decompose a conversation payload into categories with token counts and cost estimates.

```
Usage:
  ctxlens analyze [file] [flags]

Arguments:
  file    Path to conversation file (or - for stdin)

Flags:
  -f, --format string     Output format: table, json, flame (default "table")
  -p, --provider string   Override provider (anthropic, openai, google)
  -m, --model string      Override model name
  -l, --limit int         Context window limit in tokens (default 200000)
```

**Supported Input Formats:**

| Format | Detection | Example |
|--------|-----------|---------|
| Anthropic | `{"system": "...", "messages": [...]}` | Claude API requests |
| OpenAI | `{"messages": [{"role": "system", ...}]}` | GPT API requests |
| JSONL | One JSON object per line | Claude Code logs |

Format is auto-detected. No flags needed.

### `ctxlens top <dir>`

Scan a directory of conversation files and rank by token usage or cost.

```
Usage:
  ctxlens top <dir> [flags]

Flags:
  --by string   Sort by: tokens or cost (default "tokens")
  -n int        Number of results to show (default 10)
```

**Example:**

```bash
$ ctxlens top ~/.claude/projects/myproject/conversations/ --by cost -n 5

Top 5 Conversations by cost
========================================

Rank  File                            Tokens       Cost
----  -----                           ------       ----
1     session-2026-03-28.jsonl        145,230  $    0.436
2     session-2026-03-27.jsonl         98,400  $    0.295
3     session-2026-03-26.jsonl         67,800  $    0.203
4     session-2026-03-25.jsonl         45,200  $    0.136
5     session-2026-03-24.jsonl         23,100  $    0.069
```

### `ctxlens providers`

List all supported providers and their pricing.

```bash
$ ctxlens providers

Supported Providers and Pricing
================================

  anthropic:
    claude-sonnet-4           Input: $3.00/1M  Output: $15.00/1M
    claude-opus-4             Input: $15.00/1M  Output: $75.00/1M
    claude-haiku-3.5          Input: $0.80/1M  Output: $4.00/1M

  openai:
    gpt-4o                    Input: $2.50/1M  Output: $10.00/1M
    gpt-4o-mini               Input: $0.15/1M  Output: $0.60/1M
    gpt-4-turbo               Input: $10.00/1M  Output: $30.00/1M
    o1                        Input: $15.00/1M  Output: $60.00/1M
    o1-mini                   Input: $3.00/1M  Output: $12.00/1M

  google:
    gemini-2.0-flash          Input: $0.10/1M  Output: $0.40/1M
    gemini-2.0-pro            Input: $1.25/1M  Output: $10.00/1M
```

### `ctxlens version`

```bash
$ ctxlens version
ctxlens v0.1.0
```

---

## Output Formats

### Table (Default)

Human-readable table with ASCII bar charts showing relative token usage per category.

```bash
ctxlens analyze conversation.json --format table
```

### JSON

Machine-readable JSON for scripting, dashboards, or piping to `jq`.

```bash
ctxlens analyze conversation.json --format json | jq '.categories[] | select(.percent > 20)'
```

**Example JSON output:**

```json
{
  "source": "conversation.json",
  "total_tokens": 45230,
  "model_limit": 200000,
  "utilization_pct": 22.615,
  "estimated_cost_usd": 0.135,
  "provider": "anthropic",
  "model": "claude-sonnet-4",
  "is_estimate": true,
  "categories": [
    {"name": "System Prompt", "tokens": 8200, "percent": 18.1, "cost_usd": 0.025},
    {"name": "User Messages", "tokens": 5400, "percent": 11.9, "cost_usd": 0.016},
    {"name": "Tool Results", "tokens": 18430, "percent": 40.8, "cost_usd": 0.055}
  ],
  "top_segments": [
    {"label": "tool_result (web_search)", "category": "Tool Results", "tokens": 6200},
    {"label": "system_prompt (main)", "category": "System Prompt", "tokens": 5100}
  ]
}
```

### Flame

ASCII flamegraph-style output showing segments sorted by size.

```bash
ctxlens analyze conversation.json --format flame
```

```
Context Flamegraph -- conversation.json
Total: 45,230 tokens

tool_result (web_search)            6,200 (13.7%) ||||||||||||||||
system_prompt (main)                5,100 (11.3%) ||||||||||||||
tool_result (read_file)             4,800 (10.6%) |||||||||||||
user_message #3                     3,200  (7.1%) |||||||||
assistant_message #5                2,900  (6.4%) ||||||||
...
```

---

## Configuration

### Custom Pricing

Create `~/.ctxlens/config.yaml` to override default pricing or add custom providers:

```yaml
providers:
  anthropic:
    claude-sonnet-4:
      input: 3.0      # USD per 1M input tokens
      output: 15.0     # USD per 1M output tokens
    claude-opus-4:
      input: 15.0
      output: 75.0
  openai:
    gpt-4o:
      input: 2.5
      output: 10.0
  # Add custom providers
  my-provider:
    my-model:
      input: 5.0
      output: 20.0
```

Custom config is merged with defaults -- you only need to specify overrides.

---

## Architecture

```
cmd/ctxlens/main.go          -- CLI entry point (cobra)
internal/parser/              -- Format detection + message parsing
  parser.go                   -- Interface + auto-detection
  anthropic.go                -- Anthropic format parser
  openai.go                   -- OpenAI format parser
  jsonl.go                    -- JSONL (Claude Code) parser
internal/tokenizer/           -- Token counting
  tokenizer.go                -- 4-chars-per-token estimator
internal/analyzer/            -- Decomposition logic
  analyzer.go                 -- Categorize + count per segment
  segment.go                  -- Segment types and data structures
internal/display/             -- Output formatting
  table.go                    -- Table output with ASCII bars
  flame.go                    -- ASCII flamegraph
  json.go                     -- JSON output
internal/pricing/             -- Cost estimation
  pricing.go                  -- Provider pricing + calculation
  config.go                   -- YAML config file loading
```

### Token Categories

| Category | Description | Typical % |
|----------|-------------|-----------|
| System Prompt | System instructions, persona, rules | 10-25% |
| User Messages | User input text | 10-20% |
| Assistant Msgs | Model response text | 15-30% |
| Tool Calls | Function/tool invocations | 2-5% |
| Tool Results | Tool/function return values | 20-50% |
| Code Blocks | Code within triple backticks | 5-15% |

### Token Counting

ctxlens uses a 4-characters-per-token estimation heuristic. This is a widely-used approximation for cl100k_base encoding (GPT-4, Claude). Results are labeled as **estimated** in the output.

The estimator is conservative -- it tends to slightly overcount, which is better for cost planning than undercounting.

---

## Supported Input Formats

### Anthropic Format

```json
{
  "system": "You are a helpful assistant.",
  "messages": [
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": [
      {"type": "text", "text": "Hi there!"},
      {"type": "tool_use", "id": "t1", "name": "search", "input": {"q": "test"}}
    ]},
    {"role": "user", "content": [
      {"type": "tool_result", "tool_use_id": "t1", "content": "Results here"}
    ]}
  ]
}
```

### OpenAI Format

```json
{
  "messages": [
    {"role": "system", "content": "You are helpful."},
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi!", "tool_calls": [
      {"id": "call_1", "type": "function", "function": {"name": "search", "arguments": "{}"}}
    ]},
    {"role": "tool", "tool_call_id": "call_1", "content": "Results"}
  ]
}
```

### JSONL (Claude Code Logs)

```jsonl
{"role": "system", "content": "You are helpful.", "model": "claude-sonnet-4"}
{"role": "user", "content": "Hello"}
{"role": "assistant", "content": "Hi there!"}
```

---

## Security

- **No network calls** -- all processing is local
- **No secrets in output** -- conversation content is summarized, not echoed
- **Config path restricted** -- only reads from `~/.ctxlens/`
- **Input size limit** -- 50MB per file to prevent OOM

---

## Use Cases

### 1. Cost Optimization

Find which parts of your conversation consume the most tokens and cost the most money.

```bash
# Which conversations cost the most?
ctxlens top ./conversations/ --by cost

# Drill into the expensive one
ctxlens analyze ./conversations/expensive-session.json
```

### 2. System Prompt Auditing

Check if your system prompt is too long and eating into your context budget.

```bash
ctxlens analyze api-request.json | grep "System Prompt"
# System Prompt       12,500   38.2%  $0.038   ||||||||||||||||||||
```

### 3. Tool Result Trimming

Discover that tool results (file reads, search results) dominate your context.

```bash
ctxlens analyze session.json --format json | jq '.categories[] | select(.name == "Tool Results")'
```

### 4. CI/CD Token Budgeting

Add token budget checks to your CI pipeline.

```bash
TOKENS=$(ctxlens analyze prompt.json --format json | jq '.total_tokens')
if [ "$TOKENS" -gt 100000 ]; then
  echo "WARNING: Prompt exceeds 100K token budget ($TOKENS tokens)"
  exit 1
fi
```

### 5. Comparing Conversation Strategies

Compare token usage between different prompt strategies.

```bash
ctxlens analyze strategy-a.json --format json > a.json
ctxlens analyze strategy-b.json --format json > b.json
diff <(jq '.categories' a.json) <(jq '.categories' b.json)
```

---

## Development

### Build

```bash
go build -o ctxlens ./cmd/ctxlens/
```

### Test

```bash
go test ./... -v
```

### Lint

```bash
go vet ./...
```

---

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [cobra](https://github.com/spf13/cobra) for CLI framework
- [yaml.v3](https://github.com/go-yaml/yaml) for configuration parsing
- Inspired by the need to understand where AI coding tool context budgets go
