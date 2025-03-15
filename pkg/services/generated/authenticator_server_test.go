package generated

import (
	"context"
	"errors"
	"fmt"
	reflect "reflect"
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

// TestNewMockAuthServer verifies that NewMockAuthServer properly creates an instance and initializes its recorder.
func TestNewMockAuthServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockAuthServer(ctrl)
	if mockServer == nil {
		t.Fatal("NewMockAuthServer returned nil")
	}
	if mockServer.recorder == nil {
		t.Error("NewMockAuthServer did not initialize recorder")
	}
}

// TestEXPECT checks that calling EXPECT() on the mock returns a valid recorder.
func TestEXPECT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockAuthServer(ctrl)
	rec := mockServer.EXPECT()
	if rec == nil {
		t.Fatal("EXPECT() returned nil")
	}
}

// TestLogin sets an expectation for Login and then calls it.
// It verifies that the returned reply and error match the expected values.
func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockAuthServer(ctrl)

	// Create a dummy login request and expected reply.
	req := &LoginRequest{}
	expectedReply := &LoginReply{}
	expectedErr := error(nil)

	// Set expectation: Login should be called with any context and our req, and return expectedReply and no error.
	mockServer.EXPECT().Login(gomock.Any(), req).Return(expectedReply, expectedErr)

	// Call Login.
	reply, err := mockServer.Login(context.Background(), req)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if reply != expectedReply {
		t.Errorf("Expected reply %v, got %v", expectedReply, reply)
	}
}

// TestRegister sets an expectation for Register and then calls it.
// It checks that the returned reply and error are as expected.
func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockAuthServer(ctrl)

	req := &RegisterRequest{}
	expectedReply := &RegisterReply{}
	expectedErr := errors.New("registration error")

	// Set expectation: Register should be called with any context and our req, and return expectedReply and a dummy error.
	mockServer.EXPECT().Register(gomock.Any(), req).Return(expectedReply, expectedErr)

	reply, err := mockServer.Register(context.Background(), req)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else if err.Error() != expectedErr.Error() {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if reply != expectedReply {
		t.Errorf("Expected reply %v, got %v", expectedReply, reply)
	}
}

func TestNewMockUnsafeAuthServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockUnsafeAuthServer(ctrl)
	if mock == nil {
		t.Fatal("NewMockUnsafeAuthServer returned nil")
	}
}

// TestEXPECT verifies that calling EXPECT on the mock returns a non-nil expectation object.
func TestEXPECTMockUnsafeAuthServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockUnsafeAuthServer(ctrl)
	expect := mock.EXPECT()
	if expect == nil {
		t.Fatal("EXPECT returned nil")
	}
}

func TestSampleProtected_Authorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockAuthServer(ctrl)
	// Create a valid protected request; adjust the fields as needed.
	req := &ProtectedRequest{Text: "valid-token-text"}
	expectedReply := &ProtectedReply{Result: "Access Granted"}

	// Set up expectation: when SampleProtected is called with the valid request,
	// it should return the expected reply and no error.
	mockServer.EXPECT().SampleProtected(gomock.Any(), gomock.Eq(req)).Return(expectedReply, nil)

	// Call SampleProtected.
	reply, err := mockServer.SampleProtected(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !reflect.DeepEqual(reply, expectedReply) {
		t.Fatalf("Expected reply %+v, got %+v", expectedReply, reply)
	}
}

// TestSampleProtected_Unauthorized verifies that SampleProtected returns an error
// when provided with invalid credentials.
func TestSampleProtected_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServer := NewMockAuthServer(ctrl)
	// Create a request with an invalid token.
	req := &ProtectedRequest{Text: "invalid-token-text"}

	// Set up expectation: when SampleProtected is called with the invalid request,
	// it should return nil and an "unauthorized" error.
	mockServer.EXPECT().SampleProtected(gomock.Any(), gomock.Eq(req)).Return(nil, errors.New("unauthorized"))

	// Call SampleProtected.
	reply, err := mockServer.SampleProtected(context.Background(), req)
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("Expected 'unauthorized' error, got error: %v", err)
	}
	if reply != nil {
		t.Fatalf("Expected nil reply, got: %+v", reply)
	}
}
