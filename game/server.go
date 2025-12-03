package game

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	MaxPlayers      = 2
	TurnTimeSeconds = 100 //time limit per turn
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
	GamesPlayed    int
	WinsByPlayer   map[int]int
	LossesByPlayer map[int]int
	GuessFrequency map[int]int
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
		WinsByPlayer:   make(map[int]int),
		LossesByPlayer: make(map[int]int),
		GuessFrequency: make(map[int]int),
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
		writeToClient(conn, fmt.Sprintf("Welcome Player %d! Waiting for others...\n", player.id))
	}

	broadcast(players, "All players connected. Game starting now!\n")

	currentTurn := 0
	consecutiveTimeouts := 0

	for { // GAME LOOP (infinite rounds)

		secret := GenerateSecretCode()
		log.Printf("DEBUG NEW SECRET: %d\n", secret)

		broadcast(players, "New game started!\n")

		for {
			currentPlayer := players[currentTurn]

			// Notify turns
			for _, p := range players {
				if p.id == currentPlayer.id {
					writeToClient(p.conn, "TURN Your turn! Enter your guess:\n")
				} else {
					writeToClient(p.conn, fmt.Sprintf("WAIT Waiting for Player %d...\n", currentPlayer.id))
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
				broadcast(players, msg)

				consecutiveTimeouts++

				// ✅ GLOBAL RECOVERY MODE (ALL PLAYERS TIMED OUT)
				if consecutiveTimeouts >= len(players) {

					broadcast(players,
						"All players timed out. Waiting for ANY player to resume...\n",
					)

					type recoveryInput struct {
						player *Player
						data   string
					}

					recoveryCh := make(chan recoveryInput, 1)

					// ✅ Wait for input from ANY player
					for _, p := range players {
						go func(p *Player) {
							_ = p.conn.SetReadDeadline(time.Time{})

							buf := make([]byte, 1024)
							n, err := p.conn.Read(buf)
							if err != nil {
								return
							}

							guess := strings.TrimSpace(string(buf[:n]))
							recoveryCh <- recoveryInput{
								player: p,
								data:   guess,
							}
						}(p)
					}

					resume := <-recoveryCh

					numGuess, err := ValidateGuess(resume.data)
					if err != nil {
						writeToClient(resume.player.conn, "Invalid input: "+err.Error()+"\n")
						consecutiveTimeouts = 0
						currentTurn = resume.player.id - 1
						continue
					}

					feedback := GenerateFeedback(secret, numGuess, gameRng)
					analytics.GuessFrequency[numGuess]++

					if feedback.CorrectPlace == CodeLength {

						winMsg := fmt.Sprintf(
							"Player %d won! Secret was %d\n",
							resume.player.id,
							secret,
						)

						analytics.GamesPlayed++
						analytics.WinsByPlayer[resume.player.id]++

						for _, p := range players {
							if p.id != resume.player.id {
								analytics.LossesByPlayer[p.id]++
							}
						}

						prefix := GenerateTimestampPrefix()
						broadcast(players, prefix+winMsg)
						broadcast(players, "NEWGAME New game starting in 3 seconds...\n")
						printAnalytics(analytics)

						time.Sleep(3 * time.Second)

						consecutiveTimeouts = 0
						currentTurn = resume.player.id % len(players)
						break // ✅ EXIT INNER LOOP → STARTS NEW GAME
					}

					/* ✅ NORMAL RECOVERY RESULT PATH */
					msg := fmt.Sprintf(
						ColorBlue+"RESULT player: %d\n"+
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
					broadcast(players, prefix+msg)

					consecutiveTimeouts = 0
					currentTurn = resume.player.id % len(players)
					continue
				}

				// ✅ NORMAL TIMEOUT FLOW
				currentTurn = (currentTurn + 1) % len(players)
				drainLateInput(currentPlayer.conn)
				continue
			}

			if err != nil {
				log.Printf("Player %d disconnected\n", currentPlayer.id)
				return
			}

			_ = currentPlayer.conn.SetReadDeadline(time.Time{})

			guess := strings.TrimSpace(string(buffer[:n]))
			numGuess, err := ValidateGuess(guess)
			if err != nil {
				writeToClient(currentPlayer.conn, "Invalid input: "+err.Error()+"\n")
				continue
			}

			consecutiveTimeouts = 0 // ✅ VALID INPUT BREAKS TIMEOUT CHAIN

			feedback := GenerateFeedback(secret, numGuess, gameRng)
			analytics.GuessFrequency[numGuess]++

			var msg string

			if feedback.CorrectPlace == CodeLength {

				msg = fmt.Sprintf(
					"Player %d won! Secret was %d\n",
					currentPlayer.id,
					secret,
				)

				analytics.GamesPlayed++
				analytics.WinsByPlayer[currentPlayer.id]++

				for _, p := range players {
					if p.id != currentPlayer.id {
						analytics.LossesByPlayer[p.id]++
					}
				}

				prefix := GenerateTimestampPrefix()
				broadcast(players, prefix+msg)
				broadcast(players, "NEWGAME New game starting in 3 seconds...\n")
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
			broadcast(players, prefix+msg)

			currentTurn = (currentTurn + 1) % len(players)
		}
	}
}

// ✅ Proper socket drain
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

// Broadcast helper
func broadcast(players []*Player, msg string) {
	for _, p := range players {
		writeToClient(p.conn, msg)
	}
}

func writeToClient(conn net.Conn, s string) {
	_, err := conn.Write([]byte(s))
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

	hardest := make([]int, 0)
	for num, freq := range a.GuessFrequency {
		if freq >= 5 {
			hardest = append(hardest, num)
		}
	}

	if len(hardest) > 0 {
		log.Printf("Hardest-to-guess candidates (freq >=5): %v\n", hardest)
	} else {
		log.Println("No hardest numbers identified yet")
	}

	log.Println("============================")
}
