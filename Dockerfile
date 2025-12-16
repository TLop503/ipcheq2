# --------------------
# Build stage
# --------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ipcheq2 .

# --------------------
# Final stage
# --------------------
FROM alpine:latest

# Install runtime deps + python
RUN apk --no-cache add \
    ca-certificates \
    python3

# Create non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

# Copy binary and assets
COPY --from=builder /app/ipcheq2 .
COPY --from=builder /app/web ./web

# Copy data directory (includes update_icloud_relays.py)
COPY data ./data

# Copy config
COPY vpnid_config.txt ./vpnid_config.txt

# Run the python script from the data directory
WORKDIR /app/data
RUN python3 update_icloud_relays.py

# Restore working directory
WORKDIR /app

# Fix ownership
RUN chown -R appuser:appuser /app

# Drop privileges
USER appuser

EXPOSE 8080
CMD ["./ipcheq2"]
