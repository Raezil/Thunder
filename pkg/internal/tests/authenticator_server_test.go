package tests

import (
	"context"
	"errors"
	"fmt"
	. "generated"
	generated "generated"
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
	req := &generated.LoginRequest{}
	expectedReply := &generated.LoginReply{}
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

	req := &generated.RegisterRequest{}
	expectedReply := &generated.RegisterReply{}
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

// TestRegisterWithNilRequest verifies that calling Register with a nil request returns an error.
func TestRegisterWithNilRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	mockAuthClient.EXPECT().Register(gomock.Any(), nil, gomock.Any()).Return(nil, fmt.Errorf("register request is nil"))

	res, err := mockAuthClient.Register(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// TestLoginWithNilRequest verifies that calling Login with a nil request returns an error.
func TestLoginWithNilRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	mockAuthClient.EXPECT().Login(gomock.Any(), nil, gomock.Any()).Return(nil, fmt.Errorf("login request is nil"))

	res, err := mockAuthClient.Login(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// TestSampleProtectedNilRequest verifies that calling SampleProtected with a nil request returns an error.
func TestSampleProtectedNilRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	mockAuthClient.EXPECT().SampleProtected(gomock.Any(), nil).Return(nil, fmt.Errorf("protected request is nil"))

	res, err := mockAuthClient.SampleProtected(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// TestLoginWithNilContext verifies that calling Login with a nil context returns an error.
func TestLoginWithNilContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)
	loginReq := &LoginRequest{Email: "john_doe@example.com", Password: "password"}

	mockAuthClient.EXPECT().Login(nil, gomock.Eq(loginReq), gomock.Any()).Return(nil, fmt.Errorf("context is nil"))

	res, err := mockAuthClient.Login(nil, loginReq)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// TestRegisterWithInvalidAge verifies that a registration request with an invalid age returns an error.
func TestRegisterWithInvalidAge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)
	invalidAgeReq := &RegisterRequest{
		Email:    "jane_doe@example.com",
		Password: "password",
		Name:     "Jane",
		Surname:  "Doe",
		Age:      -5, // invalid age
	}

	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(invalidAgeReq), gomock.Any()).Return(nil, fmt.Errorf("invalid age"))

	res, err := mockAuthClient.Register(context.Background(), invalidAgeReq)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// TestDoubleLogin simulates two successive login attempts for the same user that return different tokens.
func TestDoubleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)
	loginReq := &LoginRequest{Email: "john_doe@example.com", Password: "password"}
	expectedLoginRes1 := &LoginReply{Token: "token-1"}
	expectedLoginRes2 := &LoginReply{Token: "token-2"}

	// Set expectations for two successive calls
	mockAuthClient.EXPECT().Login(gomock.Any(), gomock.Eq(loginReq), gomock.Any()).Return(expectedLoginRes1, nil).Times(1)
	mockAuthClient.EXPECT().Login(gomock.Any(), gomock.Eq(loginReq), gomock.Any()).Return(expectedLoginRes2, nil).Times(1)

	// First login attempt
	res1, err1 := mockAuthClient.Login(context.Background(), loginReq)
	assert.NoError(t, err1)
	assert.Equal(t, expectedLoginRes1, res1)

	// Second login attempt
	res2, err2 := mockAuthClient.Login(context.Background(), loginReq)
	assert.NoError(t, err2)
	assert.Equal(t, expectedLoginRes2, res2)
}

// TestMultipleRegistrationsDifferentUsers verifies that two different users can register successfully.
func TestMultipleRegistrationsDifferentUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthClient := NewMockAuthClient(ctrl)

	regReq1 := &RegisterRequest{
		Email:    "alice@example.com",
		Password: "password1",
		Name:     "Alice",
		Surname:  "Smith",
		Age:      30,
	}
	regReply1 := fmt.Sprintf("Congratulations, User email: %s got created!", "alice@example.com")
	expectedRegRes1 := &RegisterReply{Reply: regReply1}

	regReq2 := &RegisterRequest{
		Email:    "bob@example.com",
		Password: "password2",
		Name:     "Bob",
		Surname:  "Brown",
		Age:      35,
	}
	regReply2 := fmt.Sprintf("Congratulations, User email: %s got created!", "bob@example.com")
	expectedRegRes2 := &RegisterReply{Reply: regReply2}

	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(regReq1), gomock.Any()).Return(expectedRegRes1, nil)
	mockAuthClient.EXPECT().Register(gomock.Any(), gomock.Eq(regReq2), gomock.Any()).Return(expectedRegRes2, nil)

	res1, err1 := mockAuthClient.Register(context.Background(), regReq1)
	assert.NoError(t, err1)
	assert.Equal(t, expectedRegRes1, res1)

	res2, err2 := mockAuthClient.Register(context.Background(), regReq2)
	assert.NoError(t, err2)
	assert.Equal(t, expectedRegRes2, res2)
}
