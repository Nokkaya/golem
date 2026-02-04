package channel

import (
    "context"
    "log/slog"
    "sync"

    "github.com/MEKXH/golem/internal/bus"
)

// Manager coordinates all channels
type Manager struct {
    channels map[string]Channel
    bus      *bus.MessageBus
    mu       sync.RWMutex
}

// NewManager creates a channel manager
func NewManager(msgBus *bus.MessageBus) *Manager {
    return &Manager{
        channels: make(map[string]Channel),
        bus:      msgBus,
    }
}

// Register adds a channel
func (m *Manager) Register(ch Channel) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.channels[ch.Name()] = ch
}

// Names returns registered channel names
func (m *Manager) Names() []string {
    m.mu.RLock()
    defer m.mu.RUnlock()

    names := make([]string, 0, len(m.channels))
    for name := range m.channels {
        names = append(names, name)
    }
    return names
}

// StartAll starts all channels
func (m *Manager) StartAll(ctx context.Context) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for name, ch := range m.channels {
        go func(n string, c Channel) {
            slog.Info("starting channel", "name", n)
            if err := c.Start(ctx); err != nil {
                slog.Error("channel error", "name", n, "error", err)
            }
        }(name, ch)
    }
}

// RouteOutbound sends outbound messages to appropriate channels
func (m *Manager) RouteOutbound(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-m.bus.Outbound():
            m.mu.RLock()
            if ch, ok := m.channels[msg.Channel]; ok {
                go ch.Send(ctx, msg)
            }
            m.mu.RUnlock()
        }
    }
}

// StopAll stops all channels
func (m *Manager) StopAll(ctx context.Context) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for _, ch := range m.channels {
        _ = ch.Stop(ctx)
    }
}
