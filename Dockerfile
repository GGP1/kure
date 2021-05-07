FROM golang:1.16.4-alpine3.13 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download && go mod verify

RUN apk add --update --no-cache git make

COPY . .

RUN make install

# ---------------------------------------------

FROM alpine:3.13.5

RUN apk add --update --no-cache vim

COPY --from=builder /go/bin/kure /usr/bin/

CMD ["/usr/bin/kure"]