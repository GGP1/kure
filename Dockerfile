FROM golang:1.16.0-alpine3.13 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download

COPY . .

RUN go install -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.13.2

RUN apk add --update \
        vim \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/kure /usr/bin/

CMD ["/usr/bin/kure"]