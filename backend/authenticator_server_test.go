package backend

import (
	"context"
	"db"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/tamathecxder/randomail"
	"go.uber.org/zap"
)

func CleaningUpDatabase(client *db.PrismaClient, email string) {
	client.User.FindUnique(
		db.User.Email.Equals(email),
	).Delete().Exec(context.Background())
}

func TestAuthenticator_Register_Login(t *testing.T) {
	client, _, _ := db.NewMock()
	authServer := AuthenticatorServer{
		PrismaClient: client,
		Logger:       zap.L().Sugar(),
	}
	email := randomail.GenerateRandomEmails(1)
	_, err := authServer.Register(context.Background(), &RegisterRequest{
		Email:    email[0],
		Password: "password",
		Name:     "test",
		Surname:  "test",
		Age:      27,
	})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	reply, err := authServer.Login(context.Background(), &LoginRequest{
		Email:    email[0],
		Password: "password",
	})
	if err != nil {
		t.Fatalf("failed to log in user: %v", err)
	}
	if reply.Token != "" {
		assert.NotEqual(t, reply.Token, "")
		claims, err := VerifyJWT(reply.Token)
		if err != nil {
			t.Fatalf("Token is invalid")
		}
		assert.Equal(t, claims.Email, email[0])

	}
	CleaningUpDatabase(client, email[0])

}
