FROM golang:1.21-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl jq tree ripgrep fd-find \
    && rm -rf /var/lib/apt/lists/*

# fd-find installs as fdfind, symlink to fd
RUN ln -s /usr/bin/fdfind /usr/local/bin/fd

WORKDIR /workspace
