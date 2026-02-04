# Golem

Lightweight personal AI assistant built with Go and Eino.

[中文文档](README.zh-CN.md)

## Installation

```bash
go install github.com/MEKXH/golem/cmd/golem@latest
```

## Quick Start

```bash
# Initialize configuration
golem init

# Edit config to add API key
# ~/.golem/config.json

# Chat interactively
golem chat

# Send single message
golem chat "What is 2+2?"

# Start server with Telegram
golem run
```

## Configuration

Config file: `~/.golem/config.json`

### Workspace

Controls where sessions, context, and the exec tool operate.

`agents.defaults.workspace_mode`:
- `default` (default): always use `~/.golem/workspace`
- `cwd`: use the current working directory when launching golem
- `path`: use the explicit `agents.defaults.workspace` value (required)

Example:

```json
{
  "agents": {
    "defaults": {
      "workspace_mode": "path",
      "workspace": "D:/Work/my-project"
    }
  }
}
```

### Providers

Supports: OpenRouter, Claude, OpenAI, DeepSeek, Gemini, Ollama, and more.

### Channels

- Telegram (implemented)
- More coming soon

## License

MIT
