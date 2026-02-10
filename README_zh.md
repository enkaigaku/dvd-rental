# DVD 租赁微服务系统

基于 Go、gRPC 和 PostgreSQL 构建的微服务 DVD 租赁系统。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[English](README.md)** | **[日本語](README_ja.md)**

## 系统架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                              客户端                                 │
│                  ┌──────────────┬──────────────┐                   │
│                  │   顾客 App   │   管理后台   │                   │
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
- **2 个 BFF 服务**：面向顾客（17 个端点）和管理员（50 个端点）的 REST API
- **JWT 认证**：双 Token（Access 15 分钟 + Refresh 7 天）配合 Redis 存储
- **全文搜索**：PostgreSQL tsvector 影片搜索
- **分区表**：按月分区的支付表
- **容器化部署**：多阶段 Alpine 构建，docker-compose 本地开发

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.25+ |
| RPC | gRPC + Protocol Buffers v3, buf |
| HTTP | 标准库 `net/http`（Go 1.22+ ServeMux）|
| 数据库 | PostgreSQL 17, sqlc |
| 缓存 | Redis 7 |
| 认证 | JWT (golang-jwt/jwt/v5), bcrypt |
| 构建 | Docker, docker-compose |

## 前置要求

| 工具 | 用途 | 安装 |
|------|------|------|
| Go 1.25+ | 语言运行时 | [go.dev/dl](https://go.dev/dl/) |
| Docker & Docker Compose | 容器运行时 | [docs.docker.com](https://docs.docker.com/get-docker/) |
| buf | Protobuf 代码生成 | `go install github.com/bufbuild/buf/cmd/buf@latest` |
| sqlc | SQL 代码生成 | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |

## 快速开始

### 方式 A：Docker 一键启动（最快）

构建并启动全部 7 个服务 + PostgreSQL + Redis。

```bash
# 克隆仓库
git clone https://github.com/enkaigaku/dvd-rental.git
cd dvd-rental

# 安装依赖并生成代码（proto + sqlc）
go mod download
make generate

# 启动所有服务
make up

# 确认所有服务运行正常
make ps

# 查看日志
make logs
```

启动后可通过以下地址访问：
- 顾客 BFF: http://localhost:8080
- 管理 BFF: http://localhost:8081

### 方式 B：本地开发（推荐日常开发使用）

基础设施用 Docker 运行，业务服务在本地运行以便快速迭代。

#### 第 1 步：启动基础设施

```bash
# 仅启动 PostgreSQL 和 Redis
# 数据库会通过 migrations/ 目录自动初始化 Schema 和示例数据
make infra-up
```

启动后：
- PostgreSQL: `localhost:5432`（用户: `dvdrental`, 密码: `dvdrental`, 数据库: `dvdrental`）
- Redis: `localhost:6379`

#### 第 2 步：安装依赖并生成代码

```bash
go mod download
make generate    # 生成 proto + sqlc 代码到 gen/
```

#### 第 3 步：启动 gRPC 服务

打开 5 个终端，分别启动各个服务：

```bash
# 终端 1 - 门店服务
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-store

# 终端 2 - 影片服务
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-film

# 终端 3 - 客户服务
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-customer

# 终端 4 - 租赁服务
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-rental

# 终端 5 - 支付服务
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-payment
```

#### 第 4 步：启动 BFF 服务

再打开 2 个终端：

```bash
# 终端 6 - 顾客 BFF
JWT_SECRET="dev-secret-change-me" REDIS_URL="redis://localhost:6379" make run-customer-bff

# 终端 7 - 管理 BFF
JWT_SECRET="dev-secret-change-me" REDIS_URL="redis://localhost:6379" make run-admin-bff
```

#### 第 5 步：验证

```bash
# 获取影片列表（无需认证）
curl http://localhost:8080/api/v1/films

# 顾客登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "mary.smith@sakilacustomer.org", "password": "password"}'
```

### 停止服务

```bash
# 停止 Docker 基础设施（本地开发）
make infra-down

# 或停止全部（方式 A）
make down
```

## 项目结构

```
dvd-rental/
├── cmd/                          # 服务入口
│   ├── store-service/            #   门店 gRPC 服务
│   ├── film-service/             #   影片 gRPC 服务
│   ├── customer-service/         #   客户 gRPC 服务
│   ├── rental-service/           #   租赁 gRPC 服务
│   ├── payment-service/          #   支付 gRPC 服务
│   ├── customer-bff/             #   顾客 REST API
│   └── admin-bff/                #   管理 REST API
├── internal/                     # 私有代码
│   ├── store/                    #   handler → service → repository
│   ├── film/                     #   handler → service → repository
│   ├── customer/                 #   handler → service → repository
│   ├── rental/                   #   handler → service → repository
│   ├── payment/                  #   handler → service → repository
│   └── bff/                      #   BFF 层（HTTP → gRPC 转换）
│       ├── customer/             #     config + handler + router
│       └── admin/                #     config + handler + router
├── pkg/                          # 公共包
│   ├── auth/                     #   JWT、bcrypt、Refresh Token 存储
│   ├── middleware/               #   Auth、CORS、日志、Recovery、响应
│   └── grpcutil/                 #   gRPC 客户端连接工具
├── proto/                        # Protocol Buffer 定义（.proto）
├── gen/                          # 生成代码
│   ├── proto/                    #   gRPC/protobuf 生成的 Go 代码
│   └── sqlc/                     #   sqlc 生成的 Go 代码
├── sql/                          # SQL 查询定义（sqlc 用）
├── migrations/                   # 数据库 Schema + 种子数据
│   ├── 001_initial_schema.sql    #   核心表定义
│   ├── 002_pagila_schema.sql     #   Pagila Schema 扩展
│   ├── 003_pagila_data.sql       #   示例数据
│   ├── 004_customer_auth.sql     #   客户密码哈希
│   └── 005_staff_auth.sql        #   员工密码哈希
├── deployments/                  # Docker 配置
│   ├── docker-compose.yml
│   └── dockerfiles/              #   各服务 Dockerfile
├── buf.yaml                      # Buf 配置
├── buf.gen.yaml                  # Buf 代码生成配置
├── sqlc.yaml                     # sqlc 配置
├── Makefile                      # 开发命令
└── go.mod
```

## API 端点

### 顾客 BFF（端口 8080）

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/login` | - | 顾客登录 |
| POST | `/api/v1/auth/refresh` | - | 刷新 Token |
| POST | `/api/v1/auth/logout` | - | 注销 |
| GET | `/api/v1/films` | - | 影片列表 |
| GET | `/api/v1/films/search` | - | 搜索影片 |
| GET | `/api/v1/films/category/{id}` | - | 按分类筛选 |
| GET | `/api/v1/films/actor/{id}` | - | 按演员筛选 |
| GET | `/api/v1/films/{id}` | - | 影片详情 |
| GET | `/api/v1/categories` | - | 分类列表 |
| GET | `/api/v1/actors` | - | 演员列表 |
| GET | `/api/v1/rentals` | JWT | 我的租赁 |
| GET | `/api/v1/rentals/{id}` | JWT | 租赁详情 |
| POST | `/api/v1/rentals` | JWT | 创建租赁 |
| POST | `/api/v1/rentals/{id}/return` | JWT | 归还 |
| GET | `/api/v1/payments` | JWT | 我的支付记录 |
| GET | `/api/v1/profile` | JWT | 我的资料 |
| PUT | `/api/v1/profile` | JWT | 更新资料 |

### 管理 BFF（端口 8081）

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/login` | - | 员工登录 |
| POST | `/api/v1/auth/refresh` | - | 刷新 Token |
| POST | `/api/v1/auth/logout` | - | 注销 |
| | `/api/v1/stores/**` | JWT | 门店管理（CRUD）|
| | `/api/v1/staff/**` | JWT | 员工管理（CRUD）|
| | `/api/v1/customers/**` | JWT | 客户管理（CRUD）|
| | `/api/v1/films/**` | JWT | 影片管理（CRUD）|
| | `/api/v1/actors/**` | JWT | 演员管理（CRUD）|
| | `/api/v1/categories/**` | JWT | 分类管理（CRUD）|
| | `/api/v1/languages/**` | JWT | 语言管理（CRUD）|
| | `/api/v1/inventory/**` | JWT | 库存管理（CRUD）|
| | `/api/v1/rentals/**` | JWT | 租赁管理（CRUD）|
| | `/api/v1/payments/**` | JWT | 支付管理（CRUD）|

## 环境变量

### gRPC 服务（store、film、customer、rental、payment）

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `DATABASE_URL` | 是 | - | PostgreSQL 连接字符串 |
| `GRPC_PORT` | 否 | 各服务不同 | gRPC 监听端口（50051-50055）|
| `LOG_LEVEL` | 否 | `info` | 日志级别（debug, info, warn, error）|

### BFF 服务（customer-bff、admin-bff）

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `HTTP_PORT` | 否 | `8080` / `8081` | HTTP 监听端口 |
| `JWT_SECRET` | 是 | - | JWT 签名密钥 |
| `JWT_ACCESS_DURATION` | 否 | `15m` | Access Token 有效期 |
| `JWT_REFRESH_DURATION` | 否 | `168h`（7 天）| Refresh Token 有效期 |
| `REDIS_URL` | 是 | - | Redis 连接地址（存储 Refresh Token）|
| `GRPC_STORE_ADDR` | 否 | `localhost:50051` | 门店服务地址（仅 admin-bff）|
| `GRPC_FILM_ADDR` | 否 | `localhost:50052` | 影片服务地址 |
| `GRPC_CUSTOMER_ADDR` | 否 | `localhost:50053` | 客户服务地址 |
| `GRPC_RENTAL_ADDR` | 否 | `localhost:50054` | 租赁服务地址 |
| `GRPC_PAYMENT_ADDR` | 否 | `localhost:50055` | 支付服务地址 |
| `LOG_LEVEL` | 否 | `info` | 日志级别 |

## 开发命令

```bash
# 代码生成
make proto-gen          # 生成 protobuf Go 代码（需要 buf）
make sqlc-gen           # 生成 sqlc Go 代码（需要 sqlc）
make generate           # 同时执行 proto-gen 和 sqlc-gen

# 构建
make build-all          # 构建全部 7 个服务

# 测试 & 检查
make test               # 运行测试（启用 race 检测）
make lint               # 运行 golangci-lint
make fmt                # 格式化代码（gofmt + goimports）

# 基础设施（本地开发，仅 PostgreSQL + Redis）
make infra-up           # 启动 PostgreSQL + Redis
make infra-down         # 停止基础设施
make infra-logs         # 查看基础设施日志
make infra-ps           # 查看基础设施状态

# 全部服务（Docker）
make up                 # 启动所有服务（基础设施 + gRPC + BFF）
make down               # 停止所有服务
make logs               # 查看所有日志
make ps                 # 查看所有服务状态

# 运行服务（本地）
make run-store          # 启动门店服务
make run-film           # 启动影片服务
make run-customer       # 启动客户服务
make run-rental         # 启动租赁服务
make run-payment        # 启动支付服务
make run-customer-bff   # 启动顾客 BFF
make run-admin-bff      # 启动管理 BFF
```

## 许可证

MIT License - 详见 [LICENSE](LICENSE)
