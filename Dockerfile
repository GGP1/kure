FROM golang:1.17.6-alpine3.15 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download && go mod verify

RUN apk add --update --no-cache git make

COPY . .

RUN make install

# ---------------------------------------------

FROM alpine:3.14.1

RUN apk add --update --no-cache vim

COPY --from=builder /go/bin/kure /usr/bin/

CMD ["/usr/bin/kure"]