package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestGenerateJWT(t *testing.T) {
	email := "test@example.com"
	token, err := GenerateJWT(email)

	assert.NoError(t, err, "expected no error from GenerateJWT")
	assert.NotEmpty(t, token, "expected token to be non-empty")

	// Verify token can be parsed and contains correct email
	claims, err := VerifyJWT(token)
	assert.NoError(t, err, "expected no error from VerifyJWT")
	assert.Equal(t, email, claims.Email, "expected email in claims to match")
}

func TestVerifyJWT(t *testing.T) {
	email := "verify@test.com"
	token, _ := GenerateJWT(email)

	claims, err := VerifyJWT(token)
	assert.NoError(t, err, "expected no error from VerifyJWT")
	assert.Equal(t, email, claims.Email, "expected email in claims to match")

	// Test with invalid token
	_, err = VerifyJWT("invalid.token")
	assert.Error(t, err, "expected error from VerifyJWT with invalid token")
}

func TestCurrentUser(t *testing.T) {
	email := "user@test.com"
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("current_user", email))

	currentUser, err := CurrentUser(ctx)
	assert.NoError(t, err, "expected no error from CurrentUser")
	assert.Equal(t, email, currentUser, "expected current user to match email")

	// Test without metadata
	_, err = CurrentUser(context.Background())
	assert.Error(t, err, "expected error from CurrentUser with missing metadata")
}
