VERSION = $(shell git tag --points-at HEAD)
COMMIT = $(shell git rev-parse --short HEAD)

install:
	@go install -ldflags="-s -w" .

test:
	go test ./...

test-race:
	go test ./... -race

proto: pb/*.proto
	@for file in $^ ; do \
		protoc -I. --go_out=./pb $$file ; \
	done

docker-build:
	docker build -t kure .

docker-run:
	docker run -it --rm kure sh

completion:
	@cd docs && go build main.go && ./main --completion