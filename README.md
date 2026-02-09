# DVD Rental Microservices

A microservices-based DVD rental system built with Go, gRPC, and PostgreSQL.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[中文文档](README.zh-CN.md)** | **[日本語ドキュメント](README.ja.md)**

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                            Clients                                  │
│                  ┌──────────────┬──────────────┐                   │
│                  │  Customer    │    Admin     │                   │
│                  │    App       │   Console    │                   │
│                  └──────────────┴──────────────┘                   │
└─────────────────────────────────────────────────────────────────────┘
                         │                 │
                         ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         BFF Layer (REST)                           │
│            ┌────────────────┐    ┌────────────────┐                │
│            │ customer-bff   │    │   admin-bff    │                │
│            │    (:8080)     │    │    (:8081)     │                │
│            │   17 endpoints │    │   50 endpoints │                │
│            └────────────────┘    └────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
                         │                 │
                         ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      Core Services (gRPC)                          │
│  ┌────────┐ ┌────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐        │
│  │ store  │ │  film  │ │ customer │ │ rental │ │ payment │        │
│  │ :50051 │ │ :50052 │ │  :50053  │ │ :50054 │ │  :50055 │        │
│  └────────┘ └────────┘ └──────────┘ └────────┘ └─────────┘        │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Data Layer                                  │
│            ┌────────────────┐    ┌────────────────┐                │
│            │   PostgreSQL   │    │     Redis      │                │
│            │   (15 tables)  │    │ (Token Store)  │                │
│            └────────────────┘    └────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
```

## Features

- **5 gRPC Microservices**: Store, Film, Customer, Rental, Payment
- **2 BFF Services**: Customer-facing and Admin REST APIs
- **JWT Authentication**: Dual token (Access + Refresh) with Redis storage
- **Full-text Search**: PostgreSQL tsvector for film search
- **Partitioned Tables**: Monthly payment partitions
- **Docker Ready**: Multi-stage Alpine builds, docker-compose for local dev

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| gRPC | Protocol Buffers v3 |
| HTTP | Standard library `net/http` |
| Database | PostgreSQL 17, sqlc |
| Cache | Redis 7 |
| Auth | JWT (golang-jwt/jwt/v5) |
| Build | Docker, docker-compose |

## Quick Start

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- buf (for proto generation)
- sqlc (for SQL code generation)

### Run with Docker

```bash
# Start all services
docker compose -f deployments/docker-compose.yml up -d

# View logs
make infra-logs

# Check status
make infra-ps
```

### Local Development

```bash
# Start infrastructure only
make infra-up

# Run gRPC services (each in a separate terminal)
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-store
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-film
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-customer
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-rental
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-payment

# Run BFF services
JWT_SECRET="your-secret" REDIS_URL="redis://localhost:6379" make run-customer-bff
JWT_SECRET="your-secret" REDIS_URL="redis://localhost:6379" make run-admin-bff
```

## Project Structure

```
dvd-rental/
├── cmd/                          # Service entry points
│   ├── store-service/
│   ├── film-service/
│   ├── customer-service/
│   ├── rental-service/
│   ├── payment-service/
│   ├── customer-bff/
│   └── admin-bff/
├── internal/                     # Private application code
│   ├── store/                    # Store service implementation
│   ├── film/                     # Film service implementation
│   ├── customer/                 # Customer service implementation
│   ├── rental/                   # Rental service implementation
│   ├── payment/                  # Payment service implementation
│   └── bff/                      # BFF implementations
│       ├── customer/
│       └── admin/
├── pkg/                          # Shared packages
│   ├── auth/                     # JWT & password utilities
│   ├── middleware/               # HTTP middleware
│   └── grpcutil/                 # gRPC client utilities
├── proto/                        # Protocol buffer definitions
├── gen/                          # Generated code (proto + sqlc)
├── migrations/                   # Database migrations
├── deployments/                  # Docker & compose files
└── docs/                         # Documentation
```

## API Endpoints

### Customer BFF (Port 8080)

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/auth/login` | POST | - | Customer login |
| `/api/v1/films` | GET | - | List/search films |
| `/api/v1/rentals` | GET/POST | JWT | Manage rentals |
| `/api/v1/payments` | GET | JWT | View payments |
| `/api/v1/profile` | GET/PUT | JWT | Manage profile |

### Admin BFF (Port 8081)

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/v1/auth/login` | POST | - | Staff login |
| `/api/v1/stores` | CRUD | JWT | Manage stores |
| `/api/v1/staff` | CRUD | JWT | Manage staff |
| `/api/v1/customers` | CRUD | JWT | Manage customers |
| `/api/v1/films` | CRUD | JWT | Manage films |
| `/api/v1/inventory` | CRUD | JWT | Manage inventory |
| `/api/v1/rentals` | CRUD | JWT | Manage rentals |
| `/api/v1/payments` | CRUD | JWT | Manage payments |

## Development

```bash
# Generate protobuf code
make proto-gen

# Generate sqlc code
make sqlc-gen

# Build all services
make build-all

# Run tests
make test

# Format code
make fmt

# Run linter
make lint
```

## Configuration

All services are configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `GRPC_PORT` | gRPC server port | Service-specific |
| `HTTP_PORT` | HTTP server port | 8080/8081 |
| `JWT_SECRET` | JWT signing secret | Required (BFF) |
| `REDIS_URL` | Redis connection URL | Required (BFF) |
| `LOG_LEVEL` | Logging level | info |

## License

MIT License - see [LICENSE](LICENSE) for details.
