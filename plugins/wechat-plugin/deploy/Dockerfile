FROM golang:1.25-alpine AS builder

WORKDIR /app/repo

RUN apk add --no-cache git

COPY backend/shared /app/repo/backend/shared
COPY plugins/wechat-plugin /app/repo/plugins/wechat-plugin

WORKDIR /app/repo/plugins/wechat-plugin

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o wechat-plugin main.go

FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/repo/plugins/wechat-plugin/wechat-plugin /app/
COPY --from=builder /app/repo/plugins/wechat-plugin/config/config.yaml /app/config/

EXPOSE 20005

CMD ["./wechat-plugin"]
