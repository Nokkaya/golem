package channel

import (
    "context"

    "github.com/MEKXH/golem/internal/bus"
)

// Channel interface for chat platforms
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Send(ctx context.Context, msg *bus.OutboundMessage) error
    IsAllowed(senderID string) bool
}

// BaseChannel provides common functionality
type BaseChannel struct {
    Bus       *bus.MessageBus
    AllowList map[string]bool
}

// IsAllowed checks if sender is permitted
func (b *BaseChannel) IsAllowed(senderID string) bool {
    if len(b.AllowList) == 0 {
        return true
    }
    return b.AllowList[senderID]
}

// PublishInbound sends message to bus
func (b *BaseChannel) PublishInbound(msg *bus.InboundMessage) {
    b.Bus.PublishInbound(msg)
}
