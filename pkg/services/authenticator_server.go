package services

import (
	"context"
	"db"
	"fmt"
	. "generated"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthenticatorServer is your gRPC server.
type AuthServiceServer struct {
	UnimplementedAuthServer
	PrismaClient *db.PrismaClient
	Logger       *zap.SugaredLogger
}

// SampleProtected is a protected endpoint.
func (s *AuthServiceServer) SampleProtected(ctx context.Context, in *ProtectedRequest) (*ProtectedReply, error) {
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		s.Logger.Warnw("Failed to retrieve current user", "error", err)
		return nil, status.Errorf(codes.Unauthenticated, "failed to retrieve current user: %v", err)
	}
	return &ProtectedReply{
		Result: in.Text + " " + currentUser,
	}, nil
}

// Login verifies the user's credentials and returns a JWT token.
func (s *AuthServiceServer) Login(ctx context.Context, in *LoginRequest) (*LoginReply, error) {
	s.Logger.Infof("Login attempt for email: %s", in.Email)

	user, err := s.PrismaClient.User.FindUnique(
		db.User.Email.Equals(in.Email),
	).Exec(ctx)

	// Handle user not found (or any error retrieving the user).
	if err != nil || user == nil {
		s.Logger.Warnw("Login failed: user not found", "email", in.Email, "error", err)
		return nil, status.Errorf(codes.Unauthenticated, "incorrect email or password")
	}

	// Compare the stored hashed password with the password provided.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		s.Logger.Warnw("Invalid password attempt", "email", in.Email)
		return nil, status.Errorf(codes.Unauthenticated, "Invalid credentials: %v", err)
	}

	token, err := GenerateJWT(in.Email)
	if err != nil {
		s.Logger.Errorw("Error generating token", "email", in.Email, "error", err)
		return nil, status.Errorf(codes.Internal, "could not generate token: %v", err)
	}

	s.Logger.Infof("Generated token for email %s", in.Email)
	return &LoginReply{
		Token: token,
	}, nil
}

// Register creates a new user after ensuring the email is unique and hashing the password.
func (s *AuthServiceServer) Register(ctx context.Context, in *RegisterRequest) (*RegisterReply, error) {
	// Check if a user with the given email already exists.
	s.Logger.Debugw("Register request received", "email", in.Email)
	existingUser, err := s.PrismaClient.User.FindUnique(
		db.User.Email.Equals(in.Email),
	).Exec(ctx)
	if err == nil && existingUser != nil {
		s.Logger.Warnw("Registration failed: email already in use", "email", in.Email)
		return nil, status.Errorf(codes.AlreadyExists, "failed to register user: email already in use")
	}

	const bcryptCost = 12 // Recommended: 12-14 for production
	// Hash the password using bcrypt with the default cost.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		s.Logger.Errorw("Failed to hash password", "email", in.Email, "error", err)
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
		s.Logger.Errorw("Failed to create user", "email", in.Email, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	s.Logger.Infow("User registered successfully", "email", obj.Email)
	return &RegisterReply{
		Reply: fmt.Sprintf("Congratulations, User email: %s got created!", obj.Email),
	}, nil
}
