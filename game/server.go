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
	TurnTimeSeconds = 15 // time limit per turn (seconds)
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
	players := acceptPlayers(listener)

	analytics := &Analytics{
		WinsByPlayer:    make(map[int]int),
		LossesByPlayer:  make(map[int]int),
		GuessesUntilWin: make(map[int]int),
	}

	broadcast(players, INFO, "All players connected. Game starting now!\n")

	currentTurn := gameRng.Intn(len(players))
	consecutiveTimeouts := 0

	for { // outer game loop
		// Generate secret
		SetCodeDigits(CodeLength)
		secret := GenerateSecretCodeWithDifficulty(CodeLength, difficulty)
		log.Printf("DEBUG NEW SECRET: %d\n", secret)

		currentGameGuesses := 0
		broadcast(players, NEWGAME, "New game started!\n")

		for { // inner round loop
			currentPlayer := players[currentTurn]
			notifyTurns(players, currentPlayer)

			numGuess, timeoutOccurred, disconnected := readPlayerGuess(currentPlayer)
			if disconnected {
				log.Printf("Player %d disconnected\n", currentPlayer.id)
				return
			}

			if timeoutOccurred {
				broadcast(players, TIMEOUT,
					fmt.Sprintf("Player %d ran out of time and forfeited the turn!\n", currentPlayer.id),
				)
				consecutiveTimeouts++

				if consecutiveTimeouts >= len(players) {
					handleRecovery(players, &currentTurn, &consecutiveTimeouts, secret, &currentGameGuesses, analytics, gameRng)
					break // exit inner loop → start new game
				}

				currentTurn = (currentTurn + 1) % len(players)
				drainLateInput(currentPlayer.conn)
				continue
			}

			consecutiveTimeouts = 0 // valid guess breaks timeout chain

			feedback := GenerateFeedback(secret, numGuess, gameRng)
			currentGameGuesses++

			if feedback.CorrectPlace == CodeLength {
				handleWin(players, &currentTurn, currentPlayer, secret, currentGameGuesses, analytics)
				break
			}

			printGuessResult(players, currentPlayer.id, numGuess, feedback)
			currentTurn = (currentTurn + 1) % len(players)
		}
	}
}

// -------------------- Helper Functions --------------------

func acceptPlayers(listener net.Listener) []*Player {
	players := []*Player{}
	for len(players) < MaxPlayers {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}
		player := &Player{id: len(players) + 1, conn: conn}
		players = append(players, player)

		log.Printf("Player %d connected\n", player.id)
		writeToClient(conn, INFO, fmt.Sprintf("Welcome Player %d! Waiting for others...\n", player.id))
	}
	return players
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

func readPlayerGuess(player *Player) (numGuess int, timeout bool, disconnected bool) {
	_ = player.conn.SetReadDeadline(time.Now().Add(time.Second * TurnTimeSeconds))
	buffer := make([]byte, 1024)
	n, err := player.conn.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0, true, false
		}
		return 0, false, true
	}
	_ = player.conn.SetReadDeadline(time.Time{})
	guess := strings.TrimSpace(string(buffer[:n]))
	numGuess, err = ValidateGuess(guess)
	if err != nil {
		writeToClient(player.conn, INFO, "Invalid input: "+err.Error()+"\n")
		return readPlayerGuess(player)
	}
	return numGuess, false, false
}

func handleRecovery(players []*Player, currentTurn *int, consecutiveTimeouts *int, secret int, currentGameGuesses *int, analytics *Analytics, rng *rand.Rand) {
	broadcast(players, RECOVERY, "All players timed out. Waiting for ANY player to resume...\n")

	type recoveryInput struct {
		player *Player
		data   string
	}

	var resume recoveryInput
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
			guess := strings.TrimSpace(string(buf[:n]))
			resume = recoveryInput{player: p, data: guess}
			found = true
			break
		}
	}

	numGuess, err := ValidateGuess(resume.data)
	if err != nil {
		writeToClient(resume.player.conn, INFO, "Invalid input: "+err.Error()+"\n")
		*consecutiveTimeouts = 0
		*currentTurn = resume.player.id - 1
		return
	}

	feedback := GenerateFeedback(secret, numGuess, rng)
	*currentGameGuesses++

	if feedback.CorrectPlace == CodeLength {
		handleWin(players, currentTurn, resume.player, secret, *currentGameGuesses, analytics)
		*consecutiveTimeouts = 0
		*currentTurn = resume.player.id % len(players)
		return
	}

	printGuessResult(players, resume.player.id, numGuess, feedback)
	*consecutiveTimeouts = 0
	*currentTurn = resume.player.id % len(players)
}

func handleWin(players []*Player, currentTurn *int, winner *Player, secret int, guesses int, analytics *Analytics) {
	winMsg := fmt.Sprintf("Player %d won! Secret was %d\n", winner.id, secret)
	analytics.GamesPlayed++
	analytics.WinsByPlayer[winner.id]++
	analytics.GuessesUntilWin[secret] = guesses

	for _, p := range players {
		if p.id != winner.id {
			analytics.LossesByPlayer[p.id]++
		}
	}

	prefix := GenerateTimestampPrefix()
	broadcast(players, WIN, prefix+winMsg)
	broadcast(players, NEWGAME, "New game starting in 3 seconds...\n")
	printAnalytics(analytics)
	time.Sleep(3 * time.Second)
}

func printGuessResult(players []*Player, playerID int, guess int, feedback Feedback) {
	msg := fmt.Sprintf(
		ColorBlue+"player: %d\n"+ColorCyan+"Number guessed: %d\n"+
			ColorGreen+"Correctly placed: %d\n"+ColorYellow+"Wrongly placed: %d\n"+
			ColorPurple+"Hint: %s\n"+ColorReset,
		playerID,
		guess,
		feedback.CorrectPlace,
		feedback.WrongPlace,
		feedback.Hint,
	)
	prefix := GenerateTimestampPrefix()
	broadcast(players, RESULT, prefix+msg)
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

func broadcast(players []*Player, msgType MessageType, text string) {
	for _, p := range players {
		writeToClient(p.conn, msgType, text)
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

	type hardEntry struct{ secret, guesses int }
	hardList := []hardEntry{}
	for secret, guesses := range a.GuessesUntilWin {
		hardList = append(hardList, hardEntry{secret, guesses})
	}

	sort.Slice(hardList, func(i, j int) bool { return hardList[i].guesses > hardList[j].guesses })

	limit := 5
	if len(hardList) < limit {
		limit = len(hardList)
	}
	if limit > 0 {
		log.Println("Top hardest secrets (by guesses until win):")
		for i := 0; i < limit; i++ {
			log.Printf("#%d Secret: %d | Guesses until win: %d\n",
				i+1, hardList[i].secret, hardList[i].guesses)
		}
	} else {
		log.Println("No completed games yet to determine hardest secrets.")
	}
	log.Println("============================")
}
