# Build stage: Go + Alpine
FROM golang:1.24.1-alpine3.20 AS build-stage

# Установим необходимые утилиты
RUN apk update && apk add bash && apk add build-base

# Установим рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum, загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные исходники
COPY . .

# Сборка бинарника auth
RUN CGO_ENABLED=0 go build -o /build/auth ./cmd/auth/main.go

# Release stage: минимальный дистрибутив
FROM gcr.io/distroless/base-debian11 AS build-release-stage

# Копируем бинарник из build-stage
COPY --from=build-stage /build/auth /build/auth

# Открываем порты (если используется HTTP или gRPC)
EXPOSE 8080 50051

# Стартуем сервис
CMD ["/build/auth"]
