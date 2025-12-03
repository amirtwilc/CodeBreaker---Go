package game

type Message struct {
	Type string `json:"type"` // TURN, WAIT, RESULT, WIN, TIMEOUT, NEWGAME, INFO
	Text string `json:"text"` // Human readable message
}
