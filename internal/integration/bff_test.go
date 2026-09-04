package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	cv1 "github.com/enkaigaku/dvd-rental/gen/proto/customer/v1"
	pv1 "github.com/enkaigaku/dvd-rental/gen/proto/payment/v1"
	rv1 "github.com/enkaigaku/dvd-rental/gen/proto/rental/v1"
	ah "github.com/enkaigaku/dvd-rental/internal/bff/admin/handler"
	ar "github.com/enkaigaku/dvd-rental/internal/bff/admin/router"
	ch "github.com/enkaigaku/dvd-rental/internal/bff/customer/handler"
	cr "github.com/enkaigaku/dvd-rental/internal/bff/customer/router"
	customerhandler "github.com/enkaigaku/dvd-rental/internal/customer/handler"
	paymenthandler "github.com/enkaigaku/dvd-rental/internal/payment/handler"
	rentalhandler "github.com/enkaigaku/dvd-rental/internal/rental/handler"
	"github.com/enkaigaku/dvd-rental/pkg/auth"
	"github.com/enkaigaku/dvd-rental/pkg/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type bffs struct {
	admin, customer           http.Handler
	staffToken, customerToken string
	payment                   pv1.PaymentServiceClient
}

func (f *fixture) bffs(t *testing.T) bffs {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	cv1.RegisterCustomerServiceServer(server, customerhandler.NewCustomerHandler(f.customers))
	pv1.RegisterPaymentServiceServer(server, paymenthandler.NewPaymentHandler(f.payments))
	rv1.RegisterRentalServiceServer(server, rentalhandler.NewRentalHandler(f.rentals))
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("gRPC server: %v", err)
		}
	}()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	})
	customers, payments, rentals := cv1.NewCustomerServiceClient(conn), pv1.NewPaymentServiceClient(conn), rv1.NewRentalServiceClient(conn)
	jwt, err := auth.NewJWTManager("test-secret", time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	mw := middleware.NewAuthMiddleware(jwt)
	staffToken, err := jwt.GenerateAccessToken(f.staff, auth.RoleStaff, "staff")
	if err != nil {
		t.Fatal(err)
	}
	customerToken, err := jwt.GenerateAccessToken(f.customer, auth.RoleCustomer, "customer")
	if err != nil {
		t.Fatal(err)
	}
	return bffs{
		admin:      ar.NewRouter(nil, nil, nil, ah.NewCustomerHandler(customers), nil, nil, ah.NewRentalHandler(rentals), ah.NewPaymentHandler(payments), mw),
		customer:   cr.NewRouter(nil, nil, ch.NewRentalHandler(rentals), ch.NewPaymentHandler(payments), ch.NewProfileHandler(customers), mw),
		staffToken: staffToken, customerToken: customerToken, payment: payments,
	}
}
func request(t *testing.T, h http.Handler, method, path, token string, want int) map[string]any {
	t.Helper()
	r := httptest.NewRequest(method, path, nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != want {
		t.Fatalf("%s %s: status %d, want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body
}
func TestBFFReadEndpoints(t *testing.T) {
	f := newFixture(t)
	f.rental(t, 73*time.Hour)
	b := f.bffs(t)
	for _, endpoint := range []string{"balance", "standing", "summary"} {
		t.Run(endpoint, func(t *testing.T) {
			customerPath := "/api/v1/" + endpoint
			adminPath := fmt.Sprintf("/api/v1/customers/%d/%s", f.customer, endpoint)
			request(t, b.customer, "GET", customerPath, "", 401)
			request(t, b.customer, "GET", customerPath, b.staffToken, 403)
			request(t, b.admin, "GET", adminPath, "", 401)
			request(t, b.admin, "GET", adminPath, b.customerToken, 403)
			own := request(t, b.customer, "GET", customerPath+"?customer_id=1", b.customerToken, 200)
			admin := request(t, b.admin, "GET", adminPath, b.staffToken, 200)
			field := "outstanding_balance"
			if endpoint == "balance" {
				field = "balance"
			}
			if own[field] != "3.99" || admin[field] != "3.99" || own["customer_id"] != float64(f.customer) {
				t.Fatalf("wrong customer/balance: %+v %+v", own, admin)
			}
			request(t, b.admin, "GET", "/api/v1/customers/2147483647/"+endpoint, b.staffToken, 404)
			request(t, b.admin, "GET", "/api/v1/customers/invalid/"+endpoint, b.staffToken, 400)
			request(t, b.admin, "GET", "/api/v1/customers/0/"+endpoint, b.staffToken, 400)
		})
	}
}
func TestBFFReturns(t *testing.T) {
	for _, admin := range []bool{false, true} {
		for _, late := range []bool{false, true} {
			t.Run(fmt.Sprintf("admin=%v/late=%v", admin, late), func(t *testing.T) {
				f := newFixture(t)
				age := time.Hour
				fee := "0.00"
				days := float64(0)
				if late {
					age = 73 * time.Hour
					fee = "1.00"
					days = 1
				}
				id := f.rental(t, age)
				b := f.bffs(t)
				h, token := b.customer, b.customerToken
				if admin {
					h, token = b.admin, b.staffToken
				}
				path := fmt.Sprintf("/api/v1/rentals/%d/return", id)
				body := request(t, h, "POST", path, token, 200)
				if body["late_fee"] != fee || body["days_overdue"] != days {
					t.Fatalf("return contract: %+v", body)
				}
				request(t, h, "POST", path, token, 404)
			})
		}
	}
	t.Run("ownership", func(t *testing.T) {
		f := newFixture(t)
		other := newFixture(t)
		id := other.rental(t, time.Hour)
		b := f.bffs(t)
		request(t, b.customer, "POST", fmt.Sprintf("/api/v1/rentals/%d/return", id), b.customerToken, 403)
	})
}
func TestRevenueEndpoints(t *testing.T) {
	f := newFixture(t)
	id := f.rental(t, 73*time.Hour)
	b := f.bffs(t)
	start := time.Now().Add(-time.Minute)
	if _, err := f.rentals.ReturnRental(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/reports/revenue?start_date=" + url.QueryEscape(start.Format(time.RFC3339)) + "&end_date=" + url.QueryEscape(time.Now().Add(time.Minute).Format(time.RFC3339))
	request(t, b.admin, "GET", path, b.customerToken, 403)
	body := request(t, b.admin, "GET", path, b.staffToken, 200)
	if body["total_revenue"] != "1.00" {
		t.Fatalf("revenue=%+v", body)
	}
	for _, query := range []string{"", "?start_date=bad&end_date=bad", "?start_date=2026-01-02T00:00:00Z&end_date=2026-01-01T00:00:00Z"} {
		request(t, b.admin, "GET", "/api/v1/reports/revenue"+query, b.staffToken, 400)
	}
	for _, invalid := range []*timestamppb.Timestamp{{Seconds: 253402300800}, {Seconds: 1, Nanos: -1}} {
		_, err := b.payment.GetRevenueByStore(context.Background(), &pv1.GetRevenueByStoreRequest{StartDate: invalid, EndDate: timestamppb.Now()})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid timestamp error=%v", err)
		}
	}
	// End is exclusive, including when a payment sits exactly on the boundary.
	var paidAt time.Time
	if err := f.pool.QueryRow(context.Background(), `SELECT payment_date FROM payment WHERE rental_id=$1`, id).Scan(&paidAt); err != nil {
		t.Fatal(err)
	}
	result, err := b.payment.GetRevenueByStore(context.Background(), &pv1.GetRevenueByStoreRequest{StartDate: timestamppb.New(start), EndDate: timestamppb.New(paidAt)})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalRevenue != "0.00" {
		t.Fatalf("exclusive end revenue=%s", result.TotalRevenue)
	}
}

func TestRentalPolicyHTTPStatus(t *testing.T) {
	f := newFixture(t)
	for i := 0; i < 5; i++ {
		f.rental(t, time.Hour)
	}
	b := f.bffs(t)
	body := fmt.Sprintf(`{"inventory_id":%d,"customer_id":%d,"staff_id":%d}`, f.item(t), f.customer, f.staff)
	r := httptest.NewRequest("POST", "/api/v1/rentals", strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+b.staffToken)
	w := httptest.NewRecorder()
	b.admin.ServeHTTP(w, r)
	if w.Code != http.StatusPreconditionFailed {
		t.Fatalf("policy status=%d: %s", w.Code, w.Body.String())
	}
}
