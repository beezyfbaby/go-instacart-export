.DEFAULT_GOAL := build
.PHONY: fmt vet test build run clean
fmt:
	go fmt ./...
vet:
	go vet ./...
test:
	go test -race ./...
build:
	go build -trimpath -o bin/instacart-export ./cmd/instacart-export
run:
	go run ./cmd/instacart-export
clean:
	rm -rf bin dist
