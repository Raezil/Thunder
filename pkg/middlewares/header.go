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
			authValue := string(authHeader)
			if !strings.HasPrefix(strings.ToLower(authValue), "bearer ") &&
				!strings.HasPrefix(strings.ToLower(authValue), "basic ") {
				authValue = "Bearer " + authValue
			}
			// Set the canonical Authorization header
			ctx.Request.Header.Set("Authorization", authValue)
			// Ensure it's also set as lowercase for consistency with gRPC metadata
			ctx.Request.Header.Set("authorization", authValue)
		}

		next(ctx)
	}
}
