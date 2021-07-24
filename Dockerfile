FROM golang:1.16-alpine as build

WORKDIR /app
ADD . /app
RUN go build -o main .

EXPOSE 9666

ENTRYPOINT ["/app/main"]
