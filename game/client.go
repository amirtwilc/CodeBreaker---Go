package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

func StartClient(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Code Breaker server. Waiting for game updates...")

	decoder := json.NewDecoder(conn)

	// Channel for server messages
	serverMsgCh := make(chan Message)

	// Channel for user input
	inputCh := make(chan string)

	// Server reader goroutine (never blocks main loop)
	go func() {
		for {
			var msg Message
			if err := decoder.Decode(&msg); err != nil {
				close(serverMsgCh)
				return
			}
			serverMsgCh <- msg
		}
	}()

	// Stdin reader goroutine (never blocks main loop)
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
	lastPrinted := Message{}

	// Single event loop — handles BOTH server + user input
	for {
		select {

		// Server message
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
			case "TURN":
				isMyTurn = true
				fmt.Print("Your guess: ")

			case "RECOVERY":
				// Allow ANY player to type
				isMyTurn = true
				fmt.Print("Recovery guess allowed: ")

			case "WAIT", "TIMEOUT", "RESULT", "WIN", "NEWGAME", "INFO":
				isMyTurn = false
			}

		// User input
		case guess, ok := <-inputCh:
			if !ok {
				return nil
			}

			if !isMyTurn {
				// Input when it's NOT your turn → safely ignore
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
