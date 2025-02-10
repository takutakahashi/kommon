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

# Download and install Goose binary
RUN curl -fsSL https://github.com/block/goose/releases/download/stable/download_cli.sh -o install.sh \
    && chmod +x install.sh \
    && ./install.sh \
    && rm install.sh

# Add the local bin directory to PATH
ENV PATH="/root/.local/bin:${PATH}"

# Create necessary directories for Goose
RUN mkdir -p /root/.config/goose

# Verify installation
RUN goose --version

ENTRYPOINT ["goose"]
CMD ["session"]