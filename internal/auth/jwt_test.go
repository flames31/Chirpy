package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWTAndValidateJWT_ValidToken(t *testing.T) {
	secret := "super-secret-key"
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	parsedUserID, err := ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	if parsedUserID != userID {
		t.Errorf("Expected userID %v, got %v", userID, parsedUserID)
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	secret := "super-secret-key"
	wrongSecret := "wrong-key"
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = ValidateJWT(tokenString, wrongSecret)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "super-secret-key"
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = ValidateJWT(tokenString, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateJWT_InvalidUUID(t *testing.T) {
	secret := "super-secret-key"
	expiresIn := time.Minute * 10

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "not-a-valid-uuid",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	_, err = ValidateJWT(tokenString, secret)
	if err == nil {
		t.Error("Expected error for invalid UUID, got nil")
	}
}
