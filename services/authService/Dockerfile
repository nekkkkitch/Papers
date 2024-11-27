FROM nekkkkitch/docker AS builder
WORKDIR /app
COPY . .
RUN GOOS=linux go build -a -o aus ./services/authService/cmd/app/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/aus .
COPY ./services/authService/cfg.yml .
CMD ["./aus"]