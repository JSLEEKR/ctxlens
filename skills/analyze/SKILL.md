---
name: analyze
description: Analyze LLM conversation token usage. Use when investigating token costs, context window utilization, or optimizing prompt sizes.
---

# Analyze Context Window

Profile token usage in LLM conversation payloads.

## Usage
```bash
# Analyze a conversation file
ctxlens analyze conversation.json

# Pipe from stdin
cat payload.json | ctxlens analyze -

# JSON output
ctxlens analyze conversation.json --format json

# Find biggest consumers
ctxlens top ./conversations/
```

Shows breakdown by: system prompt, user messages, assistant messages, tool calls, tool results, code blocks.
