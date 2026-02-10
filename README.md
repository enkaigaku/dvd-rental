# DVD Rental Microservices

A microservices-based DVD rental system built with Go, gRPC, and PostgreSQL.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[中文文档](README_zh.md)** | **[日本語ドキュメント](README_ja.md)**

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
- **2 BFF Services**: Customer-facing (17 endpoints) and Admin (50 endpoints) REST APIs
- **JWT Authentication**: Dual token (Access 15min + Refresh 7d) with Redis storage
- **Full-text Search**: PostgreSQL tsvector for film search
- **Partitioned Tables**: Monthly payment partitions
- **Docker Ready**: Multi-stage Alpine builds, docker-compose for local dev

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| gRPC | Protocol Buffers v3, buf |
| HTTP | Standard library `net/http` (Go 1.22+ ServeMux) |
| Database | PostgreSQL 17, sqlc |
| Cache | Redis 7 |
| Auth | JWT (golang-jwt/jwt/v5), bcrypt |
| Build | Docker, docker-compose |

## Prerequisites

| Tool | Purpose | Install |
|------|---------|---------|
| Go 1.25+ | Language runtime | [go.dev/dl](https://go.dev/dl/) |
| Docker & Docker Compose | Container runtime | [docs.docker.com](https://docs.docker.com/get-docker/) |
| buf | Protobuf code generation | `go install github.com/bufbuild/buf/cmd/buf@latest` |
| sqlc | SQL code generation | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |

## Getting Started

### Option A: Run Everything with Docker (Quickest)

This builds and starts all 7 services + PostgreSQL + Redis in containers.

```bash
# Clone the repository
git clone https://github.com/enkaigaku/dvd-rental.git
cd dvd-rental

# Start all services
make up

# Verify all services are running
make ps

# View logs
make logs
```

Once started, the APIs are available at:
- Customer BFF: http://localhost:8080
- Admin BFF: http://localhost:8081

### Option B: Local Development (Recommended for Dev)

Run infrastructure in Docker, services locally for fast iteration.

#### Step 1: Start Infrastructure

```bash
# Start PostgreSQL and Redis only
# Database is auto-initialized with schema + sample data via migrations/
make infra-up
```

This starts:
- PostgreSQL on `localhost:5432` (user: `dvdrental`, password: `dvdrental`, db: `dvdrental`)
- Redis on `localhost:6379`

#### Step 2: Install Go Dependencies

```bash
go mod download
```

#### Step 3: Start gRPC Services

Open 5 separate terminals and run one service each:

```bash
# Terminal 1 - Store Service
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-store

# Terminal 2 - Film Service
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-film

# Terminal 3 - Customer Service
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-customer

# Terminal 4 - Rental Service
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-rental

# Terminal 5 - Payment Service
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-payment
```

#### Step 4: Start BFF Services

Open 2 more terminals:

```bash
# Terminal 6 - Customer BFF
JWT_SECRET="dev-secret-change-me" REDIS_URL="redis://localhost:6379" make run-customer-bff

# Terminal 7 - Admin BFF
JWT_SECRET="dev-secret-change-me" REDIS_URL="redis://localhost:6379" make run-admin-bff
```

#### Step 5: Verify

```bash
# List films (no auth required)
curl http://localhost:8080/api/v1/films

# Customer login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "mary.smith@sakilacustomer.org", "password": "password"}'
```

### Stopping Services

```bash
# Stop Docker infrastructure (local dev)
make infra-down

# Or stop everything (if using Option A)
make down
```

## Project Structure

```
dvd-rental/
├── cmd/                          # Service entry points
│   ├── store-service/            #   Store gRPC server
│   ├── film-service/             #   Film gRPC server
│   ├── customer-service/         #   Customer gRPC server
│   ├── rental-service/           #   Rental gRPC server
│   ├── payment-service/          #   Payment gRPC server
│   ├── customer-bff/             #   Customer-facing REST API
│   └── admin-bff/                #   Admin REST API
├── internal/                     # Private application code
│   ├── store/                    #   handler → service → repository
│   ├── film/                     #   handler → service → repository
│   ├── customer/                 #   handler → service → repository
│   ├── rental/                   #   handler → service → repository
│   ├── payment/                  #   handler → service → repository
│   └── bff/                      #   BFF layer (HTTP → gRPC translation)
│       ├── customer/             #     config + handler + router
│       └── admin/                #     config + handler + router
├── pkg/                          # Shared packages
│   ├── auth/                     #   JWT, bcrypt, refresh token store
│   ├── middleware/               #   Auth, CORS, logging, recovery, response
│   └── grpcutil/                 #   gRPC client dial helper
├── proto/                        # Protocol Buffer definitions (.proto)
├── gen/                          # Generated code
│   ├── proto/                    #   Generated gRPC/protobuf Go code
│   └── sqlc/                     #   Generated sqlc Go code
├── sql/                          # SQL query definitions (for sqlc)
├── migrations/                   # Database schema + seed data
│   ├── 001_initial_schema.sql    #   Core table definitions
│   ├── 002_pagila_schema.sql     #   Pagila schema extensions
│   ├── 003_pagila_data.sql       #   Sample data
│   ├── 004_customer_auth.sql     #   Customer password hashes
│   └── 005_staff_auth.sql        #   Staff password hashes
├── deployments/                  # Docker & compose files
│   ├── docker-compose.yml
│   └── dockerfiles/              #   Per-service Dockerfiles
├── buf.yaml                      # Buf configuration
├── buf.gen.yaml                  # Buf code generation config
├── sqlc.yaml                     # sqlc configuration
├── Makefile                      # Development commands
└── go.mod
```

## API Endpoints

### Customer BFF (Port 8080)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/login` | - | Customer login |
| POST | `/api/v1/auth/refresh` | - | Refresh token |
| POST | `/api/v1/auth/logout` | - | Logout |
| GET | `/api/v1/films` | - | List films |
| GET | `/api/v1/films/search` | - | Search films |
| GET | `/api/v1/films/category/{id}` | - | Films by category |
| GET | `/api/v1/films/actor/{id}` | - | Films by actor |
| GET | `/api/v1/films/{id}` | - | Film detail |
| GET | `/api/v1/categories` | - | List categories |
| GET | `/api/v1/actors` | - | List actors |
| GET | `/api/v1/rentals` | JWT | My rentals |
| GET | `/api/v1/rentals/{id}` | JWT | Rental detail |
| POST | `/api/v1/rentals` | JWT | Create rental |
| POST | `/api/v1/rentals/{id}/return` | JWT | Return rental |
| GET | `/api/v1/payments` | JWT | My payments |
| GET | `/api/v1/profile` | JWT | My profile |
| PUT | `/api/v1/profile` | JWT | Update profile |

### Admin BFF (Port 8081)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/login` | - | Staff login |
| POST | `/api/v1/auth/refresh` | - | Refresh token |
| POST | `/api/v1/auth/logout` | - | Logout |
| | `/api/v1/stores/**` | JWT | Store management (CRUD) |
| | `/api/v1/staff/**` | JWT | Staff management (CRUD) |
| | `/api/v1/customers/**` | JWT | Customer management (CRUD) |
| | `/api/v1/films/**` | JWT | Film management (CRUD) |
| | `/api/v1/actors/**` | JWT | Actor management (CRUD) |
| | `/api/v1/categories/**` | JWT | Category management (CRUD) |
| | `/api/v1/languages/**` | JWT | Language management (CRUD) |
| | `/api/v1/inventory/**` | JWT | Inventory management (CRUD) |
| | `/api/v1/rentals/**` | JWT | Rental management (CRUD) |
| | `/api/v1/payments/**` | JWT | Payment management (CRUD) |

## Environment Variables

### gRPC Services (store, film, customer, rental, payment)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `GRPC_PORT` | No | Service-specific | gRPC listen port (50051-50055) |
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |

### BFF Services (customer-bff, admin-bff)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_PORT` | No | `8080` / `8081` | HTTP listen port |
| `JWT_SECRET` | Yes | - | Secret key for JWT signing |
| `JWT_ACCESS_DURATION` | No | `15m` | Access token TTL |
| `JWT_REFRESH_DURATION` | No | `168h` (7d) | Refresh token TTL |
| `REDIS_URL` | Yes | - | Redis URL for refresh token storage |
| `GRPC_STORE_ADDR` | No | `localhost:50051` | Store service address (admin-bff only) |
| `GRPC_FILM_ADDR` | No | `localhost:50052` | Film service address |
| `GRPC_CUSTOMER_ADDR` | No | `localhost:50053` | Customer service address |
| `GRPC_RENTAL_ADDR` | No | `localhost:50054` | Rental service address |
| `GRPC_PAYMENT_ADDR` | No | `localhost:50055` | Payment service address |
| `LOG_LEVEL` | No | `info` | Log level |

## Development Commands

```bash
# Code generation
make proto-gen          # Generate protobuf Go code (requires buf)
make sqlc-gen           # Generate sqlc Go code (requires sqlc)
make generate           # Run both proto-gen and sqlc-gen

# Build
make build-all          # Build all 7 services

# Test & Lint
make test               # Run tests with race detector
make lint               # Run golangci-lint
make fmt                # Format code (gofmt + goimports)

# Infrastructure (local dev, PostgreSQL + Redis only)
make infra-up           # Start PostgreSQL + Redis
make infra-down         # Stop infrastructure
make infra-logs         # Tail infrastructure logs
make infra-ps           # Show infrastructure status

# All services (Docker)
make up                 # Start all services (infra + gRPC + BFF)
make down               # Stop all services
make logs               # Tail all logs
make ps                 # Show all service status

# Run services (local)
make run-store          # Start store-service
make run-film           # Start film-service
make run-customer       # Start customer-service
make run-rental         # Start rental-service
make run-payment        # Start payment-service
make run-customer-bff   # Start customer-bff
make run-admin-bff      # Start admin-bff
```

## License

MIT License - see [LICENSE](LICENSE) for details.
