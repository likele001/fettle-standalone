FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/shared /app/shared
COPY backend/chat-service /app/chat-service

WORKDIR /app/chat-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-service main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/chat-service/chat-service /app/

EXPOSE 20004

CMD ["./chat-service"]
