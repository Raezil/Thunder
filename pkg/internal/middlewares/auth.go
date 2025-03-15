package middlewares

import (
	"context"
	"fmt"
	pb "services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// middleware verifies JWT tokens in the request context.
// Rejects unauthorized requests with a detailed log entry.
func AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/authenticator.Auth/Login" || info.FullMethod == "/authenticator.Auth/Register" {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}
	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	claims, err := pb.VerifyJWT(token[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
	}
	// Set timeout for database operations to prevent hanging requests
	md = metadata.Join(md, metadata.Pairs("current_user", claims.Email))
	ctx = metadata.NewIncomingContext(ctx, md)
	return handler(ctx, req)
}
