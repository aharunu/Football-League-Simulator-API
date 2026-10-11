# Use official Golang image as build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
RUN go build -o league-simulator .

# Use a minimal image for running
FROM alpine:latest

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/league-simulator .

# Listen on the container interface; host port mapping can remain local.
ENV LISTEN_ADDR=0.0.0.0:8080

# Expose port (change if your app uses a different port)
EXPOSE 8080

# Run the binary
CMD ["./league-simulator"]
