package auth

import "testing"

func TestHashPasswordMatchesOriginalOnly(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "password123" {
		t.Fatal("password hash must not equal the plaintext password")
	}
	if !CheckPassword(hash, "password123") {
		t.Fatal("expected hash to match the original password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("expected wrong password to be rejected")
	}
}
