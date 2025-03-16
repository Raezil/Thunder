package middlewares

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimiter structure
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewRateLimiter initializes a rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter gets or creates a rate limiter for a specific client
func (r *RateLimiter) GetLimiter(clientID string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limiter, exists := r.limiters[clientID]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(r.rate, r.burst)
	r.limiters[clientID] = limiter

	// Cleanup old limiters after a timeout
	go func() {
		time.Sleep(10 * time.Minute)
		r.mu.Lock()
		delete(r.limiters, clientID)
		r.mu.Unlock()
	}()

	return limiter
}

// RateLimiterInterceptor applies rate limiting
func (r *RateLimiter) RateLimiterInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	// Extract client identifier (can use IP or API key)
	// Extract the client's IP address from the context using the peer package.
	p, ok := peer.FromContext(ctx)
	var clientID string
	if ok {
		clientID = p.Addr.String()
		// Optionally, further parse clientID if needed (e.g., remove port information)
	} else {
		clientID = "unknown"
	}
	limiter := r.GetLimiter(clientID)
	if !limiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "Too many requests, slow down")
	}

	// Proceed to the next handler
	return handler(ctx, req)
}
