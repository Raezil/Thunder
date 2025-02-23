package backend

import (
	"context"
	"db"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthenticatorServer is your gRPC server.
type AuthenticatorServer struct {
	UnimplementedAuthServer
	PrismaClient *db.PrismaClient
	Logger       *zap.SugaredLogger
}

// SampleProtected is a protected endpoint.
func (s *AuthenticatorServer) SampleProtected(ctx context.Context, in *ProtectedRequest) (*ProtectedReply, error) {
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		s.Logger.Error(err)
		return nil, err
	}
	return &ProtectedReply{
		Result: in.Text + " " + currentUser,
	}, nil
}

// Login verifies the user's credentials and returns a JWT token.
func (s *AuthenticatorServer) Login(ctx context.Context, in *LoginRequest) (*LoginReply, error) {
	s.Logger.Infof("Login attempt for email: %s", in.Email)

	user, err := s.PrismaClient.User.FindUnique(
		db.User.Email.Equals(in.Email),
	).Exec(ctx)

	// Handle user not found (or any error retrieving the user).
	if err != nil || user == nil {
		s.Logger.Errorw("Failed to get user", "error", err)
		return nil, status.Errorf(codes.Unauthenticated, "incorrect email or password")
	}

	// Compare the stored hashed password with the password provided.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		s.Logger.Errorf("Invalid password for email %s", in.Email)
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials: %v", err)
	}

	token, err := GenerateJWT(in.Email)
	if err != nil {
		s.Logger.Errorf("Error generating token for email %s: %v", in.Email, err)
		return nil, status.Errorf(codes.Internal, "could not generate token: %v", err)
	}

	s.Logger.Infof("Generated token for email %s", in.Email)
	return &LoginReply{
		Token: token,
	}, nil
}

// Register creates a new user after ensuring the email is unique and hashing the password.
func (s *AuthenticatorServer) Register(ctx context.Context, in *RegisterRequest) (*RegisterReply, error) {
	// Check if a user with the given email already exists.
	existingUser, err := s.PrismaClient.User.FindUnique(
		db.User.Email.Equals(in.Email),
	).Exec(ctx)
	if err == nil && existingUser != nil {
		s.Logger.Errorf("User with email %s already exists", in.Email)
		return nil, status.Errorf(codes.AlreadyExists, "failed to register user: email already in use")
	}

	const bcryptCost = 12 // Recommended: 12-14 for production
	// Hash the password using bcrypt with the default cost.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		s.Logger.Errorf("failed to hash password: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	obj, err := s.PrismaClient.User.CreateOne(
		db.User.Name.Set(in.Name),
		// Store the hashed password instead of plaintext.
		db.User.Password.Set(string(hashedPassword)),
		db.User.Email.Set(in.Email),
		db.User.Age.Set(int(in.Age)),
	).Exec(ctx)
	if err != nil {
		s.Logger.Errorf("failed to create user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	s.Logger.Infof("User registered successfully with email: %s", obj.Email)
	return &RegisterReply{
		Reply: fmt.Sprintf("Congratulations, User email: %s got created!", obj.Email),
	}, nil
}
