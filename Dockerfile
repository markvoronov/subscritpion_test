# ---- Builder ----
FROM golang:1.23-alpine AS builder

WORKDIR /subscription/

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /subscription_server ./cmd/main.go

# ---- Runtime ----
FROM alpine:latest

WORKDIR /home/app

# Select which env file to bake in (default .env)
ARG CONFIG_FILE=.env

# Copy binary and chosen env file
COPY --from=builder /subscription_server .
COPY ${CONFIG_FILE} .env

# Copy migrations (now at repo root) into runtime image
# If you keep migrations at /internal/migrations, change the source path accordingly.
COPY --from=builder /subscription/migrations ./migrations

# Копируем Swagger-документацию
COPY --from=builder /subscription/docs ./docs

EXPOSE 8080
CMD ["/home/app/subscription_server"]
