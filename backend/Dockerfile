FROM node:20.11.0-alpine3.14
FROM go:1.22-alpine

RUN mkdir -p /app
WORKDIR /app

COPY . .

RUN go build -o main ./cmd/main.go
EXPOSE 8080

CMD ["./main"]
