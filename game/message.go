package game

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

type Message struct {
	Type MessageType `json:"type"` // TURN, RECOVERY, WAIT, RESULT, WIN, TIMEOUT, NEWGAME, INFO
	Text string      `json:"text"` // Human readable message
}
