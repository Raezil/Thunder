package tests

import (
	"context"
	"fmt"
	. "generated"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAndLogin(t *testing.T) {
	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock of the AuthClient
	mockAuthClient := NewMockAuthClient(ctrl)

	// Define the input and expected output for registration
	regReq := &RegisterRequest{
		Email:    "john_doe@example.com",
		Password: "password",
		Name:     "test",
		Surname:  "test",
		Age:      27,
	}
	regReply := fmt.Sprintf("Congratulations, User email: %s got created!", "john_doe@example.com")
	expectedRegRes := &RegisterReply{Reply: regReply}

	// Set up the mock to expect a Register call and return the expected response
	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(regReq), gomock.Any()).
		Return(expectedRegRes, nil)

	// Call the Register method
	regRes, regErr := mockAuthClient.Register(context.Background(), regReq)

	// Validate the registration response
	assert.NoError(t, regErr)
	assert.Equal(t, expectedRegRes, regRes)

	// Define the input and expected output for login
	loginReq := &LoginRequest{Email: "john_doe@example.com", Password: "securepass"}
	expectedLoginRes := &LoginReply{Token: "some-jwt-token"}

	// Set up the mock to expect a Login call and return the expected response
	mockAuthClient.EXPECT().Login(gomock.Any(), gomock.Eq(loginReq), gomock.Any()).
		Return(expectedLoginRes, nil)

	// Call the Login method
	loginRes, loginErr := mockAuthClient.Login(context.Background(), loginReq)

	// Validate the login response
	assert.NoError(t, loginErr)
	assert.Equal(t, expectedLoginRes, loginRes)

	mockAuthClient.EXPECT().SampleProtected(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	sampleRes, sampleErr := mockAuthClient.SampleProtected(context.Background(), &ProtectedRequest{})
	assert.Error(t, sampleErr)
	assert.Nil(t, sampleRes)
}
