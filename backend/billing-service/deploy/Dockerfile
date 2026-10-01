FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY backend/shared /app/shared
COPY backend/billing-service/go.mod backend/billing-service/go.sum /app/billing-service/
WORKDIR /app/billing-service
RUN go mod download

COPY backend/billing-service/ /app/billing-service/
RUN CGO_ENABLED=0 GOOS=linux go build -o billing-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/billing-service .

EXPOSE 9600
CMD ["./billing-service"]
