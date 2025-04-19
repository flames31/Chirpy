package auth

import (
	"testing"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "supersecret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error from HashPassword, got %v", err)
	}

	err = CheckPasswordHash(hash, password)
	if err != nil {
		t.Errorf("expected password to match hash, got error: %v", err)
	}
}

func TestCheckPasswordHash_InvalidPassword(t *testing.T) {
	password := "supersecret123"
	wrongPassword := "nottherightone"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("error hashing password: %v", err)
	}

	err = CheckPasswordHash(hash, wrongPassword)
	if err == nil {
		t.Error("expected error for invalid password, got nil")
	}
}
