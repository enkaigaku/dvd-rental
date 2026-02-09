package router

import (
	"net/http"

	"github.com/tokyoyuan/dvd-rental/internal/bff/customer/handler"
	"github.com/tokyoyuan/dvd-rental/pkg/middleware"
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
	mux.Handle("GET /api/v1/rentals/{id}", authMw.Require(http.HandlerFunc(rentalH.GetRental)))
	mux.Handle("GET /api/v1/rentals", authMw.Require(http.HandlerFunc(rentalH.ListRentals)))
	mux.Handle("POST /api/v1/rentals/{id}/return", authMw.Require(http.HandlerFunc(rentalH.ReturnRental)))
	mux.Handle("POST /api/v1/rentals", authMw.Require(http.HandlerFunc(rentalH.CreateRental)))

	// --- Protected: Payments ---
	mux.Handle("GET /api/v1/payments", authMw.Require(http.HandlerFunc(paymentH.ListPayments)))

	// --- Protected: Profile ---
	mux.Handle("GET /api/v1/profile", authMw.Require(http.HandlerFunc(profileH.GetProfile)))
	mux.Handle("PUT /api/v1/profile", authMw.Require(http.HandlerFunc(profileH.UpdateProfile)))

	// Apply middleware chain: Recovery → Logging → CORS → router.
	return middleware.Recovery(
		middleware.Logging(
			middleware.CORS(middleware.DefaultCORSConfig())(mux),
		),
	)
}
