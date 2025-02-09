FROM rust:1.75-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    pkg-config \
    libssl-dev \
    git \
    bzip2 \
    cmake \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Clone Goose repository
RUN git clone https://github.com/block/goose.git .

# Initialize git submodules if any
RUN git submodule update --init --recursive

# Set environment variable for verbose Cargo output
ENV RUST_BACKTRACE=1
ENV CARGO_TERM_VERBOSE=true

# Build Goose (with verbose output)
RUN cargo build --release -vv

# Create necessary directories for Goose
RUN mkdir -p /root/.config/goose

# Add the binary to PATH
ENV PATH="/app/target/release:${PATH}"

# Command to run Goose
CMD ["goose", "session"]