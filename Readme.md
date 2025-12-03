# Code Breaker Multiplayer Game

**Author:** Amir Twil-Cohen 
**Language:** Go  
**Project Type:** Multiplayer TCP Game  
**Mode:** Turn-based, time-limited guessing game  

---

## 🧠 Overview

**Code Breaker** is a multiplayer, turn-based guessing game built in Go.  
Players connect to a TCP server and take turns guessing a secret 4-digit code. After each guess, players receive:

- Number of correct digits in the correct position
- Number of correct digits in the wrong position
- A smart hint (e.g., whether a correct digit is in the first or second half)
- A time-limited turn – if you don’t guess in time, you lose your turn
- Automatic new game after a win
- Server-side analytics

The game supports:
- Multiple players
- Turn-based logic
- Timeout enforcement
- Automatic rematches
- Server analytics

---

## How to start with docker

1. In one terminal:
   docker run --rm -p 8080:8080 codebreaker-game

2. For each player open a new terminal and run:
go run main.go client

