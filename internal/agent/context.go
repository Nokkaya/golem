package agent

import (
    "github.com/cloudwego/eino/schema"
    "github.com/MEKXH/golem/internal/session"
)

// ContextBuilder builds LLM context
// Minimal stub for loop compilation; expanded in Task 13.
type ContextBuilder struct {
    workspacePath string
}

// NewContextBuilder creates a context builder
func NewContextBuilder(workspacePath string) *ContextBuilder {
    return &ContextBuilder{workspacePath: workspacePath}
}

// BuildSystemPrompt assembles the system prompt
func (c *ContextBuilder) BuildSystemPrompt() string {
    return ""
}

// BuildMessages constructs the full message list
func (c *ContextBuilder) BuildMessages(history []*session.Message, current string, media []string) []*schema.Message {
    return []*schema.Message{{
        Role:    schema.User,
        Content: current,
    }}
}
