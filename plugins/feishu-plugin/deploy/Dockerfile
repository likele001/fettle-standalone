FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY plugins/feishu-plugin /app/feishu-plugin
WORKDIR /app/feishu-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o feishu-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/feishu-plugin/feishu-plugin /app/
COPY --from=builder /app/feishu-plugin/config/config.yaml /app/config/
EXPOSE 20024
CMD ["./feishu-plugin"]
