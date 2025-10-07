FROM golang:1.25.2

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o main ./api

CMD ["./main"]