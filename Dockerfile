FROM golang:1.16.0-alpine3.13 as builder

COPY . /kure

WORKDIR /kure

RUN go install -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.13.2

COPY --from=builder /go/bin/kure /usr/bin/

RUN apk add --update \
        vim \
    && rm -rf /var/cache/apk/*

CMD ["/usr/bin/kure"]