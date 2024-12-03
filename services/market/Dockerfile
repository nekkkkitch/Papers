FROM nekkkkitch/docker AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -a -o market ./services/market/cmd/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/market .
COPY ./services/market/cfg.yml .
CMD ["./market"]