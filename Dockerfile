FROM golang:1.23-alpine AS builder

WORKDIR /subscription/

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /subscription_server ./cmd/main.go

FROM alpine:latest

WORKDIR /home/app
COPY --from=builder /subscription_server .
COPY .env .

EXPOSE 8080

CMD ["./subscription_server"]