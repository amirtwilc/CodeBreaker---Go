package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sort"
	"strings"
	"time"
)

const (
	MaxPlayers      = 3
	CodeLength      = 4
	difficulty      = DifficultyHard
	TurnTimeSeconds = 15
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorPurple = "\033[35m"
)

type Player struct {
	conn net.Conn
	id   int
}

type Analytics struct {
	GamesPlayed     int
	WinsByPlayer    map[int]int
	LossesByPlayer  map[int]int
	GuessesUntilWin map[int]int
}

func StartServer() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server started. Waiting for %d players...\n", MaxPlayers)

	gameRng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var players []*Player

	analytics := &Analytics{
		WinsByPlayer:    make(map[int]int),
		LossesByPlayer:  make(map[int]int),
		GuessesUntilWin: make(map[int]int),
	}

	// Accept players
	for len(players) < MaxPlayers {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		player := &Player{
			id:   len(players) + 1,
			conn: conn,
		}
		players = append(players, player)

		log.Printf("Player %d connected\n", player.id)
		writeToClient(conn, INFO, fmt.Sprintf("Welcome Player %d! Waiting for others...\n", player.id))
	}

	broadcast(players, INFO, "All players connected. Game starting now!\n")

	currentTurn := gameRng.Intn(len(players))
	consecutiveTimeouts := 0

	for { // Game loop: infinite rounds
		SetCodeDigits(CodeLength)
		secret := GenerateSecretCodeWithDifficulty(CodeLength, difficulty)
		log.Printf("DEBUG NEW SECRET: %d\n", secret)
		currentGameGuesses := 0

		broadcast(players, NEWGAME, "New game started!\n")

		for { // Single round loop
			currentPlayer := players[currentTurn]
			notifyTurns(players, currentPlayer)

			_ = currentPlayer.conn.SetReadDeadline(time.Now().Add(time.Second * TurnTimeSeconds))
			buffer := make([]byte, 1024)
			n, err := currentPlayer.conn.Read(buffer)

			// Timeout case
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				broadcast(players, TIMEOUT, fmt.Sprintf("Player %d ran out of time and forfeited the turn!\n", currentPlayer.id))
				consecutiveTimeouts++

				// All players timed out → Recovery Mode
				if consecutiveTimeouts >= len(players) {
					handleRecovery(players, secret, &currentTurn, &consecutiveTimeouts, analytics, gameRng, &currentGameGuesses)
					break
				}

				// Normal timeout → move to next player
				currentTurn = (currentTurn + 1) % len(players)
				drainLateInput(currentPlayer.conn)
				continue
			}

			if err != nil {
				log.Printf("Player %d disconnected: %v\n", currentPlayer.id, err)
				return
			}

			_ = currentPlayer.conn.SetReadDeadline(time.Time{})

			guess := strings.TrimSpace(string(buffer[:n]))
			numGuess, err := ValidateGuess(guess)
			if err != nil {
				writeToClient(currentPlayer.conn, INFO, "Invalid input: "+err.Error()+"\n")
				continue
			}

			// Valid guess
			consecutiveTimeouts = 0
			currentGameGuesses++

			feedback := GenerateFeedback(secret, numGuess, gameRng)
			if feedback.CorrectPlace == CodeLength {
				handleWin(players, currentPlayer, secret, analytics, currentGameGuesses)
				time.Sleep(3 * time.Second)
				currentTurn = (currentTurn + 1) % len(players)
				break
			}

			// Normal feedback
			msg := fmt.Sprintf(
				ColorBlue+"player: %d\n"+ColorCyan+"Number guessed: %d\n"+ColorGreen+"Correctly placed: %d\n"+ColorYellow+"Wrongly placed: %d\n"+ColorPurple+"Hint: %s\n"+ColorReset,
				currentPlayer.id, numGuess, feedback.CorrectPlace, feedback.WrongPlace, feedback.Hint,
			)
			broadcast(players, RESULT, GenerateTimestampPrefix()+msg)
			currentTurn = (currentTurn + 1) % len(players)
		}
	}
}

func handleRecovery(players []*Player, secret int, currentTurn *int, consecutiveTimeouts *int, analytics *Analytics, gameRng *rand.Rand, currentGameGuesses *int) {
	broadcast(players, RECOVERY, "All players timed out. Waiting for ANY player to resume...\n")

	var resumePlayer *Player
	var guess string
	found := false

	for !found {
		for _, p := range players {
			_ = p.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			buf := make([]byte, 1024)
			n, err := p.conn.Read(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Error during recovery read from player %d: %v", p.id, err)
				continue
			}

			_ = p.conn.SetReadDeadline(time.Time{})
			resumePlayer = p
			guess = strings.TrimSpace(string(buf[:n]))
			found = true
			break
		}
	}

	numGuess, err := ValidateGuess(guess)
	if err != nil {
		writeToClient(resumePlayer.conn, INFO, "Invalid input: "+err.Error()+"\n")
		*consecutiveTimeouts = 0
		*currentTurn = resumePlayer.id - 1
		return
	}

	*currentGameGuesses++
	feedback := GenerateFeedback(secret, numGuess, gameRng)
	if feedback.CorrectPlace == CodeLength {
		handleWin(players, resumePlayer, secret, analytics, *currentGameGuesses)
		time.Sleep(3 * time.Second)
		*currentTurn = resumePlayer.id % len(players)
		*consecutiveTimeouts = 0
		return
	}

	// Normal feedback
	msg := fmt.Sprintf(
		ColorBlue+"player: %d\n"+ColorCyan+"Number guessed: %d\n"+ColorGreen+"Correctly placed: %d\n"+ColorYellow+"Wrongly placed: %d\n"+ColorPurple+"Hint: %s\n"+ColorReset,
		resumePlayer.id, numGuess, feedback.CorrectPlace, feedback.WrongPlace, feedback.Hint,
	)
	broadcast(players, RESULT, GenerateTimestampPrefix()+msg)
	*consecutiveTimeouts = 0
	*currentTurn = resumePlayer.id % len(players)
}

func handleWin(players []*Player, winner *Player, secret int, analytics *Analytics, currentGameGuesses int) {
	winMsg := fmt.Sprintf("Player %d won! Secret was %d\n", winner.id, secret)

	analytics.GamesPlayed++
	analytics.WinsByPlayer[winner.id]++
	analytics.GuessesUntilWin[secret] = currentGameGuesses

	for _, p := range players {
		if p.id != winner.id {
			analytics.LossesByPlayer[p.id]++
		}
	}

	broadcast(players, WIN, GenerateTimestampPrefix()+winMsg)
	broadcast(players, NEWGAME, "New game starting in 3 seconds...\n")
	printAnalytics(analytics)
}

func notifyTurns(players []*Player, currentPlayer *Player) {
	for _, p := range players {
		if p.id == currentPlayer.id {
			writeToClient(p.conn, TURN, "Your turn!\n")
		} else {
			writeToClient(p.conn, WAIT, fmt.Sprintf("Waiting for Player %d...\n", currentPlayer.id))
		}
	}
}

func drainLateInput(conn net.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}
	}
	_ = conn.SetReadDeadline(time.Time{})
}

func broadcast(players []*Player, msgType MessageType, msg string) {
	for _, p := range players {
		writeToClient(p.conn, msgType, msg)
	}
}

func writeToClient(conn net.Conn, msgType MessageType, text string) {
	msg := Message{Type: msgType, Text: text}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	_, err = conn.Write(append(data, '\n'))
	if err != nil {
		log.Printf("Error writing to client: %v", err)
	}
}

func printAnalytics(a *Analytics) {
	log.Println("====== GAME ANALYTICS ======")
	log.Printf("Games Played: %d\n", a.GamesPlayed)
	for pid, wins := range a.WinsByPlayer {
		log.Printf("Player %d Wins: %d\n", pid, wins)
	}
	for pid, losses := range a.LossesByPlayer {
		log.Printf("Player %d Losses: %d\n", pid, losses)
	}

	type hardEntry struct {
		secret  int
		guesses int
	}
	hardList := make([]hardEntry, 0)
	for secret, guesses := range a.GuessesUntilWin {
		hardList = append(hardList, hardEntry{secret: secret, guesses: guesses})
	}

	sort.Slice(hardList, func(i, j int) bool { return hardList[i].guesses > hardList[j].guesses })
	limit := 5
	if len(hardList) < limit {
		limit = len(hardList)
	}

	if limit > 0 {
		log.Println("Top hardest secrets (by guesses until win):")
		for i := 0; i < limit; i++ {
			log.Printf("#%d Secret: %d | Guesses until win: %d\n", i+1, hardList[i].secret, hardList[i].guesses)
		}
	} else {
		log.Println("No completed games yet to determine hardest secrets.")
	}
	log.Println("============================")
}
