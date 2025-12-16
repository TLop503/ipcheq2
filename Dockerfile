# --------------------
# Build stage
# --------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install python ONLY in builder
RUN apk --no-cache add python3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run the data update script natively
WORKDIR /app/data
RUN python3 update_icloud_relays.py

# Build the Go binary (cross-compiled by buildx)
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ipcheq2 .

# --------------------
# Final stage
# --------------------
FROM alpine:latest

# Runtime-only deps
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

# Copy binary and assets
COPY --from=builder /app/ipcheq2 .
COPY --from=builder /app/web ./web
COPY --from=builder /app/data ./data

COPY vpnid_config.txt ./vpnid_config.txt

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
CMD ["./ipcheq2"]
