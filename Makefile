.DEFAULT_GOAL := help

.PHONY: build test lint tidy clean help

build:
	@mkdir -p bin
	go build -o bin/hub ./cmd/hub
	go build -o bin/agent ./cmd/agent

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ dist/

help:
	@echo "Available targets:"
	@echo "  build  Build hub and agent binaries"
	@echo "  test   Run tests with the race detector and coverage"
	@echo "  lint   Run golangci-lint"
	@echo "  tidy   Tidy Go module dependencies"
	@echo "  clean  Remove build artifacts"
	@echo "  help   Show this help"
