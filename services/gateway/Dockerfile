FROM nekkkkitch/docker AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -a -o gateway ./services/gateway/cmd/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gateway .
COPY ./services/gateway/cfg.yml .
CMD ["./gateway"]