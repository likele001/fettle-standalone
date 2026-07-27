FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY plugins/dingtalk-plugin /app/dingtalk-plugin
WORKDIR /app/dingtalk-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o dingtalk-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/dingtalk-plugin/dingtalk-plugin /app/
COPY --from=builder /app/dingtalk-plugin/config/config.yaml /app/config/
EXPOSE 20025
CMD ["./dingtalk-plugin"]
