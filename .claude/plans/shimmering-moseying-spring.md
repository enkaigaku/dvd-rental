# Rental Service 实现计划

## Context

这是 DVD 租赁微服务项目的第四个核心服务。Rental Service 管理 `rental` 和 `inventory` 两张表，是整个系统的核心业务——DVD 的租借和归还。rental 表引用 customer、inventory、staff；inventory 表引用 film、store。所有类型均为标准 PostgreSQL 类型（integer, timestamptz），无特殊类型。

与前三个服务不同的是，rental-service 涉及**业务状态转换**（租出 → 归还）和**库存可用性检查**，而不仅仅是 CRUD。

## 关键设计决策

1. **两个 gRPC 服务**：RentalService（核心业务）+ InventoryService（库存管理），同一进程端口 50054
2. **GetRental 返回 RentalDetail**：通过额外的 SQL lookup 查询聚合 customer 姓名、film 标题、store_id。使用单独的轻量 SQL 查询（`GetCustomerName`、`GetFilmTitleByInventory`）而非跨服务 gRPC 调用——因为是同一个数据库
3. **ReturnRental 替代 UpdateRental**：rental 的主要变更是归还（设置 return_date），不提供通用 UpdateRental
4. **CreateRental 的业务校验**：
   - 验证 inventory 存在
   - 验证库存可用（无未归还的 rental 记录，即 return_date IS NULL）
   - 验证 customer_id / staff_id > 0（FK 约束兜底）
   - rental_date 由服务端设置为 now()
5. **库存可用性检查**：使用 SQL 查询 `SELECT NOT EXISTS (SELECT 1 FROM rental WHERE inventory_id = $1 AND return_date IS NULL)` 而非调用 PG 函数 `inventory_in_stock()`——保持 sqlc 兼容
6. **return_date 的 NULL 处理**：DB NULL → pgtype.Timestamptz{Valid: false} → Go time.Time 零值 → proto nil Timestamp（handler 层检查 IsZero）
7. **ListOverdueRentals**：查询 return_date IS NULL 的租赁记录，按 rental_date 升序
8. **DeleteRental / DeleteInventory 处理 FK 约束**：payment → rental RESTRICT，rental → inventory RESTRICT
9. **ListAvailableInventory**：按 film_id + store_id 查找可租借的库存项
10. **默认端口 50054**

## 实现步骤

### Step 1: Proto 定义 + SQL 查询 + 工具配置

**新增文件：**
- `proto/rental/v1/rental.proto` — 两个 gRPC 服务定义
- `sql/queries/rental/rental.sql` — Rental 查询
- `sql/queries/rental/inventory.sql` — Inventory 查询

**修改文件：**
- `sqlc.yaml` — 新增 rental sql block
- `Makefile` — 添加 `run-rental` target

**Proto 服务定义概要：**
```protobuf
service RentalService {
  rpc GetRental(GetRentalRequest) returns (RentalDetail);
  rpc ListRentals(ListRentalsRequest) returns (ListRentalsResponse);
  rpc ListRentalsByCustomer(ListRentalsByCustomerRequest) returns (ListRentalsResponse);
  rpc ListRentalsByInventory(ListRentalsByInventoryRequest) returns (ListRentalsResponse);
  rpc ListOverdueRentals(ListOverdueRentalsRequest) returns (ListRentalsResponse);
  rpc CreateRental(CreateRentalRequest) returns (Rental);
  rpc ReturnRental(ReturnRentalRequest) returns (Rental);
  rpc DeleteRental(DeleteRentalRequest) returns (google.protobuf.Empty);
}

service InventoryService {
  rpc GetInventory(GetInventoryRequest) returns (Inventory);
  rpc ListInventory(ListInventoryRequest) returns (ListInventoryResponse);
  rpc ListInventoryByFilm(ListInventoryByFilmRequest) returns (ListInventoryResponse);
  rpc ListInventoryByStore(ListInventoryByStoreRequest) returns (ListInventoryResponse);
  rpc CheckInventoryAvailability(CheckInventoryAvailabilityRequest) returns (CheckInventoryAvailabilityResponse);
  rpc ListAvailableInventory(ListAvailableInventoryRequest) returns (ListInventoryResponse);
  rpc CreateInventory(CreateInventoryRequest) returns (Inventory);
  rpc DeleteInventory(DeleteInventoryRequest) returns (google.protobuf.Empty);
}
```

**SQL 查询清单：**

| 文件 | 查询名 | 类型 | 说明 |
|------|--------|------|------|
| rental.sql | GetRental | :one | 基础租赁记录 |
| rental.sql | ListRentals | :many | 分页列表 |
| rental.sql | CountRentals | :one | 总数 |
| rental.sql | ListRentalsByCustomer | :many | 按客户过滤 |
| rental.sql | CountRentalsByCustomer | :one | 按客户计数 |
| rental.sql | ListRentalsByInventory | :many | 按库存项过滤 |
| rental.sql | CountRentalsByInventory | :one | 按库存项计数 |
| rental.sql | ListOverdueRentals | :many | 未归还记录 |
| rental.sql | CountOverdueRentals | :one | 未归还计数 |
| rental.sql | CreateRental | :one | 创建租赁 |
| rental.sql | ReturnRental | :one | 归还（设置 return_date） |
| rental.sql | DeleteRental | :exec | 删除 |
| rental.sql | GetCustomerName | :one | 查询客户姓名（供 RentalDetail 聚合） |
| rental.sql | GetFilmTitleByInventory | :one | 通过 inventory 查 film 标题和 store_id |
| rental.sql | IsInventoryAvailable | :one | 检查库存是否可租 |
| inventory.sql | GetInventory | :one | |
| inventory.sql | ListInventory | :many | |
| inventory.sql | CountInventory | :one | |
| inventory.sql | ListInventoryByFilm | :many | 按 film_id 过滤 |
| inventory.sql | CountInventoryByFilm | :one | |
| inventory.sql | ListInventoryByStore | :many | 按 store_id 过滤 |
| inventory.sql | CountInventoryByStore | :one | |
| inventory.sql | ListAvailableInventory | :many | 按 film_id+store_id 查可用库存 |
| inventory.sql | CountAvailableInventory | :one | |
| inventory.sql | CreateInventory | :one | |
| inventory.sql | DeleteInventory | :exec | |

**数据类型处理：**

| DB 类型 | Proto 类型 | Go 模型类型 | 转换说明 |
|---------|-----------|------------|---------|
| timestamptz NOT NULL | google.protobuf.Timestamp | time.Time | pgtype.Timestamptz → time.Time |
| timestamptz NULL (return_date) | google.protobuf.Timestamp | time.Time | NULL → zero Time → proto nil |
| integer | int32 | int32 | 直接映射 |

### Step 2: 运行代码生成

执行 `make generate`，验证 store/film/customer 生成代码未变化。

### Step 3: Repository 层

**新增文件：**
- `internal/rental/model/model.go` — Rental, RentalDetail, Inventory 领域模型
- `internal/rental/repository/rental_repository.go` — RentalRepository 接口 + 实现
- `internal/rental/repository/inventory_repository.go` — InventoryRepository 接口 + 实现
- `internal/rental/repository/helpers.go` — timestamptzToTime（含 NULL 处理）

**领域模型：**
```go
type Rental struct {
    RentalID    int32
    RentalDate  time.Time
    InventoryID int32
    CustomerID  int32
    ReturnDate  time.Time  // zero value = not returned
    StaffID     int32
    LastUpdate  time.Time
}

type RentalDetail struct {
    Rental
    CustomerName string  // first_name || ' ' || last_name
    FilmTitle    string  // film.title via inventory
    StoreID      int32   // inventory.store_id
}

type Inventory struct {
    InventoryID int32
    FilmID      int32
    StoreID     int32
    LastUpdate  time.Time
}
```

**Repository 接口方法：**
- RentalRepository（15 个）：GetRental, ListRentals, CountRentals, ListRentalsByCustomer, CountRentalsByCustomer, ListRentalsByInventory, CountRentalsByInventory, ListOverdueRentals, CountOverdueRentals, CreateRental, ReturnRental, DeleteRental, GetCustomerName, GetFilmTitleByInventory, IsInventoryAvailable
- InventoryRepository（11 个）：GetInventory, ListInventory, CountInventory, ListInventoryByFilm, CountInventoryByFilm, ListInventoryByStore, CountInventoryByStore, ListAvailableInventory, CountAvailableInventory, CreateInventory, DeleteInventory

### Step 4: Service 层

**新增文件：**
- `internal/rental/service/rental_service.go` — RentalService
- `internal/rental/service/inventory_service.go` — InventoryService
- `internal/rental/service/errors.go` — 哨兵错误
- `internal/rental/service/pgutil.go` — PG 错误检查
- `internal/rental/service/pagination.go` — clampPagination

**RentalService 业务规则：**
- GetRental：获取 rental 后聚合 customer name + film title + store_id → RentalDetail
- CreateRental：校验 inventory_id/customer_id/staff_id > 0，校验库存可用（IsInventoryAvailable），FK 违约 → ErrInvalidArgument
- ReturnRental：校验 rental 存在，校验 return_date 为 NULL（未归还），设置 return_date = now()
- DeleteRental：FK 违约 → ErrForeignKey（被 payment 引用）
- ListOverdueRentals：分页列表

**InventoryService 业务规则：**
- CheckInventoryAvailability：调用 repo.IsInventoryAvailable
- ListAvailableInventory：按 film_id + store_id 查可用库存
- CreateInventory：校验 film_id/store_id > 0，FK 违约 → ErrInvalidArgument
- DeleteInventory：FK 违约 → ErrForeignKey（被 rental 引用）

### Step 5: Handler 层

**新增文件：**
- `internal/rental/handler/rental_handler.go` — RentalService 8 个 RPC
- `internal/rental/handler/inventory_handler.go` — InventoryService 8 个 RPC
- `internal/rental/handler/convert.go` — toGRPCError + proto 转换（含 return_date nil 处理）

**return_date proto 转换关键逻辑：**
```go
func rentalToProto(r model.Rental) *rentalv1.Rental {
    proto := &rentalv1.Rental{
        RentalId:    r.RentalID,
        RentalDate:  timestamppb.New(r.RentalDate),
        InventoryId: r.InventoryID,
        CustomerId:  r.CustomerID,
        StaffId:     r.StaffID,
        LastUpdate:  timestamppb.New(r.LastUpdate),
    }
    if !r.ReturnDate.IsZero() {
        proto.ReturnDate = timestamppb.New(r.ReturnDate)
    }
    return proto
}
```

### Step 6: Main 入口 + 配置

**新增文件：**
- `internal/rental/config/config.go` — DATABASE_URL, GRPC_PORT（默认 "50054"）, LOG_LEVEL
- `cmd/rental-service/main.go` — DI（2 repo → 2 service → 2 handler）、gRPC server、health check、reflection、graceful shutdown

## 文件清单

### 新增文件（17 个）

| 文件 | 说明 |
|------|------|
| `proto/rental/v1/rental.proto` | 2 个 gRPC 服务定义 |
| `sql/queries/rental/rental.sql` | Rental 查询 + 跨表 lookup |
| `sql/queries/rental/inventory.sql` | Inventory CRUD + 可用性查询 |
| `internal/rental/model/model.go` | 领域模型 |
| `internal/rental/repository/rental_repository.go` | RentalRepository |
| `internal/rental/repository/inventory_repository.go` | InventoryRepository |
| `internal/rental/repository/helpers.go` | pgtype 转换辅助 |
| `internal/rental/service/rental_service.go` | Rental 业务逻辑 |
| `internal/rental/service/inventory_service.go` | Inventory 业务逻辑 |
| `internal/rental/service/errors.go` | 哨兵错误 |
| `internal/rental/service/pgutil.go` | PG 错误检查 |
| `internal/rental/service/pagination.go` | 分页辅助 |
| `internal/rental/handler/rental_handler.go` | RentalService gRPC |
| `internal/rental/handler/inventory_handler.go` | InventoryService gRPC |
| `internal/rental/handler/convert.go` | 转换 + 错误映射 |
| `internal/rental/config/config.go` | 配置加载 |
| `cmd/rental-service/main.go` | 服务入口 |

### 修改文件（2 个）

| 文件 | 变更 |
|------|------|
| `sqlc.yaml` | 新增 rental sql block |
| `Makefile` | 添加 `run-rental` target |

## 验证方式

1. `make generate` — 验证所有 4 组代码正确生成
2. `go build ./cmd/{store,film,customer}-service/` — 确认未被破坏
3. `go build ./cmd/rental-service/` — 编译通过
4. `go vet ./internal/rental/...` — 无告警
