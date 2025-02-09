FROM debian:bookworm-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    bzip2 \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Download and install Goose binary
RUN curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh -o install.sh \
    && chmod +x install.sh \
    && ./install.sh \
    && rm install.sh

# Create necessary directories for Goose
RUN mkdir -p /root/.config/goose

# Verify installation
RUN goose --version

# Command to run Goose
CMD ["goose", "session"]