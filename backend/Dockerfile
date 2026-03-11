FROM golang:1.25.2

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o main ./cmd

ENV GIN_MODE=release
ENV NO_DOTENV=1

CMD ["./main"]