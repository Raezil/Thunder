package helpers

import (
	"log"
	"net/http"
	"strings"

	graphqlruntime "github.com/ysugimoto/grpc-graphql-gateway/runtime"
	"google.golang.org/grpc/metadata"
)

// GraphqlServeMux wraps graphqlruntime.ServeMux and adds an incoming header matcher.
type GraphqlServeMux struct {
	*graphqlruntime.ServeMux
	incomingHeaderMatcher func(string) (string, bool)
}

// NewGraphqlServeMux creates a new GraphqlServeMux.
func NewGraphqlServeMux() *GraphqlServeMux {
	return &GraphqlServeMux{
		ServeMux:              graphqlruntime.NewServeMux(),
		incomingHeaderMatcher: defaultHeaderMatcher,
	}
}

func defaultHeaderMatcher(key string) (string, bool) {
	// Here you could map headers as needed.
	// For example, return "Authorization" (with proper case) for any form of the auth header.
	if strings.ToLower(key) == "authorization" {
		return "Authorization", true
	}
	return strings.ToLower(key), true
}

func (c *GraphqlServeMux) SetIncomingHeaderMatcher(matcher func(string) (string, bool)) {
	c.incomingHeaderMatcher = matcher
}

// ServeHTTP converts HTTP headers to gRPC metadata and logs it.
// ServeHTTP converts HTTP headers to gRPC metadata and logs it.
func (c *GraphqlServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log incoming headers for debugging
	log.Printf("Incoming HTTP headers: %v", r.Header)

	// Build metadata map from HTTP headers
	mdMap := make(map[string][]string)

	// Directly copy the Authorization header to ensure it's preserved
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Store both capitalized and lowercase versions to ensure compatibility
		mdMap["Authorization"] = []string{authHeader}
		mdMap["authorization"] = []string{authHeader}
	}

	// Process remaining headers
	for key, values := range r.Header {
		// Skip Authorization since it was already processed
		if len(values) > 0 && strings.ToLower(key) == "authorization" {
			continue
		}
		if mappedKey, ok := c.incomingHeaderMatcher(key); ok {
			mdMap[mappedKey] = []string{values[0]}
		}
	}

	// Convert the metadata map into a flat list of pairs
	pairs := []string{}
	for k, vs := range mdMap {
		for _, v := range vs {
			pairs = append(pairs, k, v)
		}
	}

	md := metadata.Pairs(pairs...)

	// Create a new context with the metadata and update the request
	newCtx := metadata.NewIncomingContext(r.Context(), md)
	newRequest := r.WithContext(newCtx)

	// Log the metadata for debugging
	log.Printf("GraphQL metadata being forwarded: %v", md)

	// Forward the request to the underlying GraphQL handler
	c.ServeMux.ServeHTTP(w, newRequest)
}
