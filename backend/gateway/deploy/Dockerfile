FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/shared /app/shared
COPY backend/gateway /app/gateway

WORKDIR /app/gateway

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/gateway/gateway /app/
COPY --from=builder /app/gateway/config.yaml /app/

EXPOSE 20001

CMD ["./gateway"]
