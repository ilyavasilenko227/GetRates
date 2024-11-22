FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app/main /main

COPY --from=builder /app/migrations /app/migrations

CMD ["./main"]