# Build stage
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with server tag
ARG BUILD_TAGS=server
RUN CGO_ENABLED=0 GOOS=linux go build -tags ${BUILD_TAGS} -o cryptoquant-server .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/cryptoquant-server .

# Create a non-root user and set up permissions
RUN adduser -D -g '' appuser && \
    mkdir -p /app/log && \
    chown -R appuser:appuser /app/log && \
    chmod 755 /app/log

USER appuser

# Expose port if needed
# EXPOSE 8080

# Command to run the application
CMD ["./cryptoquant-server"] 