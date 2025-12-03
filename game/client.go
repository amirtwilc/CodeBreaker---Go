package game

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func StartClient(address string) error {
	// Connect to the server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to Code Breaker server. Waiting for game updates...")

	reader := bufio.NewReader(os.Stdin)

	for {
		// Read message from server
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return fmt.Errorf("error reading from server: %v", err)
		}

		serverMsg := string(buffer[:n])
		fmt.Print(serverMsg) // print whatever the server sent

		lines := strings.Split(serverMsg, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}

			switch fields[0] {

			case "TURN":
				fmt.Print("Your turn. Enter your guess: ")

				guess, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %v", err)
				}
				guess = strings.TrimSpace(guess)

				if guess == "exit" {
					fmt.Println("Exiting the game.")
					return nil
				}

				_, err = conn.Write([]byte(guess))
				if err != nil {
					return fmt.Errorf("error sending guess: %v", err)
				}

			case "WAIT":
				// Do nothing, just wait silently

			case "RESULT":
				// Already printed above

			case "TIMEOUT":
				// Already printed above

			case "WIN":
				fmt.Println("Game won! Waiting for restart...")

			case "NEWGAME":
				fmt.Println("New game starting...")

			}
		}
	}
}
