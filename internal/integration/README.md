These tests exercise PostgreSQL repositories, service transactions, real gRPC calls,
and both HTTP BFF routers. Use a disposable PostgreSQL database initialized with
all files in `migrations/`, including the Pagila seed data and migration 006.
Tests create their own customers, films, inventory, rentals, and payments and
remove those fixture rows afterward. Run this suite without other writers in the
database, because revenue assertions cover a time range across all stores.

From the repository root:

```sh
make generate
TEST_DATABASE_URL='postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=disable' go test -race ./...
go vet ./...
```

Without `TEST_DATABASE_URL`, database tests skip; the decimal unit tests still run.

The billing tests preserve the PR's automatic late-fee payment behavior. Charges
include base rental rates and $1 per started 24-hour day overdue. Returning a
rental freezes its accrued late days and records the matching payment atomically.
Overdue reporting starts at the due time; checkout blocking starts strictly more
than seven days later. Standing also reports inactive accounts and unpaid balances.
