#FROM golang:1.23.4-alpine AS builder

#WORKDIR /usr/local/src

#RUN apk --no-cache add bash git make gcc gettext musl-dev

#COPY ["go.mod", "go.sum", "./"]
#RUN go mod download

#build
#COPY ./ ./
#RUN go build -o ./bin/app cmd/main.go

#FROM alpine

#COPY --from=builder

# Этап 1: Сборка приложения
FROM golang:1.23.4-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd


FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/myapp .

# Копируем goose из этапа сборки
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Копируем директорию с миграциями (если она есть)
COPY migrations ./migrations

# Копируем директорию с файлами переводов
COPY internal/locale ./internal/locale

# Экспонируем порт, на котором будет работать приложение
EXPOSE 8083

# Команда для запуска приложения
CMD ["./myapp"]