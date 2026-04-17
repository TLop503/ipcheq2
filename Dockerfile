# --------------------
# Builder stage
# --------------------
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run ./cmd/update-icloud -data-dir ./internal/data

WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ipcheq2 ./cmd/server

# --------------------
# Final stage (NO apk, NO shell)
# --------------------
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/ipcheq2 .

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/ipcheq2"]
