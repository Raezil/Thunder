package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// main.go
func GetUserName(ctx context.Context, client *PrismaClient, postID string) (string, error) {
	user, err := client.User.FindUnique(
		User.ID.Equals(postID),
	).Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("error fetching user: %w", err)
	}

	return user.Name, nil
}

func TestGetUserName_error(t *testing.T) {
	client, mock, ensure := NewMock()
	defer ensure(t)

	mock.User.Expect(
		client.User.FindUnique(
			User.ID.Equals("123"),
		),
	).Errors(ErrNotFound)

	_, err := GetUserName(context.Background(), client, "123")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error expected to return ErrNotFound but is %s", err)
	}
}

func TestGetUserName_returns(t *testing.T) {
	// create a new mock
	// this returns a mock prisma `client` and a `mock` object to set expectations
	client, mock, ensure := NewMock()
	// defer calling ensure, which makes sure all of the expectations were met and actually called
	// calling this makes sure that an error is returned if there was no query happening for a given expectation
	// and makes sure that all of them succeeded
	defer ensure(t)

	expected := UserModel{
		InnerUser: InnerUser{
			ID:       "123",
			Name:     "foo",
			Email:    "kmosc@protonmail.com",
			Password: "password",
		},
	}

	// start the expectation
	mock.User.Expect(
		// define your exact query as in your tested function
		// call it with the exact arguments which you expect the function to be called with
		// you can copy and paste this from your tested function, and just put specific values into the arguments
		client.User.FindUnique(
			User.ID.Equals("123"),
		),
	).Returns(expected) // sets the object which should be returned in the function call

	// mocking set up is done; let's define the actual test now
	name, err := GetUserName(context.Background(), client, "123")
	if err != nil {
		t.Fatal(err)
	}

	if name != "foo" {
		t.Fatalf("name expected to be foo but is %s", name)
	}
}
