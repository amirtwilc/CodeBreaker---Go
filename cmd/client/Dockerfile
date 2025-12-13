# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/client
RUN go build -o client .

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/cmd/client/client .
ENTRYPOINT ["/app/client"]
