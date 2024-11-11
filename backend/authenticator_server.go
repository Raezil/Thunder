package backend

import (
	"context"
	"db"
	"fmt"
	"log"

	"go.uber.org/zap"
)

type AuthenticatorServer struct {
	UnimplementedAuthServer
	PrismaClient *db.PrismaClient
	Logger       *zap.SugaredLogger
}

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

func (s *AuthenticatorServer) Login(ctx context.Context, in *LoginRequest) (*LoginReply, error) {
	log.Println("Login attempt for email:", in.Email)

	user, err := s.PrismaClient.User.FindUnique(
		db.User.Email.Equals(in.Email),
	).Exec(ctx)

	if err != nil {
		s.Logger.Error("User not found: %v", err)

		return nil, fmt.Errorf("incorrect email or password")
	}

	if user.Password != in.Password {
		s.Logger.Error("Invalid password")
		return nil, fmt.Errorf("incorrect email or password")
	}

	token, err := GenerateJWT(in.Email)
	if err != nil {
		s.Logger.Error("Error generating token: %v", err)
		return nil, fmt.Errorf("could not generate token: %v", err)
	}

	s.Logger.Info("Generated token: %s", token)

	return &LoginReply{
		Token: token,
	}, nil
}

func (s *AuthenticatorServer) Register(ctx context.Context, in *RegisterRequest) (*RegisterReply, error) {
	obj, err := s.PrismaClient.User.CreateOne(
		db.User.Name.Set(in.Name),
		db.User.Password.Set(in.Password),
		db.User.Email.Set(in.Email),
		db.User.Age.Set(int(in.Age)),
	).Exec(ctx)

	if err != nil {
		s.Logger.Error("failed to create user: %v", err)
		return nil, fmt.Errorf("failed to register user")
	}

	return &RegisterReply{
		Reply: fmt.Sprintf("Congratulations, User email: %s got created!", obj.Email),
	}, nil
}
