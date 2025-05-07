# Сборка Go-приложения
FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o http-load-balancer ./cmd/balancer

# Минимальный образ для запуска
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/http-load-balancer .
EXPOSE 8080
CMD ["./http-load-balancer"]


