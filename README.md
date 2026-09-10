<div align="center">
<h1>cdc-trigger</h1>

[![Auth](https://img.shields.io/badge/Auth-chihqiang-ff69b4)](https://github.com/chihqiang)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/chihqiang/cdc-trigger)](https://github.com/chihqiang/cdc-trigger/pulls)
[![Release](https://img.shields.io/github/release/chihqiang/cdc-trigger.svg?style=flat-square)](https://github.com/chihqiang/cdc-trigger/releases)
[![GitHub Pull Requests](https://img.shields.io/github/stars/chihqiang/cdc-trigger)](https://github.com/chihqiang/cdc-trigger/stargazers)
[![HitCount](https://views.whatilearened.today/views/github/chihqiang/cdc-trigger.svg)](https://github.com/chihqiang/cdc-trigger)
[![GitHub license](https://img.shields.io/github/license/chihqiang/cdc-trigger)](https://github.com/chihqiang/cdc-trigger/blob/main/LICENSE)

<p>
 cdc-trigger is an efficient Go-based Change Data Capture (CDC) tool that real-time monitors database changes, parses and processes events, and sends them to message queues or other downstream systems.
</p>

</div>

## Project Overview

## Features

- **Real-time Capture**: Monitor database change events in real-time through binlog parsing
- **Unified Event Format**: Convert changes from different databases into a consistent JSON format
- **Multiple Output Support**: Send events to various downstream systems including stdout, Redis, Kafka, RabbitMQ, and RocketMQ
- **Checkpoint Resumption**: Store synchronization positions to achieve breakpoint resumption
- **Extensible Architecture**: Easy to extend with new data sources and output types
- **Worker Pool Processing**: Process events efficiently with worker goroutines
- **Graceful Shutdown**: Properly handle context cancellation and resource cleanup

## Supported Components

### Data Sources

- MySQL (via binlog parsing)

### Outputs

- Standard Output (stdout)
- [Redis](https://redis.io/)
- [Kafka](https://kafka.apache.org/)
- [RabbitMQ](https://www.rabbitmq.com/)
- [RocketMQ](https://rocketmq.apache.org/)
- [Pulsar](https://pulsar.apache.org/)

### Storage

- File Storage
- Redis Storage

## Quick Start

### Prerequisites

- Go 1.23+ environment
- MySQL server with binary logging enabled
- Correct database access permissions (MySQL user needs binlog read permissions)
- If using other output components, ensure corresponding services are available

### Installation

### Prerequisites

- Go **1.23+** (latest version recommended)
- `$GOPATH/bin` added to your `$PATH`

### Option 1: Install via `go install` (Recommended)

This is the simplest and recommended way to install **cdc-trigger**:

```bash
go install github.com/chihqiang/cdc-trigger/cmd/cdc-trigger@latest
```

### Option 2: Build from Source

If you want to modify the source code or contribute to development, build from source:

```bash
# Clone the repository
git clone https://github.com/chihqiang/cdc-trigger.git
cd cdc-trigger && make build
cp ./cdc-trigger /usr/local/bin/
```

### Usage Example

1. Create configuration file `config.yml`: (see Configuration File Description section for details)

2. Run the program:

```bash
# Using the default config.yml
cdc-trigger
# Using a specific config file
cdc-trigger -c path/to/config.yml
# Explicitly using the listen command
cdc-trigger listen -c path/to/config.yml
```

## Configuration File Description

The configuration file uses YAML format and consists of three main parts: `store` (offset storage), `source` (data source), and `output` (output destination).

### Example Configuration

```yaml
# ==========================================
# cdc-trigger Configuration File Example (YAML)
# ==========================================

# ---------- Offset Storage Configuration ----------
store:
  type: "file"                # Storage type: file / redis

  file:
    dir: "runtime"            # Directory for storing offset files

  redis:
    addr: "127.0.0.1:6379"   # Redis address
    password: ""              # Redis password (leave empty if none)
    db: 0                     # Redis database number (default 0)

# ---------- Data Source Configuration ----------
source:
  type: "mysql"               # Data source type: mysql

  mysql:
    addr: "127.0.0.1:3306"   # Database address (host:port)
    user: "root"              # Database username (recommended to use a dedicated account in production)
    password: ""              # Database password
    exclude_table_regex:      # Tables to exclude (regex patterns)
      - "mysql.*"
      - "information_schema.*"
      - "performance_schema.*"
      - "sys.*"
    include_table_regex:      # Tables to include (regex patterns, empty = all except excluded)
      - "cdctrigger.*"        # Example: only listen to cdctrigger tables

# ---------- Output Configuration ----------
output:
  type: "stdout"              # Output type: stdout / kafka / redis / rabbitmq / rocketmq / pulsar

  # Kafka settings
  kafka:
    brokers:
      - "127.0.0.1:9092"      # Kafka broker list
    topic: "cdc-trigger-events"  # Kafka topic name

  # RabbitMQ settings
  rabbitmq:
    url: "amqp://guest:guest@127.0.0.1:5672/" # RabbitMQ connection URL
    exchange: "cdc-trigger-exchange" # Exchange name
    queue: "cdc-trigger-events"      # Queue name
    durable: true              # Whether the queue should survive server restarts
    auto_delete: false         # Whether the queue should auto-delete when unused
    auto_ack: false            # Whether to auto-acknowledge messages
    exclusive: false           # Whether the queue is exclusive to this connection
    no_wait: false             # Whether to wait for the server to confirm queue declaration

  # Redis settings
  redis:
    addr: "127.0.0.1:6379"     # Redis address
    password: ""               # Redis password
    db: 0                      # Redis database number
    key: "cdc-trigger-events"  # Redis key for storing events

  # RocketMQ settings
  rocketmq:
    servers:
      - "127.0.0.1:9876"       # RocketMQ NameServer address
    topic: "cdc-trigger-events"  # RocketMQ topic name
    group: "cdc-trigger-group"   # Producer group name
    namespace: ""              # Namespace
    access_key: ""             # Access key
    secret_key: ""             # Secret key
    retry: 3                   # Retry count on failure

  # Pulsar settings
  pulsar:
    url: "pulsar://127.0.0.1:6650"  # Pulsar broker URL
    topic: "cdc-trigger-events"       # Pulsar topic name
    token: "YOUR_PULSAR_TOKEN"      # Optional authentication token
    operation_timeout: 30           # Operation timeout in seconds
    connection_timeout: 30          # Connection timeout in seconds
```

## Docker Deployment

You can use Docker to run cdc-trigger in containerized environments. Here's how to build and run cdc-trigger with Docker:

```bash
# =========================
# 1️⃣ MySQL only (read from MySQL)
# =========================
docker run -it --rm \
    --name cdc-trigger \
    -e SOURCE_MYSQL_ADDR="127.0.0.1:3306" \
    -e SOURCE_MYSQL_USER="root" \
    -e SOURCE_MYSQL_PASSWORD="123456" \
    zhiqiangwang/app:cdc-trigger

# =========================
# 2️⃣ MySQL → Redis & Redis
# =========================
docker run -it --rm \
    --name cdc-trigger \
    -e SOURCE_MYSQL_ADDR="127.0.0.1:3306" \
    -e SOURCE_MYSQL_USER="root" \
    -e SOURCE_MYSQL_PASSWORD="123456" \
    -e STORE_TYPE="redis" \
    -e STORE_REDIS_ADDR="127.0.0.1:6379" \
    -e STORE_REDIS_PASSWORD="123456" \
    -e STORE_REDIS_DB="1" \
    -e OUTPUT_TYPE="redis" \
    -e OUTPUT_REDIS_ADDR="127.0.0.1:6379" \
    -e OUTPUT_REDIS_PASSWORD="123456" \
    -e OUTPUT_REDIS_DB="1" \
    -e OUTPUT_REDIS_KEY="cdc-trigger-events" \
    zhiqiangwang/app:cdc-trigger
```

## Notes

1. **MySQL Configuration Requirements**:
   - Binary logging must be enabled (`log-bin=ON`)
   - Server ID must be set (`server-id=1`)
   - Binlog format should be `ROW` (`binlog_format=ROW`)
2. **Permission Requirements**: When using MySQL data source, ensure the database user has sufficient permissions:

```sql
-- Create an account
CREATE USER 'cdctrigger'@'%' IDENTIFIED BY 'strong_password';

-- Authorization (the REPLICATION permission is required to read the binlog)
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'cdctrigger'@'%';

-- If cdc-trigger needs to do metadata queries, it also needs read permissions
GRANT SELECT ON *.* TO 'cdctrigger'@'%';

-- Refresh permissions
FLUSH PRIVILEGES;
```
