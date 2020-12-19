FROM golang:1.15-alpine as builder

COPY . /go/src/github.com/GGP1/kure

WORKDIR /go/src/github.com/GGP1/kure

RUN go get -d -v ./...

RUN go build -o kure -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.12.1

COPY --from=builder /go/src/github.com/GGP1/kure/kure /usr/bin/

CMD ["kure"]