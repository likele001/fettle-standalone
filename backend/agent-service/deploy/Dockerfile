FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/shared /app/shared
COPY backend/agent-service /app/agent-service

WORKDIR /app/agent-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-service main.go

FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/agent-service/agent-service /app/
COPY --from=builder /app/agent-service/config/config.yaml /app/config/

EXPOSE 9300

CMD ["./agent-service"]
