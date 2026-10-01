FROM golang:1.25-alpine AS builder
WORKDIR /app/repo
RUN apk add --no-cache git
COPY backend/shared /app/repo/backend/shared
COPY plugins/feishu-plugin /app/repo/plugins/feishu-plugin
WORKDIR /app/repo/plugins/feishu-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o feishu-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/repo/plugins/feishu-plugin/feishu-plugin /app/
COPY --from=builder /app/repo/plugins/feishu-plugin/config/config.yaml /app/config/
EXPOSE 20024
CMD ["./feishu-plugin"]
