# Copilot Instructions — DVD Rental Microservices

Go 1.25 microservices system (module `github.com/enkaigaku/dvd-rental`) built on the Pagila/Sakila sample database. Five gRPC core services sit behind two REST BFFs, backed by PostgreSQL 17 and Redis 7.

## Architecture

| Layer | Services | Transport | Ports |
|-------|----------|-----------|-------|
| BFF | `customer-bff`, `admin-bff` | REST/JSON (`net/http` ServeMux) | 8080, 8081 |
| Core | `store`, `film`, `customer`, `rental`, `payment` | gRPC | 50051–50055 |
| Data | PostgreSQL (shared DB), Redis (refresh tokens) | — | 5432, 6379 |

- Clients only talk to the BFFs. BFFs call core services over gRPC; core services never call each other. They share one database and each reads the tables it needs through its own sqlc package.
- `customer-bff` is the public/customer API. `admin-bff` is the staff console API. Both issue JWTs (15 min access token plus 7 day refresh token, with the refresh token tracked in Redis).

## Directory layout

```
cmd/<service>/main.go        # Wiring only: config → pgxpool → repos → services → handlers → server
internal/<domain>/           # store, film, customer, rental, payment
  config/                    # envconfig-based Config + Load()
  model/                     # Plain domain structs (no proto/sqlc types)
  repository/                # Interfaces + impls wrapping sqlc Queries; sqlc rows → model
  service/                   # Business logic, validation, pagination, sentinel errors
  handler/                   # gRPC server impls; model ↔ proto conversion; error → status code
internal/bff/{customer,admin}/
  config/ handler/ router/   # HTTP handlers call gRPC clients; router wires routes + middleware
pkg/auth/                    # JWT manager, bcrypt, Redis refresh-token store
pkg/middleware/              # Auth, CORS, logging, recovery, WriteJSON/WriteJSONError
pkg/grpcutil/                # gRPC client Dial helper
proto/<domain>/v1/*.proto    # Source of truth for gRPC APIs
sql/queries/<domain>/*.sql   # Source of truth for sqlc queries
migrations/*.sql             # Schema + seed data (also the sqlc schema input)
gen/                         # GENERATED, gitignored: gen/proto/... and gen/sqlc/<domain>
```

## Code generation

`gen/` is **not committed**. Run `make generate` (`buf generate` + `sqlc generate`) after cloning and after any change to `proto/`, `sql/queries/`, or `migrations/`.

- Never hand-edit files under `gen/`. Change the `.proto` or `.sql` source and regenerate.
- Proto packages are `<domain>.v1` with `go_package = ".../gen/proto/<domain>/v1;<domain>v1"`. Import them as `rentalv1`, `filmv1`, and so on.
- sqlc packages are named `<domain>sqlc` (for example `rentalsqlc`). They use `pgx/v5`, so nullable columns are `pgtype.*` types. Convert them in the repository layer and don't leak them upward.
- sqlc query annotations follow the `-- name: GetRental :one` / `:many` / `:exec` convention. Pair every paginated `List*` query (`LIMIT $n OFFSET $m`) with a `Count*` query.
- `buf lint` uses the STANDARD rules and breaking checks use FILE. Keep proto changes backward compatible.

## Layering rules for core services

Keep the flow **handler → service → repository**, with each layer depending only on the one below it.

**Repository**
- Define a `XxxRepository` interface and an unexported struct holding `q *<domain>sqlc.Queries`. The constructor is `NewXxxRepository(pool *pgxpool.Pool) XxxRepository`.
- Map `pgx.ErrNoRows` to the package's `repository.ErrNotFound`. Wrap other errors with context: `fmt.Errorf("get rental: %w", err)`.
- Convert sqlc rows to `model` structs through `toXxxModel` / `toXxxModels` helpers. A NULL timestamp becomes the zero `time.Time`.

**Service**
- Use concrete structs (`*RentalService`) whose constructors take repository interfaces.
- Validate inputs first. IDs must be positive: `fmt.Errorf("rental_id must be positive: %w", ErrInvalidArgument)`.
- Use the sentinel errors in `service/errors.go`: `ErrNotFound`, `ErrInvalidArgument`, `ErrAlreadyExists`, `ErrForeignKey`. Translate `repository.ErrNotFound` into `service.ErrNotFound`. Map Postgres errors with `isUniqueViolation` / `isForeignKeyViolation` (`pgutil.go`).
- Paginate with `clampPagination(pageSize, page)` (default 20, max 100, page is 1-based) and `offset := (page - 1) * pageSize`. List methods return `([]model.X, total int64, error)`.

**Handler (gRPC)**
- Embed `<domain>v1.UnimplementedXxxServiceServer`. Read request fields through the generated `req.GetXxx()` getters.
- Send every error through `toGRPCError` in `handler/convert.go`: NotFound → `NotFound`, InvalidArgument → `InvalidArgument`, AlreadyExists → `AlreadyExists`, ForeignKey → `FailedPrecondition`, anything else → `Internal` with the generic message `"internal error"`. Never expose raw DB errors.
- Convert with `xxxToProto` helpers using `timestamppb.New`. Leave optional timestamps nil when the model value is zero (for example an unreturned `ReturnDate`).

**main.go**
- Keep the existing pattern: `run() error`, `pgxpool` plus `Ping`, gRPC health server with per-service `SetServingStatus`, `reflection.Register`, and graceful shutdown on SIGINT/SIGTERM.

## BFF conventions

- Use Go 1.22+ method-and-pattern routes: `mux.HandleFunc("GET /api/v1/films/{id}", h.GetFilm)`, reading path params with `r.PathValue("id")` through `parseID`. Register more specific routes before general ones.
- Wrap protected routes with `authMw.Require(http.HandlerFunc(...))`. Read identity with `middleware.GetClaims(r.Context())` and nil-check the result.
- **Customer BFF ownership:** always take the customer ID from `claims.UserID`, never from the request body or query. When fetching a single resource by ID (rental, payment, and so on), check that it belongs to the caller before returning or mutating it. See `internal/bff/customer/handler/rental.go`.
- Give each downstream gRPC call its own timeout: `ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second); defer cancel()`.
- Map gRPC errors to HTTP with `grpcToHTTPError`. Write responses only with `middleware.WriteJSON` and `middleware.WriteJSONError(w, status, "CODE", msg)`. The error body shape is `{"error":{"code","message"}}`.
- Decode request bodies with `readJSON` (it calls `DisallowUnknownFields`). Parse query paging with `parsePagination`.
- Define JSON DTOs as unexported structs in the handler file with `snake_case` json tags. List responses include `total_count`, `page`, and `page_size`. Format timestamps as RFC3339 via `timestampToString`.
- Keep the middleware chain order Recovery → Logging → CORS → mux.

## Configuration

- Config structs use `github.com/kelseyhightower/envconfig` tags (`envconfig:"DATABASE_URL" required:"true"`, `default:"50054"`) and expose a `Load() (*Config, error)` function.
- Core services read `DATABASE_URL` (required), `GRPC_PORT`, and `LOG_LEVEL`. BFFs read `JWT_SECRET` (required), `REDIS_URL` (required), `HTTP_PORT`, `JWT_*_DURATION`, and `GRPC_<SERVICE>_ADDR`.
- Never hard-code secrets or commit `.env` files.

## Style

- Idiomatic Go, formatted with `gofmt` and `goimports`. Imports come in three groups: stdlib, third-party, then `github.com/enkaigaku/dvd-rental/...`.
- Exported identifiers get doc comments starting with the identifier name. Each package has a package comment where one already exists.
- Wrap errors with `%w` and short lowercase context. Don't log and return the same error.
- Accept `context.Context` as the first parameter of every repository, service, and handler method.
- Use `int32` for entity IDs and page values to match the proto and sqlc types.
- Avoid new dependencies. The stack is intentionally small: stdlib `net/http`, grpc, pgx, go-redis, golang-jwt, envconfig, and x/crypto.

## Adding a feature (checklist)

1. Schema change: add a new numbered file to `migrations/` rather than editing applied ones.
2. Add or modify queries in `sql/queries/<domain>/`, then update `proto/<domain>/v1/` if the API changes.
3. Run `make generate`.
4. Implement the change through repository → service → handler, then expose it via the BFF handler and router.
5. Update the endpoint tables in all three READMEs (`README.md`, `README_zh.md`, `README_ja.md`) when the public API changes.

## Commands

```bash
make infra-up      # Postgres + Redis in Docker (DB auto-seeded from migrations/)
make generate      # buf + sqlc codegen into gen/
make build-all     # build all 7 binaries
make test          # go test -v -race ./...
make lint          # golangci-lint run ./...
make fmt           # go fmt + goimports
make up / down     # run the whole stack in Docker
make run-<svc>     # run one service locally (store, film, customer, rental, payment, customer-bff, admin-bff)
```

## Testing

The repo has no tests yet. When adding them:
- Write table-driven tests with the stdlib `testing` package, placed next to the code (`*_test.go`).
- Unit-test services with hand-written fakes of the repository interfaces (no mocking framework).
- Test BFF handlers with `httptest` and fake gRPC client interfaces.
- Tests must pass under `-race`.
