package router

import (
	"net/http"

	"github.com/tokyoyuan/dvd-rental/internal/bff/admin/handler"
	"github.com/tokyoyuan/dvd-rental/pkg/middleware"
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

	// --- Public: Auth ---
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authH.Logout)

	// --- Protected: Stores ---
	mux.Handle("GET /api/v1/stores", authMw.Require(http.HandlerFunc(storeH.ListStores)))
	mux.Handle("GET /api/v1/stores/{id}", authMw.Require(http.HandlerFunc(storeH.GetStore)))
	mux.Handle("POST /api/v1/stores", authMw.Require(http.HandlerFunc(storeH.CreateStore)))
	mux.Handle("PUT /api/v1/stores/{id}", authMw.Require(http.HandlerFunc(storeH.UpdateStore)))
	mux.Handle("DELETE /api/v1/stores/{id}", authMw.Require(http.HandlerFunc(storeH.DeleteStore)))

	// --- Protected: Staff ---
	mux.Handle("GET /api/v1/staff", authMw.Require(http.HandlerFunc(staffH.ListStaff)))
	mux.Handle("GET /api/v1/staff/{id}", authMw.Require(http.HandlerFunc(staffH.GetStaff)))
	mux.Handle("POST /api/v1/staff", authMw.Require(http.HandlerFunc(staffH.CreateStaff)))
	mux.Handle("PUT /api/v1/staff/{id}", authMw.Require(http.HandlerFunc(staffH.UpdateStaff)))
	mux.Handle("POST /api/v1/staff/{id}/deactivate", authMw.Require(http.HandlerFunc(staffH.DeactivateStaff)))
	mux.Handle("PUT /api/v1/staff/{id}/password", authMw.Require(http.HandlerFunc(staffH.UpdateStaffPassword)))

	// --- Protected: Customers ---
	mux.Handle("GET /api/v1/customers", authMw.Require(http.HandlerFunc(customerH.ListCustomers)))
	mux.Handle("GET /api/v1/customers/{id}", authMw.Require(http.HandlerFunc(customerH.GetCustomer)))
	mux.Handle("POST /api/v1/customers", authMw.Require(http.HandlerFunc(customerH.CreateCustomer)))
	mux.Handle("PUT /api/v1/customers/{id}", authMw.Require(http.HandlerFunc(customerH.UpdateCustomer)))
	mux.Handle("DELETE /api/v1/customers/{id}", authMw.Require(http.HandlerFunc(customerH.DeleteCustomer)))

	// --- Protected: Films ---
	mux.Handle("GET /api/v1/films", authMw.Require(http.HandlerFunc(filmH.ListFilms)))
	mux.Handle("GET /api/v1/films/{id}", authMw.Require(http.HandlerFunc(filmH.GetFilm)))
	mux.Handle("POST /api/v1/films", authMw.Require(http.HandlerFunc(filmH.CreateFilm)))
	mux.Handle("PUT /api/v1/films/{id}", authMw.Require(http.HandlerFunc(filmH.UpdateFilm)))
	mux.Handle("DELETE /api/v1/films/{id}", authMw.Require(http.HandlerFunc(filmH.DeleteFilm)))
	mux.Handle("POST /api/v1/films/{id}/actors", authMw.Require(http.HandlerFunc(filmH.AddActorToFilm)))
	mux.Handle("DELETE /api/v1/films/{id}/actors/{actorId}", authMw.Require(http.HandlerFunc(filmH.RemoveActorFromFilm)))
	mux.Handle("POST /api/v1/films/{id}/categories", authMw.Require(http.HandlerFunc(filmH.AddCategoryToFilm)))
	mux.Handle("DELETE /api/v1/films/{id}/categories/{categoryId}", authMw.Require(http.HandlerFunc(filmH.RemoveCategoryFromFilm)))

	// --- Protected: Actors ---
	mux.Handle("GET /api/v1/actors", authMw.Require(http.HandlerFunc(filmH.ListActors)))
	mux.Handle("GET /api/v1/actors/{id}", authMw.Require(http.HandlerFunc(filmH.GetActor)))
	mux.Handle("POST /api/v1/actors", authMw.Require(http.HandlerFunc(filmH.CreateActor)))
	mux.Handle("PUT /api/v1/actors/{id}", authMw.Require(http.HandlerFunc(filmH.UpdateActor)))
	mux.Handle("DELETE /api/v1/actors/{id}", authMw.Require(http.HandlerFunc(filmH.DeleteActor)))

	// --- Protected: Categories (read-only) ---
	mux.Handle("GET /api/v1/categories", authMw.Require(http.HandlerFunc(filmH.ListCategories)))

	// --- Protected: Languages (read-only) ---
	mux.Handle("GET /api/v1/languages", authMw.Require(http.HandlerFunc(filmH.ListLanguages)))

	// --- Protected: Inventory ---
	mux.Handle("GET /api/v1/inventory", authMw.Require(http.HandlerFunc(inventoryH.ListInventory)))
	mux.Handle("GET /api/v1/inventory/{id}", authMw.Require(http.HandlerFunc(inventoryH.GetInventory)))
	mux.Handle("POST /api/v1/inventory", authMw.Require(http.HandlerFunc(inventoryH.CreateInventory)))
	mux.Handle("DELETE /api/v1/inventory/{id}", authMw.Require(http.HandlerFunc(inventoryH.DeleteInventory)))
	mux.Handle("GET /api/v1/inventory/{id}/available", authMw.Require(http.HandlerFunc(inventoryH.CheckAvailability)))

	// --- Protected: Rentals ---
	mux.Handle("GET /api/v1/rentals", authMw.Require(http.HandlerFunc(rentalH.ListRentals)))
	mux.Handle("GET /api/v1/rentals/overdue", authMw.Require(http.HandlerFunc(rentalH.ListOverdueRentals)))
	mux.Handle("GET /api/v1/rentals/{id}", authMw.Require(http.HandlerFunc(rentalH.GetRental)))
	mux.Handle("POST /api/v1/rentals", authMw.Require(http.HandlerFunc(rentalH.CreateRental)))
	mux.Handle("POST /api/v1/rentals/{id}/return", authMw.Require(http.HandlerFunc(rentalH.ReturnRental)))
	mux.Handle("DELETE /api/v1/rentals/{id}", authMw.Require(http.HandlerFunc(rentalH.DeleteRental)))

	// --- Protected: Payments ---
	mux.Handle("GET /api/v1/payments", authMw.Require(http.HandlerFunc(paymentH.ListPayments)))
	mux.Handle("GET /api/v1/payments/{id}", authMw.Require(http.HandlerFunc(paymentH.GetPayment)))
	mux.Handle("POST /api/v1/payments", authMw.Require(http.HandlerFunc(paymentH.CreatePayment)))
	mux.Handle("DELETE /api/v1/payments/{id}", authMw.Require(http.HandlerFunc(paymentH.DeletePayment)))

	// Apply middleware chain: Recovery → Logging → CORS → router.
	return middleware.Recovery(
		middleware.Logging(
			middleware.CORS(middleware.DefaultCORSConfig())(mux),
		),
	)
}
