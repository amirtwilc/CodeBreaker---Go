# ---- Build stage ----
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod files and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build a statically linked binary
RUN go build -o /codebreaker ./main.go

# ---- Runtime stage ----
FROM alpine:3.18

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /codebreaker /app/codebreaker

# Environment (optional, here just as example)
ENV CODEBREAKER_PORT=8080

# Expose the server port
EXPOSE 8080

# Entry point: run the Go binary.
# We keep the binary as entrypoint and pass "server" by default as CMD,
# so we can override it to "client" if we ever want to.
ENTRYPOINT ["./codebreaker"]
CMD ["server"]
