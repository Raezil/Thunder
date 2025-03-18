package middlewares

import "github.com/valyala/fasthttp"

func HeaderForwarderMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		authHeader := ctx.Request.Header.Peek("Authorization")
		if authHeader != nil {
			ctx.Request.Header.Set("Authorization", string(authHeader)) // ✅ Forward Authorization Header
		}
		next(ctx)
	}
}
