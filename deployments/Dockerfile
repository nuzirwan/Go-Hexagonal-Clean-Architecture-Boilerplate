# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates
RUN go install go.opentelemetry.io/auto/otelc@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux otelc go build -ldflags="-s -w" -o /app/myservice ./main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/myservice .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 8081

ENTRYPOINT ["./myservice"]
CMD ["serve", "http"]
