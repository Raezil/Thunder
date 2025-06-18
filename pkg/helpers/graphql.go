package helpers

import (
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
	if strings.ToLower(key) == "authorization" {
		return "authorization", true
	}
	return strings.ToLower(key), true
}

func (c *GraphqlServeMux) SetIncomingHeaderMatcher(matcher func(string) (string, bool)) {
	c.incomingHeaderMatcher = matcher
}

// Custom handler that intercepts the request and manually sets up metadata
func (c *GraphqlServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get the authorization header
	authHeader := r.Header.Get("Authorization")

	// Create a custom context with metadata
	ctx := r.Context()

	if authHeader != "" {
		// Create metadata with the authorization header
		md := metadata.Pairs("authorization", authHeader)

		// Set both incoming and outgoing metadata
		ctx = metadata.NewIncomingContext(ctx, md)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Update the request with the new context
		r = r.WithContext(ctx)
	}

	// Call the original GraphQL handler
	c.ServeMux.ServeHTTP(w, r)
}
