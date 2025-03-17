package middlewares

import (
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

func HeaderForwarderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Define the headers you want to forward
		headersToForward := []string{"Authorization", "X-Custom-Header"}

		// Initialize a metadata map
		md := metadata.New(nil)

		for _, header := range headersToForward {
			if values, ok := r.Header[header]; ok {
				// gRPC metadata keys must be lowercase
				key := strings.ToLower(header)
				for _, value := range values {
					md.Append(key, value)
				}
			}
		}

		// Create a new context with the metadata
		ctx := metadata.NewOutgoingContext(r.Context(), md)

		// Create a new request with the updated context
		r = r.WithContext(ctx)

		// Call the next handler with the updated request
		next.ServeHTTP(w, r)
	})
}
