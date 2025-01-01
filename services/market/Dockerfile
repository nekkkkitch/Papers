FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN GOOS=linux go build -o market ./cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/market /app/market
COPY cfg.yml .
EXPOSE 6378
CMD ["./market"]