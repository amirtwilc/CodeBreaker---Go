package main

import (
	"code_breaker/internal/net"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

func main() {
	players := netpkg.MaxPlayers // default

	if len(os.Args) >= 2 {
		n, err := strconv.Atoi(os.Args[1])
		if err == nil && n > 0 {
			players = n
		}
	}

	fmt.Printf("Starting CodeBreaker with %d players...\n", players)

	openTerminal("go run ./cmd/server")

	time.Sleep(2 * time.Second)

	for i := 1; i <= players; i++ {
		openTerminal("go run ./cmd/client")
	}
}

func openTerminal(command string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		cmd = exec.Command(
			"cmd",
			"/C",
			"start",
			"",
			"cmd",
			"/K",
			command,
		)

	case "darwin": // macOS (was not tested)
		cmd = exec.Command(
			"osascript",
			"-e",
			fmt.Sprintf(`tell app "Terminal" to do script "%s"`, command),
		)

	default: // Linux (was not tested)
		cmd = exec.Command(
			"x-terminal-emulator",
			"-e",
			"bash",
			"-c",
			command,
		)
	}

	_ = cmd.Start()
}
