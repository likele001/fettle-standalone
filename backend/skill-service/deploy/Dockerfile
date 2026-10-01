FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY backend/shared /app/shared
COPY backend/skill-service/go.mod backend/skill-service/go.sum /app/skill-service/
WORKDIR /app/skill-service
RUN go mod download

COPY backend/skill-service/ /app/skill-service/
RUN CGO_ENABLED=0 GOOS=linux go build -o skill-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/skill-service .

EXPOSE 9500
CMD ["./skill-service"]
