FROM golang:latest AS build

WORKDIR /app

# Optimization to cache dependencies
ADD go.* ./
RUN go mod download

ADD . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kure ./cmd

FROM scratch
COPY --from=build /app/kure /
ENTRYPOINT ["/kure"]