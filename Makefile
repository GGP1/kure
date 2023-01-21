VERSION = $(shell git tag --points-at HEAD)
COMMIT = $(shell git rev-parse --short HEAD)

install:
	@go install -ldflags="-s -w" .

test:
	go test ./...

test-race:
	go test ./... -race

proto:
	@cd pb && for type in card entry file totp ; do \
		protoc -I. --go_out=. $$type.proto ; \
	done

docker-build:
	docker build -t kure .

docker-run:
	docker run -it --rm kure sh

completion:
	@cd docs && go build main.go && ./main --completion