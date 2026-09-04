package router

import (
	"net/http"

	"github.com/enkaigaku/dvd-rental/internal/bff/customer/handler"
	"github.com/enkaigaku/dvd-rental/pkg/auth"
	"github.com/enkaigaku/dvd-rental/pkg/middleware"
)

// NewRouter creates the HTTP handler with all routes and middleware.
func NewRouter(
	authH *handler.AuthHandler,
	filmH *handler.FilmHandler,
	rentalH *handler.RentalHandler,
	paymentH *handler.PaymentHandler,
	profileH *handler.ProfileHandler,
	authMw *middleware.AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()
	requireRole := func(h http.Handler) http.Handler {
		return authMw.Require(authMw.RequireRole(auth.RoleCustomer)(h))
	}

	// --- Public: Auth ---
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", authH.Logout)

	// --- Public: Films (read-only) ---
	mux.HandleFunc("GET /api/v1/films/search", filmH.SearchFilms)
	mux.HandleFunc("GET /api/v1/films/category/{id}", filmH.ListFilmsByCategory)
	mux.HandleFunc("GET /api/v1/films/actor/{id}", filmH.ListFilmsByActor)
	mux.HandleFunc("GET /api/v1/films/{id}", filmH.GetFilm)
	mux.HandleFunc("GET /api/v1/films", filmH.ListFilms)
	mux.HandleFunc("GET /api/v1/categories", filmH.ListCategories)
	mux.HandleFunc("GET /api/v1/actors", filmH.ListActors)

	// --- Protected: Rentals ---
	mux.Handle("GET /api/v1/rentals/{id}", requireRole(http.HandlerFunc(rentalH.GetRental)))
	mux.Handle("GET /api/v1/rentals", requireRole(http.HandlerFunc(rentalH.ListRentals)))
	mux.Handle("POST /api/v1/rentals/{id}/return", requireRole(http.HandlerFunc(rentalH.ReturnRental)))
	mux.Handle("POST /api/v1/rentals", requireRole(http.HandlerFunc(rentalH.CreateRental)))

	// --- Protected: Payments ---
	mux.Handle("GET /api/v1/payments", requireRole(http.HandlerFunc(paymentH.ListPayments)))
	mux.Handle("GET /api/v1/balance", requireRole(http.HandlerFunc(paymentH.GetMyBalance)))

	// --- Protected: Profile ---
	mux.Handle("GET /api/v1/profile", requireRole(http.HandlerFunc(profileH.GetProfile)))
	mux.Handle("PUT /api/v1/profile", requireRole(http.HandlerFunc(profileH.UpdateProfile)))
	mux.Handle("GET /api/v1/standing", requireRole(http.HandlerFunc(profileH.GetMyStanding)))
	mux.Handle("GET /api/v1/summary", requireRole(http.HandlerFunc(profileH.GetMySummary)))

	// Apply middleware chain: Recovery → Logging → CORS → router.
	return middleware.Recovery(
		middleware.Logging(
			middleware.CORS(middleware.DefaultCORSConfig())(mux),
		),
	)
}
