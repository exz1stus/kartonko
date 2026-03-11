# Stage 1: build
FROM golang:1.25.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o main ./cmd

# Stage 2: minimal image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/main .
ENV GIN_MODE=release
ENV NO_DOTENV=1
CMD ["./main"]