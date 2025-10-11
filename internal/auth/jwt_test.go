package auth

import (
	"testing"
)

func TestGenerateToken_WithValidInputs_ReturnsToken(t *testing.T) {
	// Given
	userID := int64(123)
	username := "testuser"
	secret := "testsecret"

	// When
	token, err := GenerateToken(userID, username, secret)

	// Then
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}
