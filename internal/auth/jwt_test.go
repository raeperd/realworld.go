package auth_test

import (
	"testing"

	"github.com/raeperd/test"

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
	test.Nil(t, err)
	test.NotZero(t, token)
}

func TestParseToken_WithValidToken_ReturnsCorrectClaims(t *testing.T) {
	t.Parallel()

	// Given
	userID := int64(456)
	username := "parseuser"
	secret := "parsesecret"

	// Generate a token first
	token, err := auth.GenerateToken(userID, username, secret)
	test.Nil(t, err)

	// When
	claims, err := auth.ParseToken(token, secret)
	// Then
	test.Nil(t, err)
	test.Equal(t, userID, claims.UserID)
	test.Equal(t, username, claims.Username)
}
