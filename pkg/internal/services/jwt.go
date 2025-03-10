package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

// getJWTSecret fetches the JWT secret securely from environment variables.
func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set in environment variables")
	}
	return []byte(secret), nil
}

// Claims struct for JWT payload with added issuer claim.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// NewClaims creates a new Claims object with expiration time and an issuer.
func NewClaims(email string) *Claims {
	expirationTime := time.Now().Add(24 * time.Hour)
	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "default-issuer" // Fallback issuer, but best to set it in env.
	}
	return &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    issuer,
			// Optionally, add Audience, Subject, etc.
		},
	}
}

// GenerateJWT generates a JWT token securely with additional claims.
func GenerateJWT(email string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	claims := NewClaims(email)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return tokenString, nil
}

// VerifyJWT verifies a JWT token, strictly checks the signing method, and extracts claims.
func VerifyJWT(tokenStr string) (*Claims, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method conforms to expected signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}
	return claims, nil
}

// CurrentUser extracts the user email from context metadata safely.
func CurrentUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	currentUser, exists := md["current_user"]
	if !exists || len(currentUser) == 0 {
		return "", fmt.Errorf("current_user metadata is missing")
	}
	return currentUser[0], nil
}
