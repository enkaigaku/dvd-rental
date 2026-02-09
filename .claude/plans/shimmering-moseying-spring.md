# Customer Service 实现计划

## Context

这是 DVD 租赁微服务项目的第三个核心服务。Customer Service 管理 `customer`、`address`、`city`、`country` 共 4 张表。customer → address → city → country 构成层级关系链。所有表均为标准 PostgreSQL 类型（text, integer, boolean, date, timestamptz），无特殊类型（ENUM/DOMAIN/ARRAY），比 film-service 简单。实现模式复用已建立的三层架构。

## 关键设计决策

1. **四个 gRPC 服务在一个 proto 文件中**：CustomerService（核心 CRUD）、AddressService（CRUD）、CityService（只读）、CountryService（只读），同一进程端口 50053
2. **GetCustomer 返回 CustomerDetail**：聚合 customer + address + city name + country name，减少 BFF 调用轮次；ListCustomers 返回基础 Customer（含 address_id/store_id）
3. **Address 需要完整 CRUD**：客户创建/更新需要管理地址；City/Country 为只读参考数据
4. **ListCustomersByStore**：按 store_id 过滤客户列表，支持多门店场景
5. **active 状态**：`activebool` (boolean) 为源，`active` (integer) 为冗余字段，proto 暴露为 `bool active`，Create/Update 同时写入两个字段保持一致
6. **password_hash 不暴露**：存在于 DB 模型但不出现在 proto 响应中；认证功能留给后续独立服务
7. **email 可选但唯一**：proto 中为 string（空串表示无），service 层校验非空 email 的唯一性（unique violation → AlreadyExists）
8. **DeleteCustomer 处理 FK 约束**：rental/payment 表 FK RESTRICT，删除失败时映射为 FailedPrecondition
9. **DeleteAddress 处理 FK 约束**：customer 表 FK RESTRICT，删除失败时映射为 FailedPrecondition
10. **create_date 只读**：由 DB DEFAULT CURRENT_DATE 生成，不接受客户端输入

## 实现步骤

### Step 1: Proto 定义 + SQL 查询 + 工具配置

**新增文件：**
- `proto/customer/v1/customer.proto` — 四个 gRPC 服务定义
- `sql/queries/customer/customer.sql` — Customer 查询（8 个）
- `sql/queries/customer/address.sql` — Address 查询（5 个）
- `sql/queries/customer/city.sql` — City 查询（3 个）
- `sql/queries/customer/country.sql` — Country 查询（3 个）

**修改文件：**
- `sqlc.yaml` — 新增 customer sql block
- `Makefile` — 添加 `run-customer` target

**Proto 服务定义概要：**
```protobuf
service CustomerService {
  rpc GetCustomer(GetCustomerRequest) returns (CustomerDetail);
  rpc ListCustomers(ListCustomersRequest) returns (ListCustomersResponse);
  rpc ListCustomersByStore(ListCustomersByStoreRequest) returns (ListCustomersResponse);
  rpc CreateCustomer(CreateCustomerRequest) returns (Customer);
  rpc UpdateCustomer(UpdateCustomerRequest) returns (Customer);
  rpc DeleteCustomer(DeleteCustomerRequest) returns (google.protobuf.Empty);
}

service AddressService {
  rpc GetAddress(GetAddressRequest) returns (Address);
  rpc ListAddresses(ListAddressesRequest) returns (ListAddressesResponse);
  rpc CreateAddress(CreateAddressRequest) returns (Address);
  rpc UpdateAddress(UpdateAddressRequest) returns (Address);
  rpc DeleteAddress(DeleteAddressRequest) returns (google.protobuf.Empty);
}

service CityService {
  rpc GetCity(GetCityRequest) returns (City);
  rpc ListCities(ListCitiesRequest) returns (ListCitiesResponse);
}

service CountryService {
  rpc GetCountry(GetCountryRequest) returns (Country);
  rpc ListCountries(ListCountriesRequest) returns (ListCountriesResponse);
}
```

**SQL 查询清单：**

| 文件 | 查询名 | 类型 |
|------|--------|------|
| customer.sql | GetCustomer | :one |
| customer.sql | ListCustomers | :many |
| customer.sql | CountCustomers | :one |
| customer.sql | ListCustomersByStore | :many |
| customer.sql | CountCustomersByStore | :one |
| customer.sql | CreateCustomer | :one |
| customer.sql | UpdateCustomer | :one |
| customer.sql | DeleteCustomer | :exec |
| address.sql | GetAddress | :one |
| address.sql | ListAddresses | :many |
| address.sql | CountAddresses | :one |
| address.sql | CreateAddress | :one |
| address.sql | UpdateAddress | :one |
| address.sql | DeleteAddress | :exec (注意：不需要，address 没有独立删除需求 → 保留以防万一) |
| city.sql | GetCity | :one |
| city.sql | ListCities | :many |
| city.sql | CountCities | :one |
| country.sql | GetCountry | :one |
| country.sql | ListCountries | :many |
| country.sql | CountCountries | :one |

**数据类型处理（比 film 简单）：**

| DB 类型 | Proto 类型 | Go 模型类型 | 转换说明 |
|---------|-----------|------------|---------|
| text NOT NULL | string | string | 直接映射 |
| text NULL | string | string | pgtype.Text ↔ string，NULL→"" |
| boolean | bool | bool | 直接映射 |
| date | google.protobuf.Timestamp | time.Time | pgtype.Date ↔ time.Time |
| timestamptz | google.protobuf.Timestamp | time.Time | pgtype.Timestamptz ↔ time.Time |
| integer NULL (active) | — | int32 | 不暴露在 proto，内部同步 |

### Step 2: 运行代码生成

执行 `make generate`（buf generate + sqlc generate），生成：
- `gen/proto/customer/v1/customer.pb.go` + `customer_grpc.pb.go`
- `internal/customer/repository/sqlcgen/*.go`

验证 store-service 和 film-service 的生成代码未变化。

### Step 3: Repository 层

**新增文件：**
- `internal/customer/model/model.go` — Customer, CustomerDetail, Address, City, Country 领域模型
- `internal/customer/repository/customer_repository.go` — CustomerRepository 接口 + 实现
- `internal/customer/repository/address_repository.go` — AddressRepository 接口 + 实现
- `internal/customer/repository/city_repository.go` — CityRepository 接口 + 实现（只读）
- `internal/customer/repository/country_repository.go` — CountryRepository 接口 + 实现（只读）
- `internal/customer/repository/helpers.go` — pgtype 转换辅助函数

**领域模型：**
```go
type Customer struct {
    CustomerID int32
    StoreID    int32
    FirstName  string
    LastName   string
    Email      string
    AddressID  int32
    Active     bool
    CreateDate time.Time
    LastUpdate time.Time
}

type CustomerDetail struct {
    Customer
    Address     string  // address.address
    Address2    string  // address.address2
    District    string
    CityName    string
    CountryName string
    PostalCode  string
    Phone       string
}

type Address struct {
    AddressID  int32
    Address    string
    Address2   string
    District   string
    CityID     int32
    PostalCode string
    Phone      string
    LastUpdate time.Time
}

type City struct {
    CityID    int32
    City      string
    CountryID int32
    LastUpdate time.Time
}

type Country struct {
    CountryID  int32
    Country    string
    LastUpdate time.Time
}
```

**Repository 接口方法：**
- CustomerRepository（8 个）：GetCustomer, ListCustomers, CountCustomers, ListCustomersByStore, CountCustomersByStore, CreateCustomer, UpdateCustomer, DeleteCustomer
- AddressRepository（5 个）：GetAddress, ListAddresses, CountAddresses, CreateAddress, UpdateAddress
- CityRepository（3 个）：GetCity, ListCities, CountCities
- CountryRepository（3 个）：GetCountry, ListCountries, CountCountries

**helpers.go 辅助函数：**
- textToString / stringToText（pgtype.Text ↔ string）
- dateToTime（pgtype.Date → time.Time）
- timestamptzToTime（pgtype.Timestamptz → time.Time）
- boolToActive（bool → int32，true=1, false=0，用于同步 active 字段）

### Step 4: Service 层

**新增文件：**
- `internal/customer/service/customer_service.go` — CustomerService（持有 4 个 repo）
- `internal/customer/service/address_service.go` — AddressService
- `internal/customer/service/city_service.go` — CityService
- `internal/customer/service/country_service.go` — CountryService
- `internal/customer/service/errors.go` — ErrNotFound, ErrInvalidArgument, ErrAlreadyExists, ErrForeignKey
- `internal/customer/service/pgutil.go` — isUniqueViolation + isForeignKeyViolation
- `internal/customer/service/pagination.go` — clampPagination

**CustomerService 业务规则：**
- GetCustomer：获取 customer 后聚合 address + city + country → CustomerDetail
- CreateCustomer：校验 first_name/last_name 非空、store_id > 0、address_id 存在；email 非空时检查格式合理性；active 字段同步（activebool=true → active=1）
- UpdateCustomer：同上校验 + customer 存在性检查
- DeleteCustomer：FK 违约 → ErrForeignKey（"customer is referenced by rental/payment records"）
- ListCustomersByStore：校验 store_id > 0

**AddressService 业务规则：**
- CreateAddress：校验 address/district/phone 非空、city_id 存在
- UpdateAddress：同上 + address 存在性检查
- FK 违约（city_id 不存在）→ ErrInvalidArgument

### Step 5: Handler 层

**新增文件：**
- `internal/customer/handler/customer_handler.go` — CustomerService 6 个 RPC
- `internal/customer/handler/address_handler.go` — AddressService 4 个 RPC（不含 Delete）
- `internal/customer/handler/city_handler.go` — CityService 2 个 RPC
- `internal/customer/handler/country_handler.go` — CountryService 2 个 RPC
- `internal/customer/handler/convert.go` — toGRPCError + proto 转换函数

### Step 6: Main 入口 + 配置

**新增文件：**
- `internal/customer/config/config.go` — DATABASE_URL（必填）, GRPC_PORT（默认 "50053"）, LOG_LEVEL
- `cmd/customer-service/main.go` — DI（4 repo → 4 service → 4 handler）、gRPC server、health check、reflection、graceful shutdown

## 文件清单

### 新增文件（22 个）

| 文件 | 说明 |
|------|------|
| `proto/customer/v1/customer.proto` | 4 个 gRPC 服务定义 |
| `sql/queries/customer/customer.sql` | Customer CRUD + 按 store 过滤 |
| `sql/queries/customer/address.sql` | Address CRUD |
| `sql/queries/customer/city.sql` | City 只读查询 |
| `sql/queries/customer/country.sql` | Country 只读查询 |
| `internal/customer/model/model.go` | 领域模型 |
| `internal/customer/repository/customer_repository.go` | CustomerRepository |
| `internal/customer/repository/address_repository.go` | AddressRepository |
| `internal/customer/repository/city_repository.go` | CityRepository |
| `internal/customer/repository/country_repository.go` | CountryRepository |
| `internal/customer/repository/helpers.go` | pgtype 转换辅助 |
| `internal/customer/service/customer_service.go` | Customer 业务逻辑 |
| `internal/customer/service/address_service.go` | Address 业务逻辑 |
| `internal/customer/service/city_service.go` | City 业务逻辑 |
| `internal/customer/service/country_service.go` | Country 业务逻辑 |
| `internal/customer/service/errors.go` | 哨兵错误 |
| `internal/customer/service/pgutil.go` | PG 错误检查 |
| `internal/customer/service/pagination.go` | 分页辅助 |
| `internal/customer/handler/customer_handler.go` | CustomerService gRPC |
| `internal/customer/handler/address_handler.go` | AddressService gRPC |
| `internal/customer/handler/city_handler.go` | CityService gRPC |
| `internal/customer/handler/country_handler.go` | CountryService gRPC |
| `internal/customer/handler/convert.go` | 转换 + 错误映射 |
| `internal/customer/config/config.go` | 配置加载 |
| `cmd/customer-service/main.go` | 服务入口 |

### 修改文件（2 个）

| 文件 | 变更 |
|------|------|
| `sqlc.yaml` | 新增 customer sql block |
| `Makefile` | 添加 `run-customer` target |

## 验证方式

1. `make sqlc-gen` — 验证 store + film + customer 三组代码均正确生成
2. `make proto-gen` — 生成 customer proto 代码
3. `go build ./cmd/store-service/` — 确认未被破坏
4. `go build ./cmd/film-service/` — 确认未被破坏
5. `go build ./cmd/customer-service/` — 编译通过
6. `make infra-up` + `go run ./cmd/customer-service/` — 启动服务
7. 用 grpcurl 测试 Customer/Address/City/Country CRUD
