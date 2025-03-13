package middlewares

import "github.com/valyala/fasthttp"

// CORSMiddleware adds CORS headers to fasthttp requests.
func CORSMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Set CORS headers on the response.
		header := ctx.Response.Header
		header.Set("Access-Control-Allow-Origin", "*") // Or specify a particular domain.
		header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request.
		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		// Continue processing the request.
		next(ctx)
	}
}
