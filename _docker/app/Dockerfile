FROM golang:1.23.4-alpine AS builder

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/vektra/mockery/v2@latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/myapp .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations
COPY internal/locale ./internal/locale

EXPOSE 8083

CMD ["./myapp"]