package middlewares

import (
	"strings"

	"github.com/valyala/fasthttp"
)

func HeaderForwarderMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Get the Authorization header
		authHeader := ctx.Request.Header.Peek("Authorization")
		if authHeader != nil {
			// Ensure the header is preserved in the expected format
			// This ensures consistency regardless of how the client sent it
			authValue := string(authHeader)
			if !strings.HasPrefix(strings.ToLower(authValue), "bearer ") && !strings.HasPrefix(strings.ToLower(authValue), "basic ") {
				authValue = "Bearer " + authValue
			}
			// Set the canonical Authorization header
			ctx.Request.Header.Set("Authorization", authValue)
		}
		next(ctx)
	}
}
