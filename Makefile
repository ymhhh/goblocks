.PHONY: test lint clean

test:
	go test ./... -race -count=1

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -rf bin/

.DEFAULT_GOAL := test
