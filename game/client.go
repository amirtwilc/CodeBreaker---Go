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

	// ✅ Channel for server messages
	serverMsgCh := make(chan Message)

	// ✅ Channel for user input
	inputCh := make(chan string)

	// ✅ SERVER READER GOROUTINE (never blocks main loop)
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

	// ✅ STDIN READER GOROUTINE (never blocks main loop)
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

	// ✅ SINGLE EVENT LOOP — handles BOTH server + user input
	for {
		select {

		// ✅ SERVER MESSAGE
		case msg, ok := <-serverMsgCh:
			if !ok {
				return fmt.Errorf("server disconnected")
			}

			// Always print server text
			fmt.Print(msg.Text)

			switch msg.Type {
			case "TURN":
				isMyTurn = true
				fmt.Print("Your guess: ")

			case "RECOVERY":
				// ✅ SPECIAL MODE: allow ANY player to type
				isMyTurn = true
				fmt.Print("Recovery guess allowed: ")

			case "WAIT", "TIMEOUT", "RESULT", "WIN", "NEWGAME", "INFO":
				isMyTurn = false
			}

		// ✅ USER INPUT
		case guess, ok := <-inputCh:
			if !ok {
				return nil
			}

			if !isMyTurn {
				// ❌ Input when it's NOT your turn → safely ignore
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
