package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

// HashPassword generates a bcrypt hash from a plaintext password.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

// ComparePassword compares a plaintext password with a bcrypt hash.
// Returns nil on match, an error otherwise.
func ComparePassword(hashedPassword, plainPassword string) error {
	if hashedPassword == "" || plainPassword == "" {
		return fmt.Errorf("password and hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("incorrect password")
		}
		return fmt.Errorf("compare password: %w", err)
	}

	return nil
}
