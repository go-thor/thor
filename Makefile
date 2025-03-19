.PHONY: all build clean test lint proto install

# Default target
all: test build

# Build the project
build:
	go build -v ./...

# Build the protoc-gen-thor binary
build-protoc:
	go build -o bin/protoc-gen-thor ./cmd/protoc-gen-thor

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linting
lint:
	golangci-lint run ./...

# Generate protobuf code for examples
proto: build-protoc
	cd examples/greeter && protoc --proto_path=proto --go_out=paths=source_relative:proto --thor_out=paths=source_relative:proto proto/greeter.proto

# Install the protoc-gen-thor generator
install:
	go install ./cmd/protoc-gen-thor

# Update dependencies
deps:
	go get -u ./...
	go mod tidy

# Format code
fmt:
	gofmt -s -w .

# Run the example
run-example-server:
	go run examples/greeter/server/main.go

run-example-client:
	go run examples/greeter/client/main.go