FROM golang:alpine AS builder
WORKDIR /app
ADD go.mod .
COPY . .
RUN GOOS=linux go build -o aus ./cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/aus /app/aus
COPY cfg.yml .
EXPOSE 50051
CMD ["./aus"]