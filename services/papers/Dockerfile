FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN GOOS=linux go build -o papers ./cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/papers /app/papers
COPY cfg.yml .
EXPOSE 50053
CMD ["./papers"]