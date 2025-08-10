# ---- Builder ----
FROM golang:1.23-alpine AS builder

WORKDIR /subscription/

# Кеш
COPY go.mod go.sum ./
RUN go mod download

# Исходники
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /subscription_server ./cmd/main.go

# ---- Runtime ----
FROM alpine:3.20

WORKDIR /home/app

# Бинарь
COPY --from=builder /subscription_server .

# Миграции
COPY migrations ./migrations

# Копируем Swagger-документацию
COPY docs ./docs

ARG CONFIG_FILE=./config/config.yaml
COPY ${CONFIG_FILE} ./config/config.yaml

EXPOSE 8080
CMD ["/home/app/subscription_server"]
