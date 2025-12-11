package main

import (
	"fmt"
	"os"

	netpkg "code_breaker/internal/net"
)

func main() {
	addr := os.Getenv("CODEBREAKER_ADDR")
	if addr == "" {
		addr = "localhost:8080"
	}
	if err := netpkg.StartClient(addr); err != nil {
		fmt.Println("client error:", err)
	}
}
