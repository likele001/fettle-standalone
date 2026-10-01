FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/shared /app/shared
COPY backend/user-service /app/user-service

WORKDIR /app/user-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/user-service/user-service /app/
COPY --from=builder /app/user-service/config/config.yaml /app/config/

EXPOSE 9200

CMD ["./user-service"]
