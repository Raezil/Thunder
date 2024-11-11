package backend_test

import (
	"backend" // Import your backend package here
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock of the AuthClient
	mockAuthClient := backend.NewMockAuthClient(ctrl)

	// Define the input and expected output
	req := &backend.RegisterRequest{
		Email:    "john_doe@example.com",
		Password: "password",
		Name:     "test",
		Surname:  "test",
		Age:      27,
	}
	reply := fmt.Sprintf("Congratulations, User email: %s got created!", "john_doe@example.com")
	expectedRes := &backend.RegisterReply{Reply: reply}

	// Set up the mock to expect a Register call and return the expected response
	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(req), gomock.Any()).
		Return(expectedRes, nil)

	// Call the method you want to test
	res, err := mockAuthClient.Register(context.Background(), req)

	// Validate the response
	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res)
}

func TestLogin(t *testing.T) {
	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock of the AuthClient
	mockAuthClient := backend.NewMockAuthClient(ctrl)

	// Define the input and expected output
	loginReq := &backend.LoginRequest{Email: "john_doe@example.com", Password: "securepass"}
	expectedLoginRes := &backend.LoginReply{Token: "some-jwt-token"}

	// Set up the mock to expect a Login call and return the expected response
	mockAuthClient.EXPECT().Login(gomock.Any(), gomock.Eq(loginReq), gomock.Any()).
		Return(expectedLoginRes, nil)

	// Call the method you want to test
	loginRes, err := mockAuthClient.Login(context.Background(), loginReq)

	// Validate the response
	assert.NoError(t, err)
	assert.Equal(t, expectedLoginRes, loginRes)
}
