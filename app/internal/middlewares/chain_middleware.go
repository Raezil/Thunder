package middlewares

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryInterceptors manually chains multiple gRPC Unary Interceptors
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Recursive call to chain interceptors
		var chainHandler grpc.UnaryHandler
		chainHandler = handler

		// Apply interceptors in reverse order (last one runs first)
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			next := chainHandler
			chainHandler = func(c context.Context, r interface{}) (interface{}, error) {
				return interceptor(c, r, info, next)
			}
		}

		// Call the first interceptor
		return chainHandler(ctx, req)
	}
}
