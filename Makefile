.DEFAULT_GOAL := build

.PHONY: fmt vet build run test clean

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o bin/take-home ./cmd/take-home

run:
	go run ./cmd/take-home/main.go

test:
	go test -v ./... ./test/...

clean:
	go clean
	rm -rf bin/