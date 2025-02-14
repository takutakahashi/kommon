FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o kommon

FROM debian:bookworm-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    bzip2 \
    libxcb1 \
    libdbus-1-3 \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /root

# Copy kommon binary from builder
COPY --from=builder /app/kommon /usr/local/bin/

# Download and install Goose binary
RUN curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh -o install.sh \
    && chmod +x install.sh \
    && ./install.sh \
    && rm install.sh

# Add the local bin directory to PATH
ENV PATH="/root/.local/bin:${PATH}"

# Create necessary directories for Goose
RUN mkdir -p /root/.config/goose /root/.local/share/goose

# Verify installation
RUN goose --version

ENTRYPOINT ["kommon"]
