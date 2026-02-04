package channel

import (
    "context"
    "testing"

    "github.com/MEKXH/golem/internal/bus"
)

type mockChannel struct {
    BaseChannel
    name string
}

func (m *mockChannel) Name() string { return m.name }
func (m *mockChannel) Start(ctx context.Context) error { return nil }
func (m *mockChannel) Stop(ctx context.Context) error  { return nil }
func (m *mockChannel) Send(ctx context.Context, msg *bus.OutboundMessage) error {
    return nil
}

func TestBaseChannel_IsAllowed(t *testing.T) {
    msgBus := bus.NewMessageBus(1)
    ch := &mockChannel{
        BaseChannel: BaseChannel{Bus: msgBus, AllowList: map[string]bool{"u1": true}},
        name:        "mock",
    }

    if ch.IsAllowed("u1") != true {
        t.Fatalf("expected u1 allowed")
    }
    if ch.IsAllowed("u2") != false {
        t.Fatalf("expected u2 denied")
    }
}
