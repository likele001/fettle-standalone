FROM golang:1.25-alpine AS builder
WORKDIR /app/repo
RUN apk add --no-cache git
COPY backend/shared /app/repo/backend/shared
COPY plugins/douyin-plugin /app/repo/plugins/douyin-plugin
WORKDIR /app/repo/plugins/douyin-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o douyin-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/repo/plugins/douyin-plugin/douyin-plugin /app/
COPY --from=builder /app/repo/plugins/douyin-plugin/config/config.yaml /app/config/
EXPOSE 20023
CMD ["./douyin-plugin"]
