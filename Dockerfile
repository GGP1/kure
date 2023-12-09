FROM golang:1.21-alpine3.18 as builder

WORKDIR /kure

COPY go.mod .

RUN go mod download && go mod verify

RUN apk add --update --no-cache git

COPY . .

RUN CGO_ENABLED=0 go install -ldflags="-s -w" .

# ---------------------------------------------

FROM alpine:3.18

RUN apk add --update --no-cache vim

COPY --from=builder /go/bin/kure /usr/bin/

CMD ["/usr/bin/kure"]
