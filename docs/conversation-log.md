# DVD 租赁微服务 - 需求分析对话记录

## 2026-02-08

### 对话 1：项目启动 & 数据库分析

**用户**：根据 migrations 目录下的 DB 数据，我想使用 Go 语言搭建一套 DVD 租赁服务，采用微服务架构。

**助手**：已完整分析数据库结构，基于 Pagila（PostgreSQL 版 Sakila）的 DVD 租赁数据库，包含 5 个迁移文件，目前项目中没有任何 Go 代码。

#### 数据库概览

**核心表（15张）**

| 领域 | 表 |
|------|-----|
| 影片 | `film`, `actor`, `category`, `language`, `film_actor`, `film_category` |
| 门店/人员 | `store`, `staff`, `customer` |
| 地址 | `country`, `city`, `address` |
| 业务交易 | `inventory`, `rental`, `payment`（按月分区） |

**关键特性**
- 全文搜索（film 表的 tsvector）
- payment 表按月范围分区（2022年1-7月）
- 所有表有 `last_update` 自动更新触发器
- 客户和员工都有 bcrypt 密码哈希（认证支持）
- 丰富的业务函数（库存检查、客户余额计算、奖励报告等）
- 7 个预建视图用于报表查询

---

### 对话 2：架构设计关键问题

**助手**提出了以下需要确认的问题：

1. **微服务划分**：如何划分微服务？比如按域模型划分（电影管理服务、租赁服务、支付服务、客户服务等）？
2. **通信方式**：各微服务之间倾向于用 gRPC、REST API，还是消息队列（如 Kafka/RabbitMQ）？
3. **技术栈偏好**：
   - 数据库操作：GORM、sqlc、还是 ent？
   - HTTP 框架：gin、echo、还是 fiber？
4. **API 网关**：是否需要统一的 API 网关来路由外部请求？
5. **认证授权**：是否需要独立的认证服务？用 JWT 还是其他方案？
6. **部署方式**：是否需要 Docker 容器化 + docker-compose？需要考虑 Kubernetes 吗？
7. **事件驱动**：是否需要事件机制（比如租赁/归还/支付时发送通知）？
8. **架构模式**：Go 项目结构偏好 Clean Architecture、六边形架构，还是其他？

**用户**：首先关于微服务的划分，一般都有哪些方式？

---

### 对话 3：微服务划分方式讨论

**助手**：介绍了 5 种主流的微服务划分方式：

#### 1. 按业务域划分（Domain-Driven Design）
最主流的方式，基于 DDD 的限界上下文（Bounded Context）拆分。以 DVD 租赁为例：
- **影片服务**：film, actor, category, language
- **客户服务**：customer, address, city, country
- **租赁服务**：rental, inventory
- **支付服务**：payment
- **门店服务**：store, staff

优点：高内聚低耦合，团队可独立负责各自域
缺点：域边界划分需要经验，划分不当会导致频繁跨服务调用

#### 2. 按功能/能力划分（Business Capability）
围绕业务能力而非数据模型拆分：
- 目录浏览服务（搜索、浏览影片）
- 会员管理服务（注册、认证、个人信息）
- 交易服务（租赁 + 支付一体）
- 运营管理服务（门店、员工、库存）

优点：更贴近实际业务流程
缺点：可能导致单个服务过大

#### 3. 按数据所有权划分（Database per Service）
每个服务拥有自己的数据库/表，严格禁止直接访问其他服务的数据。

优点：数据隔离彻底，服务完全独立
缺点：跨服务查询复杂，需要额外处理数据一致性

#### 4. 按读写分离划分（CQRS）
将读操作和写操作拆成独立服务：
- 查询服务：影片搜索、报表统计、列表展示
- 命令服务：创建租赁、处理支付、更新库存

优点：读写可独立扩缩容，适合读多写少场景
缺点：架构复杂度高，数据同步有延迟

#### 5. 按用户角色/前端划分（BFF - Backend for Frontend）
按不同客户端或角色拆分：
- 客户端 BFF（面向租客的 APP/网站）
- 管理端 BFF（面向员工/管理员的后台）
- 底层共享核心服务

优点：各端体验可独立优化
缺点：可能有重复逻辑

#### 推荐方案
对于 DVD 租赁项目，最推荐 **方式 1（按业务域）+ 方式 5（BFF）** 的组合。业务域划分作为核心服务层，BFF 作为面向不同端的聚合层。

**用户**：微服务划分暂时按推荐方案（按业务域 + BFF）。接下来的服务间通信方式，是不是要根据服务间的关系来做取舍？

---

### 对话 4：服务间通信方式讨论

**已确认决策**：微服务划分采用 **按业务域（DDD）+ BFF** 方案。

**助手**：根据服务间的交互场景，建议混合使用三种通信方式：

#### 通信方式选择原则

| 通信方式 | 适用场景 | 特点 |
|---------|---------|------|
| REST API | 对外暴露接口、简单请求-响应 | 通用性强、调试方便、性能一般 |
| gRPC | 服务间内部调用、高频低延迟 | 高性能、强类型（protobuf）、流式支持 |
| 消息队列 | 异步处理、事件通知、解耦 | 削峰填谷、最终一致性、架构解耦 |

#### 同步调用（gRPC）—— 需要即时响应的场景
- 租赁服务 → 影片服务：实时查询影片信息和价格
- 租赁服务 → 客户服务：验证客户身份和状态
- 租赁服务 → 门店服务：检查库存
- 支付服务 → 租赁服务：支付时关联租赁记录
- BFF → 各核心服务：聚合数据返回给前端

#### 异步消息（消息队列）—— 不需要即时响应的场景
- 租赁服务 → 支付服务：创建租赁后异步发起支付
- 支付服务 → 通知服务：支付完成后发送确认通知
- 租赁服务 → 库存服务：归还 DVD 后异步更新库存
- 任意服务 → 报表/统计：业务事件异步写入

#### 对外接口（REST API）
- BFF 层 → 前端客户端：面向浏览器/APP 的 HTTP API
- API 网关 → BFF：外部流量入口

#### 架构总览
```
前端/APP
  │ (REST)
API 网关
  │ (REST)
┌─────────────┐  ┌─────────────┐
│ 客户端 BFF  │  │ 管理端 BFF  │
└──────┬──────┘  └──────┬──────┘
       │ (gRPC)         │ (gRPC)
       ▼                ▼
┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
│影片  │ │客户  │ │租赁  │ │支付  │ │门店  │
│服务  │ │服务  │ │服务  │ │服务  │ │服务  │
└──────┘ └──────┘ └──┬───┘ └──┬───┘ └──────┘
                     │        │
                     ▼        ▼ (消息队列)
                  ┌──────────────┐
                  │  事件总线     │
                  │ (Kafka/NATS) │
                  └──────────────┘
```

#### 消息队列选型建议
推荐 **NATS** 而非 Kafka：轻量级、Go 原生实现、部署简单，对中小规模项目足够。

**用户**：就用这个方案。技术栈方面，数据库操作不太想使用 ORM 框架，想用原生 SQL，主要考虑性能问题。

---

### 对话 5：数据库操作方案讨论（原生 SQL）

**已确认决策**：
- 微服务划分：按业务域（DDD）+ BFF
- 服务间通信：gRPC（同步）+ NATS（异步）+ REST（对外）

**助手**：不使用 ORM，原生 SQL 在 Go 中有三种方案：

#### 1. database/sql（标准库）
最底层，手写 Scan 映射。零依赖、性能最佳，但样板代码多、SQL 错误只能运行时发现。

#### 2. sqlx（jmoiron/sqlx）
标准库的轻量封装，增加结构体自动映射。API 与标准库一致、几乎无性能损耗，但 SQL 仍是字符串无法编译时检查。

#### 3. sqlc（编译时生成）
写 SQL 文件，自动生成类型安全的 Go 代码。原生 SQL + 编译时类型安全、零运行时开销，但需要代码生成步骤、复杂动态查询不太方便。

#### 对比

| 特性 | database/sql | sqlx | sqlc |
|------|-------------|------|------|
| 性能 | 最佳 | 几乎无损 | 几乎无损 |
| 类型安全 | 无 | 运行时 | 编译时 |
| 样板代码 | 多 | 少 | 最少（自动生成） |
| SQL 控制力 | 完全 | 完全 | 完全 |
| 动态查询 | 灵活 | 灵活 | 不太方便 |

#### 推荐
**sqlc** 为主 + **sqlx** 补充动态查询场景。

**用户**：另外根据业务领域，是不是考虑不同的数据存储策略比较好？比如这个业务适合关系型数据存储，那个业务适合 NoSQL 类的数据存储。

---

### 对话 6：多语言持久化（Polyglot Persistence）策略讨论

**助手**：分析了各服务的数据特点和推荐存储方案：

#### 各服务数据特点

| 服务 | 数据特点 | 读写模式 |
|------|---------|---------|
| 影片服务 | 结构化 + 全文搜索需求 | 读多写少，搜索频繁 |
| 客户服务 | 结构化，关联关系明确 | 读多写少 |
| 租赁服务 | 事务性强，需要强一致性 | 读写均衡，事务密集 |
| 支付服务 | 事务性极强，财务数据 | 写多，需审计追溯 |
| 门店服务 | 结构化，数据量小 | 读多写极少 |

#### 存储方案

- **PostgreSQL（主力）**：租赁、支付、客户、门店服务 —— 关系型数据、事务要求
- **PostgreSQL + Elasticsearch**：影片服务 —— 结构化存储 + 强大搜索
- **Redis（共用缓存层）**：影片缓存、库存状态缓存、会话/Token、分布式锁

#### 渐进式策略（推荐）
1. **第一阶段**：全部 PostgreSQL + Redis 缓存，影片搜索用 PG 自带 tsvector
2. **第二阶段**：搜索需求复杂化时，再引入 Elasticsearch

**用户**：采用渐进式方案。HTTP 框架方面，Go 生态真的需要吗？

---

### 对话 7：HTTP 框架讨论

**已确认决策**：
- 微服务划分：按业务域（DDD）+ BFF
- 服务间通信：gRPC（同步）+ NATS（异步）+ REST（对外）
- 数据库操作：sqlc 为主 + sqlx 补充（原生 SQL）
- 数据存储：渐进式方案（第一阶段 PostgreSQL + Redis，第二阶段按需加 ES）

**助手**：Go 1.22+ 的标准库 `net/http` 已经支持方法匹配、路径参数、通配符，框架提供的主要是便利性封装。

#### 标准库 vs 框架对比

| 功能 | net/http | gin/echo |
|------|----------|----------|
| 路由（含方法+路径参数） | 1.22+ 原生支持 | 支持 |
| 中间件链 | 手动组合，但很简单 | 内置 |
| 请求参数绑定/校验 | 手动解析 | 自动绑定到结构体 |
| JSON 响应封装 | json.NewEncoder | c.JSON(200, obj) |
| 路由分组 | 手动前缀 | 内置 Group |

#### 结论
- **核心服务**（影片/客户/租赁/支付/门店）：不需要 HTTP 框架，纯 gRPC
- **BFF 层**两种选择：
  1. 纯标准库 net/http —— 零依赖、完全掌控，多写一点工具函数
  2. 轻量路由库 chi —— 完全兼容 net/http 接口，只添加路由分组和中间件链

**用户**：倾向于纯标准库 net/http。认证授权使用 JWT 就足够了。

---

### 对话 8：认证授权方案讨论（JWT）

**已确认决策**：
- 微服务划分：按业务域（DDD）+ BFF
- 服务间通信：gRPC（同步）+ NATS（异步）+ REST（对外）
- 数据库操作：sqlc 为主 + sqlx 补充（原生 SQL）
- 数据存储：渐进式（第一阶段 PostgreSQL + Redis，第二阶段按需加 ES）
- HTTP 框架：BFF 层使用纯标准库 net/http，核心服务纯 gRPC

**助手**：JWT 完全足够。方案设计如下：

#### 双角色体系
数据库已有 customer 和 staff 两套认证（都有 password_hash）：

| 角色 | 登录方式 | Token 载荷 |
|------|---------|-----------|
| 客户 | email + password | role:"customer", customer_id, store_id |
| 员工 | username + password | role:"staff", staff_id, store_id, permissions |

#### Token 策略：Access Token + Refresh Token
- Access Token：短期（15-30 分钟），携带在 Authorization 头
- Refresh Token：长期（7 天），存 Redis（支持主动失效/登出）

#### 认证服务归属
- **方案 A**：独立 Auth Service —— 职责单一但多一个服务
- **方案 B（推荐）**：认证逻辑内嵌 BFF 层 —— 架构简单，共享 `pkg/auth` 包

#### 鉴权流程
1. 登录 → BFF 验证密码 → 签发 Access + Refresh Token
2. 后续请求 → BFF 中间件验证 Token → 解析用户信息
3. BFF → 核心服务：通过 gRPC metadata 传递 user_id/role
4. 核心服务信任 BFF 传来的身份信息（内网可信）
5. Token 过期 → Refresh Token 刷新 → Redis 检查失效状态

#### 权限控制（简单 RBAC）
- 客户：只能操作自己的数据
- 员工：门店范围内的数据
- 管理员（store manager）：门店级管理权限

**用户**：采用方案 B（认证逻辑内嵌 BFF 层）。部署方面要容器化，本地开发用 docker-compose 启动各种服务，不需要本地 Kubernetes。生产环境部署到 AWS 等云平台，使用平台提供的容器编排服务。

---

### 对话 9：部署方案讨论

**已确认决策**：
- 微服务划分：按业务域（DDD）+ BFF
- 服务间通信：gRPC（同步）+ NATS（异步）+ REST（对外）
- 数据库操作：sqlc 为主 + sqlx 补充（原生 SQL）
- 数据存储：渐进式（第一阶段 PostgreSQL + Redis，第二阶段按需加 ES）
- HTTP 框架：BFF 层使用纯标准库 net/http，核心服务纯 gRPC
- 认证授权：JWT（方案 B，内嵌 BFF 层），Access + Refresh Token，简单 RBAC

**助手**：部署方案确认：

#### 本地开发：Docker Compose
docker-compose 一键启动：
- 基础设施：PostgreSQL、Redis、NATS
- 核心服务：影片、客户、租赁、支付、门店
- BFF 层：客户端 BFF、管理端 BFF
- 本地开发也可只启动基础设施，Go 服务直接 `go run` 调试

#### 生产部署：云平台容器编排
- 容器镜像推送到 ECR / Docker Hub
- 使用 AWS ECS / EKS 等平台编排
- 不在本地搞 Kubernetes

**用户**：六边形架构是什么样的？

---

### 对话 10：Go 项目架构模式讨论

**助手**：介绍了六边形架构（Hexagonal Architecture / Ports & Adapters）：

#### 三层结构
| 层 | 职责 | 示例 |
|---|------|------|
| 业务核心（Domain） | 纯业务逻辑，零外部依赖 | 租赁规则、价格计算、库存检查 |
| 端口（Ports） | Go 接口定义，业务核心的边界 | FilmRepository 接口、RentalService 接口 |
| 适配器（Adapters） | 端口的具体实现，连接外部世界 | PostgreSQL 实现、gRPC handler、NATS publisher |

#### 依赖方向
适配器 → 端口(接口) ← 业务核心（从外向内）

#### 与 Clean Architecture 对比
- 六边形：3 层，简单直观，Go 接口天然就是端口
- Clean Architecture：4 层，模板代码更多，小项目可能过度

#### 适合本项目的原因
1. Go 接口 = 端口，天然契合
2. 微服务体量不大，三层刚好不过度
3. 换技术实现只需写新适配器
4. 测试只需 mock 接口

**用户**：微服务体量不大，采用标准 Go 布局更好。

**助手**：同意。标准 Go 布局方案：

#### 每个服务内部三层
```
handler → service → repository
```

#### 整体采用 monorepo 结构
```
dvd-rental/
├── cmd/                      # 所有服务入口
│   ├── film-service/
│   ├── customer-service/
│   ├── rental-service/
│   ├── payment-service/
│   ├── store-service/
│   ├── customer-bff/
│   └── admin-bff/
├── internal/                 # 各服务私有代码
│   ├── film/
│   ├── customer/
│   ├── rental/
│   ├── payment/
│   ├── store/
│   ├── bff/
│   │   ├── customer/
│   │   └── admin/
│   └── common/
├── pkg/                      # 可复用公共包（auth、middleware、grpcutil）
├── proto/                    # protobuf 定义
├── sql/                      # sqlc 查询文件
├── migrations/               # 数据库迁移（已有）
├── deployments/              # Dockerfile + docker-compose.yml
├── go.mod
└── Makefile
```

monorepo 优势：依赖统一管理、proto 集中管理、公共包直接引用、本地开发方便。

**用户**：就这个架构。把所有技术决策整理总结。

---

## 技术决策总结

> 以下为 2026-02-08 需求分析阶段确认的全部技术决策。

### 1. 微服务划分

**方案**：按业务域（DDD）+ BFF

**核心服务（5 个）**：

| 服务 | 负责的表 |
|------|---------|
| 影片服务（film-service） | film, actor, category, language, film_actor, film_category |
| 客户服务（customer-service） | customer, address, city, country |
| 租赁服务（rental-service） | rental, inventory |
| 支付服务（payment-service） | payment |
| 门店服务（store-service） | store, staff |

**BFF 层（2 个）**：
- 客户端 BFF（customer-bff）：面向租客的 APP/网站
- 管理端 BFF（admin-bff）：面向员工/管理员的后台

### 2. 服务间通信

| 场景 | 方式 | 说明 |
|------|------|------|
| BFF → 核心服务 | gRPC | 同步调用，低延迟，强类型 |
| 核心服务 → 核心服务 | gRPC | 需即时响应的内部调用 |
| 异步事件 | NATS | 租赁→支付、支付→通知、归还→库存更新等 |
| 对外接口 | REST（net/http） | BFF 暴露给前端的 HTTP API |

### 3. 数据库操作

**方案**：原生 SQL（不使用 ORM）
- **sqlc** 为主：编译时生成类型安全的 Go 代码，零运行时开销
- **sqlx** 补充：处理动态查询场景

### 4. 数据存储（渐进式）

**第一阶段**：
- **PostgreSQL**：所有服务的主数据存储
- **Redis**：缓存层（影片缓存、库存状态、JWT Refresh Token、分布式锁）

**第二阶段（按需）**：
- **Elasticsearch**：影片全文搜索（替代 PostgreSQL tsvector）

### 5. HTTP 框架

- **核心服务**：纯 gRPC，不需要 HTTP 框架
- **BFF 层**：Go 标准库 `net/http`（Go 1.22+ 已支持方法匹配和路径参数）

### 6. 认证授权

**方案**：JWT，认证逻辑内嵌 BFF 层（共享 `pkg/auth` 包）

- **双角色**：客户（email 登录）、员工（username 登录）
- **双 Token**：Access Token（15-30 分钟）+ Refresh Token（7 天，存 Redis）
- **权限**：简单 RBAC（客户 / 员工 / 管理员）
- **传递方式**：BFF 验证后通过 gRPC metadata 传递用户信息，核心服务信任上游

### 7. 部署方式

- **本地开发**：Docker Compose（基础设施 + 各服务一键启动）
- **生产环境**：容器镜像推送至 ECR/Docker Hub，使用 AWS ECS/EKS 等云平台编排
- **不在本地使用 Kubernetes**

### 8. 项目架构

**方案**：标准 Go 项目布局 + Monorepo

**服务内部三层**：`handler → service → repository`

**整体目录结构**：
```
dvd-rental/
├── cmd/                      # 所有服务入口
│   ├── film-service/
│   ├── customer-service/
│   ├── rental-service/
│   ├── payment-service/
│   ├── store-service/
│   ├── customer-bff/
│   └── admin-bff/
├── internal/                 # 各服务私有代码
│   ├── film/
│   │   ├── handler/
│   │   ├── service/
│   │   └── repository/
│   ├── customer/
│   ├── rental/
│   ├── payment/
│   ├── store/
│   ├── bff/
│   │   ├── customer/
│   │   └── admin/
│   └── common/
├── pkg/                      # 可复用公共包
│   ├── auth/                 #   JWT 签发/验证
│   ├── middleware/            #   HTTP 中间件
│   └── grpcutil/             #   gRPC 工具函数
├── proto/                    # protobuf 定义
├── sql/                      # sqlc 查询文件
├── migrations/               # 数据库迁移（已有）
├── deployments/
│   ├── docker/               #   各服务 Dockerfile
│   └── docker-compose.yml
├── docs/                     # 项目文档
├── go.mod
└── Makefile
```

### 9. 架构总览图

```
                        前端/APP
                           │ (REST)
                       API 网关
                           │ (REST)
                ┌──────────┴──────────┐
          ┌─────┴──────┐        ┌─────┴──────┐
          │客户端 BFF   │        │管理端 BFF   │
          │(net/http)  │        │(net/http)  │
          │+ JWT 认证   │        │+ JWT 认证   │
          └─────┬──────┘        └─────┬──────┘
                │ (gRPC)              │ (gRPC)
     ┌──────┬──┴───┬──────┬──────┐   │
     ▼      ▼      ▼      ▼      ▼   │
  ┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
  │影片  ││客户  ││租赁  ││支付  ││门店  │
  │服务  ││服务  ││服务  ││服务  ││服务  │
  │(gRPC)││(gRPC)││(gRPC)││(gRPC)││(gRPC)│
  └──┬───┘└──┬───┘└──┬───┘└──┬───┘└──┬───┘
     │       │       │       │       │
     ▼       ▼       ▼       ▼       ▼
  PostgreSQL (各服务独立 schema/database)
                     │       │
                     ▼       ▼
                  NATS (异步事件总线)
                     │
                     ▼
                   Redis (缓存 + Token + 分布式锁)
```

### 10. 技术栈一览

| 类别 | 选型 |
|------|------|
| 语言 | Go 1.22+ |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 消息队列 | NATS |
| 服务间通信 | gRPC + Protobuf |
| 对外接口 | REST（net/http 标准库） |
| 数据库操作 | sqlc + sqlx（原生 SQL） |
| 认证 | JWT（Access + Refresh Token） |
| 容器化 | Docker + Docker Compose |
| 生产部署 | AWS ECS/EKS 等云平台 |
| 项目结构 | 标准 Go 布局，Monorepo |
| 搜索引擎 | 第一阶段 PG tsvector，第二阶段 Elasticsearch |

---

*状态：需求分析阶段完成 —— 所有技术决策已确认，准备进入实施阶段*

---

## 实施阶段

### 阶段 1：基础设施搭建

**用户**：先从 infrastructure 部分开始，postgres 容器镜像版本使用 17.6-alpine。代码版本管理用 git，也要使用 .gitignore。每完成一个阶段 commit 一下，commit 信息格式采用 git workflow，内容用英文。

**助手**：完成基础设施搭建，包括：

#### 创建的文件
- **`.gitignore`** — Go、IDE、OS、Docker、环境变量等忽略规则
- **`go.mod`** — 模块 `github.com/tokyoyuan/dvd-rental`，Go 1.25.7
- **`deployments/docker-compose.yml`** — 三个基础设施容器：
  - PostgreSQL 17.6-alpine（端口 5432，用户/密码 dvdrental，挂载 migrations）
  - Redis 7-alpine（端口 6379）
  - NATS 2-alpine + JetStream（端口 4222/8222）
  - 所有容器配有 healthcheck
- **`Makefile`** — `infra-up`、`infra-down`、`infra-logs`、`infra-ps` targets
- **目录结构** — `cmd/`、`internal/`、`pkg/`、`proto/`、`sql/`、`gen/`、`deployments/docker/`、`docs/` 各放 `.gitkeep` 占位

#### Git 提交
- `git init` + 首次提交：`chore: initial project setup with infrastructure`

---

### 阶段 2：Store Service 实施

**用户**：从相对容易的开始，每完成一个步骤都让我检查一下。

**助手**：选择门店服务（store-service）作为第一个实现的微服务。原因：只管理 `store` 和 `staff` 两张表，数据量小、逻辑简单，适合建立可复用的模式。按 6 步计划实施。

#### Step 1：Proto 定义 + 代码生成工具配置

**新增文件：**

1. **`proto/store/v1/store.proto`** — gRPC 服务定义
   - `StoreService`：GetStore、ListStores、CreateStore、UpdateStore、DeleteStore
   - `StaffService`：GetStaff、GetStaffByUsername、ListStaff、ListStaffByStore、CreateStaff、UpdateStaff、DeactivateStaff、UpdateStaffPassword
   - 关键设计：
     - Staff 不做物理删除（被 rental/payment FK RESTRICT），只 DeactivateStaff（soft delete）
     - `picture`（BYTEA）仅在 GetStaff 单条查询时返回，List 不返回
     - `password_hash` 仅在 GetStaffByUsername 中返回（供 BFF 认证用）

2. **`buf.yaml`** + **`buf.gen.yaml`** — Buf v2 工具链配置
   - proto 路径：`proto`，lint: STANDARD，breaking: FILE
   - 插件：protocolbuffers/go + grpc/go，输出到 `gen/proto`

3. **`sqlc.yaml`** — sqlc 配置
   - 引擎：postgresql，查询：`sql/queries`，schema：`migrations`
   - 输出：`internal/store/repository/sqlcgen`，sql_package：pgx/v5

4. **`sql/queries/store.sql`** — 6 个查询（Get/List/Count/Create/Update/Delete）

5. **`sql/queries/staff.sql`** — 16 个查询，按需选择不同列：
   - GetStaff：含 picture，不含 password_hash
   - GetStaffByUsername：含 password_hash，不含 picture
   - List 系列：不含 picture 和 password_hash
   - Count 系列：4 个变体（全部/活跃/按门店/按门店活跃）
   - Create/Update/Deactivate/UpdatePassword

6. **`Makefile`** — 新增 `proto-gen`、`sqlc-gen`、`generate` targets

**Git 提交**：`feat(store): add proto definitions and code generation config`

#### Step 2：运行代码生成

安装工具（`brew install bufbuild/buf/buf sqlc`），执行代码生成：

- `make proto-gen`（buf generate）→ 生成 `gen/proto/store/v1/store.pb.go` + `store_grpc.pb.go`
- `make sqlc-gen`（sqlc generate）→ 生成 `internal/store/repository/sqlcgen/` 下的 `db.go`、`models.go`、`store.sql.go`、`staff.sql.go`

**Git 提交**：`feat(store): generate protobuf and sqlc code`

#### Step 3：Repository 层

**新增文件：**

1. **`internal/store/model/model.go`** — 领域模型
   ```go
   type Store struct {
       StoreID, ManagerStaffID, AddressID int32
       LastUpdate time.Time
   }
   type Staff struct {
       StaffID int32; FirstName, LastName string; AddressID int32
       Email string; StoreID int32; Active bool; Username string
       PasswordHash string; Picture []byte; LastUpdate time.Time
   }
   ```

2. **`internal/store/repository/store_repository.go`**
   - 接口 `StoreRepository`（6 个方法）
   - 实现包装 sqlcgen，pgtype → Go 标准类型转换
   - `ErrNotFound` 哨兵错误映射 `pgx.ErrNoRows`

3. **`internal/store/repository/staff_repository.go`**
   - 接口 `StaffRepository`（14 个方法）
   - `CreateStaffParams` / `UpdateStaffParams` 参数结构体
   - 每种 sqlc 行类型有独立的转换函数（因为不同查询返回不同列）

**设计要点**：
- Repository 层负责 pgtype ↔ Go 标准类型转换，上层无需关心数据库类型
- 循环 FK（store ↔ staff）通过两个 repo 分别注入 service 层解决

**Git 提交**：`feat(store): add repository layer with domain models`

#### Step 4：Service 层（业务逻辑）

**新增文件：**

1. **`internal/store/service/errors.go`** — 三个哨兵错误
   - `ErrNotFound`、`ErrInvalidArgument`、`ErrAlreadyExists`

2. **`internal/store/service/pgutil.go`**
   - `isUniqueViolation(err)` 检查 PostgreSQL 错误码 "23505"

3. **`internal/store/service/store_service.go`**
   - `StoreService` 同时持有 storeRepo 和 staffRepo（交叉验证）
   - CreateStore/UpdateStore 验证 manager staff 存在且 active
   - 分页：defaultPageSize=20，maxPageSize=100，`clampPagination()` 工具函数
   - 唯一约束冲突 → ErrAlreadyExists

4. **`internal/store/service/staff_service.go`**
   - `StaffService` 同时持有 staffRepo 和 storeRepo（交叉验证）
   - CreateStaff/UpdateStaff 验证 store 存在
   - ListStaff/ListStaffByStore 支持 activeOnly 过滤
   - UpdateStaffPassword 先验证 staff 存在

**Git 提交**：`feat(store): add service layer with business logic`

#### Step 5：Handler 层（gRPC）

**新增文件：**

1. **`internal/store/handler/convert.go`** — 转换工具
   - `toGRPCError()`：哨兵错误 → gRPC status code 映射
     - ErrNotFound → codes.NotFound
     - ErrInvalidArgument → codes.InvalidArgument
     - ErrAlreadyExists → codes.AlreadyExists
     - default → codes.Internal
   - `storeToProto()` / `staffToProto()`：领域模型 → proto 消息

2. **`internal/store/handler/store_handler.go`**
   - 嵌入 `UnimplementedStoreServiceServer`，实现 5 个 RPC
   - 薄层：委托 service 层，转换错误

3. **`internal/store/handler/staff_handler.go`**
   - 嵌入 `UnimplementedStaffServiceServer`，实现 8 个 RPC
   - `toStaffListResponse()` 辅助函数处理列表响应

**Git 提交**：`feat(store): add gRPC handler layer`

#### Step 6：Main 入口 + 配置

**新增文件：**

1. **`internal/store/config/config.go`** — 环境变量配置
   - `DATABASE_URL`（必填）
   - `GRPC_PORT`（默认 "50051"）
   - `LOG_LEVEL`（默认 "info"）

2. **`cmd/store-service/main.go`** — 服务入口
   - `run()` 函数返回 error，`main()` 只负责调用和 `log.Fatal`
   - `pgxpool.New()` + `pool.Ping()` 建立数据库连接
   - 依赖注入：repos → services → handlers
   - gRPC server 注册 StoreService + StaffService
   - 标准 gRPC Health Check（`grpc.health.v1.Health`）
   - gRPC Reflection（开发调试用）
   - Graceful Shutdown：监听 SIGINT/SIGTERM → 设置 health NOT_SERVING → GracefulStop()

**修改文件：**
- **`go.mod`** — 添加依赖：google.golang.org/grpc v1.78.0, google.golang.org/protobuf v1.36.11, github.com/jackc/pgx/v5 v5.8.0
- **`Makefile`** — 新增 `run-store`、`test`、`lint`、`fmt` targets
- **`.gitignore`** — 添加各服务二进制文件名（store-service, film-service 等）

**Git 提交**：`feat(store): add main entry point, config, and dependency wiring`

#### 遇到的问题及解决

1. **buf/sqlc 未安装**：通过 Homebrew 安装 `brew install bufbuild/buf/buf sqlc`
2. **buf.yaml 多余依赖**：初始配置中包含 `buf.build/googleapis/googleapis` 依赖，实际只用了内置的 well-known types（Timestamp、Empty），移除后解决警告
3. **docker-compose migrations 路径**：`deployments/docker-compose.yml` 中 migrations 挂载路径应为 `../migrations`（一层），非 `../../migrations`
4. **go mod tidy**：添加 repository 代码后需先执行 `go mod tidy` 再编译
5. **二进制文件误入 git**：`go build ./cmd/store-service/` 在项目根目录生成了 `store-service` 二进制，被 `git add .` 暂存。通过 `git reset HEAD store-service` + `rm -f store-service` 清理，并更新 `.gitignore` 添加所有服务二进制名

#### Store Service 文件清单

```
cmd/store-service/
└── main.go                              # 服务入口、依赖注入、graceful shutdown

internal/store/
├── config/
│   └── config.go                        # 环境变量配置加载
├── handler/
│   ├── convert.go                       # proto ↔ 领域模型转换、错误映射
│   ├── store_handler.go                 # StoreService gRPC 实现
│   └── staff_handler.go                 # StaffService gRPC 实现
├── model/
│   └── model.go                         # Store、Staff 领域模型
├── repository/
│   ├── store_repository.go              # StoreRepository 接口 + pgx 实现
│   ├── staff_repository.go              # StaffRepository 接口 + pgx 实现
│   └── sqlcgen/                         # sqlc 自动生成（勿手动修改）
│       ├── db.go
│       ├── models.go
│       ├── store.sql.go
│       └── staff.sql.go
└── service/
    ├── errors.go                        # ErrNotFound, ErrInvalidArgument, ErrAlreadyExists
    ├── pgutil.go                        # isUniqueViolation() PostgreSQL 错误检查
    ├── store_service.go                 # Store 业务逻辑（含交叉验证）
    └── staff_service.go                 # Staff 业务逻辑（含交叉验证）

proto/store/v1/
└── store.proto                          # gRPC 服务定义

gen/proto/store/v1/
├── store.pb.go                          # protobuf 生成代码
└── store_grpc.pb.go                     # gRPC 生成代码

sql/queries/
├── store.sql                            # Store 表 SQL 查询
└── staff.sql                            # Staff 表 SQL 查询
```

#### 可复用模式（供后续服务参考）

1. **三层架构**：handler（proto 转换 + 错误映射）→ service（业务逻辑 + 验证）→ repository（接口 + pgx/sqlc 实现）
2. **领域模型与 DB 类型隔离**：pgtype 转换封装在 repository 层
3. **哨兵错误体系**：service 层定义哨兵错误，handler 层统一映射为 gRPC status code
4. **分页模式**：defaultPageSize + maxPageSize + clampPagination()
5. **main.go 模式**：run() 返回 error + graceful shutdown + health check + reflection
6. **sqlc 按需选列**：敏感字段（password_hash）和大字段（picture）在 List 查询中排除

---

*状态：Store Service 实施完成 —— 准备实施下一个微服务*
