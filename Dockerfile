# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.24.5-alpine AS builder

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache ca-certificates tzdata

# Создаем пользователя для безопасности
RUN adduser -D -g '' appuser

# Устанавливаем рабочую директорию
WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код (без .git)
COPY . .

# Аргументы сборки для версии
ARG VERSION=""
ARG BUILD_DATE=""
ARG GIT_COMMIT=""

# Собираем приложение с инжектом версии
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags "-s -w -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
  -o main ./cmd/app

# Финальный образ с Alpine (с оболочкой для отладки)
FROM alpine:latest

# Устанавливаем минимальный набор утилит
RUN apk add --no-cache ca-certificates tzdata curl wget

# Создаем пользователя для безопасности
RUN adduser -D -g '' appuser

# Копируем собранное приложение
COPY --from=builder /build/main /app/main

# Переключаемся на непривилегированного пользователя
USER appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Запускаем приложение
CMD ["./main"]
