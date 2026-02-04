# Golem

使用 Go 与 Eino 构建的轻量级个人 AI 助手。

[English](README.md)

## 安装

```bash
go install github.com/MEKXH/golem/cmd/golem@latest
```

## 快速开始

```bash
# 初始化配置
golem init

# 编辑配置，填写 API Key
# ~/.golem/config.json

# 交互式聊天
golem chat

# 发送单条消息
golem chat "2+2 等于多少？"

# 启动服务（含 Telegram）
golem run
```

## 配置

配置文件：`~/.golem/config.json`

### 工作区

用于控制会话、上下文与 exec 工具的默认工作目录。

`agents.defaults.workspace_mode`：
- `default`（默认）：始终使用 `~/.golem/workspace`
- `cwd`：使用启动 golem 时的当前终端目录
- `path`：使用 `agents.defaults.workspace` 指定的路径（必填）

示例：

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

### 模型提供商

支持：OpenRouter、Claude、OpenAI、DeepSeek、Gemini、Ollama 等。

### 渠道

- Telegram（已实现）
- 更多渠道持续更新

## 许可证

MIT
