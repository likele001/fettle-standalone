FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY backend/skill-service/go.mod backend/skill-service/go.sum ./
RUN go mod download

COPY backend/skill-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o skill-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/skill-service .

EXPOSE 9500
CMD ["./skill-service"]
