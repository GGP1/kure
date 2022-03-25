FROM golang:1.18.0-alpine3.15 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download && go mod verify

RUN apk add --update --no-cache git

COPY . .

RUN go install -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.15

ENV USER=gandalf
ENV UID=10001

COPY --from=builder /go/bin/kure /usr/bin/

RUN adduser $USER -D -g "" -s "/sbin/nologin" -u $UID \
    && chown $USER /usr/bin/kure \
    && chmod 0700 /usr/bin/kure

USER $USER

CMD ["/usr/bin/kure"]