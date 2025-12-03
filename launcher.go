package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

func main() {
	players := 2 // default

	if len(os.Args) >= 2 {
		n, err := strconv.Atoi(os.Args[1])
		if err == nil && n > 0 {
			players = n
		}
	}

	fmt.Printf("Starting CodeBreaker with %d players...\n", players)

	openTerminal("go run main.go server")

	time.Sleep(2 * time.Second)

	for i := 1; i <= players; i++ {
		openTerminal("go run main.go client")
	}
}

func openTerminal(command string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		// ✅ IMPORTANT FIX:
		// "" is an EMPTY WINDOW TITLE so the command actually runs
		cmd = exec.Command(
			"cmd",
			"/C",
			"start",
			"",
			"cmd",
			"/K",
			command,
		)

	case "darwin": // macOS
		cmd = exec.Command(
			"osascript",
			"-e",
			fmt.Sprintf(`tell app "Terminal" to do script "%s"`, command),
		)

	default: // Linux
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
