FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN GOOS=linux go build -o gateway ./cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/gateway /app/gateway
COPY cfg.yml .
EXPOSE 6370
CMD ["./gateway"]