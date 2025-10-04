# dbxgo

## Project Overview

dbxgo is an efficient Go-based Change Data Capture (CDC) tool that real-time monitors database changes, parses and processes events, and sends them to message queues or other downstream systems.

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

```bash
# Clone the repository
git clone https://github.com/chihqiang/dbxgo.git
cd dbxgo

# Build the project
make build

# Or use Go command directly
go build -o dbxgo ./cmd/dbxgo/
```

### Usage Example

1. Create configuration file `config.yml`: (see Configuration File Description section for details)

2. Run the program:

```bash
# Using the default config.yml
./dbxgo

# Using a specific config file
./dbxgo -c path/to/config.yml

# Explicitly using the listen command
./dbxgo listen -c path/to/config.yml
```

## Configuration File Description

The configuration file uses YAML format and consists of three main parts: `store` (offset storage), `source` (data source), and `output` (output destination).

### Example Configuration

```yaml
# ---------- Offset Storage Configuration ----------
store:
   type: file                # Storage type: file / redis
   file:
      dir: "runtime"        # Directory for storing offset files
   redis:
      addr: "127.0.0.1:6379" # Redis address
      password: "123456"     # Redis password (leave empty if none)
      db: 0                  # Redis database number (default 0)

# ---------- Data Source Configuration ----------
source:
   type: "mysql"             # Data source type: mysql
   mysql:
      addr: "127.0.0.1:3306" # Database address (host:port)
      user: "root"           # Database username (recommended to use a dedicated account in production)
      password: "123456"     # Database password
      exclude_table_regex:   # Tables to exclude (regex patterns)
         - "mysql.*"
         - "information_schema.*"
         - "performance_schema.*"
         - "sys.*"
      include_table_regex:   # Tables to include (regex patterns, empty = all except excluded)
      # - "test_db.users"    # Example: only listen to "users" table in "test_db"

# ---------- Output Configuration ----------
output:
   type: stdout              # Output type: stdout / kafka / redis / rabbitmq / rocketmq

   # Kafka settings
   kafka:
      brokers:
         - "127.0.0.1:9092"  # Kafka broker list
      topic: "dbxgo_events"  # Kafka topic name

   # RabbitMQ settings
   rabbitmq:
      url: "amqp://guest:guest@127.0.0.1:5672/" # RabbitMQ connection URL
      queue: "dbxgo_queue"   # Queue name
      durable: true          # Whether the queue should survive server restarts
      auto_ack: false        # Whether to auto-acknowledge messages
      exclusive: false       # Whether the queue is exclusive to this connection
      no_wait: false         # Whether to wait for the server to confirm queue declaration

   # Redis settings
   redis:
      addr: "127.0.0.1:6379" # Redis address
      password: "123456"     # Redis password
      db: 0                  # Redis database number
      key: "dbxgo_events"    # Redis key for storing events

   # RocketMQ settings
   rocketmq:
      servers:
         - "127.0.0.1:9876"   # RocketMQ NameServer address
      topic: "dbxgo_events"  # RocketMQ topic name
      group: "dbxgo_group"   # Producer group name
      namespace: "test"      # Namespace
      access_key: "RocketMQ" # Access key
      secret_key: "12345678" # Secret key
   # pulsar settings
   pulsar:
      url: "pulsar://127.0.0.1:6650"   # Pulsar broker URL
      topic: "dbxgo_events"            # Pulsar topic name
      token: "YOUR_AUTH_TOKEN"         # Optional authentication token
      operation_timeout: 30            # Operation timeout in seconds
      connection_timeout: 30           # Connection timeout in seconds
```

## Notes

1. **MySQL Configuration Requirements**:
   - Binary logging must be enabled (`log-bin=ON`)
   - Server ID must be set (`server-id=1`)
   - Binlog format should be `ROW` (`binlog_format=ROW`)
2. **Permission Requirements**: When using MySQL data source, ensure the database user has sufficient permissions:

```sql
-- Create an account
CREATE USER 'dbxgo'@'%' IDENTIFIED BY 'strong_password';

-- Authorization (the REPLICATION permission is required to read the binlog)
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'dbxgo'@'%';

-- If dbxgo needs to do metadata queries, it also needs read permissions
GRANT SELECT ON *.* TO 'dbxgo'@'%';

-- Refresh permissions
FLUSH PRIVILEGES;
```
