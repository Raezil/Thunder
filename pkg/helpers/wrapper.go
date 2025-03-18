package helpers

import (
	"log"
	"net/http"
	"strings"

	graphqlruntime "github.com/ysugimoto/grpc-graphql-gateway/runtime"
	"google.golang.org/grpc/metadata"
)

// CustomServeMux wraps graphqlruntime.ServeMux and adds an incoming header matcher.
type GraphqlServeMux struct {
	*graphqlruntime.ServeMux
	incomingHeaderMatcher func(string) (string, bool)
}

func NewGraphqlServeMux() *GraphqlServeMux {
	return &GraphqlServeMux{
		ServeMux:              graphqlruntime.NewServeMux(),
		incomingHeaderMatcher: defaultHeaderMatcher,
	}
}

func defaultHeaderMatcher(key string) (string, bool) {
	return strings.ToLower(key), true
}

func (c *GraphqlServeMux) SetIncomingHeaderMatcher(matcher func(string) (string, bool)) {
	c.incomingHeaderMatcher = matcher
}

func (c *GraphqlServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a fresh metadata map for this request
	mdMap := make(map[string][]string)

	// Process Authorization header specifically
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Store as a slice since metadata uses string slices
		mdMap["authorization"] = []string{authHeader}
	}

	// Process other headers
	for key, values := range r.Header {
		if len(values) > 0 && strings.ToLower(key) != "authorization" {
			if mappedKey, ok := c.incomingHeaderMatcher(key); ok {
				mdMap[mappedKey] = []string{values[0]}
			}
		}
	}

	// Create metadata directly using pairs to ensure format consistency
	pairs := []string{}
	for k, vs := range mdMap {
		for _, v := range vs {
			pairs = append(pairs, k, v)
		}
	}

	md := metadata.Pairs(pairs...)

	// Create a new context with the metadata
	newCtx := metadata.NewIncomingContext(r.Context(), md)

	// Update the request with the new context
	newRequest := r.WithContext(newCtx)

	log.Printf("GraphQL metadata: %v", md)

	// Forward to the underlying handler
	c.ServeMux.ServeHTTP(w, newRequest)
}
