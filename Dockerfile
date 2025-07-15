# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ipcheq2 .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user first
RUN adduser -D -s /bin/sh appuser

# Set working directory for the app user
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/ipcheq2 .

# Copy web assets
COPY --from=builder /app/web ./web

# Copy iCloud private relay prefix files in
COPY prefixes/*.txt ./prefixes/

# Change ownership to app user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./ipcheq2"]
