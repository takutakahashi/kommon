FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o kommon

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y curl

# Install system dependencies and GitHub CLI repository
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update && apt-get install -y \
    ca-certificates \
    bzip2 \
    libxcb1 \
    libdbus-1-3 \
    git \
    gh \
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

COPY assets/config.yaml /root/.config/goose/
COPY assets/.goosehints /root/.config/goose/

# Verify installation
RUN goose --version

ENTRYPOINT ["kommon"]
