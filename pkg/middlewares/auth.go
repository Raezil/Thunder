package middlewares

import (
	"context"
	"fmt"
	pb "services"
	"strings"

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
	rawToken := strings.TrimSpace(strings.TrimPrefix(token[0], "Bearer "))
	claims, err := pb.VerifyJWT(rawToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
	}
	// Set timeout for database operations to prevent hanging requests
	md = metadata.Join(md, metadata.Pairs("current_user", claims.Email))
	ctx = metadata.NewIncomingContext(ctx, md)
	return handler(ctx, req)
}

func AuthStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md["authorization"]
	if len(tokens) == 0 {
		return status.Error(codes.Unauthenticated, "missing token")
	}

	rawToken := strings.TrimSpace(strings.TrimPrefix(tokens[0], "Bearer "))
	claims, err := pb.VerifyJWT(rawToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "unauthorized: %v", err)
	}

	// 👇 Set current_user from claims
	newMD := metadata.Join(md, metadata.Pairs("current_user", claims.Email))
	newCtx := metadata.NewIncomingContext(ss.Context(), newMD)

	// 👇 Wrap the stream with overridden context
	return handler(srv, &wrappedStream{ServerStream: ss, ctx: newCtx})
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
