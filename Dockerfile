# Stage 1: Build Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o booking ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/booking .
COPY --from=builder /app/config ./config

RUN mkdir -p /app/logs

EXPOSE 8080

CMD ["./booking"]