# Этап сборки
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка приложения
RUN go build -o meet ./cmd/meet/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app

# Копируем собранный бинарник из предыдущего этапа
COPY --from=builder /app/meet .

# Копируем статические файлы
COPY --from=builder /app/web ./web

EXPOSE 8080

# Команда запуска
CMD ["./meet"]