FROM golang:latest

WORKDIR /go/src/app

COPY . .

RUN go get -d -v github.com/lib/pq
RUN go get -d -v github.com/nats-io/stan.go

CMD ["go", "run", "/go/src/app/main.go"]