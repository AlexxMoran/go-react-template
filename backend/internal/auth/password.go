package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/goapp/pkg/apperror"
)

// bcryptCost balances brute-force resistance against login latency.
const bcryptCost = 12

// HashPassword returns the bcrypt hash of a plaintext password.
func HashPassword(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", apperror.Internal(err)
	}
	return string(hash), nil
}

// CheckPassword reports whether plaintext matches the stored bcrypt hash.
func CheckPassword(hash, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}
