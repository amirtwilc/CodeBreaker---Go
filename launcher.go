package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

const (
	MaxPlayers      = 2
	CodeLength      = "5"
	Difficulty      = "easy"
	TurnTimeSeconds = "10"
)

func main() {
	os.Setenv("MAX_PLAYERS", strconv.Itoa(MaxPlayers))
	os.Setenv("CODE_LENGTH", CodeLength)
	os.Setenv("Difficulty", Difficulty)
	os.Setenv("TURN_TIME_SECONDS", TurnTimeSeconds)

	fmt.Printf("Starting CodeBreaker with %d players...\n", MaxPlayers)

	openTerminal("go run ./cmd/server")

	time.Sleep(2 * time.Second)

	for i := 1; i <= MaxPlayers; i++ {
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
