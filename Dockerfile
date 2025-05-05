FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o coding-agent

# Use a smaller image for the final image
FROM alpine:latest

# Install bash for the wrapper script
RUN apk --no-cache add bash

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/coding-agent /app/coding-agent

# Copy the wrapper script
COPY coding-agent.sh /app/coding-agent.sh

# Make the wrapper script executable
RUN chmod +x /app/coding-agent.sh

# Set the entrypoint
ENTRYPOINT ["/app/coding-agent.sh"]
