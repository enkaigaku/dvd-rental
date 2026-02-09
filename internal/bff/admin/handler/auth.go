package handler

import (
	"context"
	"net/http"
	"time"

	storev1 "github.com/tokyoyuan/dvd-rental/gen/proto/store/v1"
	"github.com/tokyoyuan/dvd-rental/pkg/auth"
)

// AuthHandler handles staff authentication endpoints.
type AuthHandler struct {
	staffClient  storev1.StaffServiceClient
	jwtManager   *auth.JWTManager
	refreshStore *auth.RefreshTokenStore
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	staffClient storev1.StaffServiceClient,
	jwtManager *auth.JWTManager,
	refreshStore *auth.RefreshTokenStore,
) *AuthHandler {
	return &AuthHandler{
		staffClient:  staffClient,
		jwtManager:   jwtManager,
		refreshStore: refreshStore,
	}
}

// --- JSON models ---

type staffLoginRequest struct {
	Username string `json:"username"`
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

// Login authenticates a staff member and returns a token pair.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req staffLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 1. Get staff by username (includes password_hash).
	staff, err := h.staffClient.GetStaffByUsername(ctx, &storev1.GetStaffByUsernameRequest{
		Username: req.Username,
	})
	if err != nil {
		// Don't reveal whether the username exists.
		writeError(w, http.StatusUnauthorized, "incorrect username or password")
		return
	}

	// 2. Verify password.
	if err := auth.ComparePassword(staff.GetPasswordHash(), req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "incorrect username or password")
		return
	}

	// 3. Check staff is active.
	if !staff.GetActive() {
		writeError(w, http.StatusForbidden, "account is disabled")
		return
	}

	// 4. Generate token pair.
	tokenPair, jti, err := h.jwtManager.GenerateTokenPair(staff.GetStaffId(), auth.RoleStaff, staff.GetUsername())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	// 5. Store refresh token in Redis.
	if err := h.refreshStore.Store(ctx, jti, auth.RefreshTokenData{
		UserID: staff.GetStaffId(),
		Role:   auth.RoleStaff,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Refresh generates a new token pair from a refresh token.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// 1. Verify refresh token.
	claims, err := h.jwtManager.VerifyToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 2. Check Redis for the refresh token.
	tokenData, err := h.refreshStore.Get(ctx, claims.ID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "refresh token not found or expired")
		return
	}

	// 3. Delete the old refresh token (rotate).
	_ = h.refreshStore.Delete(ctx, claims.ID)

	// 4. Get staff info for the new access token.
	staff, err := h.staffClient.GetStaff(ctx, &storev1.GetStaffRequest{
		StaffId: tokenData.UserID,
	})
	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// 5. Generate new token pair.
	tokenPair, jti, err := h.jwtManager.GenerateTokenPair(tokenData.UserID, tokenData.Role, staff.GetUsername())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	// 6. Store new refresh token.
	if err := h.refreshStore.Store(ctx, jti, auth.RefreshTokenData{
		UserID: tokenData.UserID,
		Role:   tokenData.Role,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store refresh token")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	})
}

// Logout invalidates a refresh token.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
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
