name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest
    env:
      GO_VERSION: '1.22'

    services:
      docker:
        image: docker:dind
        options: --privileged
        env:
          DOCKER_TLS_CERTDIR: ""
          DOCKER_HOST: tcp://0.0.0.0:2375

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Install dependencies
      run: |
        go mod download
        go mod verify

    - name: Build
      run: go build -v ./...

    - name: Run tests
      run: go test -v -race ./... -short

    - name: Build goose image for integration tests
      run: |
        docker build -t goose:latest -f Dockerfile.goose .
        docker tag goose:latest kommon-agent:latest
        docker images

    - name: Run Docker integration tests
      run: |
        go test -v ./... -run TestDocker

    - name: Run linter
      uses: golangci/golinkci-lint-action@v4
      with:
        version: v1.55.2
        skip-cache: true
        args: --timeout=5m