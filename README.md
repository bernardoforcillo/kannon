# Kannon 💥

[![CI](https://github.com/gyozatech/kannon/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/gyozatech/kannon/actions/workflows/ci.yaml)

![Kannon Logo](assets/kannonlogo.png?raw=true)

A **Cloud Native SMTP mail sender** for Kubernetes and modern infrastructure.

> [!NOTE]
> Due to limitations of AWS, GCP, etc. on port 25, this project will not work on cloud providers that block port 25.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Quickstart](#quickstart)
- [Configuration](#configuration)
- [Database Schema](#database-schema)
- [API Overview](#api-overview)
- [Deployment](#deployment)
- [Domain & DNS Setup](#domain--dns-setup)
- [Sending Mail](#sending-mail)
- [Development & Contributing](#development--contributing)
- [License](#license)

---

## Features

- Cloud-native, scalable SMTP mail sending
- gRPC API for sending HTML and templated emails
- DKIM and SPF support for deliverability
- Attachments support (JSONB in DB)
- Custom fields for emails (JSONB in DB)
- Statistics and analytics (DB + gRPC API)
- Template management (CRUD via API)
- Kubernetes-ready deployment
- Postgres-backed persistence

**Planned:**

- Multi-node sending
- Advanced analytics dashboard

## Architecture

Kannon is composed of several microservices and workers:

- **API**: gRPC server for mail, template, and domain management
- **SMTP**: Handles SMTP protocol and relays mail
- **Sender**: Sends emails from the queue
- **Dispatcher**: Manages the sending pool and delivery
- **Verifier**: Validates emails before sending
- **Bounce**: Handles bounces
- **Stats**: Collects and stores delivery statistics

All components can be enabled/disabled via CLI flags or config.

> **See [`ARCHITECTURE.md`](./ARCHITECTURE.md) for a full breakdown of modules, NATS streams, topics, consumers, and message flows.**

```mermaid
flowchart TD
    subgraph Core
        API["gRPC API"]
        SMTP["SMTP Server"]
        Sender["Sender"]
        Dispatcher["Dispatcher"]
        Verifier["Verifier"]
        Bounce["Bounce"]
        Stats["Stats"]
    end
    DB[(PostgreSQL)]
    NATS[(NATS)]
    API <--> DB
    Dispatcher <--> DB
    Sender <--> DB
    Verifier <--> DB
    Stats <--> DB
    Bounce <--> DB
    API <--> NATS
    Sender <--> NATS
    Dispatcher <--> NATS
    SMTP <--> NATS
    Stats <--> NATS
    Verifier <--> NATS
    Bounce <--> NATS
```

## Quickstart

### Prerequisites

- Go 1.24+
- Docker (optional, for containerized deployment)
- PostgreSQL database
- NATS server (for internal messaging)

### Local Run (for development)

```sh
git clone https://github.com/kannon-email/kannon.git
cd kannon
go build -o kannon .
./kannon --run-api --run-smtp --run-sender --run-dispatcher --config ./config.yaml
```

### Docker Compose

See [`examples/docker-compose/`](examples/docker-compose/) for ready-to-use files.

```sh
docker-compose -f examples/docker-compose/docker-compose.yaml up
```

- Edit `examples/docker-compose/kannon.yaml` to configure your environment.

### Makefile Targets

- `make test` — Run all tests
- `make generate` — Generate DB and proto code
- `make lint` — Run linters

## Configuration

Kannon can be configured via YAML file, environment variables, or CLI flags. Precedence: CLI > Env > YAML.

**Main config options:**

| Key / Env Var                                   | Type     | Default        | Description                            |
| ----------------------------------------------- | -------- | -------------- | -------------------------------------- |
| `database_url` / `K_DATABASE_URL`               | string   | (required)     | PostgreSQL connection string           |
| `nats_url` / `K_NATS_URL`                       | string   | (required)     | NATS server URL for internal messaging |
| `debug` / `K_DEBUG`                             | bool     | false          | Enable debug logging                   |
| `api.port` / `K_API_PORT`                       | int      | 50051          | gRPC API port                          |
| `sender.hostname` / `K_SENDER_HOSTNAME`         | string   | (required)     | Hostname for outgoing mail             |
| `sender.max_jobs` / `K_SENDER_MAX_JOBS`         | int      | 10             | Max parallel sending jobs              |
| `smtp.address` / `K_SMTP_ADDRESS`               | string   | :25            | SMTP server listen address             |
| `smtp.domain` / `K_SMTP_DOMAIN`                 | string   | localhost      | SMTP server domain                     |
| `smtp.read_timeout` / `K_SMTP_READ_TIMEOUT`     | duration | 10s            | SMTP read timeout                      |
| `smtp.write_timeout` / `K_SMTP_WRITE_TIMEOUT`   | duration | 10s            | SMTP write timeout                     |
| `smtp.max_payload` / `K_SMTP_MAX_PAYLOAD`       | size     | 1024kb         | Max SMTP message size                  |
| `smtp.max_recipients` / `K_SMTP_MAX_RECIPIENTS` | int      | 50             | Max recipients per SMTP message        |
| `run-api` / `K_RUN_API`                         | bool     | false          | Enable API server                      |
| `run-smtp` / `K_RUN_SMTP`                       | bool     | false          | Enable SMTP server                     |
| `run-sender` / `K_RUN_SENDER`                   | bool     | false          | Enable sender worker                   |
| `run-dispatcher` / `K_RUN_DISPATCHER`           | bool     | false          | Enable dispatcher worker               |
| `run-verifier` / `K_RUN_VERIFIER`               | bool     | false          | Enable verifier worker                 |
| `run-bounce` / `K_RUN_BOUNCE`                   | bool     | false          | Enable bounce worker                   |
| `run-stats` / `K_RUN_STATS`                     | bool     | false          | Enable stats worker                    |
| `config`                                        | string   | ~/.kannon.yaml | Path to config file                    |

- See [`examples/docker-compose/kannon.yaml`](examples/docker-compose/kannon.yaml) for a full example.

## Database Schema

Kannon requires a PostgreSQL database. Main tables:

- **domains**: Registered sender domains, DKIM keys
- **messages**: Outgoing messages, subject, sender, template, attachments
- **sending_pool_emails**: Email queue, status, scheduling, custom fields
- **templates**: Email templates, type, metadata
- **stats**: Delivery and open/click statistics
- **stats_keys**: Public/private keys for stats security

See [`db/migrations/`](./db/migrations/) for full schema and migrations.

## API Overview

Kannon exposes a gRPC API for sending mail, managing domains/templates, and retrieving stats.

### Services & Methods

- **Mailer API** ([proto](./proto/kannon/mailer/apiv1/mailerapiv1.proto))
  - `SendHTML`: Send a raw HTML email
  - `SendTemplate`: Send an email using a stored template
- **Admin API** ([proto](./proto/kannon/admin/apiv1/adminapiv1.proto))
  - `GetDomains`, `GetDomain`, `CreateDomain`, `RegenerateDomainKey`
  - `CreateTemplate`, `UpdateTemplate`, `DeleteTemplate`, `GetTemplate`, `GetTemplates`
- **Stats API** ([proto](./proto/kannon/stats/apiv1/statsapiv1.proto))
  - `GetStats`, `GetStatsAggregated`

### Authentication

All gRPC APIs use Basic Auth with your domain and API key:

```
token = base64(<your domain>:<your domain key>)
```

Pass this in the `Authorization` metadata for gRPC calls:

```json
{
  "Authorization": "Basic <your token>"
}
```

### Example: SendHTML Request

```json
{
  "sender": {
    "email": "no-reply@yourdomain.com",
    "alias": "Your Name"
  },
  "recipients": ["user@example.com"],
  "subject": "Test",
  "html": "<h1>Hello</h1><p>This is a test.</p>",
  "attachments": [
    { "filename": "file.txt", "content": "base64-encoded-content" }
  ],
  "fields": { "custom": "value" }
}
```

See the [proto files](./proto/kannon/) for all fields and options.

## Deployment

### Kubernetes

- See [`k8s/deployment.yaml`](k8s/deployment.yaml) for a production-ready manifest.
- Configure your environment via a mounted YAML file or environment variables.

### Docker Compose

- See [`examples/docker-compose/`](examples/docker-compose/) for local or test deployments.

## Domain & DNS Setup

To send mail, you must register a sender domain and configure DNS:

1. Register a domain via the Admin API
2. Set up DNS records:
   - **A record**: `<SENDER_NAME>` → your server IP
   - **Reverse DNS**: your server IP → `<SENDER_NAME>`
   - **SPF TXT**: `<SENDER_NAME>` → `v=spf1 ip4:<YOUR SENDER IP> -all`
   - **DKIM TXT**: `smtp._domainkey.<YOUR_DOMAIN>` → `k=rsa; p=<YOUR DKIM KEY HERE>`

## Sending Mail

- Use the gRPC API (`SendHTML` or `SendTemplate`) with Basic Auth as above.
- See the [proto file](./proto/kannon/mailer/apiv1/mailerapiv1.proto) for all fields and options.

## Development & Contributing

We welcome contributions! Please:

- Use [feature request](.github/ISSUE_TEMPLATE/feature_request.md) and [bug report](.github/ISSUE_TEMPLATE/bug_report.md) templates for issues
- Follow the [pull request template](.github/PULL_REQUEST_TEMPLATE.md)
- See the [Apache 2.0 License](./LICENSE)
- **Read our [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines, code style, and the full contribution process.**

### Local Development

- Build: `go build -o kannon .`
- Test: `make test` or `go test ./...`
- Generate code: `make generate`
- Lint: `make lint`

### Testing

- Unit and integration tests are in `internal/`, `pkg/`, and run with `go test ./...`
- Some tests use Docker to spin up a test Postgres instance

---

## License

Kannon is licensed under the Apache 2.0 License. See [LICENSE](./LICENSE) for details.
