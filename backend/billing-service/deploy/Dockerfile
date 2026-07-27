FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/shared /app/shared
COPY backend/billing-service /app/billing-service

WORKDIR /app/billing-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o billing-service main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/billing-service/billing-service /app/

EXPOSE 20006

CMD ["./billing-service"]
