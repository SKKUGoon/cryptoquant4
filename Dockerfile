# -------- Build Stage --------
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Enable Go modules
ENV GO111MODULE=on

# Build with optional tags
ARG BUILD_TAGS
RUN CGO_ENABLED=0 GOOS=linux go build -tags "${BUILD_TAGS}" -o /app/cryptoquant-server .

# -------- Final Runtime Stage --------
FROM alpine:latest

WORKDIR /app

# Create non-root user for safer container execution
RUN adduser -D -g '' appuser

# Copy only the built binary
COPY --from=builder /app/cryptoquant-server .

# Set correct ownership for execution
RUN chown appuser:appuser /app/cryptoquant-server

# Switch to non-root user
USER appuser

# Optionally expose a port (uncomment if needed)
# EXPOSE 8080

# Entry point
CMD ["./cryptoquant-server"]