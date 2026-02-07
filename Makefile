.PHONY: build test lint snapshot install-hooks

build:
	go build -o sentei .

test:
	go test -race ./...

lint:
	golangci-lint run

snapshot:
	goreleaser build --snapshot --clean

install-hooks:
	./scripts/install-hooks.sh
