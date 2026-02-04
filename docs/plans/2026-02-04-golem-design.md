# Golem - Go语言轻量级AI助手框架设计文档

## 概述

Golem 是一个使用 Go 语言和 Eino 框架重写的轻量级个人AI助手，参考 nanobot 的架构设计。

**设计原则：**
- 参考 Eino 架构但自主实现大部分组件
- 只在 LLM 调用等核心部分使用 Eino
- 保留便捷的渠道扩展性

## 1. 项目结构

```
golem/
├── cmd/
│   └── golem/
│       ├── main.go
│       └── commands/
│           ├── init.go
│           ├── chat.go
│           ├── run.go
│           └── status.go
├── internal/
│   ├── agent/
│   │   ├── loop.go              # Agent主循环
│   │   ├── context.go           # 系统提示词构建
│   │   └── subagent.go          # 子agent管理
│   ├── tools/
│   │   ├── registry.go          # 工具注册表
│   │   ├── filesystem.go        # read_file, write_file, edit_file, list_dir
│   │   ├── shell.go             # exec
│   │   ├── web.go               # web_search, web_fetch
│   │   ├── message.go           # message
│   │   └── spawn.go             # spawn
│   ├── bus/
│   │   ├── events.go            # InboundMessage, OutboundMessage
│   │   └── queue.go             # MessageBus
│   ├── channel/
│   │   ├── channel.go           # Channel接口
│   │   ├── manager.go           # 渠道管理器
│   │   └── telegram/
│   │       └── telegram.go
│   ├── session/
│   │   └── manager.go           # 会话持久化
│   ├── memory/
│   │   └── memory.go            # 记忆系统
│   ├── config/
│   │   └── config.go            # 配置管理
│   └── provider/
│       └── provider.go          # ChatModel封装
├── pkg/
│   └── skills/                  # 内置技能
└── go.mod
```

## 2. 核心数据结构

### 2.1 消息事件

```go
// InboundMessage 从渠道收到的消息
type InboundMessage struct {
    Channel   string            // "telegram", "cli"
    SenderID  string            // 用户标识
    ChatID    string            // 会话标识
    Content   string            // 消息文本
    Timestamp time.Time
    Media     []string          // 媒体文件路径
    Metadata  map[string]any    // 渠道特定数据
}

// OutboundMessage 发送到渠道的消息
type OutboundMessage struct {
    Channel  string
    ChatID   string
    Content  string
    ReplyTo  string
    Media    []string
    Metadata map[string]any
}
```

### 2.2 消息总线

```go
type MessageBus struct {
    inbound  chan *InboundMessage
    outbound chan *OutboundMessage
}
```

## 3. 配置系统

配置路径：`~/.golem/config.json`

```go
type Config struct {
    Agents    AgentsConfig
    Channels  ChannelsConfig
    Providers ProvidersConfig
    Gateway   GatewayConfig
    Tools     ToolsConfig
}

type AgentDefaults struct {
    Workspace         string  // ~/.golem/workspace
    Model             string  // anthropic/claude-sonnet-4-5
    MaxTokens         int     // 8192
    Temperature       float64 // 0.7
    MaxToolIterations int     // 20
}

type TelegramConfig struct {
    Enabled   bool
    Token     string
    AllowFrom []string
}
```

## 4. Provider支持

支持 eino-ext 的所有 ChatModel Provider：

| Provider | 说明 |
|----------|------|
| openrouter | 多模型网关（推荐） |
| claude | Anthropic Claude |
| openai | OpenAI GPT |
| deepseek | DeepSeek |
| gemini | Google Gemini |
| ark | 火山引擎 |
| qianfan | 百度千帆 |
| qwen | 阿里通义千问 |
| ollama | 本地模型 |

优先级：OpenRouter → Claude → OpenAI → DeepSeek → Gemini → Ark → Qianfan → Qwen → Ollama

## 5. Channel接口

```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Send(ctx context.Context, msg *bus.OutboundMessage) error
    IsAllowed(senderID string) bool
}
```

初期实现 Telegram，保留扩展接口支持 Discord、Slack 等。

## 6. Tool系统

使用 Eino 的 InvokableTool 接口：

```go
type InvokableTool interface {
    Info(ctx context.Context) (*schema.ToolInfo, error)
    InvokableRun(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error)
}
```

使用 `utils.InferTool` 从 Go struct 自动推断 JSON Schema。

### 内置工具

| 工具 | 说明 |
|------|------|
| read_file | 读取文件内容 |
| write_file | 写入文件 |
| edit_file | 编辑文件 |
| list_dir | 列出目录 |
| exec | 执行Shell命令 |
| web_search | Brave搜索 |
| web_fetch | 获取网页内容 |
| message | 发送消息到渠道 |
| spawn | 创建子agent |

## 7. Agent Loop

核心处理流程：

1. 从 MessageBus 接收 InboundMessage
2. 加载 Session 历史，构建上下文
3. 循环：调用 LLM → 执行 Tool → 继续，直到无 tool_calls
4. 保存 Session，发送响应到 MessageBus

```go
func (l *Loop) processMessage(ctx context.Context, msg *bus.InboundMessage) (*bus.OutboundMessage, error) {
    sess := l.sessions.GetOrCreate(msg.SessionKey())
    messages := l.context.BuildMessages(sess.GetHistory(), msg.Content, msg.Media)

    for i := 0; i < l.maxIterations; i++ {
        resp, _ := l.model.Generate(ctx, messages)

        if len(resp.ToolCalls) == 0 {
            return &bus.OutboundMessage{Content: resp.Content}, nil
        }

        messages = append(messages, resp)
        for _, tc := range resp.ToolCalls {
            result, _ := l.tools.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
            messages = append(messages, &schema.Message{
                Role: schema.Tool, Content: result, ToolCallID: tc.ID,
            })
        }
    }
}
```

## 8. Context Builder

系统提示词层次：

1. 核心身份（内置）
2. Bootstrap 文件（IDENTITY, SOUL, USER, TOOLS, AGENTS）
3. 长期记忆（MEMORY.md）
4. 最近日记（3天）
5. 技能摘要

## 9. Session与Memory

### Session
- 路径：`~/.golem/sessions/{channel}:{chat_id}.jsonl`
- 格式：JSONL 逐行追加
- 内容：role, content, timestamp

### Memory
- 长期记忆：`~/.golem/workspace/memory/MEMORY.md`
- 日记：`~/.golem/workspace/memory/YYYY-MM-DD.md`

## 10. CLI命令

```bash
golem init                 # 初始化配置向导
golem chat                 # 交互式对话
golem chat "消息内容"       # 单条消息
golem run                  # 启动服务（Telegram + 定时任务）
golem status               # 查看配置和连接状态
```

## 11. 数据流架构图

```
Telegram ──► Channel ──► MessageBus.inbound ──► Agent Loop
                                                    │
                                         ┌──────────┴──────────┐
                                         │                     │
                                    ContextBuilder      ToolRegistry
                                         │                     │
                                         └──────────┬──────────┘
                                                    │
                                                ChatModel
                                                    │
                                              Tool Loop
                                                    │
Telegram ◄── Channel ◄── MessageBus.outbound ◄─────┘
```

## 12. 依赖清单

```go
// go.mod
require (
    github.com/cloudwego/eino v0.x.x
    github.com/cloudwego/eino-ext v0.x.x
    github.com/spf13/cobra v1.x.x
    github.com/spf13/viper v1.x.x
    github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.x.x
)
```

## 13. 文件结构

```
~/.golem/
├── config.json              # 配置文件
├── sessions/                # 会话历史
│   └── telegram:123.jsonl
└── workspace/               # 工作空间
    ├── IDENTITY.md
    ├── SOUL.md
    ├── USER.md
    ├── TOOLS.md
    ├── AGENTS.md
    ├── memory/
    │   ├── MEMORY.md
    │   └── 2024-01-15.md
    └── skills/
```
