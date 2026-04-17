# --------------------
# Builder stage
# --------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Native python for data generation (no QEMU)
RUN apk --no-cache add python3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run ./cmd/update-icloud -data-dir ./data

WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ipcheq2 ./cmd/server

# --------------------
# Final stage (NO apk, NO shell)
# --------------------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/ipcheq2 .
COPY --from=builder /app/data ./data
COPY vpnid_config.txt ./vpnid_config.txt

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/ipcheq2"]
