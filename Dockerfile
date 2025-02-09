FROM rust:1.75-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    pkg-config \
    libssl-dev \
    git \
    bzip2 \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Clone Goose repository
RUN git clone https://github.com/block/goose.git .

# Install dependencies and build
RUN cargo build --release

# Add the binary to PATH
ENV PATH="/app/target/release:${PATH}"

# Create necessary directories for Goose
RUN mkdir -p /root/.config/goose

# Command to run Goose
CMD ["goose", "session"]