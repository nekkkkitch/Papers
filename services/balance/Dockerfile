FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN GOOS=linux go build -o balance ./cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/balance /app/balance
COPY cfg.yml .
EXPOSE 50052
CMD ["./balance"]