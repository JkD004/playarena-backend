# Build Stage
FROM golang:1.24-alpine AS builder

# Install build tools (git is often needed for go mod)
RUN apk add --no-cache git

WORKDIR /app

# Copy Go module files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
# Run tidy to ensure checksums match (fixes your previous "updates to go.mod needed" error)
RUN go mod tidy

# Copy the rest of the application code
COPY . .

# Build the Go application
# -o main: names the output binary "main"
RUN go build -o main .


# Final Run Stage (Small image for production)
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .
# Copy the .env file so the app can read secrets
#COPY .env .

# Expose the port your app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]