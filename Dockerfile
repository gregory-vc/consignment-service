FROM golang:1.11.5-alpine as builder

WORKDIR /go/src/github.com/gregory-vc/consignment-service

COPY . .

RUN go build

FROM alpine:latest

RUN apk --no-cache add ca-certificates iputils bash

RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/github.com/gregory-vc/consignment-service/consignment-service .

CMD ["./consignment-service"]