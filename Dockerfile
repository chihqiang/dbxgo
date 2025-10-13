# ---------- Build stage ----------
# Use Go Alpine image as builder for smaller image and faster build
FROM golang:1.23-alpine AS builder

# Set build-time argument for version
ARG DBXGO_VERSION=main

# Install dependencies for building Go project
RUN apk add --no-cache git make gcc musl-dev

# Set working directory inside container
WORKDIR /app

# Copy project files to container
COPY . .

# Build the Go binary with the specified version
RUN GOOS=linux make build VERSION=${DBXGO_VERSION}

# ---------- Runtime stage ----------
# Use minimal Debian image for runtime
FROM debian:bookworm-slim

# Set default timezone to Shanghai
ENV TZ=Asia/Shanghai

# Install runtime dependencies in a single layer
# - ca-certificates, curl, wget, jq, tzdata, mysql client
# - configure timezone
# - clean apt cache to reduce image size
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl wget jq tzdata default-mysql-client && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/*


# ==========================
# Store Configuration
# ==========================
ENV STORE_TYPE=file

# ==========================
# Source Configuration
# ==========================
ENV SOURCE_TYPE=mysql
ENV SOURCE_MYSQL_ADDR=127.0.0.1:3306
ENV SOURCE_MYSQL_USER=root
ENV SOURCE_MYSQL_PASSWORD=""
ENV SOURCE_MYSQL_INCLUDE_TABLE_REGEX=""
ENV SOURCE_MYSQL_EXCLUDE_TABLE_REGEX="mysql.*,information_schema.*,performance_schema.*,sys.*"

# ==========================
# Output Configuration
# ==========================
ENV OUTPUT_TYPE=stdout


# Create a non-root user for security
RUN useradd --system --no-create-home --shell /usr/sbin/nologin dbxgo

# Copy the built binary from builder stage
COPY --from=builder /app/dbxgo /usr/local/bin/dbxgo

# Set ownership to the non-root user
RUN chown dbxgo:dbxgo /usr/local/bin/dbxgo

# Switch to non-root user
USER dbxgo

# Set working directory
WORKDIR /app

# Default command to run the binary
CMD ["dbxgo","-c","/app/config.yaml"]