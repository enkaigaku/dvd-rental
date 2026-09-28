package router

import (
	"net/http"

	"github.com/enkaigaku/dvd-rental/internal/bff/admin/handler"
	"github.com/enkaigaku/dvd-rental/pkg/auth"
	"github.com/enkaigaku/dvd-rental/pkg/middleware"
)

// NewRouter creates the HTTP handler with all routes and middleware.
func NewRouter(
	authH *handler.AuthHandler,
	storeH *handler.StoreHandler,
	staffH *handler.StaffHandler,
	customerH *handler.CustomerHandler,
	filmH *handler.FilmHandler,
	inventoryH *handler.InventoryHandler,
	rentalH *handler.RentalHandler,
	paymentH *handler.PaymentHandler,
	authMw *middleware.AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()
	requireRole := func(h http.Handler) http.Handler {
		return authMw.Require(authMw.RequireRole(auth.RoleStaff)(h))
	}

	// --- Public: Auth ---
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authH.Logout)

	// --- Protected: Stores ---
	mux.Handle("GET /api/v1/stores", requireRole(http.HandlerFunc(storeH.ListStores)))
	mux.Handle("GET /api/v1/stores/{id}", requireRole(http.HandlerFunc(storeH.GetStore)))
	mux.Handle("POST /api/v1/stores", requireRole(http.HandlerFunc(storeH.CreateStore)))
	mux.Handle("PUT /api/v1/stores/{id}", requireRole(http.HandlerFunc(storeH.UpdateStore)))
	mux.Handle("DELETE /api/v1/stores/{id}", requireRole(http.HandlerFunc(storeH.DeleteStore)))

	// --- Protected: Staff ---
	mux.Handle("GET /api/v1/staff", requireRole(http.HandlerFunc(staffH.ListStaff)))
	mux.Handle("GET /api/v1/staff/{id}", requireRole(http.HandlerFunc(staffH.GetStaff)))
	mux.Handle("POST /api/v1/staff", requireRole(http.HandlerFunc(staffH.CreateStaff)))
	mux.Handle("PUT /api/v1/staff/{id}", requireRole(http.HandlerFunc(staffH.UpdateStaff)))
	mux.Handle("POST /api/v1/staff/{id}/deactivate", requireRole(http.HandlerFunc(staffH.DeactivateStaff)))
	mux.Handle("PUT /api/v1/staff/{id}/password", requireRole(http.HandlerFunc(staffH.UpdateStaffPassword)))

	// --- Protected: Customers ---
	mux.Handle("GET /api/v1/customers", requireRole(http.HandlerFunc(customerH.ListCustomers)))
	mux.Handle("GET /api/v1/customers/{id}", requireRole(http.HandlerFunc(customerH.GetCustomer)))
	mux.Handle("POST /api/v1/customers", requireRole(http.HandlerFunc(customerH.CreateCustomer)))
	mux.Handle("PUT /api/v1/customers/{id}", requireRole(http.HandlerFunc(customerH.UpdateCustomer)))
	mux.Handle("DELETE /api/v1/customers/{id}", requireRole(http.HandlerFunc(customerH.DeleteCustomer)))
	mux.Handle("GET /api/v1/customers/{id}/standing", requireRole(http.HandlerFunc(customerH.GetCustomerStanding)))
	mux.Handle("GET /api/v1/customers/{id}/summary", requireRole(http.HandlerFunc(customerH.GetCustomerSummary)))
	mux.Handle("GET /api/v1/customers/{id}/balance", requireRole(http.HandlerFunc(paymentH.GetCustomerBalance)))

	// --- Protected: Films ---
	mux.Handle("GET /api/v1/films", requireRole(http.HandlerFunc(filmH.ListFilms)))
	mux.Handle("GET /api/v1/films/{id}", requireRole(http.HandlerFunc(filmH.GetFilm)))
	mux.Handle("POST /api/v1/films", requireRole(http.HandlerFunc(filmH.CreateFilm)))
	mux.Handle("PUT /api/v1/films/{id}", requireRole(http.HandlerFunc(filmH.UpdateFilm)))
	mux.Handle("DELETE /api/v1/films/{id}", requireRole(http.HandlerFunc(filmH.DeleteFilm)))
	mux.Handle("POST /api/v1/films/{id}/actors", requireRole(http.HandlerFunc(filmH.AddActorToFilm)))
	mux.Handle("DELETE /api/v1/films/{id}/actors/{actorId}", requireRole(http.HandlerFunc(filmH.RemoveActorFromFilm)))
	mux.Handle("POST /api/v1/films/{id}/categories", requireRole(http.HandlerFunc(filmH.AddCategoryToFilm)))
	mux.Handle("DELETE /api/v1/films/{id}/categories/{categoryId}", requireRole(http.HandlerFunc(filmH.RemoveCategoryFromFilm)))

	// --- Protected: Actors ---
	mux.Handle("GET /api/v1/actors", requireRole(http.HandlerFunc(filmH.ListActors)))
	mux.Handle("GET /api/v1/actors/{id}", requireRole(http.HandlerFunc(filmH.GetActor)))
	mux.Handle("POST /api/v1/actors", requireRole(http.HandlerFunc(filmH.CreateActor)))
	mux.Handle("PUT /api/v1/actors/{id}", requireRole(http.HandlerFunc(filmH.UpdateActor)))
	mux.Handle("DELETE /api/v1/actors/{id}", requireRole(http.HandlerFunc(filmH.DeleteActor)))

	// --- Protected: Categories (read-only) ---
	mux.Handle("GET /api/v1/categories", requireRole(http.HandlerFunc(filmH.ListCategories)))

	// --- Protected: Languages (read-only) ---
	mux.Handle("GET /api/v1/languages", requireRole(http.HandlerFunc(filmH.ListLanguages)))

	// --- Protected: Inventory ---
	mux.Handle("GET /api/v1/inventory", requireRole(http.HandlerFunc(inventoryH.ListInventory)))
	mux.Handle("GET /api/v1/inventory/{id}", requireRole(http.HandlerFunc(inventoryH.GetInventory)))
	mux.Handle("POST /api/v1/inventory", requireRole(http.HandlerFunc(inventoryH.CreateInventory)))
	mux.Handle("DELETE /api/v1/inventory/{id}", requireRole(http.HandlerFunc(inventoryH.DeleteInventory)))
	mux.Handle("GET /api/v1/inventory/{id}/available", requireRole(http.HandlerFunc(inventoryH.CheckAvailability)))

	// --- Protected: Rentals ---
	mux.Handle("GET /api/v1/rentals", requireRole(http.HandlerFunc(rentalH.ListRentals)))
	mux.Handle("GET /api/v1/rentals/overdue", requireRole(http.HandlerFunc(rentalH.ListOverdueRentals)))
	mux.Handle("GET /api/v1/rentals/{id}", requireRole(http.HandlerFunc(rentalH.GetRental)))
	mux.Handle("POST /api/v1/rentals", requireRole(http.HandlerFunc(rentalH.CreateRental)))
	mux.Handle("POST /api/v1/rentals/{id}/return", requireRole(http.HandlerFunc(rentalH.ReturnRental)))
	mux.Handle("DELETE /api/v1/rentals/{id}", requireRole(http.HandlerFunc(rentalH.DeleteRental)))

	// --- Protected: Payments ---
	mux.Handle("GET /api/v1/payments", requireRole(http.HandlerFunc(paymentH.ListPayments)))
	mux.Handle("GET /api/v1/payments/{id}", requireRole(http.HandlerFunc(paymentH.GetPayment)))
	mux.Handle("POST /api/v1/payments", requireRole(http.HandlerFunc(paymentH.CreatePayment)))
	mux.Handle("DELETE /api/v1/payments/{id}", requireRole(http.HandlerFunc(paymentH.DeletePayment)))

	// --- Protected: Reports ---
	mux.Handle("GET /api/v1/reports/revenue", requireRole(http.HandlerFunc(paymentH.GetRevenueByStore)))

	// Apply middleware chain: Recovery → Logging → CORS → router.
	return middleware.Recovery(
		middleware.Logging(
			middleware.CORS(middleware.DefaultCORSConfig())(mux),
		),
	)
}
