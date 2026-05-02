.DEFAULT_GOAL := build

.PHONY: build test lint clean cover

build:
	@echo "==> Building stresstest..."
	go build -o stresstest ./cmd/stresstest/

test:
	@echo "==> Running tests..."
	go test -count=1 -v ./...

test-race:
	@echo "==> Running tests with race detector..."
	go test -race -count=1 ./internal/... ./test/...

bench:
	@echo "==> Running benchmarks..."
	go test -bench=. -benchmem -run=^$ ./...

lint:
	@echo "==> Running linter..."
	golangci-lint run ./...

clean:
	@echo "==> Cleaning..."
	rm -f stresstest
	go clean -cache

cover: test
	go tool cover -html=coverage.out
