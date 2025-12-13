# Code Breaker Multiplayer Game

**Author:** Amir Twil-Cohen <br>
**Language:** Go  
**Project Type:** Multiplayer TCP Game  
**Mode:** Turn-based, time-limited guessing game  

---

## Motive

This project began as an assignment during an interview process, and was later extended
for further learning and experimentation with Go, Docker, and Kubernetes.

## Overview

**Code Breaker** is a multiplayer, turn-based guessing game built in Go.  
Players connect to a TCP server and take turns guessing a secret code. 
After each guess, players receive:

- Number of correct digits in the correct position
- Number of correct digits in the wrong position
- A smart hint (e.g., whether a correct digit is in the first or second half)
- A time-limited turn – if you don’t guess in time, you lose your turn
- Server-side analytics

The game supports:
- Multiple players
- Turn-based logic
- Timeout enforcement
- Automatic rematches
- Server analytics

---

## Starting the game

This section explains the different ways to start the game.
The game behavior is controlled by several configurable settings that affect gameplay.
- **MaxPlayers** – Number of players that can play simultaneously (minimum: 1)
- **CodeLength** – Number of digits in the secret code (2–8). Hard difficulty requires more than 2 digits.
- **Difficulty** – easy / medium / hard
- **TurnTimeSeconds** – Time allowed per turn before it is skipped (irrelevant for single-player games)

### With Go

**Prerequisites:** Go 1.20 or newer  
https://go.dev/doc/install

1. Open the launcher.go file and change the const variables if desired
2. Run this command at the project's root path: <br>
   `go run launcher.go`
3. Separate terminals will open for the server and each player.

**Note:** This project was tested on Windows 11.
Although `launcher.go` includes support for macOS and Linux, those platforms were not tested.

### With Docker
* prerequisites: Docker https://docs.docker.com/desktop/

1. Open `/cmd/server/Dockerfile` and change the environment variables if desired
2. Build the Server and Client images: <br>
   `docker build -f cmd/server/Dockerfile -t codebreaker-server:latest .` <br>
   `docker build -f cmd/client/Dockerfile -t codebreaker-client:latest .`
3. Run the Server: <br>
   `docker run -it --rm -p 8080:8080 codebreaker-server:latest`
4. For each player open a terminal and run: <br>
   `docker run -it --rm -e SERVER_ADDR=host.docker.internal:8080 codebreaker-client:latest` <br><br>
   `host.docker.internal` allows Docker containers to reach the server running on the host machine.


### With Docker Compose

* prerequisites: Docker https://docs.docker.com/desktop/

1. Open `docker-compose.yaml` and change the variables under environment if desired
2. The current `docker-compose.yaml` is configured for 2 players.
   Therefore:
   - `MAX_PLAYERS` is set to 2 in the server
   - Two client services (`client1` and `client2`) are defined

    To change the number of players:
   - Update `MAX_PLAYERS`
   - Add or remove client services so their count matches `MAX_PLAYERS`
3. Run the docker compose by: <br>
   `docker compose up`
4. Each client runs in interactive mode. To play, attach a terminal to each client container: <br>
   `docker compose attach client1` <br>
   `docker compose attach client2` <br>
   **note**: to see the logs before attaching the terminal: <br>
   `docker compose logs client1`

### With Kubernetes

* prerequisites: Kubernetes. I used Docker Desktop https://docs.docker.com/desktop/ <br>
In Settings, go to Kubernetes and "Enable Kubernetes"

1. Update settings in `server-configmap.yaml` if desired. <br>
   Make sure `MAX_PLAYERS` value is equal to `replicas` value under `client-deployment.yaml`. <br>
   This will make sure enough client pods are created for each player
2. Deploy the Kubernetes resources: <br>
   `kubectl apply -f k8s/namespace.yaml` <br>
   `kubectl apply -f k8s/server-configmap.yaml` <br>
   `kubectl apply -f k8s/server-deployment.yaml` <br>
   `kubectl apply -f k8s/server-service.yaml` <br>
   `kubectl apply -f k8s/client-deployment.yaml` <br>
**note**: Can also run <br>
   `kubectl apply -f k8s/` <br>
   But if there is an error due to order, then should run again until no error received
3. Port-forward the server service. This exposes the server locally so clients can connect via localhost:8080. <br>
   `kubectl -n codebreaker port-forward svc/codebreaker-server 8080:8080`
4. Each client runs in its own pod with an interactive TTY. <br>
   Client pods are implemented using a StatefulSet to ensure predictable pod names. <br>
   Attach a terminal to each client pod to play. <br>
   `kubectl attach -it codebreaker-client-0 -n codebreaker` <br>
   `kubectl attach -it codebreaker-client-1 -n codebreaker` <br>
   `kubectl attach -it codebreaker-client-2 -n codebreaker` <br>
**note**: to see the logs before attaching the terminal: <br>
   `kubectl logs codebreaker-client-0 -n codebreaker`



