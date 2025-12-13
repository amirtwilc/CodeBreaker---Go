package game

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	MaxPlayers      int
	CodeLength      int
	Difficulty      Difficulty
	TurnTimeSeconds int
}

func LoadConfig() Config {
	return Config{
		MaxPlayers:      envInt("MAX_PLAYERS", 2),
		CodeLength:      envInt("CODE_LENGTH", 4),
		Difficulty:      envDifficulty("DIFFICULTY", DifficultyMedium),
		TurnTimeSeconds: envInt("TURN_TIME_SECONDS", 30),
	}
}

// Helpers
func envInt(name string, defaultVal int) int {
	val := os.Getenv(name)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("Invalid %s=%s, using default %d", name, val, defaultVal)
		return defaultVal
	}
	return n
}

func envDifficulty(name string, defaultVal Difficulty) Difficulty {
	val := os.Getenv(name)
	switch val {
	case "easy", "EASY":
		return DifficultyEasy
	case "medium", "MEDIUM":
		return DifficultyMedium
	case "hard", "HARD":
		return DifficultyHard
	}
	return defaultVal
}
