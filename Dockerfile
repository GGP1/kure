FROM golang:1.16.3-alpine3.13 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download

COPY . .

RUN go install -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.13.4

RUN apk add --update \
        vim \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/kure /usr/bin/

CMD ["/usr/bin/kure"]