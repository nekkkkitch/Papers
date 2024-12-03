FROM nekkkkitch/docker AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -a -o balance ./services/balance/cmd/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/balance .
COPY ./services/balance/cfg.yml .
CMD ["./balance"]