package main

import (
	"ccs_interview/game"
	"log"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <mode>")
	}

	mode := os.Args[1]

	switch mode {
	case "server":
		game.StartServer() // Start the server, handling one player for now
	case "client":
		addr := os.Getenv("CODEBREAKER_ADDR")
		if addr == "" {
			addr = "localhost:8080"
		}

		err := game.StartClient(addr)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Invalid mode. Use 'server' or 'client'.")
	}
}
