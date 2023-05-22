FROM golang:1.20 as builder

WORKDIR /app

COPY . /app

RUN go build -o main .

FROM ubuntu:devel

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /app/main /app/main

WORKDIR "/app"

CMD ["/app/main"]

EXPOSE 8080