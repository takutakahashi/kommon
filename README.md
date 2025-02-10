# kommon

kommon is developer and adviser

## Prerequisites

- Go 1.20 or later
- Docker (optional, for containerized usage)

## Installation

### Local Installation

1. Clone the repository:
```bash
git clone https://github.com/takutakahashi/kommon.git
cd kommon
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o kommon
```

### Docker Installation

1. Build the Docker image:
```bash
docker build -t kommon .
```

2. Run the container:
```bash
docker run -it kommon
```

## Development Setup

1. Fork and clone the repository
2. Install dependencies:
```bash
go mod download
```

3. Make your changes
4. Run tests:
```bash
go test ./...
```

5. Submit a Pull Request

## License

[Add license information here]