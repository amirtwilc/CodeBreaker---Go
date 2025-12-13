package netpkg // Package netpkg to void conflict with standard net package

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"code_breaker/internal/game"
)

// StartClient connects to server and runs the client loop
func StartClient(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to server: %w", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Code Breaker server. Waiting for game updates...")

	decoder := json.NewDecoder(conn)

	serverMsgCh := make(chan game.Message)
	inputCh := make(chan string)

	// Server reader
	go func() {
		for {
			var msg game.Message
			if err := decoder.Decode(&msg); err != nil {
				close(serverMsgCh)
				return
			}
			serverMsgCh <- msg
		}
	}()

	// Stdin reader
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				close(inputCh)
				return
			}
			inputCh <- strings.TrimSpace(line)
		}
	}()

	isMyTurn := false
	lastPrinted := game.Message{}

	for {
		select {
		case msg, ok := <-serverMsgCh:
			if !ok {
				return fmt.Errorf("server disconnected")
			}
			// Always print server text
			if msg.Type != lastPrinted.Type || msg.Text != lastPrinted.Text {
				fmt.Print(msg.Text)
				lastPrinted = msg
			}

			switch msg.Type {
			case game.TURN:
				isMyTurn = true
				fmt.Print("Your guess: ")
			case game.RECOVERY:
				isMyTurn = true
				fmt.Print("Recovery guess allowed: ")
			case game.WAIT, game.TIMEOUT, game.RESULT, game.WIN, game.NEWGAME, game.INFO:
				isMyTurn = false
			}

		case guess, ok := <-inputCh:
			if !ok {
				return nil
			}
			if !isMyTurn {
				// ignore when not turn
				continue
			}
			isMyTurn = false
			if guess == "exit" {
				fmt.Println("Exiting game...")
				return nil
			}
			_, err = conn.Write([]byte(guess + "\n"))
			if err != nil {
				return err
			}
		}
	}
}
