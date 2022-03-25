FROM golang:1.18.0-alpine3.15 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download && go mod verify

RUN apk add --update --no-cache git

COPY . .

RUN go install -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.15

RUN apk add --update --no-cache vim

COPY --from=builder /go/bin/kure /usr/bin/

CMD ["/usr/bin/kure"]