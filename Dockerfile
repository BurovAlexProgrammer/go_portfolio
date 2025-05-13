# Этап сборки
FROM golang:1.21 AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN go build -o /go_portfolio ./cmd

# Этап запуска
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /go_portfolio .

EXPOSE 8080

CMD ["./go_portfolio"]