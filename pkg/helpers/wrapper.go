package helpers

import (
	"log"
	"net/http"
	"strings"

	graphqlruntime "github.com/ysugimoto/grpc-graphql-gateway/runtime"
	"google.golang.org/grpc/metadata"
)

// CustomServeMux wraps graphqlruntime.ServeMux and adds an incoming header matcher.
type CustomServeMux struct {
	*graphqlruntime.ServeMux
	// incomingHeaderMatcher maps incoming header keys to their desired form.
	incomingHeaderMatcher func(string) (string, bool)
}

// NewCustomServeMux returns a new instance of CustomServeMux.
func NewCustomServeMux() *CustomServeMux {
	return &CustomServeMux{
		ServeMux:              graphqlruntime.NewServeMux(),
		incomingHeaderMatcher: defaultHeaderMatcher,
	}
}

// defaultHeaderMatcher is a simple implementation that converts keys to lower-case.
func defaultHeaderMatcher(key string) (string, bool) {
	return strings.ToLower(key), true
}

// SetIncomingHeaderMatcher allows setting a custom header matcher.
func (c *CustomServeMux) SetIncomingHeaderMatcher(matcher func(string) (string, bool)) {
	c.incomingHeaderMatcher = matcher
}

// ServeHTTP intercepts incoming requests, applies the header matcher, injects matching headers as gRPC metadata,
// and then delegates to the underlying ServeMux.
func (c *CustomServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create an empty metadata map.
	mdMap := make(map[string]string)

	// Iterate over incoming headers.
	for key, values := range r.Header {
		if len(values) > 0 {
			// If the header is an authorization header (any case), map it to "Authorization"
			if strings.ToLower(key) == "authorization" {
				log.Println(values[0])
				mdMap["Authorization"] = values[0]
			} else {
				mdMap[strings.ToLower(key)] = values[0]
			}
		}
	}

	// Create gRPC metadata from the map.
	md := metadata.New(mdMap)

	// Attach the metadata to the incoming context.
	newCtx := metadata.NewIncomingContext(r.Context(), md)

	// Replace the request's context with the new context.
	r = r.WithContext(newCtx)

	// Delegate to the underlying ServeMux.
	c.ServeMux.ServeHTTP(w, r)
}
