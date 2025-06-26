FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o email-api main.go

# Final image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/email-api /app/email-api

ENV PORT=8080
EXPOSE 8080

CMD ["/app/email-api"]