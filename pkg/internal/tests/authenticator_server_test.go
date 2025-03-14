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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	// Valid registration request
	regReq := &RegisterRequest{
		Email:    "john_doe@example.com",
		Password: "password",
		Name:     "test",
		Surname:  "test",
		Age:      27,
	}
	regReply := fmt.Sprintf("Congratulations, User email: %s got created!", "john_doe@example.com")
	expectedRegRes := &RegisterReply{Reply: regReply}

	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(regReq), gomock.Any()).Return(expectedRegRes, nil)

	// Call Register
	regRes, regErr := mockAuthClient.Register(context.Background(), regReq)
	assert.NoError(t, regErr)
	assert.Equal(t, expectedRegRes, regRes)

	// Valid login request
	loginReq := &LoginRequest{Email: "john_doe@example.com", Password: "password"}
	expectedLoginRes := &LoginReply{Token: "some-jwt-token"}

	mockAuthClient.EXPECT().Login(gomock.Any(), gomock.Eq(loginReq), gomock.Any()).Return(expectedLoginRes, nil)

	// Call Login
	loginRes, loginErr := mockAuthClient.Login(context.Background(), loginReq)
	assert.NoError(t, loginErr)
	assert.Equal(t, expectedLoginRes, loginRes)
}

func TestRegisterWithEmptyFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	invalidReq := &RegisterRequest{
		Email:    "",
		Password: "",
		Name:     "test",
		Surname:  "test",
		Age:      27,
	}
	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(invalidReq), gomock.Any()).
		Return(nil, fmt.Errorf("email and password cannot be empty"))

	// Call Register with empty fields
	res, err := mockAuthClient.Register(context.Background(), invalidReq)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestLoginWithIncorrectPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	invalidLoginReq := &LoginRequest{Email: "john_doe@example.com", Password: "wrongpass"}

	mockAuthClient.EXPECT().Login(gomock.Any(), gomock.Eq(invalidLoginReq), gomock.Any()).
		Return(nil, fmt.Errorf("invalid credentials"))

	// Call Login with wrong password
	res, err := mockAuthClient.Login(context.Background(), invalidLoginReq)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestRegisterExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	existingUserReq := &RegisterRequest{
		Email:    "john_doe@example.com",
		Password: "password",
		Name:     "test",
		Surname:  "test",
		Age:      27,
	}

	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(existingUserReq), gomock.Any()).
		Return(nil, fmt.Errorf("user already exists"))

	// Attempt to register the same user twice
	res, err := mockAuthClient.Register(context.Background(), existingUserReq)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestSampleProtectedSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	expectedRes := &ProtectedReply{Result: "Access Granted"}
	mockAuthClient.EXPECT().SampleProtected(gomock.Any(), gomock.Any()).Return(expectedRes, nil)

	// Call SampleProtected
	res, err := mockAuthClient.SampleProtected(context.Background(), &ProtectedRequest{})
	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res)
}

func TestSampleProtectedFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	mockAuthClient.EXPECT().SampleProtected(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("authentication failed"))

	// Call SampleProtected
	res, err := mockAuthClient.SampleProtected(context.Background(), &ProtectedRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}
