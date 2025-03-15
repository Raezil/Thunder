package middlewares

import (
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

// LoggingMiddleware logs request method, path, status code, and duration.
func LoggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()

		next(ctx) // Execute the next handler

		duration := time.Since(start)

		log.Printf("[HTTP] %s %s %d %s",
			string(ctx.Method()), ctx.Path(), ctx.Response.StatusCode(), duration)
	}
}
