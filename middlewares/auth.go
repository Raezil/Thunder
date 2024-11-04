package middlewares

import (
	pb "backend"
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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
		return nil, fmt.Errorf("missing token")
	}

	claims, err := pb.VerifyJWT(token[0])
	ctx = metadata.AppendToOutgoingContext(ctx, "current_user", claims.Email)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %v", err)
	}
	return handler(ctx, req)
}
