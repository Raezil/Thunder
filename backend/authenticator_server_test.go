package backend

import (
	"context"
	"db"
	"testing"

	"github.com/tamathecxder/randomail"
	"go.uber.org/zap"
)

func CleaningDatabase(client *db.PrismaClient, email string) {
	client.User.FindUnique(
		db.User.Email.Equals(email),
	).Delete().Exec(context.Background())
}

func TestAuthenticator_Register_Login(t *testing.T) {
	client := db.NewClient()
	client.Connect()
	defer client.Disconnect()
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
	_, err = authServer.Login(context.Background(), &LoginRequest{
		Email:    email[0],
		Password: "password",
	})
	if err != nil {
		t.Fatalf("failed to log in user: %v", err)
	}
	CleaningDatabase(client, email[0])

}
