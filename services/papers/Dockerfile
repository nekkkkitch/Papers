FROM nekkkkitch/docker AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -a -o papers ./services/papers/cmd/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/papers .
COPY ./services/papers/cfg.yml .
CMD ["./papers"]