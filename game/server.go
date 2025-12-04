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
	MaxPlayers      = 2
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
		writeToClient(conn, "INFO",
			fmt.Sprintf("Welcome Player %d! Waiting for others...\n", player.id))
	}

	broadcast(players, "INFO", "All players connected. Game starting now!\n")

	currentTurn := 0
	consecutiveTimeouts := 0

	for { // GAME LOOP (infinite rounds)

		SetCodeDigits(CodeLength)
		secret := GenerateSecretCodeWithDifficulty(CodeLength, difficulty)
		log.Printf("DEBUG NEW SECRET: %d\n", secret)
		currentGameGuesses := 0

		broadcast(players, "NEWGAME", "New game started!\n")

		for { // single round loop
			currentPlayer := players[currentTurn]

			// Notify turns
			for _, p := range players {
				if p.id == currentPlayer.id {
					writeToClient(p.conn, "TURN",
						"Your turn!\n")
				} else {
					writeToClient(p.conn, "WAIT",
						fmt.Sprintf("Waiting for Player %d...\n", currentPlayer.id))
				}
			}

			// Apply turn deadline
			_ = currentPlayer.conn.SetReadDeadline(
				time.Now().Add(time.Second * TurnTimeSeconds),
			)

			buffer := make([]byte, 1024)
			n, err := currentPlayer.conn.Read(buffer)

			// ✅ TIMEOUT CASE
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				msg := fmt.Sprintf(
					"Player %d ran out of time and forfeited the turn!\n",
					currentPlayer.id,
				)
				broadcast(players, "TIMEOUT", msg)

				consecutiveTimeouts++

				// ✅ GLOBAL RECOVERY MODE (ALL PLAYERS TIMED OUT)
				if consecutiveTimeouts >= len(players) {

					broadcast(players, "RECOVERY",
						"All players timed out. Waiting for ANY player to resume...\n",
					)

					type recoveryInput struct {
						player *Player
						data   string
					}

					var resume recoveryInput
					found := false

					// ✅ Synchronous, no goroutine spam, no shared-conn reads
					for !found {
						for _, p := range players {
							// small per-player poll timeout so we can rotate players
							_ = p.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

							buf := make([]byte, 1024)
							n, err := p.conn.Read(buf)
							if err != nil {
								if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
									// just move to next player
									continue
								}
								// some other error, log and skip
								log.Printf("Error during recovery read from player %d: %v", p.id, err)
								continue
							}

							// reset deadline after successful read
							_ = p.conn.SetReadDeadline(time.Time{})

							guess := strings.TrimSpace(string(buf[:n]))
							resume = recoveryInput{
								player: p,
								data:   guess,
							}
							found = true
							break
						}
					}

					// ✅ Process the recovery guess exactly like before
					numGuess, err := ValidateGuess(resume.data)
					if err != nil {
						writeToClient(resume.player.conn, "INFO",
							"Invalid input: "+err.Error()+"\n")
						consecutiveTimeouts = 0
						// same behavior: give turn back to this player index
						currentTurn = resume.player.id - 1
						continue
					}

					feedback := GenerateFeedback(secret, numGuess, gameRng)
					currentGameGuesses++

					// ✅ WIN CASE in recovery
					if feedback.CorrectPlace == CodeLength {

						winMsg := fmt.Sprintf(
							"Player %d won! Secret was %d\n",
							resume.player.id,
							secret,
						)

						analytics.GamesPlayed++
						analytics.WinsByPlayer[currentPlayer.id]++
						analytics.GuessesUntilWin[secret] = currentGameGuesses

						for _, p := range players {
							if p.id != resume.player.id {
								analytics.LossesByPlayer[p.id]++
							}
						}

						prefix := GenerateTimestampPrefix()
						broadcast(players, "WIN", prefix+winMsg)
						broadcast(players, "NEWGAME",
							"New game starting in 3 seconds...\n")
						printAnalytics(analytics)

						time.Sleep(3 * time.Second)

						consecutiveTimeouts = 0
						// next turn after winner
						currentTurn = resume.player.id % len(players)
						break // ✅ EXIT INNER LOOP → STARTS NEW GAME
					}

					// ✅ NORMAL RECOVERY RESULT PATH (no win)
					msg := fmt.Sprintf(
						ColorBlue+"player: %d\n"+
							ColorCyan+"Number guessed: %d\n"+
							ColorGreen+"Correctly placed: %d\n"+
							ColorYellow+"Wrongly placed: %d\n"+
							ColorPurple+"Hint: %s\n"+ColorReset,
						resume.player.id,
						numGuess,
						feedback.CorrectPlace,
						feedback.WrongPlace,
						feedback.Hint,
					)

					prefix := GenerateTimestampPrefix()
					broadcast(players, "RESULT", prefix+msg)

					consecutiveTimeouts = 0
					currentTurn = resume.player.id % len(players)
					continue
				}

				// ✅ NORMAL TIMEOUT FLOW (not all timed out yet)
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
				writeToClient(currentPlayer.conn, "INFO",
					"Invalid input: "+err.Error()+"\n")
				continue
			}

			consecutiveTimeouts = 0 // ✅ VALID INPUT BREAKS TIMEOUT CHAIN

			feedback := GenerateFeedback(secret, numGuess, gameRng)
			currentGameGuesses++

			var msg string

			if feedback.CorrectPlace == CodeLength {

				msg = fmt.Sprintf(
					"Player %d won! Secret was %d\n",
					currentPlayer.id,
					secret,
				)
				analytics.GamesPlayed++
				analytics.WinsByPlayer[currentPlayer.id]++
				analytics.GuessesUntilWin[secret] = currentGameGuesses

				for _, p := range players {
					if p.id != currentPlayer.id {
						analytics.LossesByPlayer[p.id]++
					}
				}

				prefix := GenerateTimestampPrefix()
				broadcast(players, "WIN", prefix+msg)
				broadcast(players, "NEWGAME",
					"New game starting in 3 seconds...\n")
				printAnalytics(analytics)

				time.Sleep(3 * time.Second)
				currentTurn = (currentTurn + 1) % len(players)
				break
			}

			msg = fmt.Sprintf(
				ColorBlue+"RESULT player: %d\n"+
					ColorCyan+"Number guessed: %d\n"+
					ColorGreen+"Correctly placed: %d\n"+
					ColorYellow+"Wrongly placed: %d\n"+
					ColorPurple+"Hint: %s\n"+ColorReset,
				currentPlayer.id,
				numGuess,
				feedback.CorrectPlace,
				feedback.WrongPlace,
				feedback.Hint,
			)

			prefix := GenerateTimestampPrefix()
			broadcast(players, "RESULT", prefix+msg)

			currentTurn = (currentTurn + 1) % len(players)
		}
	}
}

// ✅ Proper socket drain (unchanged logic)
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

// ✅ JSON broadcast helper: keeps your strings & colors as-is
func broadcast(players []*Player, msgType, msg string) {
	for _, p := range players {
		writeToClient(p.conn, msgType, msg)
	}
}

func writeToClient(conn net.Conn, msgType string, text string) {
	msg := Message{
		Type: msgType,
		Text: text,
	}
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

	// ✅ Convert map to sortable slice
	type hardEntry struct {
		secret  int
		guesses int
	}

	hardList := make([]hardEntry, 0)

	for secret, guesses := range a.GuessesUntilWin {
		hardList = append(hardList, hardEntry{
			secret:  secret,
			guesses: guesses,
		})
	}

	// ✅ Sort by guesses DESCENDING (hardest first)
	sort.Slice(hardList, func(i, j int) bool {
		return hardList[i].guesses > hardList[j].guesses
	})

	// ✅ Print ONLY top 5
	limit := 5
	if len(hardList) < limit {
		limit = len(hardList)
	}

	if limit > 0 {
		log.Println("Top hardest secrets (by guesses until win):")
		for i := 0; i < limit; i++ {
			log.Printf(
				"#%d Secret: %d | Guesses until win: %d\n",
				i+1,
				hardList[i].secret,
				hardList[i].guesses,
			)
		}
	} else {
		log.Println("No completed games yet to determine hardest secrets.")
	}

	log.Println("============================")
}
