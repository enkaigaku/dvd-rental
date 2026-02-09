// Package auth provides authentication utilities including JWT and password hashing.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// ErrInvalidToken is returned when a token is malformed or invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a token has expired.
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidSigningKey is returned when the signing key is empty.
	ErrInvalidSigningKey = errors.New("invalid signing key")
)

// Role represents user roles in the system.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleStaff    Role = "staff"
)

// Claims represents JWT custom claims.
type Claims struct {
	UserID   int32  `json:"user_id"`
	Role     Role   `json:"role"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents an access token and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// JWTManager handles JWT operations.
type JWTManager struct {
	secretKey            []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(secretKey string, accessTokenDuration, refreshTokenDuration time.Duration) (*JWTManager, error) {
	if secretKey == "" {
		return nil, ErrInvalidSigningKey
	}
	return &JWTManager{
		secretKey:            []byte(secretKey),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}, nil
}

// GenerateAccessToken creates a short-lived access token.
func (m *JWTManager) GenerateAccessToken(userID int32, role Role, identifier string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	if role == RoleCustomer {
		claims.Email = identifier
	} else {
		claims.Username = identifier
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// GenerateRefreshToken creates a long-lived refresh token. Returns the signed
// token string and its JTI (for storing in Redis).
func (m *JWTManager) GenerateRefreshToken(userID int32, role Role) (string, string, error) {
	now := time.Now()
	jti := uuid.New().String()

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secretKey)
	return signed, jti, err
}

// GenerateTokenPair generates both access and refresh tokens.
func (m *JWTManager) GenerateTokenPair(userID int32, role Role, identifier string) (*TokenPair, string, error) {
	accessToken, err := m.GenerateAccessToken(userID, role, identifier)
	if err != nil {
		return nil, "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, jti, err := m.GenerateRefreshToken(userID, role)
	if err != nil {
		return nil, "", fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.accessTokenDuration.Seconds()),
	}, jti, nil
}

// VerifyToken validates and parses a JWT token.
func (m *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ExtractToken extracts the token from an Authorization header value ("Bearer <token>").
func ExtractToken(authHeader string) (string, error) {
	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", errors.New("authorization header must use 'Bearer <token>' format")
	}
	return authHeader[len(prefix):], nil
}
