# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/server
RUN go build -o server .

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cmd/server/server .
ENV MAX_PLAYERS=3
ENV CODE_LENGTH=4
ENV DIFFICULTY=easy
ENV TURN_TIME_SECONDS=30
EXPOSE 8080
ENTRYPOINT ["/app/server"]
