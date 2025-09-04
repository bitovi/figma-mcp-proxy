FROM golang:1.21.0-alpine

WORKDIR /app

COPY . .

RUN go build -o main .

EXPOSE 3845 3846

CMD ["./main"]