package middlewares

import (
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

// Test CORS Middleware with a simulated OPTIONS request
func TestCORSMiddleware(t *testing.T) {
	mockHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(200)
	}

	cors := CORSMiddleware(mockHandler)

	// Simulate an OPTIONS request
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/test")
	ctx.Request.Header.SetMethod("OPTIONS")

	cors(ctx)

	// Debug headers
	t.Logf("Response Headers:\n%s", ctx.Response.Header.String())

	// Expected CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	// Check if headers are correctly set
	for key, expectedValue := range expectedHeaders {
		value := string(ctx.Response.Header.Peek(key))
		if value != expectedValue {
			t.Errorf("Expected %s: %s, but got %s", key, expectedValue, value)
		}
	}

	// Ensure correct status code for preflight requests
	if ctx.Response.StatusCode() != fasthttp.StatusNoContent {
		t.Errorf("Expected status 204, got %d", ctx.Response.StatusCode())
	}
}

// Test Rate Limiting Middleware
func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(1, 1) // 1 request per second

	clientID := "test-client"

	// First request should pass
	if !limiter.GetLimiter(clientID).Allow() {
		t.Error("Expected first request to pass")
	}

	// Second request should be rate-limited
	if limiter.GetLimiter(clientID).Allow() {
		t.Error("Expected second request to be blocked")
	}

	// Wait for rate limit to reset
	time.Sleep(time.Second)

	if !limiter.GetLimiter(clientID).Allow() {
		t.Error("Expected request after reset to pass")
	}
}
