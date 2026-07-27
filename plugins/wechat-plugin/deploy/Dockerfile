FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY plugins/wechat-plugin /app/wechat-plugin

WORKDIR /app/wechat-plugin

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o wechat-plugin main.go

FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/wechat-plugin/wechat-plugin /app/
COPY --from=builder /app/wechat-plugin/config/config.yaml /app/config/

EXPOSE 20005

CMD ["./wechat-plugin"]
