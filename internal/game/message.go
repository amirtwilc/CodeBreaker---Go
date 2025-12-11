package game

// MessageType represents the type of message sent to clients.
type MessageType string

const (
	TURN     MessageType = "TURN"
	WAIT     MessageType = "WAIT"
	INFO     MessageType = "INFO"
	RESULT   MessageType = "RESULT"
	WIN      MessageType = "WIN"
	NEWGAME  MessageType = "NEWGAME"
	TIMEOUT  MessageType = "TIMEOUT"
	RECOVERY MessageType = "RECOVERY"
)

// Message is the JSON-serializable message sent to clients.
type Message struct {
	Type MessageType `json:"type"`
	Text string      `json:"text"`
}
