FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制公共库
COPY backend/shared /app/shared

# 复制服务代码
COPY backend/gateway /app/gateway

WORKDIR /app/gateway

# 下载依赖并构建
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway main.go

FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates

# 从构建阶段复制二进制文件
COPY --from=builder /app/gateway/gateway /app/
COPY --from=builder /app/gateway/config/config.yaml /app/config/

EXPOSE 9100

CMD ["./gateway"]
