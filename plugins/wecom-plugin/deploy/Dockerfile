FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY plugins/wecom-plugin /app/wecom-plugin
WORKDIR /app/wecom-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o wecom-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/wecom-plugin/wecom-plugin /app/
COPY --from=builder /app/wecom-plugin/config/config.yaml /app/config/
EXPOSE 20022
CMD ["./wecom-plugin"]
