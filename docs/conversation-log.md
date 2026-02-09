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

---

### 阶段 3：Film Service 实施

**用户**：下一个服务选 film-service，直接完成全部 6 个 step，中途无需确认。

**助手**：选择影片服务（film-service）作为第二个实现的微服务。原因：涉及 6 张表（film, actor, category, language, film_actor, film_category），包含多对多关系、全文搜索、自定义 ENUM/DOMAIN 等复杂场景，能建立更丰富的模式参考。后续 rental-service 和 BFF 层都依赖 film-service。

#### 前置变更：sqlc 查询目录重组

原有 `sqlc.yaml` 的 `queries: "sql/queries"` 会递归扫描子目录，无法按服务隔离。重组方案：

- `sql/queries/store.sql` → `sql/queries/store/store.sql`
- `sql/queries/staff.sql` → `sql/queries/store/staff.sql`
- 新建 `sql/queries/film/` 放 film 相关查询
- `sqlc.yaml` 拆分为两个 sql block，分别输出到各自的 sqlcgen 目录

#### 关键设计决策

1. **四个 gRPC 服务在一个 proto 文件中**：FilmService（核心，12 RPCs）、ActorService（6 RPCs）、CategoryService（只读，2 RPCs）、LanguageService（只读，2 RPCs）
2. **GetFilm 返回 FilmDetail**：包含 actors、categories、language 名称，减少 BFF 调用轮次
3. **多对多关联操作归 FilmService**：AddActorToFilm / RemoveActorFromFilm / AddCategoryToFilm / RemoveCategoryFromFilm
4. **货币字段用 string**：rental_rate、replacement_cost 用 string 避免浮点精度问题（pgtype.Numeric ↔ string）
5. **mpaa_rating ENUM → string**：service 层校验合法值 {"G","PG","PG-13","R","NC-17"}
6. **全文搜索用 plainto_tsquery**：安全处理用户输入，按相关度排序
7. **DeleteFilm 保留但处理 FK 约束**：inventory 表 FK RESTRICT，违约时映射为 FailedPrecondition
8. **新增 ErrForeignKey 哨兵错误**：映射到 gRPC codes.FailedPrecondition
9. **language.name 需 trim**：character(20) 列有空格填充，repository 层 strings.TrimRight
10. **默认端口 50052**：与 store-service 的 50051 区分

#### Step 1：Proto 定义 + SQL 查询 + 工具配置

**新增文件：**

1. **`proto/film/v1/film.proto`** — 四个 gRPC 服务定义
   - `FilmService`：GetFilm、ListFilms、SearchFilms、ListFilmsByCategory、ListFilmsByActor、CreateFilm、UpdateFilm、DeleteFilm、AddActorToFilm、RemoveActorFromFilm、AddCategoryToFilm、RemoveCategoryFromFilm
   - `ActorService`：GetActor、ListActors、ListActorsByFilm、CreateActor、UpdateActor、DeleteActor
   - `CategoryService`：GetCategory、ListCategories
   - `LanguageService`：GetLanguage、ListLanguages
   - 消息类型：Film（列表用）、FilmDetail（含 actors/categories/language_name）、Actor、Category、Language

2. **`sql/queries/film/films.sql`** — 13 个查询
   - Get/List/Count（基础 CRUD）
   - SearchFilms / CountSearchFilms（全文搜索，plainto_tsquery + ts_rank）
   - ListFilmsByCategory / CountFilmsByCategory（按分类过滤）
   - ListFilmsByActor / CountFilmsByActor（按演员过滤）
   - CreateFilm / UpdateFilm / DeleteFilm

3. **`sql/queries/film/actor.sql`** — 7 个查询（Get/List/Count/ListByFilm/Create/Update/Delete）

4. **`sql/queries/film/category.sql`** — 4 个查询（Get/List/Count/ListByFilm，只读）

5. **`sql/queries/film/language.sql`** — 3 个查询（Get/List/Count，只读）

6. **`sql/queries/film/film_actor.sql`** — 2 个（Add/Remove）

7. **`sql/queries/film/film_category.sql`** — 2 个（Add/Remove）

**移动文件：**
- `sql/queries/store.sql` → `sql/queries/store/store.sql`
- `sql/queries/staff.sql` → `sql/queries/store/staff.sql`

**修改文件：**
- `sqlc.yaml` — 拆分为 store + film 两个 sql block
- `Makefile` — 添加 `run-film` target

#### Step 2：运行代码生成

- `make proto-gen` → `gen/proto/film/v1/film.pb.go` + `film_grpc.pb.go`
- `make sqlc-gen` → `internal/film/repository/sqlcgen/` 下 8 个文件（db.go, models.go, films.sql.go, actor.sql.go, category.sql.go, language.sql.go, film_actor.sql.go, film_category.sql.go）
- 验证 store-service 编译不受影响

#### Step 3：Repository 层

**新增文件：**

1. **`internal/film/model/model.go`** — 领域模型
   - `Film`：FilmID, Title, Description, ReleaseYear(int32), LanguageID, OriginalLanguageID, RentalDuration(int16), RentalRate(string), Length(int16), ReplacementCost(string), Rating(string), SpecialFeatures([]string), LastUpdate
   - `FilmDetail`：嵌入 Film + LanguageName, OriginalLanguageName, Actors, Categories
   - `Actor`：ActorID, FirstName, LastName, LastUpdate
   - `Category`：CategoryID, Name, LastUpdate
   - `Language`：LanguageID, Name(trimmed), LastUpdate

2. **`internal/film/repository/helpers.go`** — 类型转换辅助函数
   - `textToString` / `stringToText`（pgtype.Text ↔ string）
   - `numericToString` / `stringToNumeric`（pgtype.Numeric ↔ string，使用 big.Float 精确转换）
   - `int4ToInt32` / `int32ToInt4`、`int2ToInt16` / `int16ToInt2`（nullable 整数）
   - `yearToInt32`（interface{} → int32，处理 year domain 的类型断言）
   - `ratingToString` / `stringToRating`（NullMpaaRating ↔ string）
   - `trimLanguageName`（去除 character(20) 的空格填充）
   - `filmFields` + `convertFilmFields` + `filmConvertedFields`（统一 film 行类型转换）

3. **`internal/film/repository/film_repository.go`**
   - 接口 `FilmRepository`（17 个方法）
   - `CreateFilmParams` / `UpdateFilmParams` 参数结构体
   - 7 个行转换函数（filmFromGetRow, filmFromCreateRow, filmFromUpdateRow, filmsFromListRows, filmsFromSearchRows, filmsFromCategoryRows, filmsFromActorRows）

4. **`internal/film/repository/actor_repository.go`** — 接口 `ActorRepository`（7 个方法）

5. **`internal/film/repository/category_repository.go`** — 接口 `CategoryRepository`（4 个方法，只读）

6. **`internal/film/repository/language_repository.go`** — 接口 `LanguageRepository`（3 个方法，只读，含 name trim）

#### Step 4：Service 层

**新增文件：**

1. **`internal/film/service/errors.go`** — 四个哨兵错误
   - `ErrNotFound`、`ErrInvalidArgument`、`ErrAlreadyExists`、`ErrForeignKey`（新增）

2. **`internal/film/service/pgutil.go`**
   - `isUniqueViolation()` — PG 错误码 23505
   - `isForeignKeyViolation()` — PG 错误码 23503（新增，用于 DeleteFilm/DeleteActor）

3. **`internal/film/service/pagination.go`** — clampPagination（同 store-service 模式）

4. **`internal/film/service/film_service.go`**
   - `FilmService` 持有全部 4 个 repo（filmRepo, actorRepo, categoryRepo, languageRepo）
   - GetFilm：获取 film 后聚合 language name、actors（ListActorsByFilm）、categories（ListCategoriesByFilm）
   - CreateFilm/UpdateFilm：调用 `validateFilmParams()` 校验 title 非空、language 存在、rating 合法、rental_rate/replacement_cost 为合法数值
   - DeleteFilm：FK 违约 → ErrForeignKey
   - AddActorToFilm/AddCategoryToFilm：验证双方存在，唯一约束冲突 → ErrAlreadyExists

5. **`internal/film/service/actor_service.go`** — CRUD + ListByFilm + DeleteActor（FK → ErrForeignKey）

6. **`internal/film/service/category_service.go`** — GetCategory + ListCategories（只读）

7. **`internal/film/service/language_service.go`** — GetLanguage + ListLanguages（只读）

#### Step 5：Handler 层

**新增文件：**

1. **`internal/film/handler/convert.go`**
   - `toGRPCError()`：新增 ErrForeignKey → codes.FailedPrecondition 映射
   - `filmToProto()` / `filmDetailToProto()` / `actorToProto()` / `categoryToProto()` / `languageToProto()`

2. **`internal/film/handler/film_handler.go`** — 12 个 RPC 实现 + `toFilmListResponse()` 辅助

3. **`internal/film/handler/actor_handler.go`** — 6 个 RPC 实现 + `toActorListResponse()` 辅助

4. **`internal/film/handler/category_handler.go`** — 2 个 RPC 实现

5. **`internal/film/handler/language_handler.go`** — 2 个 RPC 实现

#### Step 6：Main 入口 + 配置

**新增文件：**

1. **`internal/film/config/config.go`** — DATABASE_URL（必填）, GRPC_PORT（默认 "50052"）, LOG_LEVEL

2. **`cmd/film-service/main.go`**
   - 依赖注入：4 repo → 4 service → 4 handler
   - gRPC server 注册 4 个服务
   - Health check 注册 4 个服务状态
   - Reflection + graceful shutdown

**修改文件：**
- `.gitignore` — 二进制名改为根目录锚定（`/film-service` 而非 `film-service`），避免匹配 `cmd/film-service/` 目录

**Git 提交**：`feat(film): add film-service with gRPC APIs for films, actors, categories, and languages`

#### 遇到的问题及解决

1. **sqlc 静默跳过 `film.sql`**：文件名与 `film` 表名冲突，sqlc 不报错但不生成代码。改名为 `films.sql` 后解决
2. **helpers.go 类型错误**：`filmConvertedFields.LastUpdate` 误定义为 `fmt.Stringer`，应为 `time.Time`，编译时发现并修复
3. **`.gitignore` 匹配过宽**：`film-service`（无前缀 `/`）匹配了 `cmd/film-service/` 目录，导致 main.go 无法 git add。改为 `/film-service`（根目录锚定）解决。同时修复了所有 7 个服务二进制名

#### Film Service 文件清单

```
cmd/film-service/
└── main.go                              # 服务入口、依赖注入、graceful shutdown

internal/film/
├── config/
│   └── config.go                        # 环境变量配置（端口默认 50052）
├── handler/
│   ├── convert.go                       # proto ↔ 领域模型转换、错误映射
│   ├── film_handler.go                  # FilmService gRPC 实现（12 RPCs）
│   ├── actor_handler.go                 # ActorService gRPC 实现（6 RPCs）
│   ├── category_handler.go              # CategoryService gRPC 实现（2 RPCs）
│   └── language_handler.go              # LanguageService gRPC 实现（2 RPCs）
├── model/
│   └── model.go                         # Film, FilmDetail, Actor, Category, Language
├── repository/
│   ├── film_repository.go               # FilmRepository 接口 + pgx 实现（17 方法）
│   ├── actor_repository.go              # ActorRepository 接口 + pgx 实现（7 方法）
│   ├── category_repository.go           # CategoryRepository 接口（只读，4 方法）
│   ├── language_repository.go           # LanguageRepository 接口（只读，3 方法）
│   ├── helpers.go                       # pgtype 转换辅助（numeric, rating, year, text 等）
│   └── sqlcgen/                         # sqlc 自动生成（勿手动修改）
│       ├── db.go
│       ├── models.go
│       ├── films.sql.go
│       ├── actor.sql.go
│       ├── category.sql.go
│       ├── language.sql.go
│       ├── film_actor.sql.go
│       └── film_category.sql.go
└── service/
    ├── errors.go                        # ErrNotFound, ErrInvalidArgument, ErrAlreadyExists, ErrForeignKey
    ├── pgutil.go                        # isUniqueViolation + isForeignKeyViolation
    ├── pagination.go                    # clampPagination
    ├── film_service.go                  # Film 业务逻辑（含数据聚合 + 校验）
    ├── actor_service.go                 # Actor 业务逻辑
    ├── category_service.go              # Category 业务逻辑（只读）
    └── language_service.go              # Language 业务逻辑（只读）

proto/film/v1/
└── film.proto                           # 4 个 gRPC 服务定义

gen/proto/film/v1/
├── film.pb.go                           # protobuf 生成代码
└── film_grpc.pb.go                      # gRPC 生成代码

sql/queries/film/
├── films.sql                            # Film 表查询（含全文搜索）
├── actor.sql                            # Actor 表查询
├── category.sql                         # Category 表查询（只读）
├── language.sql                         # Language 表查询（只读）
├── film_actor.sql                       # 关联表 Add/Remove
└── film_category.sql                    # 关联表 Add/Remove
```

#### 新增可复用模式（相比 store-service）

1. **多对多关联管理**：junction table 的 Add/Remove 操作归主实体 Service，校验双方存在后操作
2. **数据聚合模式**：GetFilm 在 service 层聚合多个 repo 查询结果（film + language + actors + categories）→ FilmDetail
3. **特殊类型转换集中化**：helpers.go 封装所有 pgtype ↔ Go 标准类型转换，包括 numeric(金额)、ENUM、DOMAIN、ARRAY
4. **ErrForeignKey 错误体系**：新增 FK 违约哨兵错误 → gRPC FailedPrecondition，用于 Delete 操作
5. **全文搜索模式**：plainto_tsquery + ts_rank 排序，Count 变体用于分页
6. **sqlc 多服务目录结构**：sql/queries/{service}/ 子目录 + sqlc.yaml 多 sql block

#### 特殊类型处理总结

| DB 类型 | sqlc 生成类型 | Go 模型类型 | 转换方式 |
|---------|-------------|------------|---------|
| mpaa_rating ENUM | NullMpaaRating | string | ratingToString / stringToRating |
| year DOMAIN | interface{} | int32 | yearToInt32（类型断言） |
| text[] ARRAY | []string | []string | 直接映射 |
| numeric(4,2) | pgtype.Numeric | string | numericToString / stringToNumeric（big.Float） |
| tsvector | interface{} | 不暴露 | 仅在 SQL 查询中使用 |
| character(20) | string | string | trimLanguageName（TrimRight 空格） |
| text (nullable) | pgtype.Text | string | textToString / stringToText |
| int4 (nullable) | pgtype.Int4 | int32 | int4ToInt32 / int32ToInt4（0=NULL） |
| int2 (nullable) | pgtype.Int2 | int16 | int2ToInt16 / int16ToInt2（0=NULL） |

---

### 阶段 4：Customer Service 实施

**用户**：下一个服务从简单的开始。

**助手**：分析剩余三个服务的复杂度：
1. **customer-service** — 最简单：customer + address 两张表，无复杂业务逻辑，模式与 store-service 类似
2. **rental-service** — 中等：rental + inventory，涉及库存分配、归还逻辑、跨服务查询
3. **payment-service** — 较复杂：payment 分区表，涉及金额计算、跨服务关联

选择 customer-service，管理 `customer`、`address`、`city`、`country` 共 4 张表。address → city → country 构成层级关系链。所有表均为标准 PostgreSQL 类型（text, integer, boolean, date, timestamptz），无特殊类型（ENUM/DOMAIN/ARRAY），比 film-service 简单得多。

#### 关键设计决策

1. **四个 gRPC 服务在一个 proto 文件中**：CustomerService（核心 CRUD）、AddressService（CRUD）、CityService（只读）、CountryService（只读），同一进程端口 50053
2. **GetCustomer 返回 CustomerDetail**：聚合 customer + address + city name + country name，减少 BFF 调用轮次；ListCustomers 返回基础 Customer
3. **Address 需要完整 CRUD**：客户创建/更新需要管理地址；City/Country 为只读参考数据
4. **ListCustomersByStore**：按 store_id 过滤客户列表
5. **active 状态同步**：`activebool` (boolean) 为源，`active` (integer) 为冗余字段，proto 暴露为 `bool active`，Create/Update 同时写入两个字段保持一致
6. **password_hash 不暴露**：存在于 DB 但不出现在 proto 响应中；认证功能留给后续独立服务
7. **email 唯一约束**：unique violation → AlreadyExists
8. **DeleteCustomer 处理 FK 约束**：rental/payment 表 FK RESTRICT → FailedPrecondition
9. **DeleteAddress 处理 FK 约束**：customer/store 表 FK RESTRICT → FailedPrecondition
10. **create_date 只读**：由 DB DEFAULT CURRENT_DATE 生成，不接受客户端输入

#### Step 1：Proto 定义 + SQL 查询 + 工具配置

**新增文件：**

1. **`proto/customer/v1/customer.proto`** — 四个 gRPC 服务定义
   - `CustomerService`：GetCustomer、ListCustomers、ListCustomersByStore、CreateCustomer、UpdateCustomer、DeleteCustomer（6 RPCs）
   - `AddressService`：GetAddress、ListAddresses、CreateAddress、UpdateAddress、DeleteAddress（5 RPCs）
   - `CityService`：GetCity、ListCities（2 RPCs）
   - `CountryService`：GetCountry、ListCountries（2 RPCs）
   - 消息类型：Customer（列表用）、CustomerDetail（含 address/city/country 信息）、Address、City、Country

2. **`sql/queries/customer/customer.sql`** — 8 个查询
   - Get/List/Count（基础 CRUD）
   - ListCustomersByStore / CountCustomersByStore（按门店过滤）
   - CreateCustomer / UpdateCustomer（含 activebool + active 双字段）
   - DeleteCustomer

3. **`sql/queries/customer/address.sql`** — 6 个查询（Get/List/Count/Create/Update/Delete）

4. **`sql/queries/customer/city.sql`** — 3 个查询（Get/List/Count，只读）

5. **`sql/queries/customer/country.sql`** — 3 个查询（Get/List/Count，只读）

**修改文件：**
- `sqlc.yaml` — 新增 customer sql block（第三个 sql block）
- `Makefile` — 添加 `run-customer` target

#### Step 2：运行代码生成

- `make proto-gen` → `gen/proto/customer/v1/customer.pb.go` + `customer_grpc.pb.go`
- `make sqlc-gen` → `internal/customer/repository/sqlcgen/` 下 6 个文件（db.go, models.go, customer.sql.go, address.sql.go, city.sql.go, country.sql.go）
- 验证 store-service 和 film-service 编译不受影响

#### Step 3：Repository 层

**新增文件：**

1. **`internal/customer/model/model.go`** — 领域模型
   - `Customer`：CustomerID, StoreID, FirstName, LastName, Email(string), AddressID, Active(bool), CreateDate(time.Time), LastUpdate
   - `CustomerDetail`：嵌入 Customer + Address, Address2, District, CityName, CountryName, PostalCode, Phone
   - `Address`：AddressID, Address, Address2, District, CityID, PostalCode, Phone, LastUpdate
   - `City`：CityID, City, CountryID, LastUpdate
   - `Country`：CountryID, Country, LastUpdate

2. **`internal/customer/repository/helpers.go`** — 类型转换辅助函数
   - `textToString` / `stringToText`（pgtype.Text ↔ string，NULL → ""）
   - `dateToTime`（pgtype.Date → time.Time）
   - `timestamptzToTime`（pgtype.Timestamptz → time.Time）
   - `boolToActive`（bool → pgtype.Int4，true=1/false=0，同步 active 冗余字段）

3. **`internal/customer/repository/customer_repository.go`**
   - 接口 `CustomerRepository`（8 个方法）
   - `CreateCustomerParams` / `UpdateCustomerParams` 参数结构体
   - `toCustomerModel()` 通用转换函数（接收各行类型的公共字段作为参数）

4. **`internal/customer/repository/address_repository.go`** — 接口 `AddressRepository`（6 个方法）

5. **`internal/customer/repository/city_repository.go`** — 接口 `CityRepository`（3 个方法，只读）

6. **`internal/customer/repository/country_repository.go`** — 接口 `CountryRepository`（3 个方法，只读）

#### Step 4：Service 层

**新增文件：**

1. **`internal/customer/service/errors.go`** — 四个哨兵错误（同 film-service 模式）

2. **`internal/customer/service/pgutil.go`** — isUniqueViolation + isForeignKeyViolation

3. **`internal/customer/service/pagination.go`** — clampPagination

4. **`internal/customer/service/customer_service.go`**
   - `CustomerService` 持有全部 4 个 repo（customerRepo, addressRepo, cityRepo, countryRepo）
   - GetCustomer：获取 customer 后链式聚合 address → city → country → CustomerDetail
   - CreateCustomer/UpdateCustomer：校验 first_name/last_name 非空、store_id > 0、address_id 存在
   - FK 违约 → ErrInvalidArgument（store_id/address_id 不存在）
   - 唯一约束 → ErrAlreadyExists（email 重复）
   - DeleteCustomer：FK 违约 → ErrForeignKey（被 rental/payment 引用）

5. **`internal/customer/service/address_service.go`**
   - `AddressService` 持有 addressRepo + cityRepo
   - CreateAddress/UpdateAddress：校验 address/district/phone 非空、city_id 存在
   - DeleteAddress：FK 违约 → ErrForeignKey（被 customer/store 引用）

6. **`internal/customer/service/city_service.go`** — GetCity + ListCities（只读）

7. **`internal/customer/service/country_service.go`** — GetCountry + ListCountries（只读）

#### Step 5：Handler 层

**新增文件：**

1. **`internal/customer/handler/convert.go`**
   - `toGRPCError()`：同 film-service 模式（含 ErrForeignKey → FailedPrecondition）
   - `customerToProto()` / `customerDetailToProto()` / `addressToProto()` / `cityToProto()` / `countryToProto()`

2. **`internal/customer/handler/customer_handler.go`** — 6 个 RPC 实现 + `toCustomerListResponse()` 辅助

3. **`internal/customer/handler/address_handler.go`** — 5 个 RPC 实现 + `toAddressListResponse()` 辅助

4. **`internal/customer/handler/city_handler.go`** — 2 个 RPC 实现

5. **`internal/customer/handler/country_handler.go`** — 2 个 RPC 实现

#### Step 6：Main 入口 + 配置

**新增文件：**

1. **`internal/customer/config/config.go`** — DATABASE_URL（必填）, GRPC_PORT（默认 "50053"）, LOG_LEVEL

2. **`cmd/customer-service/main.go`**
   - 依赖注入：4 repo → 4 service → 4 handler
   - gRPC server 注册 4 个服务
   - Health check 注册 4 个服务状态
   - Reflection + graceful shutdown

**Git 提交**：`feat: implement customer-service with customer, address, city, country`

#### 与 Film Service 的对比

| 维度 | film-service | customer-service |
|------|-------------|-----------------|
| 表数量 | 6（含 2 张关联表） | 4（无关联表） |
| 特殊 PG 类型 | ENUM, DOMAIN, ARRAY, Numeric, tsvector | 无（全部标准类型） |
| 数据聚合模式 | 扇出式（film + language + actors + categories） | 链式（customer → address → city → country） |
| 全文搜索 | 有（plainto_tsquery + ts_rank） | 无 |
| 多对多关联 | 有（film_actor, film_category） | 无 |
| helpers.go 复杂度 | 高（numeric, rating, year, text[] 等） | 低（Text, Date, Timestamptz, boolToActive） |
| Repository 方法总数 | 31（17+7+4+3） | 20（8+6+3+3） |
| Service RPC 总数 | 22（12+6+2+2） | 15（6+5+2+2） |

#### Customer Service 文件清单

```
cmd/customer-service/
└── main.go                              # 服务入口、依赖注入、graceful shutdown

internal/customer/
├── config/
│   └── config.go                        # 环境变量配置（端口默认 50053）
├── handler/
│   ├── convert.go                       # proto ↔ 领域模型转换、错误映射
│   ├── customer_handler.go              # CustomerService gRPC 实现（6 RPCs）
│   ├── address_handler.go               # AddressService gRPC 实现（5 RPCs）
│   ├── city_handler.go                  # CityService gRPC 实现（2 RPCs）
│   └── country_handler.go              # CountryService gRPC 实现（2 RPCs）
├── model/
│   └── model.go                         # Customer, CustomerDetail, Address, City, Country
├── repository/
│   ├── customer_repository.go           # CustomerRepository 接口 + pgx 实现（8 方法）
│   ├── address_repository.go            # AddressRepository 接口 + pgx 实现（6 方法）
│   ├── city_repository.go               # CityRepository 接口（只读，3 方法）
│   ├── country_repository.go            # CountryRepository 接口（只读，3 方法）
│   ├── helpers.go                       # pgtype 转换辅助（Text, Date, Timestamptz, boolToActive）
│   └── sqlcgen/                         # sqlc 自动生成（勿手动修改）
│       ├── db.go
│       ├── models.go
│       ├── customer.sql.go
│       ├── address.sql.go
│       ├── city.sql.go
│       └── country.sql.go
└── service/
    ├── errors.go                        # ErrNotFound, ErrInvalidArgument, ErrAlreadyExists, ErrForeignKey
    ├── pgutil.go                        # isUniqueViolation + isForeignKeyViolation
    ├── pagination.go                    # clampPagination
    ├── customer_service.go              # Customer 业务逻辑（含链式数据聚合）
    ├── address_service.go               # Address 业务逻辑（含 city 存在性校验）
    ├── city_service.go                  # City 业务逻辑（只读）
    └── country_service.go              # Country 业务逻辑（只读）

proto/customer/v1/
└── customer.proto                       # 4 个 gRPC 服务定义

gen/proto/customer/v1/
├── customer.pb.go                       # protobuf 生成代码
└── customer_grpc.pb.go                  # gRPC 生成代码

sql/queries/customer/
├── customer.sql                         # Customer 表查询（含按门店过滤）
├── address.sql                          # Address 表查询
├── city.sql                             # City 表查询（只读）
└── country.sql                          # Country 表查询（只读）
```

#### 服务端口汇总

| 服务 | 端口 | 状态 |
|------|------|------|
| store-service | 50051 | 已完成 |
| film-service | 50052 | 已完成 |
| customer-service | 50053 | 已完成 |
| rental-service | 50054 | 待实现 |
| payment-service | 50055 | 待实现 |

---

*状态：Customer Service 实施完成*

---

### 阶段 5：Rental Service 实施

**用户**：继续下一个服务的实装。

**助手**：选择 rental-service 作为第四个实现的微服务。原因：payment-service 依赖 rental（payment 表 FK 引用 rental），需先完成 rental-service。管理 `rental` 和 `inventory` 两张表，涉及库存可用性检查、归还逻辑、跨表数据聚合等中等复杂度的业务场景。

#### 关键设计决策

1. **两个 gRPC 服务在一个 proto 文件中**：RentalService（核心，8 RPCs）、InventoryService（8 RPCs），同一进程端口 50054
2. **GetRental 返回 RentalDetail**：聚合 rental + customer name + film title + store_id，减少 BFF 调用轮次
3. **ReturnRental 替代 UpdateRental**：归还是 rental 唯一的状态变更操作，语义更明确
4. **CreateRental 验证库存可用性**：先检查 inventory 是否存在，再检查是否已被借出（`return_date IS NULL`）
5. **跨表查询用 SQL 而非跨服务 gRPC**：GetCustomerName、GetFilmTitleByInventory 直接查同一数据库的 customer/film 表，避免引入服务间依赖
6. **return_date 为 NULL 表示未归还**：DB NULL → pgtype.Timestamptz{Valid:false} → Go time.Time 零值 → proto nil Timestamp（handler 层 IsZero 判断）
7. **ReturnRental 防重复归还**：SQL `WHERE return_date IS NULL` 条件，已归还的 rental 返回 ErrNotFound
8. **sqlc interface{} 处理**：`first_name || ' ' || last_name` 拼接表达式 sqlc 映射为 `interface{}`，repository 层用类型断言处理
9. **ListOverdueRentals**：查询 `return_date IS NULL` 的所有租赁记录（未归还 = 逾期）
10. **InventoryService 只读 + Create/Delete**：无 UpdateInventory（库存项无需修改，只有增删）
11. **ListAvailableInventory**：按 film_id + store_id 过滤，用 `NOT EXISTS` 子查询排除已借出的库存
12. **CheckInventoryAvailability**：验证库存存在后检查可用性，返回 bool

#### Step 1：Proto 定义 + SQL 查询 + 工具配置

**新增文件：**

1. **`proto/rental/v1/rental.proto`** — 两个 gRPC 服务定义
   - `RentalService`：GetRental、ListRentals、ListRentalsByCustomer、ListRentalsByInventory、ListOverdueRentals、CreateRental、ReturnRental、DeleteRental（8 RPCs）
   - `InventoryService`：GetInventory、ListInventory、ListInventoryByFilm、ListInventoryByStore、CheckInventoryAvailability、ListAvailableInventory、CreateInventory、DeleteInventory（8 RPCs）
   - 消息类型：Rental（return_date 可选）、RentalDetail（含 customer_name/film_title/store_id）、Inventory

2. **`sql/queries/rental/rental.sql`** — 15 个查询
   - Get/List/Count（基础 CRUD）
   - ListRentalsByCustomer / CountRentalsByCustomer（按客户过滤）
   - ListRentalsByInventory / CountRentalsByInventory（按库存过滤）
   - ListOverdueRentals / CountOverdueRentals（`return_date IS NULL`）
   - CreateRental（`rental_date = now()`）
   - ReturnRental（`return_date = now() WHERE return_date IS NULL`，RETURNING *）
   - DeleteRental
   - GetCustomerName（`first_name || ' ' || last_name`，跨表查询 customer）
   - GetFilmTitleByInventory（JOIN inventory + film，返回 title + store_id）
   - IsInventoryAvailable（`NOT EXISTS` 子查询检查未归还租赁）

3. **`sql/queries/rental/inventory.sql`** — 11 个查询
   - Get/List/Count（基础 CRUD）
   - ListInventoryByFilm / CountInventoryByFilm（按影片过滤）
   - ListInventoryByStore / CountInventoryByStore（按门店过滤）
   - ListAvailableInventory / CountAvailableInventory（按 film_id + store_id + NOT EXISTS 过滤）
   - CreateInventory / DeleteInventory（无 Update）

**修改文件：**
- `sqlc.yaml` — 新增 rental sql block（第四个 sql block）
- `Makefile` — 添加 `run-rental` target

#### Step 2：运行代码生成

- `make proto-gen` → `gen/proto/rental/v1/rental.pb.go` + `rental_grpc.pb.go`
- `make sqlc-gen` → `internal/rental/repository/sqlcgen/` 下 4 个文件（db.go, models.go, rental.sql.go, inventory.sql.go）
- sqlcgen 模型：Rental 用 `pgtype.Timestamptz`（RentalDate, ReturnDate, LastUpdate）；GetCustomerName 返回 `interface{}`
- 验证 store-service、film-service、customer-service 编译不受影响

#### Step 3：Repository 层

**新增文件：**

1. **`internal/rental/model/model.go`** — 领域模型
   - `Rental`：RentalID, RentalDate(time.Time), InventoryID, CustomerID, ReturnDate(time.Time, 零值=未归还), StaffID, LastUpdate
   - `RentalDetail`：嵌入 Rental + CustomerName, FilmTitle, StoreID
   - `Inventory`：InventoryID, FilmID, StoreID, LastUpdate

2. **`internal/rental/repository/helpers.go`** — 仅 `timestamptzToTime`（比 customer 更简单，无 Date/Text 类型）

3. **`internal/rental/repository/rental_repository.go`**
   - 接口 `RentalRepository`（15 个方法）
   - `CreateRentalParams` 参数结构体（InventoryID, CustomerID, StaffID）
   - `GetCustomerName`：处理 sqlc 的 `interface{}` 返回，用 `val.(string)` 类型断言
   - `GetFilmTitleByInventory`：返回 (string, int32, error) — title 和 store_id
   - `ReturnRental`：pgx.ErrNoRows → ErrNotFound（覆盖"不存在"和"已归还"两种情况）
   - `toRentalModel` / `toRentalModels` 辅助函数

4. **`internal/rental/repository/inventory_repository.go`**
   - 接口 `InventoryRepository`（11 个方法）
   - `CreateInventoryParams` 参数结构体（FilmID, StoreID）
   - 无 Update 方法
   - `toInventoryModel` / `toInventoryModels` 辅助函数

#### Step 4：Service 层

**新增文件：**

1. **`internal/rental/service/errors.go`** — 四个哨兵错误（同前三个服务模式）

2. **`internal/rental/service/pgutil.go`** — isUniqueViolation + isForeignKeyViolation

3. **`internal/rental/service/pagination.go`** — clampPagination

4. **`internal/rental/service/rental_service.go`**
   - `RentalService` 持有 rentalRepo + inventoryRepo
   - GetRental：获取 rental 后聚合 customer name（GetCustomerName）+ film title & store_id（GetFilmTitleByInventory）→ RentalDetail
   - CreateRental：校验 inventory_id/customer_id/staff_id > 0 → 验证 inventory 存在 → 检查库存可用性（IsInventoryAvailable）→ 创建 rental
   - ReturnRental：ErrNotFound 消息为"not found or already returned"
   - DeleteRental：先验证存在，FK 违约 → ErrForeignKey（被 payment 引用）
   - 五种列表查询：ListRentals, ListRentalsByCustomer, ListRentalsByInventory, ListOverdueRentals（各带分页和总数）

5. **`internal/rental/service/inventory_service.go`**
   - `InventoryService` 持有 inventoryRepo + rentalRepo
   - CheckInventoryAvailability：先验证 inventory 存在，再调用 IsInventoryAvailable
   - ListAvailableInventory：校验 film_id + store_id > 0
   - CreateInventory：校验 film_id + store_id > 0，FK 违约 → ErrInvalidArgument
   - DeleteInventory：先验证存在，FK 违约 → ErrForeignKey（被 rental 引用）
   - 三种列表查询：ListInventory, ListInventoryByFilm, ListInventoryByStore

#### Step 5：Handler 层

**新增文件：**

1. **`internal/rental/handler/convert.go`**
   - `toGRPCError()`：同前三个服务模式
   - `rentalToProto()`：return_date 零值判断（`!r.ReturnDate.IsZero()` → 设置 proto Timestamp，否则 nil）
   - `rentalDetailToProto()` / `inventoryToProto()`

2. **`internal/rental/handler/rental_handler.go`** — 8 个 RPC 实现 + `toRentalListResponse()` 辅助

3. **`internal/rental/handler/inventory_handler.go`** — 8 个 RPC 实现 + `toInventoryListResponse()` 辅助

#### Step 6：Main 入口 + 配置

**新增文件：**

1. **`internal/rental/config/config.go`** — DATABASE_URL（必填）, GRPC_PORT（默认 "50054"）, LOG_LEVEL

2. **`cmd/rental-service/main.go`**
   - 依赖注入：2 repo → 2 service（交叉注入：RentalService 用两个 repo，InventoryService 也用两个 repo）→ 2 handler
   - gRPC server 注册 2 个服务
   - Health check 注册 2 个服务状态
   - Reflection + graceful shutdown

**Git 提交**：`feat(rental-service): implement rental and inventory gRPC services`

#### 与前三个服务的对比

| 维度 | store-service | film-service | customer-service | rental-service |
|------|-------------|-------------|-----------------|---------------|
| 表数量 | 2 | 6（含 2 关联表） | 4 | 2 |
| gRPC 服务数 | 2 | 4 | 4 | 2 |
| RPC 总数 | 13 | 22 | 15 | 16 |
| Repository 方法数 | 20 | 31 | 20 | 26（15+11） |
| 数据聚合 | 无 | 扇出式 | 链式 | 扇出式（跨表 SQL） |
| 特殊业务逻辑 | soft delete | 全文搜索、多对多 | 双字段同步 | 库存可用性、归还逻辑 |
| 跨表查询 | 无 | 无 | 无 | 有（GetCustomerName, GetFilmTitleByInventory） |
| 特殊类型处理 | 无 | ENUM, DOMAIN, ARRAY, Numeric | boolToActive | interface{} 类型断言 |

#### Rental Service 文件清单

```
cmd/rental-service/
└── main.go                              # 服务入口、依赖注入、graceful shutdown

internal/rental/
├── config/
│   └── config.go                        # 环境变量配置（端口默认 50054）
├── handler/
│   ├── convert.go                       # proto ↔ 领域模型转换、错误映射
│   ├── rental_handler.go                # RentalService gRPC 实现（8 RPCs）
│   └── inventory_handler.go             # InventoryService gRPC 实现（8 RPCs）
├── model/
│   └── model.go                         # Rental, RentalDetail, Inventory
├── repository/
│   ├── rental_repository.go             # RentalRepository 接口 + pgx 实现（15 方法）
│   ├── inventory_repository.go          # InventoryRepository 接口 + pgx 实现（11 方法）
│   ├── helpers.go                       # pgtype 转换辅助（timestamptzToTime）
│   └── sqlcgen/                         # sqlc 自动生成（勿手动修改）
│       ├── db.go
│       ├── models.go
│       ├── rental.sql.go
│       └── inventory.sql.go
└── service/
    ├── errors.go                        # ErrNotFound, ErrInvalidArgument, ErrAlreadyExists, ErrForeignKey
    ├── pgutil.go                        # isUniqueViolation + isForeignKeyViolation
    ├── pagination.go                    # clampPagination
    ├── rental_service.go                # Rental 业务逻辑（含库存可用性检查 + 跨表聚合）
    └── inventory_service.go             # Inventory 业务逻辑（含可用性查询）

proto/rental/v1/
└── rental.proto                         # 2 个 gRPC 服务定义

gen/proto/rental/v1/
├── rental.pb.go                         # protobuf 生成代码
└── rental_grpc.pb.go                    # gRPC 生成代码

sql/queries/rental/
├── rental.sql                           # Rental 表查询（含跨表查询 + 归还逻辑）
└── inventory.sql                        # Inventory 表查询（含可用库存过滤）
```

#### 新增可复用模式（相比前三个服务）

1. **跨表 SQL 查询**：同一数据库内的跨服务数据查询（GetCustomerName, GetFilmTitleByInventory），避免引入 gRPC 服务间依赖
2. **可空时间戳处理**：DB NULL → Go 零值 → proto nil 的完整管道（handler 层 `IsZero()` 判断）
3. **sqlc interface{} 类型断言**：SQL 表达式（如字符串拼接）生成 `interface{}` 时，在 repository 层用类型断言安全转换
4. **业务状态操作替代通用更新**：ReturnRental（语义明确）替代 UpdateRental（通用但模糊）
5. **可用性检查模式**：`NOT EXISTS (SELECT 1 FROM rental WHERE ... AND return_date IS NULL)` 用于库存可用性判断
6. **服务交叉注入**：RentalService 和 InventoryService 都需要对方的 repo（RentalService 需要 inventoryRepo 验证库存存在，InventoryService 需要 rentalRepo 检查可用性）

#### 服务端口汇总

| 服务 | 端口 | 状态 |
|------|------|------|
| store-service | 50051 | 已完成 |
| film-service | 50052 | 已完成 |
| customer-service | 50053 | 已完成 |
| rental-service | 50054 | 已完成 |
| payment-service | 50055 | 待实现 |

---

*状态：Rental Service 实施完成*

---

### 阶段 6：Payment Service 实施

**用户**：继续实现 payment-service。

**助手**：选择 payment-service 作为第五个（也是最后一个）核心微服务。原因：payment 是 FK 图中的叶子节点（无其他表引用 payment），依赖 rental/customer/staff 表，逻辑上应最后实现。管理 `payment` 单张分区表（按月范围分区），涉及金额处理（numeric(5,2)）、跨表数据聚合、日期范围查询（利用分区裁剪）。

#### 关键设计决策

1. **单个 gRPC 服务**：PaymentService（8 RPCs），端口 50055 — 是所有核心服务中最简单的 main（1 repo → 1 service → 1 handler）
2. **GetPayment 返回 PaymentDetail**：聚合 customer_name、staff_name、rental_date，减少 BFF 调用轮次
3. **不提供 UpdatePayment**：财务记录不可变，仅支持 Create/Delete
4. **amount 用 string**：numeric(5,2) 在 proto/Go 模型中均为 string，复用 film-service 的 numericToString/stringToNumeric（big.Float 精确转换）
5. **ListPaymentsByDateRange**：payment 独有的日期范围查询，半开区间 `[start, end)` 利用 PG 分区裁剪
6. **跨表查询用 SQL**：GetCustomerName、GetStaffName（均为 interface{} 类型断言）、GetRentalDate，同 rental-service 模式
7. **DeletePayment 无 FK 顾虑**：payment 是叶子表，无其他表引用，删除不会触发 FK 违约
8. **分区透明**：sqlc 查询对分区表透明操作，PG 自动路由分区，Go 代码无需特殊处理
9. **CreatePayment 用 `now()`**：payment_date 由服务器生成，确保正确的分区路由和时间一致性
10. **所有 List 排序**：`ORDER BY payment_date DESC, payment_id DESC` — 最新交易优先，符合财务记录自然顺序

#### 分区表特殊说明

```sql
CREATE TABLE public.payment (
    payment_id integer NOT NULL,
    customer_id integer NOT NULL,  -- FK → customer ON DELETE RESTRICT
    staff_id integer NOT NULL,     -- FK → staff ON DELETE RESTRICT
    rental_id integer NOT NULL,    -- FK → rental ON DELETE RESTRICT
    amount numeric(5,2) NOT NULL,  -- 最大 $999.99
    payment_date timestamptz NOT NULL,
    PRIMARY KEY (payment_date, payment_id)  -- 复合主键
) PARTITION BY RANGE (payment_date);
```

- 7 个月度分区：payment_p2022_01 ~ payment_p2022_07
- 每个分区有 customer_id 和 staff_id 索引
- 复合主键 (payment_date, payment_id)，按 payment_id 单字段查询仍可工作（PG 扫描所有分区的 PK 索引）

#### Step 1：Proto 定义 + SQL 查询 + 工具配置

**新增文件：**

1. **`proto/payment/v1/payment.proto`** — 单个 gRPC 服务定义
   - `PaymentService`：GetPayment、ListPayments、ListPaymentsByCustomer、ListPaymentsByStaff、ListPaymentsByRental、ListPaymentsByDateRange、CreatePayment、DeletePayment（8 RPCs）
   - 消息类型：Payment（amount=string）、PaymentDetail（含 customer_name/staff_name/rental_date）
   - `ListPaymentsByDateRangeRequest` 含 start_date/end_date（google.protobuf.Timestamp）
   - `CreatePaymentRequest` 无 payment_date（服务器 `now()` 生成）

2. **`sql/queries/payment/payment.sql`** — 17 个查询
   - Get/List/Count（基础 CRUD）
   - ListByCustomer/CountByCustomer、ListByStaff/CountByStaff、ListByRental/CountByRental
   - ListByDateRange/CountByDateRange（`payment_date >= $1 AND payment_date < $2`，半开区间）
   - CreatePayment（`payment_date = now()`）/ DeletePayment
   - GetCustomerName、GetStaffName（跨表，`first_name || ' ' || last_name` → interface{}）
   - GetRentalDate（跨表，从 rental 获取 rental_date → pgtype.Timestamptz）

**修改文件：**
- `sqlc.yaml` — 新增 payment sql block（第 5 个 sql block）
- `Makefile` — 添加 `run-payment` target

#### Step 2：运行代码生成

- `make proto-gen` → `gen/proto/payment/v1/payment.pb.go` + `payment_grpc.pb.go`
- `make sqlc-gen` → `internal/payment/repository/sqlcgen/` 下 3 个文件（db.go, models.go, payment.sql.go）
- sqlcgen 模型：Payment.Amount = `pgtype.Numeric`，PaymentDate = `pgtype.Timestamptz`
- sqlc 为分区表生成了 7 个 PaymentP2022XX 结构体（每个分区一个），但查询只操作父表 Payment
- DateRange 参数：`PaymentDate pgtype.Timestamptz` + `PaymentDate_2 pgtype.Timestamptz`（sqlc 自动命名）
- GetCustomerName / GetStaffName 返回 `interface{}`，GetRentalDate 返回 `pgtype.Timestamptz`
- 验证前 4 个服务编译不受影响

#### Step 3：Repository 层

**新增文件：**

1. **`internal/payment/model/model.go`** — 领域模型
   - `Payment`：PaymentID, CustomerID, StaffID, RentalID, Amount(string), PaymentDate(time.Time)
   - `PaymentDetail`：嵌入 Payment + CustomerName, StaffName, RentalDate(time.Time, 零值=未找到)
   - 注意：Payment 表无 LastUpdate 列（与其他表不同）

2. **`internal/payment/repository/helpers.go`** — 类型转换辅助
   - `timestamptzToTime`（同 rental-service）
   - `timeToTimestamptz`（新增，用于 DateRange 参数 time.Time → pgtype.Timestamptz 转换）
   - `numericToString` / `stringToNumeric` / `max32`（从 film-service 复制，big.Float 精确转换）

3. **`internal/payment/repository/payment_repository.go`**
   - 接口 `PaymentRepository`（16 个方法）
   - `CreatePaymentParams` 参数结构体（CustomerID, StaffID, RentalID, Amount string）
   - `toPaymentModel` 用 `numericToString(row.Amount)` 转换金额
   - `GetCustomerName` / `GetStaffName` 处理 interface{} → string 类型断言
   - `GetRentalDate` 返回 `timestamptzToTime(ts)` 
   - `ListPaymentsByDateRange` / `CountPaymentsByDateRange` 用 `timeToTimestamptz()` 转换参数

#### Step 4：Service 层

**新增文件：**

1. **`internal/payment/service/errors.go`** — 四个哨兵错误（同前四个服务模式）

2. **`internal/payment/service/pgutil.go`** — isUniqueViolation + isForeignKeyViolation

3. **`internal/payment/service/pagination.go`** — clampPagination

4. **`internal/payment/service/payment_service.go`**
   - `PaymentService` 持有单个 repo（最简单的服务依赖）
   - GetPayment：获取 payment 后聚合 customer_name（GetCustomerName）+ staff_name（GetStaffName）+ rental_date（GetRentalDate）→ PaymentDetail；聚合失败不阻断（同 rental-service 模式）
   - CreatePayment：校验 customer_id/staff_id/rental_id > 0，amount 非空；FK 违约 → ErrInvalidArgument
   - DeletePayment：先验证存在再删除（无 FK 顾虑，payment 是叶子表）
   - ListPaymentsByDateRange：校验 start_date < end_date 且均非零值
   - 五种 List 查询各带分页和总数

#### Step 5：Handler 层

**新增文件：**

1. **`internal/payment/handler/convert.go`**
   - `toGRPCError()`：同前四个服务模式
   - `paymentToProto()`：amount 直接作为 string 传递，payment_date 用 `timestamppb.New()`
   - `paymentDetailToProto()`：rental_date 零值判断（`!d.RentalDate.IsZero()` → 设置 Timestamp，否则 nil）

2. **`internal/payment/handler/payment_handler.go`** — 8 个 RPC 实现 + `toPaymentListResponse()` 辅助
   - `ListPaymentsByDateRange`：handler 层检查 nil Timestamp（`req.GetStartDate() == nil` → InvalidArgument），然后 `.AsTime()` 转换
   - 这是所有服务中唯一需要在 handler 层处理 proto Timestamp → time.Time 输入转换的 RPC

#### Step 6：Main 入口 + 配置

**新增文件：**

1. **`internal/payment/config/config.go`** — DATABASE_URL（必填）, GRPC_PORT（默认 "50055"）, LOG_LEVEL

2. **`cmd/payment-service/main.go`**
   - 最简单的 main：1 repo → 1 service → 1 handler
   - gRPC server 注册 1 个服务
   - Health check 注册 1 个服务状态
   - Reflection + graceful shutdown

**Git 提交**：`feat(payment-service): implement payment gRPC service`

#### 与前四个服务的对比

| 维度 | store | film | customer | rental | payment |
|------|-------|------|----------|--------|---------|
| 表数量 | 2 | 6（含 2 关联表） | 4 | 2 | 1（分区表） |
| gRPC 服务数 | 2 | 4 | 4 | 2 | 1 |
| RPC 总数 | 13 | 22 | 15 | 16 | 8 |
| Repository 方法数 | 20 | 31 | 20 | 26 | 16 |
| 数据聚合 | 无 | 扇出式 | 链式 | 扇出式(SQL) | 扇出式(SQL) |
| 特殊业务逻辑 | soft delete | 全文搜索、多对多 | 双字段同步 | 库存可用性 | 不可变记录 |
| 特殊类型处理 | 无 | ENUM,DOMAIN,ARRAY,Numeric | boolToActive | interface{} | Numeric + interface{} |
| 独有特性 | 循环 FK | 全文搜索 | 层级聚合 | 跨表查询 | 分区表、日期范围 |

#### Payment Service 文件清单

```
cmd/payment-service/
└── main.go                              # 服务入口（最简 main：1 repo → 1 svc → 1 handler）

internal/payment/
├── config/
│   └── config.go                        # 环境变量配置（端口默认 50055）
├── handler/
│   ├── convert.go                       # proto ↔ 领域模型转换、错误映射
│   └── payment_handler.go              # PaymentService gRPC 实现（8 RPCs）
├── model/
│   └── model.go                         # Payment, PaymentDetail
├── repository/
│   ├── payment_repository.go            # PaymentRepository 接口 + pgx 实现（16 方法）
│   ├── helpers.go                       # pgtype 转换辅助（timestamptz + numeric）
│   └── sqlcgen/                         # sqlc 自动生成（勿手动修改）
│       ├── db.go
│       ├── models.go
│       └── payment.sql.go
└── service/
    ├── errors.go                        # ErrNotFound, ErrInvalidArgument, ErrAlreadyExists, ErrForeignKey
    ├── pgutil.go                        # isUniqueViolation + isForeignKeyViolation
    ├── pagination.go                    # clampPagination
    └── payment_service.go              # Payment 业务逻辑（含跨表聚合 + 日期范围校验）

proto/payment/v1/
└── payment.proto                        # 1 个 gRPC 服务定义

gen/proto/payment/v1/
├── payment.pb.go                        # protobuf 生成代码
└── payment_grpc.pb.go                   # gRPC 生成代码

sql/queries/payment/
└── payment.sql                          # Payment 表查询（含跨表查询 + 日期范围）
```

#### 新增可复用模式（相比前四个服务）

1. **分区表透明操作**：sqlc 查询对分区表无需特殊处理，PG 自动路由分区；复合 PK (payment_date, payment_id) 不影响按 payment_id 单字段查询
2. **日期范围查询**：半开区间 `[start, end)` 是分区裁剪的最佳实践，handler 层处理 proto Timestamp 输入（nil 检查 + AsTime 转换）
3. **不可变记录模式**：财务数据只允许 Create/Delete，不提供 Update，业务层面的设计约束
4. **numeric 金额复用**：从 film-service 复制 numericToString/stringToNumeric，验证了跨服务的类型转换模式可复用性
5. **最简 main 模式**：单 repo → 单 service → 单 handler，是所有核心服务中依赖最简单的

#### 五个核心服务完成汇总

| 服务 | 端口 | 表 | gRPC 服务数 | RPC 总数 | Repository 方法数 |
|------|------|-----|-----------|---------|-----------------|
| store-service | 50051 | store, staff | 2 | 13 | 20 |
| film-service | 50052 | film, actor, category, language, film_actor, film_category | 4 | 22 | 31 |
| customer-service | 50053 | customer, address, city, country | 4 | 15 | 20 |
| rental-service | 50054 | rental, inventory | 2 | 16 | 26 |
| payment-service | 50055 | payment | 1 | 8 | 16 |
| **合计** | | **15 张表** | **13** | **74** | **113** |

---

*状态：全部 5 个核心微服务实施完成 —— 准备进入 BFF 层 + 公共模块实施阶段*
