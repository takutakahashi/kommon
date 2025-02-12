.PHONY: install_deps lint test integration-test build run docker-build vet fmt

# Go modules and golangci-lint installation
install_deps:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run go vet
vet: install_deps
	go vet ./...

# Run go fmt
fmt: install_deps
	go fmt ./...

# Run linters (includes vet and fmt check)
lint: install_deps vet fmt
	golangci-lint run ./...

# Run unit tests
test: install_deps
	go test -v -race ./...

# Run integration tests
integration-test: install_deps
	go test -v -race -tags=integration ./...

# Build the binary
build: install_deps
	go build -o bin/kommon .

# Run the application
run: build
	./bin/kommon

# Build Docker image
docker-build:
	docker build -t kommon:latest .