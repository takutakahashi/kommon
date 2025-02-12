.PHONY: install_deps lint test integration-test build run docker-build

# Go modules and golangci-lint installation
install_deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
lint:
	golangci-lint run ./...

# Run unit tests
test:
	go test -v -race ./...

# Run integration tests
integration-test:
	go test -v -race -tags=integration ./...

# Build the binary
build:
	go build -o bin/kommon .

# Run the application
run: build
	./bin/kommon

# Build Docker image
docker-build:
	docker build -t kommon:latest .