.PHONY: build test lint clean install

MODULE := github.com/ymhhh/goblocks
CLI := ./cmd/goblocks

build:
	go build -o bin/goblocks $(CLI)

install:
	go install $(CLI)

test:
	go test ./... -race -count=1

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -rf bin/

.DEFAULT_GOAL := build
