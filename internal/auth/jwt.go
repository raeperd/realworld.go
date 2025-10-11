package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int64
	Username string
}

func GenerateToken(userID int64, username string, secret string) (string, error) {
	// Create claims with user data and 7-day expiration
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	// Create token with HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (*Claims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64) // JSON numbers are float64
		if !ok {
			return nil, errors.New("invalid user_id claim")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return nil, errors.New("invalid username claim")
		}

		return &Claims{
			UserID:   int64(userID),
			Username: username,
		}, nil
	}

	return nil, errors.New("invalid token")
}
