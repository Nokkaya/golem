package channel

import (
    "context"
    "testing"
    "time"

    "github.com/MEKXH/golem/internal/bus"
)

type mockManagerChannel struct {
    BaseChannel
    name    string
    sent    int
    started bool
    stopped bool
}

func (m *mockManagerChannel) Name() string { return m.name }
func (m *mockManagerChannel) Start(ctx context.Context) error { m.started = true; return nil }
func (m *mockManagerChannel) Stop(ctx context.Context) error  { m.stopped = true; return nil }
func (m *mockManagerChannel) Send(ctx context.Context, msg *bus.OutboundMessage) error {
    m.sent++
    return nil
}

func TestManager_RouteOutbound(t *testing.T) {
    msgBus := bus.NewMessageBus(1)
    mgr := NewManager(msgBus)

    ch := &mockManagerChannel{name: "test", BaseChannel: BaseChannel{Bus: msgBus}}
    mgr.Register(ch)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go mgr.RouteOutbound(ctx)

    msgBus.PublishOutbound(&bus.OutboundMessage{Channel: "test", ChatID: "1", Content: "hi"})

    <-time.After(10 * time.Millisecond)

    if ch.sent == 0 {
        t.Fatalf("expected message sent")
    }
}
