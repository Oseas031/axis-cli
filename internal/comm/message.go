package comm

import "time"

// MessageType classifies the intent of a message.
type MessageType string

const (
	MsgTask     MessageType = "task"     // task assignment
	MsgResult   MessageType = "result"   // task completion
	MsgQuery    MessageType = "query"    // information request
	MsgNotify   MessageType = "notify"   // one-way notification
	MsgDelegate MessageType = "delegate" // forward to another actor
	MsgYield    MessageType = "yield"    // actor yields execution
)

// Message is the universal communication unit between Actors.
type Message struct {
	ID        string         `json:"id"`
	From      string         `json:"from"`
	To        string         `json:"to"`
	Type      MessageType    `json:"type"`
	Payload   map[string]any `json:"payload"`
	Timestamp time.Time      `json:"timestamp"`
	ReplyTo   string         `json:"reply_to,omitempty"`
}
