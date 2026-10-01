FROM golang:1.25-alpine AS builder
WORKDIR /app/repo
RUN apk add --no-cache git
COPY backend/shared /app/repo/backend/shared
COPY plugins/wecom-plugin /app/repo/plugins/wecom-plugin
WORKDIR /app/repo/plugins/wecom-plugin
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o wecom-plugin .

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/repo/plugins/wecom-plugin/wecom-plugin /app/
COPY --from=builder /app/repo/plugins/wecom-plugin/config/config.yaml /app/config/
EXPOSE 20022
CMD ["./wecom-plugin"]
