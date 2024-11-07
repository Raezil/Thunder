package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

var jwtSecret = []byte("your-256-bit-secret")

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func NewClaims(email string) *Claims {
	expirationTime := time.Now().Add(24 * time.Hour)
	return &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
}

func GenerateJWT(email string) (string, error) {
	claims := NewClaims(email)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	return claims, nil
}

func CurrentUser(ctx context.Context) (string, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}
	current_user := md["current_user"]
	return current_user[0], nil
}
