# DVD 租赁微服务系统

基于 Go、gRPC 和 PostgreSQL 构建的微服务 DVD 租赁系统。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[English](README.md)** | **[日本語](README.ja.md)**

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────┐
│                              客户端                                 │
│                  ┌──────────────┬──────────────┐                   │
│                  │   顾客 APP   │   管理后台   │                   │
│                  └──────────────┴──────────────┘                   │
└─────────────────────────────────────────────────────────────────────┘
                         │                 │
                         ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        BFF 层 (REST API)                           │
│            ┌────────────────┐    ┌────────────────┐                │
│            │ customer-bff   │    │   admin-bff    │                │
│            │    (:8080)     │    │    (:8081)     │                │
│            │   17 个端点    │    │   50 个端点    │                │
│            └────────────────┘    └────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
                         │                 │
                         ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       核心服务层 (gRPC)                             │
│  ┌────────┐ ┌────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐        │
│  │  门店  │ │  影片  │ │   客户   │ │  租赁  │ │  支付   │        │
│  │ :50051 │ │ :50052 │ │  :50053  │ │ :50054 │ │  :50055 │        │
│  └────────┘ └────────┘ └──────────┘ └────────┘ └─────────┘        │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                            数据层                                   │
│            ┌────────────────┐    ┌────────────────┐                │
│            │   PostgreSQL   │    │     Redis      │                │
│            │   (15 张表)    │    │  (Token 存储)  │                │
│            └────────────────┘    └────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
```

## 功能特性

- **5 个 gRPC 微服务**：门店、影片、客户、租赁、支付
- **2 个 BFF 服务**：面向顾客和管理员的 REST API
- **JWT 认证**：双 Token（Access + Refresh）配合 Redis 存储
- **全文搜索**：PostgreSQL tsvector 影片搜索
- **分区表**：按月分区的支付表
- **容器化部署**：多阶段 Alpine 构建，docker-compose 本地开发

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.25+ |
| RPC | gRPC + Protocol Buffers v3 |
| HTTP | 标准库 `net/http` |
| 数据库 | PostgreSQL 17, sqlc |
| 缓存 | Redis 7 |
| 认证 | JWT (golang-jwt/jwt/v5) |
| 构建 | Docker, docker-compose |

## 快速开始

### 前置要求

- Go 1.25+
- Docker & Docker Compose
- buf（proto 代码生成）
- sqlc（SQL 代码生成）

### Docker 启动

```bash
# 启动所有服务
docker compose -f deployments/docker-compose.yml up -d

# 查看日志
make infra-logs

# 检查状态
make infra-ps
```

### 本地开发

```bash
# 仅启动基础设施
make infra-up

# 启动 gRPC 服务（每个服务在单独终端）
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-store
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-film
# ... 其他服务

# 启动 BFF 服务
JWT_SECRET="your-secret" REDIS_URL="redis://localhost:6379" make run-customer-bff
JWT_SECRET="your-secret" REDIS_URL="redis://localhost:6379" make run-admin-bff
```

## 项目结构

```
dvd-rental/
├── cmd/                          # 服务入口
├── internal/                     # 私有代码
│   ├── store/                    # 门店服务
│   ├── film/                     # 影片服务
│   ├── customer/                 # 客户服务
│   ├── rental/                   # 租赁服务
│   ├── payment/                  # 支付服务
│   └── bff/                      # BFF 服务
├── pkg/                          # 公共包
├── proto/                        # Proto 定义
├── gen/                          # 生成代码
├── migrations/                   # 数据库迁移
└── deployments/                  # Docker 配置
```

## API 端点

### 顾客 BFF（端口 8080）

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/v1/auth/login` | POST | - | 顾客登录 |
| `/api/v1/films` | GET | - | 影片列表/搜索 |
| `/api/v1/rentals` | GET/POST | JWT | 租赁管理 |
| `/api/v1/payments` | GET | JWT | 支付记录 |
| `/api/v1/profile` | GET/PUT | JWT | 个人信息 |

### 管理 BFF（端口 8081）

| 端点 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/v1/auth/login` | POST | - | 员工登录 |
| `/api/v1/stores` | CRUD | JWT | 门店管理 |
| `/api/v1/staff` | CRUD | JWT | 员工管理 |
| `/api/v1/customers` | CRUD | JWT | 客户管理 |
| `/api/v1/films` | CRUD | JWT | 影片管理 |
| `/api/v1/inventory` | CRUD | JWT | 库存管理 |
| `/api/v1/rentals` | CRUD | JWT | 租赁管理 |
| `/api/v1/payments` | CRUD | JWT | 支付管理 |

## 开发命令

```bash
make proto-gen     # 生成 protobuf 代码
make sqlc-gen      # 生成 sqlc 代码
make build-all     # 构建所有服务
make test          # 运行测试
make fmt           # 格式化代码
make lint          # 代码检查
```

## 许可证

MIT License - 详见 [LICENSE](LICENSE)
