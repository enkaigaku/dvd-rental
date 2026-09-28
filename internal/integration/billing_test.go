package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	cr "github.com/enkaigaku/dvd-rental/internal/customer/repository"
	cs "github.com/enkaigaku/dvd-rental/internal/customer/service"
	pr "github.com/enkaigaku/dvd-rental/internal/payment/repository"
	ps "github.com/enkaigaku/dvd-rental/internal/payment/service"
	rm "github.com/enkaigaku/dvd-rental/internal/rental/model"
	rr "github.com/enkaigaku/dvd-rental/internal/rental/repository"
	rs "github.com/enkaigaku/dvd-rental/internal/rental/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Tests use a disposable database initialized with migrations/*.sql.
// Only fixture-owned rows are removed; no existing rentals are modified.
type fixture struct {
	pool                  *pgxpool.Pool
	customer, film, staff int32
	inventory             []int32
	rentals               *rs.RentalService
	customers             *cs.CustomerService
	payments              *ps.PaymentService
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated disposable PostgreSQL database")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	f := &fixture{pool: pool}
	f.rentals = rs.NewRentalService(rr.NewRentalRepository(pool), rr.NewInventoryRepository(pool))
	f.customers = cs.NewCustomerService(cr.NewCustomerRepository(pool), nil, nil, nil)
	f.payments = ps.NewPaymentService(pr.NewPaymentRepository(pool))
	ctx := context.Background()
	if err := pool.QueryRow(ctx, `INSERT INTO customer(store_id, first_name, last_name, address_id) SELECT store_id, 'PR1', 'Review', address_id FROM customer LIMIT 1 RETURNING customer_id`).Scan(&f.customer); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO film(title, language_id, rental_duration, rental_rate) VALUES ('PR1 Review', 1, 3, 2.99) RETURNING film_id`).Scan(&f.film); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT staff_id FROM staff ORDER BY staff_id LIMIT 1`).Scan(&f.staff); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, query := range []string{`DELETE FROM payment WHERE customer_id=$1`, `DELETE FROM rental WHERE customer_id=$1`, `DELETE FROM customer WHERE customer_id=$1`} {
			if _, err := pool.Exec(context.Background(), query, f.customer); err != nil {
				t.Error(err)
			}
		}
		if _, err := pool.Exec(context.Background(), `DELETE FROM inventory WHERE film_id=$1`, f.film); err != nil {
			t.Error(err)
		}
		if _, err := pool.Exec(context.Background(), `DELETE FROM film WHERE film_id=$1`, f.film); err != nil {
			t.Error(err)
		}
	})
	return f
}
func (f *fixture) item(t *testing.T) int32 {
	t.Helper()
	var id int32
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO inventory(film_id,store_id) VALUES($1,1) RETURNING inventory_id`, f.film).Scan(&id); err != nil {
		t.Fatal(err)
	}
	f.inventory = append(f.inventory, id)
	return id
}
func (f *fixture) rental(t *testing.T, age time.Duration) int32 {
	t.Helper()
	var id int32
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO rental(rental_date,inventory_id,customer_id,staff_id) VALUES($1,$2,$3,$4) RETURNING rental_id`, time.Now().Add(-age), f.item(t), f.customer, f.staff).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
func (f *fixture) checkBalance(t *testing.T, charges, payments, balance string) {
	t.Helper()
	ctx := context.Background()
	b, err := f.payments.GetCustomerBalance(ctx, f.customer)
	if err != nil {
		t.Fatal(err)
	}
	if b.TotalCharges != charges || b.TotalPayments != payments || b.Balance != balance {
		t.Fatalf("balance=%+v; want %s - %s = %s", b, charges, payments, balance)
	}
	standing, err := f.customers.GetCustomerStanding(ctx, f.customer)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := f.customers.GetCustomerSummary(ctx, f.customer)
	if err != nil {
		t.Fatal(err)
	}
	if standing.OutstandingBalance != balance || summary.OutstandingBalance != balance {
		t.Fatalf("balances disagree: %+v %+v", standing, summary)
	}
}
func TestReturnBilling(t *testing.T) {
	for _, tc := range []struct {
		name         string
		age          time.Duration
		days         int32
		charges, fee string
	}{
		{"on time", 48 * time.Hour, 0, "2.99", "0.00"},
		{"partial overdue day", 73 * time.Hour, 1, "3.99", "1.00"},
		{"multiple overdue days", 121 * time.Hour, 3, "5.99", "3.00"},
		{"long overdue", (1004*24 + 1) * time.Hour, 1002, "1004.99", "1002.00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			id := f.rental(t, tc.age)
			ctx := context.Background()
			f.checkBalance(t, tc.charges, "0.00", tc.charges)
			result, err := f.rentals.ReturnRental(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if result.DaysOverdue != tc.days || result.LateFee != tc.fee || result.Rental.ReturnDate.IsZero() {
				t.Fatalf("unexpected result: %+v", result)
			}
			f.checkBalance(t, tc.charges, tc.fee, "2.99")
			if _, err := f.rentals.ReturnRental(ctx, id); !errors.Is(err, rs.ErrNotFound) {
				t.Fatalf("repeat return error=%v", err)
			}
			var count int
			if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM payment WHERE rental_id=$1`, id).Scan(&count); err != nil {
				t.Fatal(err)
			}
			want := 0
			if tc.days > 0 {
				want = 1
			}
			if count != want {
				t.Fatalf("payment count=%d, want %d", count, want)
			}
			_, err = f.payments.CreatePayment(ctx, pr.CreatePaymentParams{CustomerID: f.customer, StaffID: f.staff, RentalID: id, Amount: "2.99"})
			if err != nil {
				t.Fatal(err)
			}
			f.checkBalance(t, tc.charges, tc.charges, "0.00")
			standing, err := f.customers.GetCustomerStanding(ctx, f.customer)
			if err != nil {
				t.Fatal(err)
			}
			if !standing.InGoodStanding {
				t.Fatalf("settled customer not in good standing: %+v", standing)
			}
		})
	}
}

// Fail a real transaction after its return update to verify rollback rather
// than relying on a mock's interpretation of transaction semantics.
type failingRepo struct {
	rr.RentalRepository
	failTerms bool
}

func (r failingRepo) WithinTx(ctx context.Context, fn func(rr.RentalRepository) error) error {
	return r.RentalRepository.WithinTx(ctx, func(tx rr.RentalRepository) error { return fn(failingRepo{tx, r.failTerms}) })
}

var errInjected = errors.New("injected persistence failure")

func (r failingRepo) GetFilmRentalTermsByInventory(ctx context.Context, id int32) (rm.FilmRentalTerms, error) {
	if r.failTerms {
		return rm.FilmRentalTerms{}, errInjected
	}
	return r.RentalRepository.GetFilmRentalTermsByInventory(ctx, id)
}
func (r failingRepo) CreateLateFeePayment(context.Context, int32, int32, int32, string) error {
	return errInjected
}
func TestReturnRollsBackOnFailure(t *testing.T) {
	for _, terms := range []bool{false, true} {
		t.Run(fmt.Sprintf("terms=%v", terms), func(t *testing.T) {
			f := newFixture(t)
			id := f.rental(t, 73*time.Hour)
			ctx := context.Background()
			svc := rs.NewRentalService(failingRepo{rr.NewRentalRepository(f.pool), terms}, nil)
			if _, err := svc.ReturnRental(ctx, id); !errors.Is(err, errInjected) {
				t.Fatalf("error=%v", err)
			}
			rental, err := rr.NewRentalRepository(f.pool).GetRental(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if !rental.ReturnDate.IsZero() {
				t.Fatal("failed return was committed")
			}
			result, err := f.rentals.ReturnRental(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if result.LateFee != "1.00" {
				t.Fatalf("retry=%+v", result)
			}
		})
	}
}
func TestOverdueAndStanding(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	current := f.rental(t, time.Hour)
	late := f.rental(t, 73*time.Hour)
	repo := rr.NewRentalRepository(f.pool)
	// Existing seed rentals are older, so inspect all pages rather than assuming
	// fixture rows are the first results.
	count, err := repo.CountOverdueRentals(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := repo.ListOverdueRentals(ctx, int32(count+1), 0)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(rows)) != count {
		t.Fatalf("list/count mismatch: %d/%d", len(rows), count)
	}
	found := false
	for _, r := range rows {
		if r.RentalID == current {
			t.Fatal("not-yet-due rental listed as overdue")
		}
		if r.RentalID == late {
			found = true
		}
	}
	if !found {
		t.Fatal("overdue rental missing")
	}
	standing, err := f.customers.GetCustomerStanding(ctx, f.customer)
	if err != nil {
		t.Fatal(err)
	}
	if standing.OverdueRentals != 1 || standing.InGoodStanding {
		t.Fatalf("standing=%+v", standing)
	}
	// The seven-day grace applies to checkout blocking, not overdue reporting.
	_, err = f.rentals.CreateRental(ctx, rr.CreateRentalParams{CustomerID: f.customer, InventoryID: f.item(t), StaffID: f.staff})
	if err != nil {
		t.Fatalf("within grace: %v", err)
	}
	f.rental(t, 241*time.Hour)
	_, err = f.rentals.CreateRental(ctx, rr.CreateRentalParams{CustomerID: f.customer, InventoryID: f.item(t), StaffID: f.staff})
	if !errors.Is(err, rs.ErrRentalPolicy) {
		t.Fatalf("past grace: %v", err)
	}
}
func TestMissingAndInactiveCustomers(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	missing := int32(2147483647)
	f.checkBalance(t, "0.00", "0.00", "0.00")
	if _, err := f.customers.GetCustomerStanding(ctx, missing); !errors.Is(err, cs.ErrNotFound) {
		t.Fatalf("standing error=%v", err)
	}
	if _, err := f.customers.GetCustomerSummary(ctx, missing); !errors.Is(err, cs.ErrNotFound) {
		t.Fatalf("summary error=%v", err)
	}
	if _, err := f.payments.GetCustomerBalance(ctx, missing); !errors.Is(err, ps.ErrNotFound) {
		t.Fatalf("balance error=%v", err)
	}
	if _, err := f.pool.Exec(ctx, `UPDATE customer SET activebool=false WHERE customer_id=$1`, f.customer); err != nil {
		t.Fatal(err)
	}
	standing, err := f.customers.GetCustomerStanding(ctx, f.customer)
	if err != nil {
		t.Fatal(err)
	}
	if standing.InGoodStanding {
		t.Fatal("inactive account in good standing")
	}
	_, err = f.rentals.CreateRental(ctx, rr.CreateRentalParams{CustomerID: f.customer, InventoryID: f.item(t), StaffID: f.staff})
	if !errors.Is(err, rs.ErrRentalPolicy) {
		t.Fatalf("inactive checkout: %v", err)
	}
}
func TestConcurrentCheckouts(t *testing.T) {
	for _, sameItem := range []bool{false, true} {
		t.Run(fmt.Sprintf("same_item=%v", sameItem), func(t *testing.T) {
			f := newFixture(t)
			g := newFixture(t)
			ctx := context.Background()
			if !sameItem {
				for i := 0; i < 4; i++ {
					f.rental(t, time.Hour)
				}
			}
			item1, item2 := f.item(t), f.item(t)
			customer2 := f.customer
			if sameItem {
				item2 = item1
				customer2 = g.customer
			}
			params := []rr.CreateRentalParams{{CustomerID: f.customer, InventoryID: item1, StaffID: f.staff}, {CustomerID: customer2, InventoryID: item2, StaffID: f.staff}}
			start := make(chan struct{})
			results := make(chan error, 2)
			var wg sync.WaitGroup
			for _, p := range params {
				wg.Add(1)
				go func(p rr.CreateRentalParams) {
					defer wg.Done()
					<-start
					_, err := f.rentals.CreateRental(ctx, p)
					results <- err
				}(p)
			}
			close(start)
			wg.Wait()
			close(results)
			success := 0
			for err := range results {
				if err == nil {
					success++
				} else if !errors.Is(err, rs.ErrRentalPolicy) && !errors.Is(err, rs.ErrInvalidArgument) {
					t.Fatal(err)
				}
			}
			if success != 1 {
				t.Fatalf("successful checkouts=%d", success)
			}
		})
	}
}
func TestConcurrentReturns(t *testing.T) {
	f := newFixture(t)
	id := f.rental(t, 73*time.Hour)
	ctx := context.Background()
	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { <-start; _, err := f.rentals.ReturnRental(ctx, id); results <- err }()
	}
	close(start)
	successes := 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			successes++
		} else if !errors.Is(err, rs.ErrNotFound) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful returns=%d", successes)
	}
	f.checkBalance(t, "3.99", "1.00", "2.99")
}
