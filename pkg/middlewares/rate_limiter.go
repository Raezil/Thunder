package middlewares

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimiter structure
type RateLimiter struct {
	mu             sync.Mutex
	limiters       map[string]*rate.Limiter
	rate           rate.Limit
	burst          int
	trustedProxies map[string]bool
	maxLimiters    int
}

// NewRateLimiter initializes a rate limiter with configurable rate, burst, and trusted proxies
func NewRateLimiter(r rate.Limit, b int, proxies []string) *RateLimiter {
	trusted := make(map[string]bool)
	for _, proxy := range proxies {
		if isValidIP(proxy) {
			trusted[proxy] = true
		}
	}

	return &RateLimiter{
		limiters:       make(map[string]*rate.Limiter),
		rate:           r,
		burst:          b,
		trustedProxies: trusted,
		maxLimiters:    10000, // Default maximum number of limiters to prevent memory leaks
	}
}

// isValidIP checks if a string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GetLimiter gets or creates a rate limiter for a specific client
func (r *RateLimiter) GetLimiter(clientID string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return existing limiter if it exists
	if limiter, exists := r.limiters[clientID]; exists {
		return limiter
	}

	// Prevent memory leaks by enforcing a maximum number of limiters
	if len(r.limiters) >= r.maxLimiters {
		// Simple eviction strategy: remove one random entry
		// For production, consider using LRU or similar algorithm
		for k := range r.limiters {
			delete(r.limiters, k)
			break
		}
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

// DefaultTrustedProxies returns a list of commonly trusted proxy IPs
func DefaultTrustedProxies() []string {
	return []string{"127.0.0.1", "::1"}
}

// RateLimiterInterceptor applies rate limiting
func (r *RateLimiter) RateLimiterInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "could not determine peer")
	}

	peerIP, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid peer address: %v", err)
	}

	var clientID string

	// Only trust proxy headers if the request is from a trusted proxy
	if r.trustedProxies[peerIP] {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			// Check X-Forwarded-For (may be a comma-separated list)
			if xff := md.Get("x-forwarded-for"); len(xff) > 0 && xff[0] != "" {
				ips := strings.Split(xff[0], ",")
				if len(ips) > 0 && strings.TrimSpace(ips[0]) != "" {
					cleanIP := strings.TrimSpace(ips[0])
					if isValidIP(cleanIP) {
						clientID = cleanIP
					}
				}
			} else if xri := md.Get("x-real-ip"); len(xri) > 0 && xri[0] != "" {
				cleanIP := strings.TrimSpace(xri[0])
				if isValidIP(cleanIP) {
					clientID = cleanIP
				}
			}
		}
	}

	// If no trusted clientID was found from headers, use the peer's IP
	if clientID == "" {
		clientID = peerIP
	}

	limiter := r.GetLimiter(clientID)
	if !limiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "Too many requests, slow down")
	}

	// Proceed to the next handler
	return handler(ctx, req)
}
