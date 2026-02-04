package bus

import "testing"

func TestInboundMessage_SessionKey(t *testing.T) {
    msg := &InboundMessage{
        Channel: "telegram",
        ChatID:  "12345",
    }

    expected := "telegram:12345"
    if got := msg.SessionKey(); got != expected {
        t.Errorf("SessionKey() = %q, want %q", got, expected)
    }
}
