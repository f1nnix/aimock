FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o aimock

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/aimock .

# Copy the config file
COPY config.json .

# Expose the port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/aimock", "-config", "config.json"]