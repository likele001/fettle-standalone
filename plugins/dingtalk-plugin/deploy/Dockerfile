FROM golang:1.25-alpine AS builder
WORKDIR /app/repo
RUN apk add --no-cache git
COPY backend/shared /app/repo/backend/shared
COPY plugins/dingtalk-plugin /app/repo/plugins/dingtalk-plugin
WORKDIR /app/repo/plugins/dingtalk-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o dingtalk-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/repo/plugins/dingtalk-plugin/dingtalk-plugin /app/
EXPOSE 20025
CMD ["./dingtalk-plugin"]
