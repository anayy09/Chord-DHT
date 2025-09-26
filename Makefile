.PHONY: build run test clean

build:
	go build -o bin/chord ./cmd/chord

run:
	go run ./cmd/chord

test:
	go test ./...

clean:
	rm -rf bin/

fmt:
	go fmt ./...

vet:
	go vet ./...