# Этап сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum сначала (для кеширования зависимостей)
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Собираем бинарник (статически скомпилированный)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o roulette .

# Финальный минимальный образ
FROM alpine:latest

# Устанавливаем полезные утилиты (необязательно)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем собранный бинарник из этапа сборки
COPY --from=builder /app/roulette .

# Копируем статику и конфиги (они будут перезаписаны volume'ами, если нужно)
COPY --from=builder /app/static ./static
COPY --from=builder /app/configs ./configs

# Пробрасываем порт
EXPOSE 3001

# Запускаем сервер
CMD ["./roulette"]