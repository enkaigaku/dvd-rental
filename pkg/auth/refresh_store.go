package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RefreshTokenData stores metadata for a refresh token in Redis.
type RefreshTokenData struct {
	UserID int32 `json:"user_id"`
	Role   Role  `json:"role"`
}

// RefreshTokenStore manages refresh tokens in Redis.
type RefreshTokenStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRefreshTokenStore creates a new refresh token store.
func NewRefreshTokenStore(client *redis.Client, ttl time.Duration) *RefreshTokenStore {
	return &RefreshTokenStore{
		client: client,
		ttl:    ttl,
	}
}

// Store saves a refresh token with its metadata.
func (s *RefreshTokenStore) Store(ctx context.Context, jti string, data RefreshTokenData) error {
	key := "refresh:" + jti

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal token data: %w", err)
	}

	if err := s.client.Set(ctx, key, b, s.ttl).Err(); err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}

	return nil
}

// Get retrieves refresh token metadata by JTI.
func (s *RefreshTokenStore) Get(ctx context.Context, jti string) (*RefreshTokenData, error) {
	key := "refresh:" + jti

	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh token not found or expired")
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal token data: %w", err)
	}

	return &data, nil
}

// Delete removes a refresh token (for logout).
func (s *RefreshTokenStore) Delete(ctx context.Context, jti string) error {
	key := "refresh:" + jti

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

// Exists checks whether a refresh token exists in the store.
func (s *RefreshTokenStore) Exists(ctx context.Context, jti string) (bool, error) {
	key := "refresh:" + jti

	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check refresh token: %w", err)
	}

	return count > 0, nil
}
