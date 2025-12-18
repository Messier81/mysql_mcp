# Dockerfile
# This builds a Docker image for the MySQL MCP server.
# We use a multi-stage build to keep the final image small.

# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
# CGO_ENABLED=0 creates a static binary that works in minimal containers
# -ldflags="-w -s" strips debug info to make the binary smaller
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mysql-mcp .

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install CA certificates for HTTPS (if needed in future)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/mysql-mcp .

# The MCP protocol uses stdin/stdout, so we don't need to expose ports
# Database credentials should be passed via environment variables

# Run the server
ENTRYPOINT ["/app/mysql-mcp"]

