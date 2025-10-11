package auth_test

import (
	"testing"

	"github.com/raeperd/realworld.go/internal/auth"
)

func TestGenerateToken_WithValidInputs_ReturnsToken(t *testing.T) {
	t.Parallel()

	// Given
	userID := int64(123)
	username := "testuser"
	secret := "testsecret"

	// When
	token, err := auth.GenerateToken(userID, username, secret)
	// Then
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestParseToken_WithValidToken_ReturnsCorrectClaims(t *testing.T) {
	t.Parallel()

	// Given
	userID := int64(456)
	username := "parseuser"
	secret := "parsesecret"

	// Generate a token first
	token, err := auth.GenerateToken(userID, username, secret)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// When
	claims, err := auth.ParseToken(token, secret)
	// Then
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %d, got %d", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("expected username %s, got %s", username, claims.Username)
	}
}
