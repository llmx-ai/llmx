.PHONY: test lint fmt vet build clean examples

# Run tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -s -w .
	goimports -w .

# Run go vet
vet:
	go vet ./...

# Build all
build:
	go build ./...

# Clean
clean:
	go clean
	rm -f coverage.txt
	find examples -type f -name 'main' -delete

# Run examples
examples:
	cd examples/chat && go run main.go
	cd examples/streaming && go run main.go

# Install tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Run all checks
check: fmt vet lint test

.DEFAULT_GOAL := test
