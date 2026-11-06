# Multi-stage Dockerfile for Go RAG Backend
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/knowledge /app/knowledge

EXPOSE 8080

ENTRYPOINT ["/app/server"]

