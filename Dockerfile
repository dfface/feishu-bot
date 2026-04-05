FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o feishu-bot ./cmd/feishu-bot

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/sh -D app

WORKDIR /app

COPY --from=builder /app/feishu-bot .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

RUN chown -R app:app /app

USER app

ENTRYPOINT ["./feishu-bot"]
