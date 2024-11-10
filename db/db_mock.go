package db

import "context"

// PrismaClientInterface defines the methods required for database operations in your tests.
type PrismaClientInterface interface {
	CreateUser(ctx context.Context, email, password, name, surname string, age int) (UserModel, error)
	LoginUser(ctx context.Context, email, password string) (UserModel, error)
	// Add other necessary methods from your PrismaClient
}

// MockPrismaClient is a mock implementation of the PrismaClientInterface.
type MockPrismaClient struct {
	CreateUserResult UserModel
	CreateUserError  error
	LoginUserResult  UserModel
	LoginUserError   error
}

// CreateUser simulates creating a user.
func (m *MockPrismaClient) CreateUser(ctx context.Context, email, password, name, surname string, age int) (UserModel, error) {
	if m.CreateUserError != nil {
		return UserModel{}, m.CreateUserError
	}
	return m.CreateUserResult, nil
}

// LoginUser simulates logging in a user.
func (m *MockPrismaClient) LoginUser(ctx context.Context, email, password string) (UserModel, error) {
	if m.LoginUserError != nil {
		return UserModel{}, m.LoginUserError
	}
	return m.LoginUserResult, nil
}
