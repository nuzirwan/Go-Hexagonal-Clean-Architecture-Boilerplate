# Build stage
FROM golang:1.26-alpine AS builder

RUN go install go.opentelemetry.io/auto/otelc@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 otelc go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /app/myservice ./main.go

# Runtime: distroless static (~2MB), includes CA certs, no shell, nonroot user.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/myservice /app/main
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080 8081

# No default command — caller decides the mode:
#   serve http:  ["/app/main", "serve", "http"]
#   serve grpc:  ["/app/main", "serve", "grpc"]
#   worker:      ["/app/main", "worker"]
#   migrate:     ["/app/main", "migrate", "up"]
ENTRYPOINT ["/app/main"]
