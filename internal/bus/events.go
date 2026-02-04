package bus

import "time"

// InboundMessage received from a channel
type InboundMessage struct {
    Channel   string
    SenderID  string
    ChatID    string
    Content   string
    Timestamp time.Time
    Media     []string
    Metadata  map[string]any
}

// SessionKey returns unique session identifier
func (m *InboundMessage) SessionKey() string {
    return m.Channel + ":" + m.ChatID
}

// OutboundMessage to send to a channel
type OutboundMessage struct {
    Channel  string
    ChatID   string
    Content  string
    ReplyTo  string
    Media    []string
    Metadata map[string]any
}
