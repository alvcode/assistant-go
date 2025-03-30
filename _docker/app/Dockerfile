FROM golang:1.23.4-alpine AS builder

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o httpApp ./cmd/http
RUN CGO_ENABLED=0 GOOS=linux go build -o cliApp ./cmd/cli


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/httpApp .
COPY --from=builder /app/cliApp .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations
COPY internal/locale ./internal/locale

EXPOSE 8083

CMD ["./httpApp"]