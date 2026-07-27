FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY plugins/douyin-plugin /app/douyin-plugin
WORKDIR /app/douyin-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o douyin-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/douyin-plugin/douyin-plugin /app/
COPY --from=builder /app/douyin-plugin/config/config.yaml /app/config/
EXPOSE 20023
CMD ["./douyin-plugin"]
