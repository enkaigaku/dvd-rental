package handler

import (
	"context"
	"net/http"
	"time"

	customerv1 "github.com/tokyoyuan/dvd-rental/gen/proto/customer/v1"
	"github.com/tokyoyuan/dvd-rental/pkg/auth"
	"github.com/tokyoyuan/dvd-rental/pkg/middleware"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	customerClient customerv1.CustomerServiceClient
	jwtManager     *auth.JWTManager
	refreshStore   *auth.RefreshTokenStore
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	customerClient customerv1.CustomerServiceClient,
	jwtManager *auth.JWTManager,
	refreshStore *auth.RefreshTokenStore,
) *AuthHandler {
	return &AuthHandler{
		customerClient: customerClient,
		jwtManager:     jwtManager,
		refreshStore:   refreshStore,
	}
}

// --- JSON models ---

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Login authenticates a customer and returns a token pair.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.Email == "" || req.Password == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "email and password are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 1. Get customer by email (includes password_hash).
	customer, err := h.customerClient.GetCustomerByEmail(ctx, &customerv1.GetCustomerByEmailRequest{
		Email: req.Email,
	})
	if err != nil {
		// Don't reveal whether the email exists.
		middleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "incorrect email or password")
		return
	}

	// 2. Verify password.
	if err := auth.ComparePassword(customer.GetPasswordHash(), req.Password); err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "incorrect email or password")
		return
	}

	// 3. Check customer is active.
	if !customer.GetActive() {
		middleware.WriteJSONError(w, http.StatusForbidden, "ACCOUNT_DISABLED", "account is disabled")
		return
	}

	// 4. Generate token pair.
	tokenPair, jti, err := h.jwtManager.GenerateTokenPair(customer.GetCustomerId(), auth.RoleCustomer, customer.GetEmail())
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate tokens")
		return
	}

	// 5. Store refresh token in Redis.
	if err := h.refreshStore.Store(ctx, jti, auth.RefreshTokenData{
		UserID: customer.GetCustomerId(),
		Role:   auth.RoleCustomer,
	}); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store refresh token")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Refresh generates a new token pair from a refresh token.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := readJSON(r, &req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.RefreshToken == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "refresh_token is required")
		return
	}

	// 1. Verify refresh token.
	claims, err := h.jwtManager.VerifyToken(req.RefreshToken)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired refresh token")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 2. Check Redis for the refresh token.
	tokenData, err := h.refreshStore.Get(ctx, claims.ID)
	if err != nil {
		middleware.WriteJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "refresh token not found or expired")
		return
	}

	// 3. Delete the old refresh token (rotate).
	_ = h.refreshStore.Delete(ctx, claims.ID)

	// 4. Get customer email for the new access token.
	customerDetail, err := h.customerClient.GetCustomer(ctx, &customerv1.GetCustomerRequest{
		CustomerId: tokenData.UserID,
	})
	if err != nil {
		grpcToHTTPError(w, err)
		return
	}

	// 5. Generate new token pair.
	tokenPair, jti, err := h.jwtManager.GenerateTokenPair(tokenData.UserID, tokenData.Role, customerDetail.GetCustomer().GetEmail())
	if err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate tokens")
		return
	}

	// 6. Store new refresh token.
	if err := h.refreshStore.Store(ctx, jti, auth.RefreshTokenData{
		UserID: tokenData.UserID,
		Role:   tokenData.Role,
	}); err != nil {
		middleware.WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store refresh token")
		return
	}

	middleware.WriteJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Logout invalidates a refresh token.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := readJSON(r, &req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.RefreshToken == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "refresh_token is required")
		return
	}

	// Verify token to get JTI; even on error return 204 (idempotent).
	claims, err := h.jwtManager.VerifyToken(req.RefreshToken)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_ = h.refreshStore.Delete(ctx, claims.ID)
	w.WriteHeader(http.StatusNoContent)
}
